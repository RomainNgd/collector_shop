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
	"poc-gin/models"
	"poc-gin/pkg/logger"
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
		if err := migrateDatabase(db.DB); err != nil {
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
	authService := services.NewAuthService(db.DB, cfg.JWT.Secret)

	authMiddleware := middlewares.NewAuthMiddleware(cfg.JWT.Secret)

	categoryHandler := controllers.NewCategoryHandler(categoryService)
	productHandler := controllers.NewProductHandler(productService, categoryService, fileService)
	authHandler := controllers.NewAuthHandler(authService)

	r := gin.Default()

	// Static files must be registered BEFORE dynamic routes
	r.Static("/upload", cfg.Upload.Dir)

	routes.SetupAuthRoutes(r, authHandler)
	routes.SetupCategoryRoutes(r, categoryHandler, authMiddleware)
	routes.SetupProductRoutes(r, productHandler, authMiddleware)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: r,
	}

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
	if err := migrateDatabase(db.DB); err != nil {
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
	logger.Info("Seeded accounts: admin@collector.local / Admin123!, user@collector.local / User123!, collector@collector.local / Collector123!")
	return nil
}

func migrateDatabase(db *gorm.DB) error {
	return db.AutoMigrate(&models.Category{}, &models.Product{}, &models.User{})
}
