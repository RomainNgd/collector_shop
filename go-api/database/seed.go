package database

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"poc-gin/models"
	"poc-gin/pkg/constants"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//go:embed fixtures/demo.json fixtures/images/*
var demoSeedFiles embed.FS

const queryByName = "name = ?"

type SeedReport struct {
	CategoriesCreated int
	CategoriesUpdated int
	ProductsCreated   int
	ProductsUpdated   int
	PromotionsCreated int
	PromotionsUpdated int
	UsersCreated      int
	UsersUpdated      int
	ImagesWritten     int
}

func (r *SeedReport) Summary() string {
	return fmt.Sprintf(
		"categories created=%d updated=%d, products created=%d updated=%d, promotions created=%d updated=%d, users created=%d updated=%d, images synced=%d",
		r.CategoriesCreated,
		r.CategoriesUpdated,
		r.ProductsCreated,
		r.ProductsUpdated,
		r.PromotionsCreated,
		r.PromotionsUpdated,
		r.UsersCreated,
		r.UsersUpdated,
		r.ImagesWritten,
	)
}

type demoFixtures struct {
	Categories []demoCategoryFixture  `json:"categories"`
	Products   []demoProductFixture   `json:"products"`
	Promotions []demoPromotionFixture `json:"promotions"`
	Users      []demoUserFixture      `json:"users"`
}

type demoCategoryFixture struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type demoProductFixture struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
}

type demoUserFixture struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type demoPromotionFixture struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Type         string   `json:"type"`
	Value        float64  `json:"value"`
	IsActive     bool     `json:"is_active"`
	AppliesToAll bool     `json:"applies_to_all"`
	Products     []string `json:"products"`
}

func SeedDemoData(db *gorm.DB, uploadDir string) (*SeedReport, error) {
	if strings.TrimSpace(uploadDir) == "" {
		return nil, errors.New("upload directory is required")
	}

	fixtures, err := loadDemoFixtures()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	imagesWritten, err := syncDemoImages(uploadDir, fixtures.Products)
	if err != nil {
		return nil, err
	}

	report := &SeedReport{ImagesWritten: imagesWritten}
	categoriesByName, err := seedCategories(db, fixtures.Categories, report)
	if err != nil {
		return nil, err
	}
	if err := seedUsers(db, fixtures.Users, report); err != nil {
		return nil, err
	}
	if err := seedProducts(db, fixtures.Products, categoriesByName, report); err != nil {
		return nil, err
	}

	productIDsByName, err := buildProductIDsByName(db, fixtures.Products)
	if err != nil {
		return nil, err
	}
	if err := seedPromotions(db, fixtures.Promotions, productIDsByName, report); err != nil {
		return nil, err
	}

	return report, nil
}

func seedCategories(db *gorm.DB, fixtures []demoCategoryFixture, report *SeedReport) (map[string]*models.Category, error) {
	categoriesByName := make(map[string]*models.Category, len(fixtures))
	for _, fixture := range fixtures {
		category, created, updated, err := upsertCategory(db, fixture)
		if err != nil {
			return nil, err
		}

		categoriesByName[fixture.Name] = category
		if created {
			report.CategoriesCreated++
		}
		if updated {
			report.CategoriesUpdated++
		}
	}
	return categoriesByName, nil
}

func seedUsers(db *gorm.DB, fixtures []demoUserFixture, report *SeedReport) error {
	for _, fixture := range fixtures {
		created, updated, err := upsertUser(db, fixture)
		if err != nil {
			return err
		}
		if created {
			report.UsersCreated++
		}
		if updated {
			report.UsersUpdated++
		}
	}
	return nil
}

func seedProducts(db *gorm.DB, fixtures []demoProductFixture, categories map[string]*models.Category, report *SeedReport) error {
	for _, fixture := range fixtures {
		category := categories[fixture.Category]
		if category == nil {
			return fmt.Errorf("fixture category %q not found for product %q", fixture.Category, fixture.Name)
		}

		created, updated, err := upsertProduct(db, fixture, category.ID)
		if err != nil {
			return err
		}
		if created {
			report.ProductsCreated++
		}
		if updated {
			report.ProductsUpdated++
		}
	}
	return nil
}

func seedPromotions(db *gorm.DB, fixtures []demoPromotionFixture, productIDs map[string]uint, report *SeedReport) error {
	for _, fixture := range fixtures {
		created, updated, err := upsertPromotion(db, fixture, productIDs)
		if err != nil {
			return err
		}
		if created {
			report.PromotionsCreated++
		}
		if updated {
			report.PromotionsUpdated++
		}
	}
	return nil
}

func loadDemoFixtures() (*demoFixtures, error) {
	data, err := demoSeedFiles.ReadFile("fixtures/demo.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read demo fixtures: %w", err)
	}

	var fixtures demoFixtures
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, fmt.Errorf("failed to parse demo fixtures: %w", err)
	}

	if err := validateDemoFixtures(&fixtures); err != nil {
		return nil, err
	}

	return &fixtures, nil
}

func validateDemoFixtures(fixtures *demoFixtures) error {
	if err := validateFixtureCollections(fixtures); err != nil {
		return err
	}

	availableImages, err := availableDemoImages()
	if err != nil {
		return err
	}
	categoryNames, err := validateCategoryFixtures(fixtures.Categories)
	if err != nil {
		return err
	}
	productNames, err := validateProductFixtures(fixtures.Products, categoryNames, availableImages)
	if err != nil {
		return err
	}
	if err := validatePromotionFixtures(fixtures.Promotions, productNames); err != nil {
		return err
	}
	return validateUserFixtures(fixtures.Users)
}

func validateFixtureCollections(fixtures *demoFixtures) error {
	if len(fixtures.Categories) == 0 {
		return errors.New("demo fixtures must include at least one category")
	}
	if len(fixtures.Products) == 0 {
		return errors.New("demo fixtures must include at least one product")
	}
	if len(fixtures.Promotions) == 0 {
		return errors.New("demo fixtures must include at least one promotion")
	}
	if len(fixtures.Users) == 0 {
		return errors.New("demo fixtures must include at least one user")
	}
	return nil
}

func availableDemoImages() (map[string]struct{}, error) {
	imageEntries, err := demoSeedFiles.ReadDir("fixtures/images")
	if err != nil {
		return nil, fmt.Errorf("failed to list demo images: %w", err)
	}

	availableImages := make(map[string]struct{}, len(imageEntries))
	for _, entry := range imageEntries {
		if entry.IsDir() {
			continue
		}
		availableImages[entry.Name()] = struct{}{}
	}
	return availableImages, nil
}

func validateCategoryFixtures(fixtures []demoCategoryFixture) (map[string]struct{}, error) {
	categoryNames := make(map[string]struct{}, len(fixtures))
	for _, category := range fixtures {
		if err := validateCategoryFixture(category, categoryNames); err != nil {
			return nil, err
		}
		categoryNames[category.Name] = struct{}{}
	}
	return categoryNames, nil
}

func validateCategoryFixture(category demoCategoryFixture, categoryNames map[string]struct{}) error {
	if strings.TrimSpace(category.Name) == "" {
		return errors.New("fixture category name cannot be empty")
	}
	if strings.TrimSpace(category.Description) == "" {
		return fmt.Errorf("fixture category %q must include a description", category.Name)
	}
	if _, exists := categoryNames[category.Name]; exists {
		return fmt.Errorf("fixture category %q is duplicated", category.Name)
	}
	return nil
}

func validateProductFixtures(fixtures []demoProductFixture, categoryNames, availableImages map[string]struct{}) (map[string]struct{}, error) {
	productNames := make(map[string]struct{}, len(fixtures))
	for _, product := range fixtures {
		if err := validateProductFixture(product, categoryNames, availableImages, productNames); err != nil {
			return nil, err
		}
		productNames[product.Name] = struct{}{}
	}
	return productNames, nil
}

func validateProductFixture(product demoProductFixture, categoryNames, availableImages, productNames map[string]struct{}) error {
	if strings.TrimSpace(product.Name) == "" {
		return errors.New("fixture product name cannot be empty")
	}
	if strings.TrimSpace(product.Description) == "" {
		return fmt.Errorf("fixture product %q must include a description", product.Name)
	}
	if strings.TrimSpace(product.Image) == "" {
		return fmt.Errorf("fixture product %q must include an image", product.Name)
	}
	if filepath.Base(product.Image) != product.Image {
		return fmt.Errorf("fixture product %q has invalid image filename %q", product.Name, product.Image)
	}
	if product.Price <= 0 {
		return fmt.Errorf("fixture product %q must include a positive price", product.Name)
	}
	if _, exists := categoryNames[product.Category]; !exists {
		return fmt.Errorf("fixture product %q references unknown category %q", product.Name, product.Category)
	}
	if _, exists := availableImages[product.Image]; !exists {
		return fmt.Errorf("fixture product %q references missing image %q", product.Name, product.Image)
	}
	if _, exists := productNames[product.Name]; exists {
		return fmt.Errorf("fixture product %q is duplicated", product.Name)
	}
	return nil
}

func validatePromotionFixtures(fixtures []demoPromotionFixture, productNames map[string]struct{}) error {
	promotionNames := make(map[string]struct{}, len(fixtures))
	for _, promotion := range fixtures {
		if err := validatePromotionFixture(promotion, productNames, promotionNames); err != nil {
			return err
		}
		promotionNames[promotion.Name] = struct{}{}
	}
	return nil
}

func validatePromotionFixture(promotion demoPromotionFixture, productNames, promotionNames map[string]struct{}) error {
	if strings.TrimSpace(promotion.Name) == "" {
		return errors.New("fixture promotion name cannot be empty")
	}
	if promotion.Type != models.PromotionTypePercentage && promotion.Type != models.PromotionTypeFixed {
		return fmt.Errorf("fixture promotion %q has invalid type %q", promotion.Name, promotion.Type)
	}
	if promotion.Value <= 0 {
		return fmt.Errorf("fixture promotion %q must include a positive value", promotion.Name)
	}
	if promotion.Type == models.PromotionTypePercentage && promotion.Value > 100 {
		return fmt.Errorf("fixture promotion %q percentage cannot exceed 100", promotion.Name)
	}
	if !promotion.AppliesToAll && len(promotion.Products) == 0 {
		return fmt.Errorf("fixture promotion %q must target at least one product", promotion.Name)
	}
	if err := validatePromotionProductReferences(promotion, productNames); err != nil {
		return err
	}
	if _, exists := promotionNames[promotion.Name]; exists {
		return fmt.Errorf("fixture promotion %q is duplicated", promotion.Name)
	}
	return nil
}

func validatePromotionProductReferences(promotion demoPromotionFixture, productNames map[string]struct{}) error {
	seenProducts := make(map[string]struct{}, len(promotion.Products))
	for _, productName := range promotion.Products {
		trimmedName := strings.TrimSpace(productName)
		if trimmedName == "" {
			return fmt.Errorf("fixture promotion %q contains an empty product reference", promotion.Name)
		}
		if _, exists := productNames[trimmedName]; !exists {
			return fmt.Errorf("fixture promotion %q references unknown product %q", promotion.Name, trimmedName)
		}
		if _, exists := seenProducts[trimmedName]; exists {
			return fmt.Errorf("fixture promotion %q references duplicate product %q", promotion.Name, trimmedName)
		}
		seenProducts[trimmedName] = struct{}{}
	}
	return nil
}

func validateUserFixtures(fixtures []demoUserFixture) error {
	userEmails := make(map[string]struct{}, len(fixtures))
	for _, user := range fixtures {
		if err := validateUserFixture(user, userEmails); err != nil {
			return err
		}
		userEmails[user.Email] = struct{}{}
	}
	return nil
}

func validateUserFixture(user demoUserFixture, userEmails map[string]struct{}) error {
	if strings.TrimSpace(user.Email) == "" {
		return errors.New("fixture user email cannot be empty")
	}
	if strings.TrimSpace(user.Password) == "" {
		return fmt.Errorf("fixture user %q must include a password", user.Email)
	}
	if user.Role != constants.RoleAdmin && user.Role != constants.RoleUser {
		return fmt.Errorf("fixture user %q has invalid role %q", user.Email, user.Role)
	}
	if _, exists := userEmails[user.Email]; exists {
		return fmt.Errorf("fixture user %q is duplicated", user.Email)
	}
	return nil
}

func syncDemoImages(uploadDir string, products []demoProductFixture) (int, error) {
	imageNames := make(map[string]struct{}, len(products))
	written := 0

	for _, product := range products {
		if _, exists := imageNames[product.Image]; exists {
			continue
		}

		imageNames[product.Image] = struct{}{}
		didWrite, err := writeDemoImage(uploadDir, product.Image)
		if err != nil {
			return 0, err
		}
		if didWrite {
			written++
		}
	}

	return written, nil
}

func writeDemoImage(uploadDir, imageName string) (bool, error) {
	imageData, err := demoSeedFiles.ReadFile(path.Join("fixtures", "images", imageName))
	if err != nil {
		return false, fmt.Errorf("failed to read demo image %q: %w", imageName, err)
	}

	destinationPath := filepath.Join(uploadDir, imageName)
	existingData, err := os.ReadFile(destinationPath)
	if err == nil && bytes.Equal(existingData, imageData) {
		return false, nil
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("failed to read existing demo image %q: %w", imageName, err)
	}

	if err := os.WriteFile(destinationPath, imageData, 0644); err != nil {
		return false, fmt.Errorf("failed to write demo image %q: %w", imageName, err)
	}

	return true, nil
}

func upsertCategory(db *gorm.DB, fixture demoCategoryFixture) (*models.Category, bool, bool, error) {
	var category models.Category
	result := db.Where(queryByName, fixture.Name).Limit(1).Find(&category)
	if result.Error != nil {
		return nil, false, false, fmt.Errorf("failed to fetch fixture category %q: %w", fixture.Name, result.Error)
	}

	if result.RowsAffected == 0 {
		category = models.Category{
			Name:        fixture.Name,
			Description: fixture.Description,
		}
		if err := db.Create(&category).Error; err != nil {
			return nil, false, false, fmt.Errorf("failed to create fixture category %q: %w", fixture.Name, err)
		}
		return &category, true, false, nil
	}

	if category.Description == fixture.Description {
		return &category, false, false, nil
	}

	if err := db.Model(&category).Update("description", fixture.Description).Error; err != nil {
		return nil, false, false, fmt.Errorf("failed to update fixture category %q: %w", fixture.Name, err)
	}

	category.Description = fixture.Description
	return &category, false, true, nil
}

func upsertUser(db *gorm.DB, fixture demoUserFixture) (bool, bool, error) {
	var user models.User
	result := db.Where("email = ?", fixture.Email).Limit(1).Find(&user)
	if result.Error != nil {
		return false, false, fmt.Errorf("failed to fetch fixture user %q: %w", fixture.Email, result.Error)
	}

	if result.RowsAffected == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(fixture.Password), bcrypt.DefaultCost)
		if err != nil {
			return false, false, fmt.Errorf("failed to hash fixture password for %q: %w", fixture.Email, err)
		}

		user = models.User{
			Email:    fixture.Email,
			Password: string(hashedPassword),
			Role:     fixture.Role,
		}
		if err := db.Create(&user).Error; err != nil {
			return false, false, fmt.Errorf("failed to create fixture user %q: %w", fixture.Email, err)
		}
		return true, false, nil
	}

	updates := make(map[string]interface{})
	if user.Role != fixture.Role {
		updates["role"] = fixture.Role
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(fixture.Password)) != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(fixture.Password), bcrypt.DefaultCost)
		if err != nil {
			return false, false, fmt.Errorf("failed to hash fixture password for %q: %w", fixture.Email, err)
		}
		updates["password"] = string(hashedPassword)
	}

	if len(updates) == 0 {
		return false, false, nil
	}

	if err := db.Model(&user).Updates(updates).Error; err != nil {
		return false, false, fmt.Errorf("failed to update fixture user %q: %w", fixture.Email, err)
	}

	return false, true, nil
}

func upsertProduct(db *gorm.DB, fixture demoProductFixture, categoryID uint) (bool, bool, error) {
	var product models.Product
	result := db.Where(queryByName, fixture.Name).Limit(1).Find(&product)
	if result.Error != nil {
		return false, false, fmt.Errorf("failed to fetch fixture product %q: %w", fixture.Name, result.Error)
	}

	if result.RowsAffected == 0 {
		product = models.Product{
			Name:        fixture.Name,
			Description: fixture.Description,
			Image:       fixture.Image,
			Price:       fixture.Price,
			CategoryID:  categoryID,
		}
		if err := db.Create(&product).Error; err != nil {
			return false, false, fmt.Errorf("failed to create fixture product %q: %w", fixture.Name, err)
		}
		return true, false, nil
	}

	updates := make(map[string]interface{})
	if product.Description != fixture.Description {
		updates["description"] = fixture.Description
	}
	if product.Image != fixture.Image {
		updates["image"] = fixture.Image
	}
	if product.Price != fixture.Price {
		updates["price"] = fixture.Price
	}
	if product.CategoryID != categoryID {
		updates["category_id"] = categoryID
	}

	if len(updates) == 0 {
		return false, false, nil
	}

	if err := db.Model(&product).Updates(updates).Error; err != nil {
		return false, false, fmt.Errorf("failed to update fixture product %q: %w", fixture.Name, err)
	}

	return false, true, nil
}

func buildProductIDsByName(db *gorm.DB, fixtures []demoProductFixture) (map[string]uint, error) {
	productNames := make([]string, 0, len(fixtures))
	for _, fixture := range fixtures {
		productNames = append(productNames, fixture.Name)
	}

	var products []models.Product
	if err := db.Where("name IN ?", productNames).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch seeded products for promotions: %w", err)
	}

	productIDsByName := make(map[string]uint, len(products))
	for _, product := range products {
		productIDsByName[product.Name] = product.ID
	}

	if len(productIDsByName) != len(fixtures) {
		return nil, errors.New("failed to resolve all seeded products for promotions")
	}

	return productIDsByName, nil
}

func upsertPromotion(db *gorm.DB, fixture demoPromotionFixture, productIDsByName map[string]uint) (bool, bool, error) {
	var promotion models.Promotion
	result := db.Where(queryByName, fixture.Name).Limit(1).Find(&promotion)
	if result.Error != nil {
		return false, false, fmt.Errorf("failed to fetch fixture promotion %q: %w", fixture.Name, result.Error)
	}

	productIDs, err := resolvePromotionProductIDs(fixture, productIDsByName)
	if err != nil {
		return false, false, err
	}

	if result.RowsAffected == 0 {
		return createFixturePromotion(db, fixture, productIDs)
	}

	updates := promotionFixtureUpdates(&promotion, fixture)
	linksChanged, err := promotionLinksChanged(db, &promotion, productIDs)
	if err != nil {
		return false, false, err
	}

	if len(updates) == 0 && !linksChanged {
		return false, false, nil
	}

	if len(updates) > 0 {
		if err := db.Model(&promotion).Updates(updates).Error; err != nil {
			return false, false, fmt.Errorf("failed to update fixture promotion %q: %w", fixture.Name, err)
		}
	}

	if linksChanged {
		if err := replacePromotionProducts(db, &promotion, productIDs); err != nil {
			return false, false, err
		}
	}

	return false, true, nil
}

func resolvePromotionProductIDs(fixture demoPromotionFixture, productIDsByName map[string]uint) ([]uint, error) {
	productIDs := make([]uint, 0, len(fixture.Products))
	for _, productName := range fixture.Products {
		productID, exists := productIDsByName[productName]
		if !exists {
			return nil, fmt.Errorf("fixture promotion %q references unresolved product %q", fixture.Name, productName)
		}
		productIDs = append(productIDs, productID)
	}
	return productIDs, nil
}

func createFixturePromotion(db *gorm.DB, fixture demoPromotionFixture, productIDs []uint) (bool, bool, error) {
	promotion := models.Promotion{
		Name:         fixture.Name,
		Description:  fixture.Description,
		Type:         fixture.Type,
		Value:        fixture.Value,
		IsActive:     fixture.IsActive,
		AppliesToAll: fixture.AppliesToAll,
	}
	if err := db.Create(&promotion).Error; err != nil {
		return false, false, fmt.Errorf("failed to create fixture promotion %q: %w", fixture.Name, err)
	}
	if err := replacePromotionProducts(db, &promotion, productIDs); err != nil {
		return false, false, err
	}
	return true, false, nil
}

func promotionFixtureUpdates(promotion *models.Promotion, fixture demoPromotionFixture) map[string]interface{} {
	updates := make(map[string]interface{})
	if promotion.Description != fixture.Description {
		updates["description"] = fixture.Description
	}
	if promotion.Type != fixture.Type {
		updates["type"] = fixture.Type
	}
	if promotion.Value != fixture.Value {
		updates["value"] = fixture.Value
	}
	if promotion.IsActive != fixture.IsActive {
		updates["is_active"] = fixture.IsActive
	}
	if promotion.AppliesToAll != fixture.AppliesToAll {
		updates["applies_to_all"] = fixture.AppliesToAll
	}
	return updates
}

func replacePromotionProducts(db *gorm.DB, promotion *models.Promotion, productIDs []uint) error {
	if promotion.AppliesToAll || len(productIDs) == 0 {
		if err := db.Model(promotion).Association("Products").Clear(); err != nil {
			return fmt.Errorf("failed to clear fixture promotion products for %q: %w", promotion.Name, err)
		}
		return nil
	}

	var products []models.Product
	if err := db.Where("id IN ?", productIDs).Order("id ASC").Find(&products).Error; err != nil {
		return fmt.Errorf("failed to fetch fixture promotion products for %q: %w", promotion.Name, err)
	}

	if err := db.Model(promotion).Association("Products").Replace(products); err != nil {
		return fmt.Errorf("failed to replace fixture promotion products for %q: %w", promotion.Name, err)
	}

	return nil
}

func promotionLinksChanged(db *gorm.DB, promotion *models.Promotion, expectedProductIDs []uint) (bool, error) {
	var currentProducts []models.Product
	if err := db.Model(promotion).Association("Products").Find(&currentProducts); err != nil {
		return false, fmt.Errorf("failed to load fixture promotion products for %q: %w", promotion.Name, err)
	}

	if len(currentProducts) != len(expectedProductIDs) {
		return true, nil
	}

	currentIDs := make(map[uint]struct{}, len(currentProducts))
	for _, product := range currentProducts {
		currentIDs[product.ID] = struct{}{}
	}

	for _, expectedID := range expectedProductIDs {
		if _, exists := currentIDs[expectedID]; !exists {
			return true, nil
		}
	}

	return false, nil
}
