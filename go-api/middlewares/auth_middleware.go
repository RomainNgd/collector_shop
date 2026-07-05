package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"poc-gin/controllers"
	"poc-gin/pkg/constants"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var errInvalidClaims = errors.New("invalid claims")

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := extractBearerToken(c.GetHeader("Authorization"))
		if !ok {
			controllers.RespondError(c, http.StatusUnauthorized, "AUTH_HEADER_INVALID", "Invalid authorization header", nil)
			c.Abort()
			return
		}

		if tokenString == "" {
			controllers.RespondError(c, http.StatusUnauthorized, "TOKEN_MISSING", "Missing token", nil)
			c.Abort()
			return
		}

		claims, err := m.parseClaims(tokenString)
		if err != nil {
			if errors.Is(err, errInvalidClaims) {
				controllers.RespondError(c, http.StatusUnauthorized, "CLAIMS_INVALID", "Invalid claims", nil)
				c.Abort()
				return
			}
			controllers.RespondError(c, http.StatusUnauthorized, "TOKEN_INVALID", "Invalid token", nil)
			c.Abort()
			return
		}

		sub, exists := claims["sub"]
		if !exists {
			controllers.RespondError(c, http.StatusUnauthorized, "TOKEN_SUB_MISSING", "Missing subject claim", nil)
			c.Abort()
			return
		}

		c.Set(constants.ContextKeyUserID, sub)

		if role, ok := claims["role"].(string); ok {
			c.Set(constants.ContextKeyUserRole, role)
		}

		c.Next()
	}
}

func extractBearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" || !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}

	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")), true
}

func (m *AuthMiddleware) parseClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errInvalidClaims
	}

	return claims, nil
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(constants.ContextKeyUserRole)
		if !exists {
			controllers.RespondError(c, http.StatusForbidden, "ROLE_MISSING", "User role missing", nil)
			c.Abort()
			return
		}

		if role != constants.RoleAdmin {
			controllers.RespondError(c, http.StatusForbidden, "ADMIN_REQUIRED", "Admin access required", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
