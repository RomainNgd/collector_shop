package middlewares

import (
	"fmt"
	"net/http"
	"poc-gin/controllers"
	"poc-gin/pkg/constants"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			controllers.RespondError(c, http.StatusUnauthorized, "AUTH_HEADER_INVALID", "Invalid authorization header", nil)
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenString == "" {
			controllers.RespondError(c, http.StatusUnauthorized, "TOKEN_MISSING", "Missing token", nil)
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			controllers.RespondError(c, http.StatusUnauthorized, "TOKEN_INVALID", "Invalid token", nil)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			controllers.RespondError(c, http.StatusUnauthorized, "CLAIMS_INVALID", "Invalid claims", nil)
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
