package services

import (
	"errors"
	"os"
	"poc-gin/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Register(user models.User) (models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return user, err
	}
	user.Password = string(hashedPassword)

	result := s.DB.Create(&user)
	return user, result.Error
}

func (s *AuthService) Login(email, password string) (string, error) {
	var user models.User

	result := s.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return "", errors.New("utilisateur introuvable")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("mot de passe incorrect")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, err
}
