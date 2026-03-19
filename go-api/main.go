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
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database: %v", err)
		}
	}()

	if cfg.Database.AutoMigrate {
		logger.Info("Running database migrations...")
		if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.User{}); err != nil {
			logger.Fatal("Database migration failed: %v", err)
		}
		logger.Info("Database migrations completed")
	}

	fileService, err := services.NewFileService(&cfg.Upload)
	if err != nil {
		logger.Fatal("Failed to initialize file service: %v", err)
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
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Info("Server stopped gracefully")
}
