package routes

import (
	"poc-gin/controllers"
	"poc-gin/middlewares"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := middlewares.NewAuthMiddleware("route-test-secret")

	SetupAuthRoutes(router, controllers.NewAuthHandler(nil), nil)
	SetupCategoryRoutes(router, controllers.NewCategoryHandler(nil), authMiddleware)
	SetupProductRoutes(router, controllers.NewProductHandler(nil, nil, nil), authMiddleware)
	SetupPromotionRoutes(router, controllers.NewPromotionHandler(nil), authMiddleware)
	SetupOrderRoutes(router, controllers.NewOrderHandler(nil, nil), authMiddleware)
	SetupPaymentRoutes(router, controllers.NewPaymentHandler(nil))

	registered := make(map[string]struct{})
	for _, route := range router.Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"POST /auth/login",
		"POST /auth/register",
		"GET /categories",
		"POST /categories",
		"GET /products",
		"POST /products",
		"GET /seller/products",
		"POST /products/:id/image",
		"GET /promotions",
		"PUT /promotions/:id",
		"GET /orders",
		"POST /orders/:id/checkout-session",
		"DELETE /orders/:id",
		"POST /payments/stripe/webhook",
	}
	for _, route := range expected {
		if _, exists := registered[route]; !exists {
			t.Errorf("expected route %q to be registered", route)
		}
	}
}
