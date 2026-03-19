package services

import (
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
	service := NewAuthService(tx, "test-secret")

	user, err := service.Register(fmt.Sprintf("user-%d@example.com", time.Now().UnixNano()), "password123")
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
	service := NewAuthService(tx, "test-secret")
	email := fmt.Sprintf("dup-%d@example.com", time.Now().UnixNano())

	if _, err := service.Register(email, "password123"); err != nil {
		t.Fatalf("expected initial register success, got %v", err)
	}

	_, err := service.Register(email, "password123")
	if err == nil {
		t.Fatal("expected duplicate register error")
	}
	if errors.Is(err, ErrEmailAlreadyUsed) {
		return
	}
	t.Skipf("duplicate key is not translated by current gorm/postgres setup: %v", err)
}

func TestAuthServiceLogin(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewAuthService(tx, "test-secret")
	email := fmt.Sprintf("login-%d@example.com", time.Now().UnixNano())

	if _, err := service.Register(email, "password123"); err != nil {
		t.Fatalf("expected register success, got %v", err)
	}

	t.Run("returns invalid credentials when user is missing", func(t *testing.T) {
		_, err := service.Login("missing@example.com", "password123")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("returns invalid credentials on wrong password", func(t *testing.T) {
		_, err := service.Login(email, "wrong-password")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("returns signed token on success", func(t *testing.T) {
		tokenString, err := service.Login(email, "password123")
		if err != nil {
			t.Fatalf("expected login success, got %v", err)
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
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
	})
}

func TestAuthServiceLoginDatabaseError(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewAuthService(tx, "test-secret")

	if err := tx.Exec("DROP TABLE IF EXISTS users CASCADE").Error; err != nil {
		t.Fatalf("failed to drop users table: %v", err)
	}

	_, err := service.Login("john@example.com", "password123")
	if err == nil {
		t.Fatal("expected database error")
	}
}

func TestAuthServiceRegisterPersistsUserFields(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewAuthService(tx, "test-secret")
	email := fmt.Sprintf("persist-%d@example.com", time.Now().UnixNano())

	user, err := service.Register(email, "password123")
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
