package controllers

import (
	"poc-gin/pkg/constants"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserIDFromContext(t *testing.T) {
	validValues := []any{uint(4), uint64(4), int(4), int64(4), float64(4), "4"}
	for _, value := range validValues {
		ctx, _ := gin.CreateTestContext(nil)
		ctx.Set(constants.ContextKeyUserID, value)
		userID, err := userIDFromContext(ctx)
		if err != nil || userID != 4 {
			t.Fatalf("expected %T value to resolve to user 4, got %d, %v", value, userID, err)
		}
	}

	invalidValues := []any{int(0), int64(-1), float64(1.5), "invalid", struct{}{}}
	for _, value := range invalidValues {
		ctx, _ := gin.CreateTestContext(nil)
		ctx.Set(constants.ContextKeyUserID, value)
		if _, err := userIDFromContext(ctx); err == nil {
			t.Fatalf("expected %T value to be rejected", value)
		}
	}

	ctx, _ := gin.CreateTestContext(nil)
	if _, err := userIDFromContext(ctx); err == nil {
		t.Fatal("expected missing user ID to be rejected")
	}
}

func TestUserRoleFromContext(t *testing.T) {
	ctx, _ := gin.CreateTestContext(nil)
	if role := userRoleFromContext(ctx); role != constants.RoleUser {
		t.Fatalf("expected missing role to default to user, got %q", role)
	}

	ctx.Set(constants.ContextKeyUserRole, 42)
	if role := userRoleFromContext(ctx); role != constants.RoleUser {
		t.Fatalf("expected invalid role to default to user, got %q", role)
	}

	ctx.Set(constants.ContextKeyUserRole, constants.RoleAdmin)
	if role := userRoleFromContext(ctx); role != constants.RoleAdmin {
		t.Fatalf("expected admin role, got %q", role)
	}
}
