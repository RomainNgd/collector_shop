package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthServiceRegister(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewAuthService(tx, newTestSecret(t), 15*time.Minute, 30*24*time.Hour)

	user, err := service.Register(context.Background(), fmt.Sprintf("user-%d@example.com", time.Now().UnixNano()), "password123")
	if err != nil {
		t.Fatalf("expected register success, got %v", err)
	}
	if user.Role != constants.RoleUser {
		t.Fatalf("expected role %s, got %s", constants.RoleUser, user.Role)
	}
	if user.Password == "password123" {
		t.Fatal("expected hashed password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123")); err != nil {
		t.Fatalf("expected password hash to match: %v", err)
	}
}

func TestAuthServiceRegisterDuplicateEmail(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewAuthService(tx, newTestSecret(t), 15*time.Minute, 30*24*time.Hour)
	email := fmt.Sprintf("dup-%d@example.com", time.Now().UnixNano())

	if _, err := service.Register(context.Background(), email, "password123"); err != nil {
		t.Fatalf("expected initial register success, got %v", err)
	}

	_, err := service.Register(context.Background(), email, "password123")
	if !errors.Is(err, ErrEmailAlreadyUsed) {
		t.Fatalf("expected ErrEmailAlreadyUsed, got %v", err)
	}
}

func TestAuthServiceLogin(t *testing.T) {
	tx := openIntegrationTx(t)
	secret := newTestSecret(t)
	service := NewAuthService(tx, secret, 15*time.Minute, 30*24*time.Hour)
	email := fmt.Sprintf("login-%d@example.com", time.Now().UnixNano())

	if _, err := service.Register(context.Background(), email, "password123"); err != nil {
		t.Fatalf("expected register success, got %v", err)
	}

	t.Run("returns invalid credentials when user is missing", func(t *testing.T) {
		_, _, err := service.Login(context.Background(), "missing@example.com", "password123")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("returns invalid credentials on wrong password", func(t *testing.T) {
		_, _, err := service.Login(context.Background(), email, "wrong-password")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("returns signed access and refresh tokens on success", func(t *testing.T) {
		tokenString, refreshToken, err := service.Login(context.Background(), email, "password123")
		if err != nil {
			t.Fatalf("expected login success, got %v", err)
		}
		if refreshToken == "" {
			t.Fatal("expected non-empty refresh token")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			t.Fatalf("expected valid jwt, got err=%v", err)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatalf("expected map claims, got %#v", token.Claims)
		}
		if claims["role"] != constants.RoleUser {
			t.Fatalf("unexpected role claim: %#v", claims["role"])
		}
		if claims["sub"] == nil {
			t.Fatal("expected subject claim")
		}
		if claims["exp"] == nil {
			t.Fatal("expected expiration claim")
		}

		var persisted models.RefreshToken
		if err := tx.Where("token_hash = ?", hashRefreshToken(refreshToken)).First(&persisted).Error; err != nil {
			t.Fatalf("expected refresh token to be persisted: %v", err)
		}
		if persisted.RevokedAt != nil {
			t.Fatal("expected freshly issued refresh token to not be revoked")
		}
	})
}

func TestAuthServiceRefreshAccessToken(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewAuthService(tx, newTestSecret(t), 15*time.Minute, 30*24*time.Hour)
	email := fmt.Sprintf("refresh-%d@example.com", time.Now().UnixNano())

	if _, err := service.Register(context.Background(), email, "password123"); err != nil {
		t.Fatalf("expected register success, got %v", err)
	}

	_, refreshToken, err := service.Login(context.Background(), email, "password123")
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}

	t.Run("returns ErrInvalidRefreshToken for unknown token", func(t *testing.T) {
		_, _, err := service.RefreshAccessToken(context.Background(), "unknown-token")
		if !errors.Is(err, ErrInvalidRefreshToken) {
			t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})

	t.Run("returns ErrRefreshTokenExpired for expired token", func(t *testing.T) {
		_, expiredRaw, err := service.Login(context.Background(), email, "password123")
		if err != nil {
			t.Fatalf("expected login success, got %v", err)
		}
		if err := tx.Model(&models.RefreshToken{}).
			Where("token_hash = ?", hashRefreshToken(expiredRaw)).
			Update("expires_at", time.Now().Add(-time.Hour)).Error; err != nil {
			t.Fatalf("failed to expire token: %v", err)
		}

		_, _, err = service.RefreshAccessToken(context.Background(), expiredRaw)
		if !errors.Is(err, ErrRefreshTokenExpired) {
			t.Fatalf("expected ErrRefreshTokenExpired, got %v", err)
		}
	})

	t.Run("rotates the token on success", func(t *testing.T) {
		newAccessToken, newRefreshToken, err := service.RefreshAccessToken(context.Background(), refreshToken)
		if err != nil {
			t.Fatalf("expected refresh success, got %v", err)
		}
		if newAccessToken == "" || newRefreshToken == "" {
			t.Fatal("expected non-empty rotated tokens")
		}
		if newRefreshToken == refreshToken {
			t.Fatal("expected a new refresh token to be issued")
		}

		var oldRecord models.RefreshToken
		if err := tx.Where("token_hash = ?", hashRefreshToken(refreshToken)).First(&oldRecord).Error; err != nil {
			t.Fatalf("expected old refresh token record, got %v", err)
		}
		if oldRecord.RevokedAt == nil {
			t.Fatal("expected old refresh token to be revoked")
		}
		if oldRecord.ReplacedByID == nil {
			t.Fatal("expected old refresh token to reference its replacement")
		}

		var newRecord models.RefreshToken
		if err := tx.Where("token_hash = ?", hashRefreshToken(newRefreshToken)).First(&newRecord).Error; err != nil {
			t.Fatalf("expected new refresh token record, got %v", err)
		}
		if newRecord.RevokedAt != nil {
			t.Fatal("expected new refresh token to not be revoked")
		}

		// The rotated (now-revoked) token must no longer work either.
		refreshToken = newRefreshToken
	})

	t.Run("reusing a revoked token revokes all sessions for the user", func(t *testing.T) {
		_, firstRefresh, err := service.Login(context.Background(), email, "password123")
		if err != nil {
			t.Fatalf("expected login success, got %v", err)
		}
		_, secondRefresh, err := service.RefreshAccessToken(context.Background(), firstRefresh)
		if err != nil {
			t.Fatalf("expected rotation success, got %v", err)
		}

		// Replaying the already-rotated (revoked) token should be detected as reuse.
		_, _, err = service.RefreshAccessToken(context.Background(), firstRefresh)
		if !errors.Is(err, ErrRefreshTokenReused) {
			t.Fatalf("expected ErrRefreshTokenReused, got %v", err)
		}

		// The legitimately rotated token must now be revoked too (session family killed),
		// without needing to consume it via another RefreshAccessToken call.
		var secondRecord models.RefreshToken
		if err := tx.Where("token_hash = ?", hashRefreshToken(secondRefresh)).First(&secondRecord).Error; err != nil {
			t.Fatalf("expected second refresh token record, got %v", err)
		}
		if secondRecord.RevokedAt == nil {
			t.Fatal("expected the rotated token to be revoked as part of the session-family kill")
		}
	})
}

func TestAuthServiceLogout(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewAuthService(tx, newTestSecret(t), 15*time.Minute, 30*24*time.Hour)
	email := fmt.Sprintf("logout-%d@example.com", time.Now().UnixNano())

	if _, err := service.Register(context.Background(), email, "password123"); err != nil {
		t.Fatalf("expected register success, got %v", err)
	}
	_, refreshToken, err := service.Login(context.Background(), email, "password123")
	if err != nil {
		t.Fatalf("expected login success, got %v", err)
	}

	if err := service.Logout(context.Background(), refreshToken); err != nil {
		t.Fatalf("expected logout success, got %v", err)
	}

	var record models.RefreshToken
	if err := tx.Where("token_hash = ?", hashRefreshToken(refreshToken)).First(&record).Error; err != nil {
		t.Fatalf("expected refresh token record, got %v", err)
	}
	if record.RevokedAt == nil {
		t.Fatal("expected refresh token to be revoked after logout")
	}

	t.Run("is idempotent for an already-revoked token", func(t *testing.T) {
		if err := service.Logout(context.Background(), refreshToken); err != nil {
			t.Fatalf("expected idempotent logout success, got %v", err)
		}
	})

	t.Run("is a no-op for an unknown token", func(t *testing.T) {
		if err := service.Logout(context.Background(), "unknown-token"); err != nil {
			t.Fatalf("expected no-op success for unknown token, got %v", err)
		}
	})

	_, _, err = service.RefreshAccessToken(context.Background(), refreshToken)
	if !errors.Is(err, ErrRefreshTokenReused) {
		t.Fatalf("expected revoked token replay to be detected as reuse, got %v", err)
	}
}

func TestAuthServiceLoginDatabaseError(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewAuthService(tx, newTestSecret(t), 15*time.Minute, 30*24*time.Hour)

	if err := tx.Exec("DROP TABLE IF EXISTS users CASCADE").Error; err != nil {
		t.Fatalf("failed to drop users table: %v", err)
	}

	_, _, err := service.Login(context.Background(), "john@example.com", "password123")
	if err == nil {
		t.Fatal("expected database error")
	}
}

func TestAuthServiceRegisterPersistsUserFields(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewAuthService(tx, newTestSecret(t), 15*time.Minute, 30*24*time.Hour)
	email := fmt.Sprintf("persist-%d@example.com", time.Now().UnixNano())

	user, err := service.Register(context.Background(), email, "password123")
	if err != nil {
		t.Fatalf("expected register success, got %v", err)
	}

	var persisted models.User
	if err := tx.First(&persisted, user.ID).Error; err != nil {
		t.Fatalf("expected persisted user, got %v", err)
	}
	if persisted.Email != email {
		t.Fatalf("expected persisted email %s, got %s", email, persisted.Email)
	}
}
