package services

import (
	"errors"
	"fmt"
	"poc-gin/models"
	"testing"
	"time"

	"gorm.io/gorm"
)

func seedProduct(t *testing.T, tx *gorm.DB, categoryID uint) *models.Product {
	t.Helper()

	product := &models.Product{
		Name:        fmt.Sprintf("Product-%d", time.Now().UnixNano()),
		Description: "Test product",
		Image:       "image.png",
		Price:       10.5,
		CategoryID:  categoryID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	return product
}

func TestProductServiceCRUDAndPreload(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	service := NewProductService(tx)

	product := &models.Product{
		Name:        fmt.Sprintf("Blue-Eyes-%d", time.Now().UnixNano()),
		Description: "Mint card",
		Image:       "blue-eyes.png",
		Price:       19.99,
		CategoryID:  category.ID,
	}

	if err := service.CreateProduct(product); err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if product.Category.ID != category.ID {
		t.Fatalf("expected preloaded category %d, got %#v", category.ID, product.Category)
	}

	found, err := service.GetProductByID(product.ID)
	if err != nil {
		t.Fatalf("expected get success, got %v", err)
	}
	if found.Category.ID != category.ID {
		t.Fatalf("expected preloaded category %d, got %#v", category.ID, found.Category)
	}

	all, err := service.GetAllProducts()
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least one product")
	}

	updated, err := service.UpdateProduct(product.ID, map[string]interface{}{
		"name":  product.Name + "-updated",
		"price": 29.99,
	})
	if err != nil {
		t.Fatalf("expected update success, got %v", err)
	}
	if updated.Price != 29.99 {
		t.Fatalf("expected updated price, got %f", updated.Price)
	}
	if updated.Category.ID != category.ID {
		t.Fatalf("expected preloaded category after update, got %#v", updated.Category)
	}

	if err := service.DeleteProduct(product.ID); err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
}

func TestProductServiceUpdateNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProductService(tx)

	_, err := service.UpdateProduct(999999, map[string]interface{}{"name": "x"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestProductServiceDeleteNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProductService(tx)

	err := service.DeleteProduct(999999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestCategoryDeletionRestrictedWhenProductReferencesIt(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	_ = seedProduct(t, tx, category.ID)
	categoryService := NewCategoryService(tx)

	err := categoryService.DeleteCategory(category.ID)
	if !errors.Is(err, ErrCategoryInUse) {
		t.Fatalf("expected ErrCategoryInUse, got %v", err)
	}
}
