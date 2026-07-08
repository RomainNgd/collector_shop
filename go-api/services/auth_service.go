package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyUsed    = errors.New("email already used")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrRefreshTokenReused  = errors.New("refresh token reused")
)

type AuthServiceInterface interface {
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	Register(ctx context.Context, email, password string) (*models.User, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
}

type AuthService struct {
	db         *gorm.DB
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(db *gorm.DB, jwtSecret string, accessTTL, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		db:         db,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func generateRefreshToken() (raw string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	raw = base64.RawURLEncoding.EncodeToString(buf)
	return raw, hashRefreshToken(raw), nil
}

func (s *AuthService) signAccessToken(user *models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(s.accessTTL).Unix(),
	})

	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*models.User, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	user := &models.User{
		Email: email,
		Role:  constants.RoleUser,
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	result := s.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, ErrEmailAlreadyUsed
		}
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var user models.User
	result := s.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("database error: %w", result.Error)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	accessToken, err := s.signAccessToken(&user)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err)
	}

	rawRefreshToken, refreshHash, err := generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	refreshRecord := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}
	if result := s.db.WithContext(ctx).Create(refreshRecord); result.Error != nil {
		return "", "", fmt.Errorf("failed to persist refresh token: %w", result.Error)
	}

	return accessToken, rawRefreshToken, nil
}

// RefreshAccessToken validates the given raw refresh token and, if valid,
// rotates it: the presented token is revoked and a brand new refresh token
// is issued alongside a new access token. Rotation lets a later replay of
// the now-revoked token be detected as a reuse/theft signal.
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	hash := hashRefreshToken(refreshToken)

	// Reuse detection is checked and handled (and committed) outside of the
	// rotation transaction below: a transaction rolls back entirely when its
	// callback returns an error, which would silently undo the "kill every
	// session" revocation if it happened inside the same transaction as the
	// ErrRefreshTokenReused return.
	var preCheck models.RefreshToken
	result := s.db.WithContext(ctx).Where("token_hash = ?", hash).First(&preCheck)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", "", ErrInvalidRefreshToken
		}
		return "", "", fmt.Errorf("database error: %w", result.Error)
	}

	if preCheck.RevokedAt != nil {
		if err := s.db.WithContext(ctx).Model(&models.RefreshToken{}).
			Where("user_id = ? AND revoked_at IS NULL", preCheck.UserID).
			Update("revoked_at", time.Now()).Error; err != nil {
			return "", "", fmt.Errorf("failed to revoke sessions: %w", err)
		}
		return "", "", ErrRefreshTokenReused
	}

	var newAccessToken, newRawRefreshToken string
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record models.RefreshToken
		result := tx.Where("token_hash = ?", hash).First(&record)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return ErrInvalidRefreshToken
			}
			return fmt.Errorf("database error: %w", result.Error)
		}

		if record.RevokedAt != nil {
			// Revoked concurrently between the pre-check above and this
			// transaction (e.g. a racing refresh call rotated it first).
			return ErrRefreshTokenReused
		}

		if record.ExpiresAt.Before(time.Now()) {
			return ErrRefreshTokenExpired
		}

		var user models.User
		if result := tx.First(&user, record.UserID); result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return fmt.Errorf("database error: %w", result.Error)
		}

		rawRefreshToken, refreshHash, err := generateRefreshToken()
		if err != nil {
			return err
		}

		newRecord := &models.RefreshToken{
			UserID:    user.ID,
			TokenHash: refreshHash,
			ExpiresAt: time.Now().Add(s.refreshTTL),
		}
		if result := tx.Create(newRecord); result.Error != nil {
			return fmt.Errorf("failed to persist refresh token: %w", result.Error)
		}

		now := time.Now()
		record.RevokedAt = &now
		record.ReplacedByID = &newRecord.ID
		if result := tx.Save(&record); result.Error != nil {
			return fmt.Errorf("failed to revoke previous refresh token: %w", result.Error)
		}

		accessToken, err := s.signAccessToken(&user)
		if err != nil {
			return fmt.Errorf("failed to sign token: %w", err)
		}

		newAccessToken = accessToken
		newRawRefreshToken = rawRefreshToken
		return nil
	})
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRawRefreshToken, nil
}

// Logout revokes the given refresh token. Unknown or already-revoked
// tokens are treated as a no-op success so the response never leaks
// whether a token was valid.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	hash := hashRefreshToken(refreshToken)

	result := s.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", time.Now())
	if result.Error != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", result.Error)
	}

	return nil
}
