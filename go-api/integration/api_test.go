package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"poc-gin/config"
	"poc-gin/controllers"
	"poc-gin/middlewares"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"poc-gin/routes"
	"poc-gin/services"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func openIntegrationTx(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		testEnv("DB_HOST", "127.0.0.1"),
		testEnv("DB_USER", "golang"),
		testEnv("DB_PASSWORD", "golang"),
		testEnv("DB_NAME", "ecommerce"),
		testEnv("DB_PORT", "5432"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}

	if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.User{}); err != nil {
		t.Fatalf("failed to migrate test schema: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to start transaction: %v", tx.Error)
	}

	t.Cleanup(func() {
		_ = tx.Rollback().Error
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return tx
}

func buildRouter(t *testing.T, tx *gorm.DB, secret string) *gin.Engine {
	t.Helper()

	fileService, err := services.NewFileService(&config.UploadConfig{
		Dir:         filepath.Join(t.TempDir(), "upload"),
		MaxFileSize: constants.MaxFileSize,
	})
	if err != nil {
		t.Fatalf("failed to init file service: %v", err)
	}

	categoryService := services.NewCategoryService(tx)
	productService := services.NewProductService(tx)
	authService := services.NewAuthService(tx, secret)
	authMiddleware := middlewares.NewAuthMiddleware(secret)

	categoryHandler := controllers.NewCategoryHandler(categoryService)
	productHandler := controllers.NewProductHandler(productService, categoryService, fileService)
	authHandler := controllers.NewAuthHandler(authService)

	router := gin.New()
	routes.SetupAuthRoutes(router, authHandler)
	routes.SetupCategoryRoutes(router, categoryHandler, authMiddleware)
	routes.SetupProductRoutes(router, productHandler, authMiddleware)

	return router
}

func performJSONRequest(t *testing.T, router http.Handler, method, target string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func signToken(t *testing.T, secret string, userID uint, role string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

func TestAuthRoutesIntegration(t *testing.T) {
	tx := openIntegrationTx(t)
	router := buildRouter(t, tx, "integration-secret")
	email := fmt.Sprintf("api-%d@example.com", time.Now().UnixNano())

	registerResp := performJSONRequest(t, router, http.MethodPost, "/auth/register", map[string]any{
		"email":    email,
		"password": "password123",
	}, "")
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 from register, got %d body=%s", registerResp.Code, registerResp.Body.String())
	}

	loginResp := performJSONRequest(t, router, http.MethodPost, "/auth/login", map[string]any{
		"email":    email,
		"password": "password123",
	}, "")
	if loginResp.Code != http.StatusOK {
		t.Fatalf("expected 200 from login, got %d body=%s", loginResp.Code, loginResp.Body.String())
	}
}

func TestCategoryRoutesIntegration(t *testing.T) {
	tx := openIntegrationTx(t)
	secret := "integration-secret"
	router := buildRouter(t, tx, secret)

	category := &models.Category{Name: fmt.Sprintf("Public-%d", time.Now().UnixNano()), Description: "Public category"}
	if err := tx.Create(category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	resp := performJSONRequest(t, router, http.MethodGet, "/categories", nil, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for public list, got %d body=%s", resp.Code, resp.Body.String())
	}

	noTokenResp := performJSONRequest(t, router, http.MethodPost, "/categories", map[string]any{
		"name":        "NoToken",
		"description": "Forbidden",
	}, "")
	if noTokenResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d body=%s", noTokenResp.Code, noTokenResp.Body.String())
	}

	user := &models.User{Email: fmt.Sprintf("user-%d@example.com", time.Now().UnixNano()), Password: "hash", Role: constants.RoleUser}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	userToken := signToken(t, secret, user.ID, user.Role)

	userResp := performJSONRequest(t, router, http.MethodPost, "/categories", map[string]any{
		"name":        "UserDenied",
		"description": "Forbidden",
	}, userToken)
	if userResp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for user token, got %d body=%s", userResp.Code, userResp.Body.String())
	}

	admin := &models.User{Email: fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()), Password: "hash", Role: constants.RoleAdmin}
	if err := tx.Create(admin).Error; err != nil {
		t.Fatalf("failed to seed admin: %v", err)
	}
	adminToken := signToken(t, secret, admin.ID, admin.Role)

	adminResp := performJSONRequest(t, router, http.MethodPost, "/categories", map[string]any{
		"name":        fmt.Sprintf("Cards-%d", time.Now().UnixNano()),
		"description": "Trading cards",
	}, adminToken)
	if adminResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 for admin create, got %d body=%s", adminResp.Code, adminResp.Body.String())
	}
}

func TestProductRoutesIntegration(t *testing.T) {
	tx := openIntegrationTx(t)
	secret := "integration-secret"
	router := buildRouter(t, tx, secret)

	admin := &models.User{Email: fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()), Password: "hash", Role: constants.RoleAdmin}
	if err := tx.Create(admin).Error; err != nil {
		t.Fatalf("failed to seed admin: %v", err)
	}
	adminToken := signToken(t, secret, admin.ID, admin.Role)

	category := &models.Category{Name: fmt.Sprintf("Cards-%d", time.Now().UnixNano()), Description: "Trading cards"}
	if err := tx.Create(category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	createResp := performJSONRequest(t, router, http.MethodPost, "/products", map[string]any{
		"name":        "Blue Eyes",
		"description": "Mint card",
		"image":       "blue-eyes.png",
		"price":       19.99,
		"category_id": category.ID,
	}, adminToken)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 for product create, got %d body=%s", createResp.Code, createResp.Body.String())
	}

	product := &models.Product{}
	if err := tx.Where("name = ?", "Blue Eyes").First(product).Error; err != nil {
		t.Fatalf("failed to fetch created product: %v", err)
	}

	getResp := performJSONRequest(t, router, http.MethodGet, fmt.Sprintf("/products/%d", product.ID), nil, "")
	if getResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for product get, got %d body=%s", getResp.Code, getResp.Body.String())
	}
	if !bytes.Contains(getResp.Body.Bytes(), []byte(`"category"`)) {
		t.Fatalf("expected category in response body, got %s", getResp.Body.String())
	}

	deleteCategoryResp := performJSONRequest(t, router, http.MethodDelete, fmt.Sprintf("/categories/%d", category.ID), nil, adminToken)
	if deleteCategoryResp.Code != http.StatusConflict {
		t.Fatalf("expected 409 when deleting referenced category, got %d body=%s", deleteCategoryResp.Code, deleteCategoryResp.Body.String())
	}
}
