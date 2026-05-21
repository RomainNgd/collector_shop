package controllers

import (
	"fmt"
	"math"
	"poc-gin/pkg/constants"
	"strconv"

	"github.com/gin-gonic/gin"
)

func userIDFromContext(c *gin.Context) (uint, error) {
	rawUserID, exists := c.Get(constants.ContextKeyUserID)
	if !exists {
		return 0, fmt.Errorf("missing user id in request context")
	}

	switch value := rawUserID.(type) {
	case uint:
		return value, nil
	case uint64:
		return uint(value), nil
	case int:
		if value <= 0 {
			return 0, fmt.Errorf("invalid user id")
		}
		return uint(value), nil
	case int64:
		if value <= 0 {
			return 0, fmt.Errorf("invalid user id")
		}
		return uint(value), nil
	case float64:
		if value <= 0 || math.Trunc(value) != value {
			return 0, fmt.Errorf("invalid user id")
		}
		return uint(value), nil
	case string:
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil || parsed == 0 {
			return 0, fmt.Errorf("invalid user id")
		}
		return uint(parsed), nil
	default:
		return 0, fmt.Errorf("invalid user id type %T", rawUserID)
	}
}

func userRoleFromContext(c *gin.Context) string {
	rawRole, exists := c.Get(constants.ContextKeyUserRole)
	if !exists {
		return constants.RoleUser
	}

	role, ok := rawRole.(string)
	if !ok || role == "" {
		return constants.RoleUser
	}

	return role
}
