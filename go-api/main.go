package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"poc-gin/config"
	"poc-gin/controllers"
	"poc-gin/database"
	"poc-gin/middlewares"
	"poc-gin/pkg/logger"
	appmetrics "poc-gin/pkg/metrics"
	"poc-gin/routes"
	"poc-gin/services"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "serve":
		if err := runServer(cfg); err != nil {
			logger.Fatal("Server startup failed: %v", err)
		}
	case "seed":
		if err := runSeed(cfg); err != nil {
			logger.Fatal("Database seed failed: %v", err)
		}
	default:
		logger.Fatal("Unknown command %q. Available commands: serve, seed", command)
	}
}

func runServer(cfg *config.Config) error {
	db, err := database.New(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database: %v", err)
		}
	}()

	if cfg.Database.AutoMigrate {
		logger.Info("Running database migrations...")
		if err := database.Migrate(&cfg.Database); err != nil {
			return fmt.Errorf("database migration failed: %w", err)
		}
		logger.Info("Database migrations completed")
	}

	fileService, err := services.NewFileService(&cfg.Upload)
	if err != nil {
		return fmt.Errorf("failed to initialize file service: %w", err)
	}

	categoryService := services.NewCategoryService(db.DB)
	productService := services.NewProductService(db.DB)
	promotionService := services.NewPromotionService(db.DB)
	authService := services.NewAuthService(
		db.DB,
		cfg.JWT.Secret,
		time.Duration(cfg.JWT.AccessExpirationMinutes)*time.Minute,
		time.Duration(cfg.JWT.RefreshExpirationDays)*24*time.Hour,
	)
	orderService := services.NewOrderService(db.DB)
	stripeService := services.NewStripeService(&cfg.Stripe)
	orderPaymentService := services.NewOrderPaymentService(db.DB, stripeService, orderService, cfg.Stripe.CheckoutAllowedOrigins)

	authMiddleware := middlewares.NewAuthMiddleware(cfg.JWT.Secret)
	authRateLimiter := middlewares.NewRateLimiter(
		cfg.RateLimit.AuthRequests,
		time.Duration(cfg.RateLimit.AuthWindowSeconds)*time.Second,
	)

	categoryHandler := controllers.NewCategoryHandler(categoryService)
	productHandler := controllers.NewProductHandler(productService, categoryService, fileService)
	promotionHandler := controllers.NewPromotionHandler(promotionService)
	authHandler := controllers.NewAuthHandler(authService)
	orderHandler := controllers.NewOrderHandler(orderService, orderPaymentService)
	paymentHandler := controllers.NewPaymentHandler(orderPaymentService)
	healthHandler := controllers.NewHealthHandler(db)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), appmetrics.Middleware())

	// Static files must be registered BEFORE dynamic routes
	r.Static("/upload", cfg.Upload.Dir)

	routes.SetupHealthRoutes(r, healthHandler)
	routes.SetupAuthRoutes(r, authHandler, authRateLimiter)
	routes.SetupCategoryRoutes(r, categoryHandler, authMiddleware)
	routes.SetupProductRoutes(r, productHandler, authMiddleware)
	routes.SetupPromotionRoutes(r, promotionHandler, authMiddleware)
	routes.SetupOrderRoutes(r, orderHandler, authMiddleware)
	routes.SetupPaymentRoutes(r, paymentHandler)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", appmetrics.Handler())
	metricsServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Server.MetricsPort),
		Handler:           metricsMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("Metrics server starting on port %s", cfg.Server.MetricsPort)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Metrics server failed: %v", err)
		}
	}()

	go func() {
		logger.Info("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}
	if err := metricsServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("metrics server forced to shutdown: %w", err)
	}

	logger.Info("Server stopped gracefully")
	return nil
}

func runSeed(cfg *config.Config) error {
	db, err := database.New(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database: %v", err)
		}
	}()

	logger.Info("Running database migrations before seed...")
	if err := database.Migrate(&cfg.Database); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	var report *database.SeedReport
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		var seedErr error
		report, seedErr = database.SeedDemoData(tx, cfg.Upload.Dir)
		return seedErr
	}); err != nil {
		return err
	}

	logger.Info("Demo data seeded successfully: %s", report.Summary())
	logger.Info("Seeded accounts")
	return nil
}
