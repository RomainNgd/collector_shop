package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

	if len(fixtures.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(fixtures.Categories))
	}
	if len(fixtures.Products) != 18 {
		t.Fatalf("expected 18 products, got %d", len(fixtures.Products))
	}
	if len(fixtures.Promotions) != 3 {
		t.Fatalf("expected 3 promotions, got %d", len(fixtures.Promotions))
	}
	if len(fixtures.Users) != 5 {
		t.Fatalf("expected 5 users, got %d", len(fixtures.Users))
	}
}

func TestValidateDemoFixturesRejectsInvalidData(t *testing.T) {
	tests := []struct {
		name          string
		mutate        func(*demoFixtures)
		errorFragment string
	}{
		{"missing categories", func(f *demoFixtures) { f.Categories = nil }, "at least one category"},
		{"missing products", func(f *demoFixtures) { f.Products = nil }, "at least one product"},
		{"missing promotions", func(f *demoFixtures) { f.Promotions = nil }, "at least one promotion"},
		{"missing users", func(f *demoFixtures) { f.Users = nil }, "at least one user"},
		{"empty category name", func(f *demoFixtures) { f.Categories[0].Name = "" }, "category name"},
		{"empty category description", func(f *demoFixtures) { f.Categories[0].Description = "" }, "include a description"},
		{"duplicate category", func(f *demoFixtures) { f.Categories[1].Name = f.Categories[0].Name }, "duplicated"},
		{"empty product name", func(f *demoFixtures) { f.Products[0].Name = "" }, "product name"},
		{"empty product description", func(f *demoFixtures) { f.Products[0].Description = "" }, "include a description"},
		{"empty product image", func(f *demoFixtures) { f.Products[0].Image = "" }, "include an image"},
		{"unsafe product image", func(f *demoFixtures) { f.Products[0].Image = "../image.png" }, "invalid image filename"},
		{"invalid product price", func(f *demoFixtures) { f.Products[0].Price = 0 }, "positive price"},
		{"unknown category", func(f *demoFixtures) { f.Products[0].Category = "unknown" }, "unknown category"},
		{"missing image", func(f *demoFixtures) { f.Products[0].Image = "missing.png" }, "missing image"},
		{"duplicate product", func(f *demoFixtures) { f.Products[1].Name = f.Products[0].Name }, "duplicated"},
		{"empty promotion name", func(f *demoFixtures) { f.Promotions[0].Name = "" }, "promotion name"},
		{"invalid promotion type", func(f *demoFixtures) { f.Promotions[0].Type = "unknown" }, "invalid type"},
		{"invalid promotion value", func(f *demoFixtures) { f.Promotions[0].Value = 0 }, "positive value"},
		{"excessive percentage", func(f *demoFixtures) {
			f.Promotions[0].Type = models.PromotionTypePercentage
			f.Promotions[0].Value = 101
		}, "cannot exceed 100"},
		{"promotion without products", func(f *demoFixtures) {
			f.Promotions[0].AppliesToAll = false
			f.Promotions[0].Products = nil
		}, "target at least one product"},
		{"empty promotion product", func(f *demoFixtures) {
			f.Promotions[0].AppliesToAll = false
			f.Promotions[0].Products = []string{""}
		}, "empty product reference"},
		{"unknown promotion product", func(f *demoFixtures) {
			f.Promotions[0].AppliesToAll = false
			f.Promotions[0].Products = []string{"unknown"}
		}, "unknown product"},
		{"duplicate promotion product", func(f *demoFixtures) {
			f.Promotions[0].AppliesToAll = false
			f.Promotions[0].Products = []string{f.Products[0].Name, f.Products[0].Name}
		}, "duplicate product"},
		{"duplicate promotion", func(f *demoFixtures) { f.Promotions[1].Name = f.Promotions[0].Name }, "duplicated"},
		{"empty user email", func(f *demoFixtures) { f.Users[0].Email = "" }, "user email"},
		{"empty user password", func(f *demoFixtures) { f.Users[0].Password = "" }, "include a password"},
		{"invalid user role", func(f *demoFixtures) { f.Users[0].Role = "unknown" }, "invalid role"},
		{"duplicate user", func(f *demoFixtures) { f.Users[1].Email = f.Users[0].Email }, "duplicated"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixtures, err := loadDemoFixtures()
			if err != nil {
				t.Fatalf("expected base fixtures to load: %v", err)
			}
			test.mutate(fixtures)

			err = validateDemoFixtures(fixtures)
			if err == nil || !strings.Contains(err.Error(), test.errorFragment) {
				t.Fatalf("expected error containing %q, got %v", test.errorFragment, err)
			}
		})
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
	if err := tx.Preload("Products").Where("name = ?", "Demo - Petits meubles malins").First(&accessoryPromotion).Error; err != nil {
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

func TestSeedReportSummary(t *testing.T) {
	report := &SeedReport{
		CategoriesCreated: 1,
		CategoriesUpdated: 2,
		ProductsCreated:   3,
		ProductsUpdated:   4,
		PromotionsCreated: 5,
		PromotionsUpdated: 6,
		UsersCreated:      7,
		UsersUpdated:      8,
		ImagesWritten:     9,
	}

	summary := report.Summary()
	if !strings.Contains(summary, "categories created=1 updated=2") ||
		!strings.Contains(summary, "products created=3 updated=4") ||
		!strings.Contains(summary, "promotions created=5 updated=6") ||
		!strings.Contains(summary, "users created=7 updated=8") ||
		!strings.Contains(summary, "images synced=9") {
		t.Fatalf("unexpected summary: %q", summary)
	}
}

func TestSeedSellerIDsByEmailReturnsErrorWhenSellerMissing(t *testing.T) {
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

	if err := tx.Where("role = ?", constants.RoleUser).Delete(&models.User{}).Error; err != nil {
		t.Fatalf("failed to clear users: %v", err)
	}

	fixtures := []demoProductFixture{{Name: "Product", Seller: "missing-seller@example.com"}}
	if _, err := seedSellerIDsByEmail(tx, fixtures); err == nil {
		t.Fatal("expected error when seller user does not exist")
	}
}

func TestUpsertProductBackfillsMissingSellerFields(t *testing.T) {
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

	category := &models.Category{Name: fmt.Sprintf("Category-%d", time.Now().UnixNano()), Description: "Test"}
	if err := tx.Create(category).Error; err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	seller := &models.User{Email: fmt.Sprintf("seller-%d@example.com", time.Now().UnixNano()), Password: "hash", Role: constants.RoleUser}
	if err := tx.Create(seller).Error; err != nil {
		t.Fatalf("failed to create seller: %v", err)
	}

	fixture := demoProductFixture{
		Name:        fmt.Sprintf("Legacy-Product-%d", time.Now().UnixNano()),
		Description: "Legacy description",
		Image:       "legacy.png",
		Price:       9.99,
		Category:    "unused",
	}

	// Simulate a pre-existing product created before seller/stock/is_active
	// were introduced: zero stock, inactive, and no seller assigned.
	existing := models.Product{
		Name:        fixture.Name,
		Description: fixture.Description,
		Image:       fixture.Image,
		Price:       fixture.Price,
		Stock:       0,
		CategoryID:  category.ID,
	}
	if err := tx.Create(&existing).Error; err != nil {
		t.Fatalf("failed to seed legacy product: %v", err)
	}
	if err := tx.Model(&existing).Updates(map[string]interface{}{"stock": 0, "is_active": false}).Error; err != nil {
		t.Fatalf("failed to force legacy product state: %v", err)
	}

	created, updated, err := upsertProduct(tx, fixture, category.ID, seller.ID)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if created {
		t.Fatal("expected existing product to be updated, not created")
	}
	if !updated {
		t.Fatal("expected existing product to require an update")
	}

	var reloaded models.Product
	if err := tx.First(&reloaded, existing.ID).Error; err != nil {
		t.Fatalf("failed to reload product: %v", err)
	}
	if reloaded.Stock != 10 {
		t.Fatalf("expected stock backfilled to 10, got %d", reloaded.Stock)
	}
	if !reloaded.IsActive {
		t.Fatal("expected product to be reactivated")
	}
	if reloaded.SellerID == nil || *reloaded.SellerID != seller.ID {
		t.Fatalf("expected seller backfilled to %d, got %#v", seller.ID, reloaded.SellerID)
	}

	// A second upsert with already-healthy fields should be a no-op.
	created2, updated2, err := upsertProduct(tx, fixture, category.ID, seller.ID)
	if err != nil {
		t.Fatalf("expected success on second upsert, got %v", err)
	}
	if created2 || updated2 {
		t.Fatalf("expected idempotent no-op, got created=%v updated=%v", created2, updated2)
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
		if os.Getenv("CI") != "" {
			t.Fatalf("postgres is required in CI: %v", err)
		}
		t.Skipf("postgres not available: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Category{},
		&models.User{},
		&models.Product{},
		&models.Promotion{},
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
