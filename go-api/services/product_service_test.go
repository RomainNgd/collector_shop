package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"testing"
	"time"

	"gorm.io/gorm"
)

func seedProduct(t *testing.T, tx *gorm.DB, categoryID uint) *models.Product {
	t.Helper()

	seller := seedUser(t, tx, "ROLE_USER")
	product := &models.Product{
		Name:        fmt.Sprintf("Product-%d", time.Now().UnixNano()),
		Description: "Test product",
		Image:       "image.png",
		Price:       10.5,
		Stock:       10,
		IsActive:    true,
		SellerID:    &seller.ID,
		CategoryID:  categoryID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	return product
}

func TestProductServiceCRUDAndPreload(t *testing.T) {
	tx := openIntegrationTx(t)
	if err := tx.Where("applies_to_all = ?", true).Delete(&models.Promotion{}).Error; err != nil {
		t.Fatalf("failed to clear global promotions: %v", err)
	}
	category := seedCategory(t, tx)
	seller := seedUser(t, tx, "ROLE_USER")
	service := NewProductService(tx)

	product := &models.Product{
		Name:        fmt.Sprintf("Blue-Eyes-%d", time.Now().UnixNano()),
		Description: "Mint card",
		Image:       "blue-eyes.png",
		Price:       19.99,
		Stock:       5,
		IsActive:    true,
		SellerID:    &seller.ID,
		CategoryID:  category.ID,
	}

	if err := service.CreateProduct(context.Background(), product); err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if product.Category.ID != category.ID {
		t.Fatalf("expected preloaded category %d, got %#v", category.ID, product.Category)
	}
	if product.EffectivePrice != product.Price || product.AppliedPromotion != nil {
		t.Fatalf("expected no promotion on create, got effective=%f promotion=%#v", product.EffectivePrice, product.AppliedPromotion)
	}

	found, err := service.GetProductByID(context.Background(), product.ID)
	if err != nil {
		t.Fatalf("expected get success, got %v", err)
	}
	if found.Category.ID != category.ID {
		t.Fatalf("expected preloaded category %d, got %#v", category.ID, found.Category)
	}
	if found.EffectivePrice != found.Price || found.AppliedPromotion != nil {
		t.Fatalf("expected no promotion on get, got effective=%f promotion=%#v", found.EffectivePrice, found.AppliedPromotion)
	}

	all, err := service.GetAllProducts(context.Background())
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least one product")
	}

	updated, err := service.UpdateProduct(context.Background(), seller.ID, seller.Role, product.ID, map[string]interface{}{
		"name":  product.Name + "-updated",
		"price": 29.99,
		"stock": 7,
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
	if updated.EffectivePrice != updated.Price || updated.AppliedPromotion != nil {
		t.Fatalf("expected no promotion on update, got effective=%f promotion=%#v", updated.EffectivePrice, updated.AppliedPromotion)
	}

	if err := service.DeleteProduct(context.Background(), seller.ID, seller.Role, product.ID); err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
}

func TestProductServiceAppliesSellerPromotion(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	seller := seedUser(t, tx, "ROLE_USER")
	service := NewProductService(tx)

	productA := &models.Product{
		Name:            fmt.Sprintf("Console-%d", time.Now().UnixNano()),
		Description:     "Limited edition",
		Image:           "console.png",
		Price:           100,
		Stock:           3,
		IsActive:        true,
		SellerID:        &seller.ID,
		CategoryID:      category.ID,
		PromotionType:   models.PromotionTypePercentage,
		PromotionValue:  20,
		PromotionActive: true,
	}
	if err := tx.Create(productA).Error; err != nil {
		t.Fatalf("failed to create product A: %v", err)
	}

	foundA, err := service.GetProductByID(context.Background(), productA.ID)
	if err != nil {
		t.Fatalf("expected product A lookup success, got %v", err)
	}
	if foundA.EffectivePrice != 80 {
		t.Fatalf("expected product A effective price 80, got %f", foundA.EffectivePrice)
	}
	if foundA.AppliedPromotion == nil || foundA.AppliedPromotion.Type != models.PromotionTypePercentage {
		t.Fatalf("expected seller promotion for product A, got %#v", foundA.AppliedPromotion)
	}
}

func TestProductServiceUpdateNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProductService(tx)

	user := seedUser(t, tx, "ROLE_USER")

	_, err := service.UpdateProduct(context.Background(), user.ID, user.Role, 999999, map[string]interface{}{"name": "x"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestProductServiceDeleteNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProductService(tx)

	user := seedUser(t, tx, "ROLE_USER")

	err := service.DeleteProduct(context.Background(), user.ID, user.Role, 999999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestCategoryDeletionRestrictedWhenProductReferencesIt(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	_ = seedProduct(t, tx, category.ID)
	categoryService := NewCategoryService(tx)

	err := categoryService.DeleteCategory(context.Background(), category.ID)
	if !errors.Is(err, ErrCategoryInUse) {
		t.Fatalf("expected ErrCategoryInUse, got %v", err)
	}
}
