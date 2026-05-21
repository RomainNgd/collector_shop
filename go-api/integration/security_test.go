package integration

import (
	"fmt"
	"net/http"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func signSecurityToken(t *testing.T, secret string, userID uint, role string, expiresAt time.Time) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  expiresAt.Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

func TestSecurityProtectedRoutesRejectInvalidTokens(t *testing.T) {
	tx := openIntegrationTx(t)
	secret := "integration-secret"
	router := buildRouter(t, tx, secret)

	user := &models.User{
		Email:    fmt.Sprintf("security-user-%d@example.com", time.Now().UnixNano()),
		Password: "hash",
		Role:     constants.RoleUser,
	}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	t.Run("expired token", func(t *testing.T) {
		token := signSecurityToken(t, secret, user.ID, user.Role, time.Now().Add(-time.Minute))
		resp := performJSONRequest(t, router, http.MethodGet, "/orders", nil, token)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for expired token, got %d body=%s", resp.Code, resp.Body.String())
		}
	})

	t.Run("token signed with wrong secret", func(t *testing.T) {
		token := signSecurityToken(t, "wrong-secret", user.ID, user.Role, time.Now().Add(time.Hour))
		resp := performJSONRequest(t, router, http.MethodGet, "/orders", nil, token)
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for wrong secret token, got %d body=%s", resp.Code, resp.Body.String())
		}
	})
}

func TestSecurityAdminRoutesRejectNonAdminUsers(t *testing.T) {
	tx := openIntegrationTx(t)
	secret := "integration-secret"
	router := buildRouter(t, tx, secret)

	user := &models.User{
		Email:    fmt.Sprintf("non-admin-%d@example.com", time.Now().UnixNano()),
		Password: "hash",
		Role:     constants.RoleUser,
	}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	token := signToken(t, secret, user.ID, user.Role)
	resp := performJSONRequest(t, router, http.MethodPost, "/categories", map[string]any{
		"name":        "Forbidden category",
		"description": "Only admins can create categories",
	}, token)
	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin category creation, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestSecurityUserCannotReadAnotherUsersOrder(t *testing.T) {
	tx := openIntegrationTx(t)
	secret := "integration-secret"
	router := buildRouter(t, tx, secret)

	owner := &models.User{
		Email:    fmt.Sprintf("order-owner-%d@example.com", time.Now().UnixNano()),
		Password: "hash",
		Role:     constants.RoleUser,
	}
	otherUser := &models.User{
		Email:    fmt.Sprintf("order-other-%d@example.com", time.Now().UnixNano()),
		Password: "hash",
		Role:     constants.RoleUser,
	}
	if err := tx.Create(owner).Error; err != nil {
		t.Fatalf("failed to seed owner: %v", err)
	}
	if err := tx.Create(otherUser).Error; err != nil {
		t.Fatalf("failed to seed other user: %v", err)
	}

	category := &models.Category{
		Name:        fmt.Sprintf("Security-%d", time.Now().UnixNano()),
		Description: "Security test category",
	}
	if err := tx.Create(category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	product := &models.Product{
		Name:        fmt.Sprintf("SecurityProduct-%d", time.Now().UnixNano()),
		Description: "Security test product",
		Image:       "security.png",
		Price:       10,
		CategoryID:  category.ID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}

	ownerToken := signToken(t, secret, owner.ID, owner.Role)
	createResp := performJSONRequest(t, router, http.MethodPost, "/orders", map[string]any{
		"items": []map[string]any{{"product_id": product.ID, "quantity": 1}},
	}, ownerToken)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected owner order creation, got %d body=%s", createResp.Code, createResp.Body.String())
	}

	var order models.Order
	if err := tx.Where("user_id = ?", owner.ID).First(&order).Error; err != nil {
		t.Fatalf("failed to fetch owner order: %v", err)
	}

	otherUserToken := signToken(t, secret, otherUser.ID, otherUser.Role)
	readResp := performJSONRequest(t, router, http.MethodGet, fmt.Sprintf("/orders/%d", order.ID), nil, otherUserToken)
	if readResp.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when reading another user's order, got %d body=%s", readResp.Code, readResp.Body.String())
	}
}
