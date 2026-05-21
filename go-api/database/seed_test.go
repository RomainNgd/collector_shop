package database

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"poc-gin/models"
	"poc-gin/pkg/constants"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestLoadDemoFixtures(t *testing.T) {
	fixtures, err := loadDemoFixtures()
	if err != nil {
		t.Fatalf("expected demo fixtures to load, got %v", err)
	}

	if len(fixtures.Categories) != 4 {
		t.Fatalf("expected 4 categories, got %d", len(fixtures.Categories))
	}
	if len(fixtures.Products) != 6 {
		t.Fatalf("expected 6 products, got %d", len(fixtures.Products))
	}
	if len(fixtures.Promotions) != 3 {
		t.Fatalf("expected 3 promotions, got %d", len(fixtures.Promotions))
	}
	if len(fixtures.Users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(fixtures.Users))
	}
}

func TestSeedDemoData(t *testing.T) {
	db := openSeedTestDB(t)

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin transaction: %v", tx.Error)
	}

	t.Cleanup(func() {
		_ = tx.Rollback().Error
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	fixtures, err := loadDemoFixtures()
	if err != nil {
		t.Fatalf("failed to load demo fixtures: %v", err)
	}

	if err := removeFixtureRecords(tx, fixtures); err != nil {
		t.Fatalf("failed to cleanup fixture records in transaction: %v", err)
	}

	uploadDir := filepath.Join(t.TempDir(), "upload")

	report, err := SeedDemoData(tx, uploadDir)
	if err != nil {
		t.Fatalf("expected demo seed to succeed, got %v", err)
	}

	if report.CategoriesCreated != len(fixtures.Categories) || report.CategoriesUpdated != 0 {
		t.Fatalf("unexpected category report: %#v", report)
	}
	if report.ProductsCreated != len(fixtures.Products) || report.ProductsUpdated != 0 {
		t.Fatalf("unexpected product report: %#v", report)
	}
	if report.PromotionsCreated != len(fixtures.Promotions) || report.PromotionsUpdated != 0 {
		t.Fatalf("unexpected promotion report: %#v", report)
	}
	if report.UsersCreated != len(fixtures.Users) || report.UsersUpdated != 0 {
		t.Fatalf("unexpected user report: %#v", report)
	}
	if report.ImagesWritten != countUniqueImages(fixtures.Products) {
		t.Fatalf("expected %d images synced, got %#v", countUniqueImages(fixtures.Products), report)
	}

	var categoryCount int64
	if err := tx.Model(&models.Category{}).Count(&categoryCount).Error; err != nil {
		t.Fatalf("failed to count categories: %v", err)
	}
	if categoryCount < int64(len(fixtures.Categories)) {
		t.Fatalf("expected at least %d categories in database, got %d", len(fixtures.Categories), categoryCount)
	}

	var productCount int64
	if err := tx.Model(&models.Product{}).Count(&productCount).Error; err != nil {
		t.Fatalf("failed to count products: %v", err)
	}
	if productCount < int64(len(fixtures.Products)) {
		t.Fatalf("expected at least %d products in database, got %d", len(fixtures.Products), productCount)
	}

	var userCount int64
	if err := tx.Model(&models.User{}).Count(&userCount).Error; err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if userCount < int64(len(fixtures.Users)) {
		t.Fatalf("expected at least %d users in database, got %d", len(fixtures.Users), userCount)
	}

	var promotionCount int64
	if err := tx.Model(&models.Promotion{}).Count(&promotionCount).Error; err != nil {
		t.Fatalf("failed to count promotions: %v", err)
	}
	if promotionCount < int64(len(fixtures.Promotions)) {
		t.Fatalf("expected at least %d promotions in database, got %d", len(fixtures.Promotions), promotionCount)
	}

	var adminFixture demoUserFixture
	foundAdminFixture := false
	for _, fixtureUser := range fixtures.Users {
		if fixtureUser.Role == constants.RoleAdmin {
			adminFixture = fixtureUser
			foundAdminFixture = true
			break
		}
	}
	if !foundAdminFixture {
		t.Fatal("expected one admin fixture user")
	}

	var admin models.User
	if err := tx.Where("email = ?", adminFixture.Email).First(&admin).Error; err != nil {
		t.Fatalf("expected seeded admin user, got %v", err)
	}
	if admin.Role != constants.RoleAdmin {
		t.Fatalf("expected admin role %s, got %s", constants.RoleAdmin, admin.Role)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(adminFixture.Password)); err != nil {
		t.Fatalf("expected seeded admin password to match fixture, got %v", err)
	}

	for _, product := range fixtures.Products {
		if _, err := os.Stat(filepath.Join(uploadDir, product.Image)); err != nil {
			t.Fatalf("expected demo image %s to exist, got %v", product.Image, err)
		}
	}

	var globalPromotion models.Promotion
	if err := tx.Where("name = ?", "Demo - Offre vitrine").First(&globalPromotion).Error; err != nil {
		t.Fatalf("expected seeded global promotion, got %v", err)
	}
	if !globalPromotion.AppliesToAll || !globalPromotion.IsActive {
		t.Fatalf("expected global promotion to be active and global, got %#v", globalPromotion)
	}

	var accessoryPromotion models.Promotion
	if err := tx.Preload("Products").Where("name = ?", "Demo - Accessoires malins").First(&accessoryPromotion).Error; err != nil {
		t.Fatalf("expected seeded accessory promotion, got %v", err)
	}
	if accessoryPromotion.AppliesToAll {
		t.Fatalf("expected accessory promotion to stay targeted, got %#v", accessoryPromotion)
	}
	if len(accessoryPromotion.Products) != 2 {
		t.Fatalf("expected 2 linked products for accessory promotion, got %#v", accessoryPromotion.Products)
	}

	secondReport, err := SeedDemoData(tx, uploadDir)
	if err != nil {
		t.Fatalf("expected second seed to succeed, got %v", err)
	}

	if secondReport.CategoriesCreated != 0 || secondReport.CategoriesUpdated != 0 {
		t.Fatalf("expected category seed to be idempotent, got %#v", secondReport)
	}
	if secondReport.ProductsCreated != 0 || secondReport.ProductsUpdated != 0 {
		t.Fatalf("expected product seed to be idempotent, got %#v", secondReport)
	}
	if secondReport.PromotionsCreated != 0 || secondReport.PromotionsUpdated != 0 {
		t.Fatalf("expected promotion seed to be idempotent, got %#v", secondReport)
	}
	if secondReport.UsersCreated != 0 || secondReport.UsersUpdated != 0 {
		t.Fatalf("expected user seed to be idempotent, got %#v", secondReport)
	}
	if secondReport.ImagesWritten != 0 {
		t.Fatalf("expected image sync to be idempotent, got %#v", secondReport)
	}
}

func openSeedTestDB(t *testing.T) *gorm.DB {
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

	return db
}

func testEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func countUniqueImages(products []demoProductFixture) int {
	images := make(map[string]struct{}, len(products))
	for _, product := range products {
		images[product.Image] = struct{}{}
	}
	return len(images)
}

func removeFixtureRecords(db *gorm.DB, fixtures *demoFixtures) error {
	productNames := make([]string, 0, len(fixtures.Products))
	for _, product := range fixtures.Products {
		productNames = append(productNames, product.Name)
	}

	promotionNames := make([]string, 0, len(fixtures.Promotions))
	for _, promotion := range fixtures.Promotions {
		promotionNames = append(promotionNames, promotion.Name)
	}

	categoryNames := make([]string, 0, len(fixtures.Categories))
	for _, category := range fixtures.Categories {
		categoryNames = append(categoryNames, category.Name)
	}

	userEmails := make([]string, 0, len(fixtures.Users))
	for _, user := range fixtures.Users {
		userEmails = append(userEmails, user.Email)
	}

	if len(productNames) > 0 {
		if err := db.Unscoped().Where("name IN ?", productNames).Delete(&models.Product{}).Error; err != nil {
			return err
		}
	}

	if len(promotionNames) > 0 {
		if err := db.Unscoped().Where("name IN ?", promotionNames).Delete(&models.Promotion{}).Error; err != nil {
			return err
		}
	}

	if len(categoryNames) > 0 {
		if err := db.Unscoped().Where("name IN ?", categoryNames).Delete(&models.Category{}).Error; err != nil {
			return err
		}
	}

	if len(userEmails) > 0 {
		orderIDsForUsers := db.Model(&models.Order{}).
			Select("id").
			Where("user_id IN (?)", db.Model(&models.User{}).Select("id").Where("email IN ?", userEmails))

		if err := db.Unscoped().Where("order_id IN (?)", orderIDsForUsers).Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("user_id IN (?)", db.Model(&models.User{}).Select("id").Where("email IN ?", userEmails)).Delete(&models.Order{}).Error; err != nil {
			return err
		}
		if err := db.Unscoped().Where("email IN ?", userEmails).Delete(&models.User{}).Error; err != nil {
			return err
		}
	}

	return nil
}
