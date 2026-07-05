package integration

import (
	"bytes"
	"context"
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

	if err := db.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.Promotion{},
		&models.User{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
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

type integrationStripeService struct{}

func (s *integrationStripeService) Enabled() bool {
	return true
}

func (s *integrationStripeService) CreateCheckoutSession(_ context.Context, input services.StripeCheckoutSessionInput) (*services.StripeCheckoutSession, error) {
	return &services.StripeCheckoutSession{
		ID:            "cs_test_integration",
		URL:           "https://checkout.stripe.com/c/pay/cs_test_integration",
		Status:        "open",
		PaymentStatus: "unpaid",
		Metadata:      input.Metadata,
	}, nil
}

func (s *integrationStripeService) GetCheckoutSession(_ context.Context, sessionID string) (*services.StripeCheckoutSession, error) {
	return &services.StripeCheckoutSession{
		ID:            sessionID,
		URL:           "https://checkout.stripe.com/c/pay/" + sessionID,
		Status:        "open",
		PaymentStatus: "unpaid",
	}, nil
}

func (s *integrationStripeService) ConstructWebhookEvent(_ []byte, _ string) (*services.StripeWebhookEvent, error) {
	return &services.StripeWebhookEvent{
		Type: "checkout.session.completed",
		CheckoutSession: services.StripeCheckoutSession{
			ID:            "cs_test_integration",
			Status:        "complete",
			PaymentStatus: "paid",
			Metadata: map[string]string{
				"order_id": "0",
			},
		},
	}, nil
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
	promotionService := services.NewPromotionService(tx)
	authService := services.NewAuthService(tx, secret)
	orderService := services.NewOrderService(tx)
	orderPaymentService := services.NewOrderPaymentService(tx, &integrationStripeService{}, orderService)
	authMiddleware := middlewares.NewAuthMiddleware(secret)

	categoryHandler := controllers.NewCategoryHandler(categoryService)
	productHandler := controllers.NewProductHandler(productService, categoryService, fileService)
	promotionHandler := controllers.NewPromotionHandler(promotionService)
	authHandler := controllers.NewAuthHandler(authService)
	orderHandler := controllers.NewOrderHandler(orderService, orderPaymentService)
	paymentHandler := controllers.NewPaymentHandler(orderPaymentService)

	router := gin.New()
	routes.SetupAuthRoutes(router, authHandler)
	routes.SetupCategoryRoutes(router, categoryHandler, authMiddleware)
	routes.SetupProductRoutes(router, productHandler, authMiddleware)
	routes.SetupPromotionRoutes(router, promotionHandler, authMiddleware)
	routes.SetupOrderRoutes(router, orderHandler, authMiddleware)
	routes.SetupPaymentRoutes(router, paymentHandler)

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
	router := buildRouter(t, tx, newTestSecret(t))
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
	secret := newTestSecret(t)
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
	secret := newTestSecret(t)
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

func TestPromotionRoutesIntegration(t *testing.T) {
	tx := openIntegrationTx(t)
	secret := newTestSecret(t)
	router := buildRouter(t, tx, secret)

	admin := &models.User{Email: fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()), Password: "hash", Role: constants.RoleAdmin}
	if err := tx.Create(admin).Error; err != nil {
		t.Fatalf("failed to seed admin: %v", err)
	}
	adminToken := signToken(t, secret, admin.ID, admin.Role)

	category := &models.Category{Name: fmt.Sprintf("Consoles-%d", time.Now().UnixNano()), Description: "Vintage consoles"}
	if err := tx.Create(category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	product := &models.Product{
		Name:        fmt.Sprintf("Console-%d", time.Now().UnixNano()),
		Description: "Collector console",
		Image:       "console.png",
		Price:       100,
		CategoryID:  category.ID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}

	noTokenResp := performJSONRequest(t, router, http.MethodGet, "/promotions", nil, "")
	if noTokenResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d body=%s", noTokenResp.Code, noTokenResp.Body.String())
	}

	createResp := performJSONRequest(t, router, http.MethodPost, "/promotions", map[string]any{
		"name":           "Launch week",
		"description":    "Console only",
		"type":           models.PromotionTypePercentage,
		"value":          10,
		"is_active":      true,
		"applies_to_all": false,
		"product_ids":    []uint{product.ID},
	}, adminToken)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 for promotion create, got %d body=%s", createResp.Code, createResp.Body.String())
	}

	listResp := performJSONRequest(t, router, http.MethodGet, "/promotions", nil, adminToken)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for promotion list, got %d body=%s", listResp.Code, listResp.Body.String())
	}
	if !bytes.Contains(listResp.Body.Bytes(), []byte(`"product_ids"`)) {
		t.Fatalf("expected product_ids in promotion response, got %s", listResp.Body.String())
	}

	productResp := performJSONRequest(t, router, http.MethodGet, fmt.Sprintf("/products/%d", product.ID), nil, "")
	if productResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for product get, got %d body=%s", productResp.Code, productResp.Body.String())
	}
	if !bytes.Contains(productResp.Body.Bytes(), []byte(`"effective_price":90`)) {
		t.Fatalf("expected effective promotion price in product response, got %s", productResp.Body.String())
	}
	if !bytes.Contains(productResp.Body.Bytes(), []byte(`"applied_promotion"`)) {
		t.Fatalf("expected applied promotion in product response, got %s", productResp.Body.String())
	}
}

func TestOrderRoutesIntegration(t *testing.T) {
	tx := openIntegrationTx(t)
	secret := newTestSecret(t)
	router := buildRouter(t, tx, secret)

	user := &models.User{Email: fmt.Sprintf("buyer-%d@example.com", time.Now().UnixNano()), Password: "hash", Role: constants.RoleUser}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	userToken := signToken(t, secret, user.ID, user.Role)

	category := &models.Category{Name: fmt.Sprintf("Orders-%d", time.Now().UnixNano()), Description: "Order category"}
	if err := tx.Create(category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	product := &models.Product{
		Name:        fmt.Sprintf("OrderProduct-%d", time.Now().UnixNano()),
		Description: "Checkout product",
		Image:       "order.png",
		Price:       25,
		CategoryID:  category.ID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}

	noTokenResp := performJSONRequest(t, router, http.MethodPost, "/orders", map[string]any{
		"items": []map[string]any{{"product_id": product.ID, "quantity": 1}},
	}, "")
	if noTokenResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d body=%s", noTokenResp.Code, noTokenResp.Body.String())
	}

	createResp := performJSONRequest(t, router, http.MethodPost, "/orders", map[string]any{
		"items": []map[string]any{{"product_id": product.ID, "quantity": 2}},
	}, userToken)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 for order create, got %d body=%s", createResp.Code, createResp.Body.String())
	}
	if !bytes.Contains(createResp.Body.Bytes(), []byte(`"status":"awaiting_payment"`)) {
		t.Fatalf("expected awaiting_payment status, got %s", createResp.Body.String())
	}

	order := &models.Order{}
	if err := tx.Where("user_id = ?", user.ID).First(order).Error; err != nil {
		t.Fatalf("failed to fetch created order: %v", err)
	}

	listResp := performJSONRequest(t, router, http.MethodGet, "/orders", nil, userToken)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected 200 for order list, got %d body=%s", listResp.Code, listResp.Body.String())
	}
	if !bytes.Contains(listResp.Body.Bytes(), []byte(`"item_count":2`)) {
		t.Fatalf("expected item_count in order list, got %s", listResp.Body.String())
	}

	payResp := performJSONRequest(t, router, http.MethodPut, fmt.Sprintf("/orders/%d", order.ID), map[string]any{
		"status": models.OrderStatusPreparation,
	}, userToken)
	if payResp.Code != http.StatusConflict {
		t.Fatalf("expected 409 for direct payment validation, got %d body=%s", payResp.Code, payResp.Body.String())
	}
	if !bytes.Contains(payResp.Body.Bytes(), []byte(`"ORDER_STATUS_TRANSITION_INVALID"`)) {
		t.Fatalf("expected transition error in body, got %s", payResp.Body.String())
	}
}

func TestOrderCheckoutSessionRouteIntegration(t *testing.T) {
	tx := openIntegrationTx(t)
	secret := newTestSecret(t)
	router := buildRouter(t, tx, secret)

	user := &models.User{Email: fmt.Sprintf("checkout-%d@example.com", time.Now().UnixNano()), Password: "hash", Role: constants.RoleUser}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	userToken := signToken(t, secret, user.ID, user.Role)

	category := &models.Category{Name: fmt.Sprintf("Checkout-%d", time.Now().UnixNano()), Description: "Checkout category"}
	if err := tx.Create(category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	product := &models.Product{
		Name:        fmt.Sprintf("CheckoutProduct-%d", time.Now().UnixNano()),
		Description: "Checkout product",
		Image:       "checkout.png",
		Price:       29.99,
		CategoryID:  category.ID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}

	createOrderResp := performJSONRequest(t, router, http.MethodPost, "/orders", map[string]any{
		"items": []map[string]any{{"product_id": product.ID, "quantity": 1}},
	}, userToken)
	if createOrderResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 for order create, got %d body=%s", createOrderResp.Code, createOrderResp.Body.String())
	}

	order := &models.Order{}
	if err := tx.Where("user_id = ?", user.ID).First(order).Error; err != nil {
		t.Fatalf("failed to fetch created order: %v", err)
	}

	checkoutResp := performJSONRequest(t, router, http.MethodPost, fmt.Sprintf("/orders/%d/checkout-session", order.ID), map[string]any{
		"success_url": fmt.Sprintf("http://localhost:5173/mes-commandes/%d?payment=processing", order.ID),
		"cancel_url":  fmt.Sprintf("http://localhost:5173/mes-commandes/%d?payment=cancelled", order.ID),
	}, userToken)
	if checkoutResp.Code != http.StatusCreated {
		t.Fatalf("expected 201 for checkout session, got %d body=%s", checkoutResp.Code, checkoutResp.Body.String())
	}
	if !bytes.Contains(checkoutResp.Body.Bytes(), []byte(`"url":"https://checkout.stripe.com/c/pay/cs_test_integration"`)) {
		t.Fatalf("expected stripe checkout url in response, got %s", checkoutResp.Body.String())
	}
}
