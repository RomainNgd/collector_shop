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
	if err := tx.Where("applies_to_all = ?", true).Delete(&models.Promotion{}).Error; err != nil {
		t.Fatalf("failed to clear global promotions: %v", err)
	}
	category := seedCategory(t, tx)
	service := NewProductService(tx)

	product := &models.Product{
		Name:        fmt.Sprintf("Blue-Eyes-%d", time.Now().UnixNano()),
		Description: "Mint card",
		Image:       "blue-eyes.png",
		Price:       19.99,
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

	updated, err := service.UpdateProduct(context.Background(), product.ID, map[string]interface{}{
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
	if updated.EffectivePrice != updated.Price || updated.AppliedPromotion != nil {
		t.Fatalf("expected no promotion on update, got effective=%f promotion=%#v", updated.EffectivePrice, updated.AppliedPromotion)
	}

	if err := service.DeleteProduct(context.Background(), product.ID); err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
}

func TestProductServiceAppliesBestPromotion(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	service := NewProductService(tx)

	productA := &models.Product{
		Name:        "Console",
		Description: "Limited edition",
		Image:       "console.png",
		Price:       100,
		CategoryID:  category.ID,
	}
	productB := &models.Product{
		Name:        "Binder",
		Description: "Premium binder",
		Image:       "binder.png",
		Price:       40,
		CategoryID:  category.ID,
	}
	if err := tx.Create(productA).Error; err != nil {
		t.Fatalf("failed to create product A: %v", err)
	}
	if err := tx.Create(productB).Error; err != nil {
		t.Fatalf("failed to create product B: %v", err)
	}

	globalPromotion := &models.Promotion{
		Name:         "Global",
		Type:         models.PromotionTypeFixed,
		Value:        5,
		IsActive:     true,
		AppliesToAll: true,
	}
	targetedPromotion := &models.Promotion{
		Name:         "Targeted",
		Type:         models.PromotionTypePercentage,
		Value:        20,
		IsActive:     true,
		AppliesToAll: false,
	}
	if err := tx.Create(globalPromotion).Error; err != nil {
		t.Fatalf("failed to create global promotion: %v", err)
	}
	if err := tx.Create(targetedPromotion).Error; err != nil {
		t.Fatalf("failed to create targeted promotion: %v", err)
	}
	if err := tx.Model(targetedPromotion).Association("Products").Append(productA); err != nil {
		t.Fatalf("failed to link targeted promotion: %v", err)
	}

	foundA, err := service.GetProductByID(context.Background(), productA.ID)
	if err != nil {
		t.Fatalf("expected product A lookup success, got %v", err)
	}
	if foundA.EffectivePrice != 80 {
		t.Fatalf("expected product A effective price 80, got %f", foundA.EffectivePrice)
	}
	if foundA.AppliedPromotion == nil || foundA.AppliedPromotion.ID != targetedPromotion.ID {
		t.Fatalf("expected targeted promotion for product A, got %#v", foundA.AppliedPromotion)
	}

	foundB, err := service.GetProductByID(context.Background(), productB.ID)
	if err != nil {
		t.Fatalf("expected product B lookup success, got %v", err)
	}
	if foundB.EffectivePrice != 35 {
		t.Fatalf("expected product B effective price 35, got %f", foundB.EffectivePrice)
	}
	if foundB.AppliedPromotion == nil || foundB.AppliedPromotion.ID != globalPromotion.ID {
		t.Fatalf("expected global promotion for product B, got %#v", foundB.AppliedPromotion)
	}
}

func TestProductServiceUpdateNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProductService(tx)

	_, err := service.UpdateProduct(context.Background(), 999999, map[string]interface{}{"name": "x"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestProductServiceDeleteNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProductService(tx)

	err := service.DeleteProduct(context.Background(), 999999)
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
