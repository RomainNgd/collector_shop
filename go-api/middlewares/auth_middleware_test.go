package middlewares

import (
	"net/http"
	"net/http/httptest"
	"poc-gin/pkg/constants"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func signedToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

func performMiddlewareRequest(middleware gin.HandlerFunc, header string, setup func(c *gin.Context)) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		ctx.Request.Header.Set("Authorization", header)
	}

	engine.Use(middleware)
	engine.GET("/", func(c *gin.Context) {
		if setup != nil {
			setup(c)
		}
		c.Status(http.StatusOK)
	})
	engine.HandleContext(ctx)
	return recorder
}

func TestAuthenticate(t *testing.T) {
	secret := newTestSecret(t)
	middleware := NewAuthMiddleware(secret)

	t.Run("returns 401 when header is missing", func(t *testing.T) {
		recorder := performMiddlewareRequest(middleware.Authenticate(), "", nil)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
	})

	t.Run("returns 401 when header is not bearer", func(t *testing.T) {
		recorder := performMiddlewareRequest(middleware.Authenticate(), "Basic abc", nil)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
	})

	t.Run("returns 401 when token is empty", func(t *testing.T) {
		recorder := performMiddlewareRequest(middleware.Authenticate(), "Bearer ", nil)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
	})

	t.Run("returns 401 when token is invalid", func(t *testing.T) {
		recorder := performMiddlewareRequest(middleware.Authenticate(), "Bearer not-a-token", nil)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
	})

	t.Run("returns 401 when subject claim is missing", func(t *testing.T) {
		token := signedToken(t, secret, jwt.MapClaims{
			"role": constants.RoleAdmin,
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		recorder := performMiddlewareRequest(middleware.Authenticate(), "Bearer "+token, nil)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
	})

	t.Run("stores user id and role on valid token", func(t *testing.T) {
		token := signedToken(t, secret, jwt.MapClaims{
			"sub":  float64(42),
			"role": constants.RoleAdmin,
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		var userID any
		var role any
		recorder := performMiddlewareRequest(middleware.Authenticate(), "Bearer "+token, func(c *gin.Context) {
			userID, _ = c.Get(constants.ContextKeyUserID)
			role, _ = c.Get(constants.ContextKeyUserRole)
		})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
		if userID != float64(42) {
			t.Fatalf("unexpected user id: %#v", userID)
		}
		if role != constants.RoleAdmin {
			t.Fatalf("unexpected role: %#v", role)
		}
	})
}

func TestOptionalAuthenticate(t *testing.T) {
	secret := newTestSecret(t)
	middleware := NewAuthMiddleware(secret)

	t.Run("continues without context when header is missing", func(t *testing.T) {
		var userID any
		recorder := performMiddlewareRequest(middleware.OptionalAuthenticate(), "", func(c *gin.Context) {
			userID, _ = c.Get(constants.ContextKeyUserID)
		})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
		if userID != nil {
			t.Fatalf("expected no user id, got %#v", userID)
		}
	})

	t.Run("continues without context when token is invalid", func(t *testing.T) {
		var userID any
		recorder := performMiddlewareRequest(middleware.OptionalAuthenticate(), "Bearer not-a-token", func(c *gin.Context) {
			userID, _ = c.Get(constants.ContextKeyUserID)
		})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
		if userID != nil {
			t.Fatalf("expected no user id, got %#v", userID)
		}
	})

	t.Run("stores user id and role on valid token", func(t *testing.T) {
		token := signedToken(t, secret, jwt.MapClaims{
			"sub":  float64(42),
			"role": constants.RoleUser,
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		var userID any
		var role any
		recorder := performMiddlewareRequest(middleware.OptionalAuthenticate(), "Bearer "+token, func(c *gin.Context) {
			userID, _ = c.Get(constants.ContextKeyUserID)
			role, _ = c.Get(constants.ContextKeyUserRole)
		})
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
		if userID != float64(42) {
			t.Fatalf("unexpected user id: %#v", userID)
		}
		if role != constants.RoleUser {
			t.Fatalf("unexpected role: %#v", role)
		}
	})
}

func TestRequireAdmin(t *testing.T) {
	middleware := NewAuthMiddleware(newTestSecret(t))

	run := func(seedRole any, setRole bool) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		_, engine := gin.CreateTestContext(recorder)
		engine.Use(func(c *gin.Context) {
			if setRole {
				c.Set(constants.ContextKeyUserRole, seedRole)
			}
		})
		engine.Use(middleware.RequireAdmin())
		engine.GET("/", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		engine.ServeHTTP(recorder, req)
		return recorder
	}

	t.Run("returns 403 when role is missing", func(t *testing.T) {
		recorder := run(nil, false)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", recorder.Code)
		}
	})

	t.Run("returns 403 when role is not admin", func(t *testing.T) {
		recorder := run(constants.RoleUser, true)
		if recorder.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 when role is admin", func(t *testing.T) {
		recorder := run(constants.RoleAdmin, true)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}
