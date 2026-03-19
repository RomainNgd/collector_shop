package services

import (
	"errors"
	"fmt"
	"poc-gin/models"
	"testing"
	"time"

	"gorm.io/gorm"
)

func seedCategory(t *testing.T, tx *gorm.DB) *models.Category {
	t.Helper()

	category := &models.Category{
		Name:        fmt.Sprintf("Category-%d", time.Now().UnixNano()),
		Description: "Test category",
	}
	if err := tx.Create(category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}
	return category
}

func TestCategoryServiceCRUD(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewCategoryService(tx)

	category := &models.Category{
		Name:        fmt.Sprintf("Cards-%d", time.Now().UnixNano()),
		Description: "Trading cards",
	}

	if err := service.CreateCategory(category); err != nil {
		t.Fatalf("expected create success, got %v", err)
	}

	found, err := service.GetCategoryByID(category.ID)
	if err != nil {
		t.Fatalf("expected category fetch success, got %v", err)
	}
	if found.Name != category.Name {
		t.Fatalf("expected category name %s, got %s", category.Name, found.Name)
	}

	all, err := service.GetAllCategories()
	if err != nil {
		t.Fatalf("expected all categories success, got %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least one category")
	}

	updated, err := service.UpdateCategory(category.ID, map[string]interface{}{
		"name":        category.Name + "-updated",
		"description": "Updated description",
	})
	if err != nil {
		t.Fatalf("expected update success, got %v", err)
	}
	if updated.Description != "Updated description" {
		t.Fatalf("expected updated description, got %s", updated.Description)
	}

	if err := service.DeleteCategory(category.ID); err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
}

func TestCategoryServiceUpdateNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewCategoryService(tx)

	_, err := service.UpdateCategory(999999, map[string]interface{}{"name": "x"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestCategoryServiceDeleteNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewCategoryService(tx)

	err := service.DeleteCategory(999999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}
