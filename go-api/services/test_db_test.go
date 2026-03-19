package services

import (
	"fmt"
	"os"
	"poc-gin/models"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func testEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func openIntegrationDB(t *testing.T) *gorm.DB {
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

	return db
}

func openIntegrationTx(t *testing.T) *gorm.DB {
	t.Helper()

	db := openIntegrationDB(t)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin test transaction: %v", tx.Error)
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
