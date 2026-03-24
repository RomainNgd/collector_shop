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

type SeedReport struct {
	CategoriesCreated int
	CategoriesUpdated int
	ProductsCreated   int
	ProductsUpdated   int
	UsersCreated      int
	UsersUpdated      int
	ImagesWritten     int
}

func (r *SeedReport) Summary() string {
	return fmt.Sprintf(
		"categories created=%d updated=%d, products created=%d updated=%d, users created=%d updated=%d, images synced=%d",
		r.CategoriesCreated,
		r.CategoriesUpdated,
		r.ProductsCreated,
		r.ProductsUpdated,
		r.UsersCreated,
		r.UsersUpdated,
		r.ImagesWritten,
	)
}

type demoFixtures struct {
	Categories []demoCategoryFixture `json:"categories"`
	Products   []demoProductFixture  `json:"products"`
	Users      []demoUserFixture     `json:"users"`
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
	categoriesByName := make(map[string]*models.Category, len(fixtures.Categories))

	for _, categoryFixture := range fixtures.Categories {
		category, created, updated, err := upsertCategory(db, categoryFixture)
		if err != nil {
			return nil, err
		}

		categoriesByName[categoryFixture.Name] = category
		if created {
			report.CategoriesCreated++
		}
		if updated {
			report.CategoriesUpdated++
		}
	}

	for _, userFixture := range fixtures.Users {
		created, updated, err := upsertUser(db, userFixture)
		if err != nil {
			return nil, err
		}

		if created {
			report.UsersCreated++
		}
		if updated {
			report.UsersUpdated++
		}
	}

	for _, productFixture := range fixtures.Products {
		category := categoriesByName[productFixture.Category]
		if category == nil {
			return nil, fmt.Errorf("fixture category %q not found for product %q", productFixture.Category, productFixture.Name)
		}

		created, updated, err := upsertProduct(db, productFixture, category.ID)
		if err != nil {
			return nil, err
		}

		if created {
			report.ProductsCreated++
		}
		if updated {
			report.ProductsUpdated++
		}
	}

	return report, nil
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
	if len(fixtures.Categories) == 0 {
		return errors.New("demo fixtures must include at least one category")
	}
	if len(fixtures.Products) == 0 {
		return errors.New("demo fixtures must include at least one product")
	}
	if len(fixtures.Users) == 0 {
		return errors.New("demo fixtures must include at least one user")
	}

	imageEntries, err := demoSeedFiles.ReadDir("fixtures/images")
	if err != nil {
		return fmt.Errorf("failed to list demo images: %w", err)
	}

	availableImages := make(map[string]struct{}, len(imageEntries))
	for _, entry := range imageEntries {
		if entry.IsDir() {
			continue
		}
		availableImages[entry.Name()] = struct{}{}
	}

	categoryNames := make(map[string]struct{}, len(fixtures.Categories))
	for _, category := range fixtures.Categories {
		if strings.TrimSpace(category.Name) == "" {
			return errors.New("fixture category name cannot be empty")
		}
		if strings.TrimSpace(category.Description) == "" {
			return fmt.Errorf("fixture category %q must include a description", category.Name)
		}
		if _, exists := categoryNames[category.Name]; exists {
			return fmt.Errorf("fixture category %q is duplicated", category.Name)
		}
		categoryNames[category.Name] = struct{}{}
	}

	productNames := make(map[string]struct{}, len(fixtures.Products))
	for _, product := range fixtures.Products {
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
		productNames[product.Name] = struct{}{}
	}

	userEmails := make(map[string]struct{}, len(fixtures.Users))
	for _, user := range fixtures.Users {
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
		userEmails[user.Email] = struct{}{}
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
	result := db.Where("name = ?", fixture.Name).Limit(1).Find(&category)
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
	result := db.Where("name = ?", fixture.Name).Limit(1).Find(&product)
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
