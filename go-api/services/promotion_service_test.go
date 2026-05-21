package services

import (
	"context"
	"errors"
	"poc-gin/models"
	"testing"

	"gorm.io/gorm"
)

func TestPromotionServiceCRUD(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)

	productA := seedProduct(t, tx, category.ID)
	productB := seedProduct(t, tx, category.ID)

	service := NewPromotionService(tx)

	created, err := service.CreatePromotion(context.Background(), PromotionInput{
		Name:         "Spring sale",
		Description:  "Selected products",
		Type:         models.PromotionTypePercentage,
		Value:        15,
		IsActive:     true,
		AppliesToAll: false,
		ProductIDs:   []uint{productB.ID, productA.ID, productA.ID},
	})
	if err != nil {
		t.Fatalf("expected create success, got %v", err)
	}
	if created.ProductCount != 2 {
		t.Fatalf("expected 2 linked products, got %#v", created)
	}
	if len(created.ProductIDs) != 2 || created.ProductIDs[0] != productA.ID || created.ProductIDs[1] != productB.ID {
		t.Fatalf("expected sorted unique product ids, got %#v", created.ProductIDs)
	}

	found, err := service.GetPromotionByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("expected get success, got %v", err)
	}
	if found.Name != "Spring sale" || !found.IsActive {
		t.Fatalf("unexpected promotion after get: %#v", found)
	}

	all, err := service.GetAllPromotions(context.Background())
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least one promotion")
	}

	updated, err := service.UpdatePromotion(context.Background(), created.ID, PromotionInput{
		Name:         "Global sale",
		Description:  "All products",
		Type:         models.PromotionTypeFixed,
		Value:        5,
		IsActive:     false,
		AppliesToAll: true,
	})
	if err != nil {
		t.Fatalf("expected update success, got %v", err)
	}
	if !updated.AppliesToAll || updated.ProductCount != 0 || len(updated.ProductIDs) != 0 {
		t.Fatalf("expected global promotion without linked products, got %#v", updated)
	}
	if updated.Type != models.PromotionTypeFixed || updated.IsActive {
		t.Fatalf("unexpected updated promotion: %#v", updated)
	}

	if err := service.DeletePromotion(context.Background(), created.ID); err != nil {
		t.Fatalf("expected delete success, got %v", err)
	}
}

func TestPromotionServiceValidationErrors(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)
	service := NewPromotionService(tx)

	testCases := []struct {
		name        string
		input       PromotionInput
		expectedErr error
	}{
		{
			name: "invalid type",
			input: PromotionInput{
				Name:         "Invalid",
				Type:         "unknown",
				Value:        5,
				IsActive:     true,
				AppliesToAll: true,
			},
			expectedErr: ErrInvalidPromotionType,
		},
		{
			name: "invalid percentage",
			input: PromotionInput{
				Name:         "Invalid value",
				Type:         models.PromotionTypePercentage,
				Value:        120,
				IsActive:     true,
				AppliesToAll: true,
			},
			expectedErr: ErrInvalidPromotionValue,
		},
		{
			name: "missing products when scoped",
			input: PromotionInput{
				Name:         "Missing products",
				Type:         models.PromotionTypeFixed,
				Value:        5,
				IsActive:     true,
				AppliesToAll: false,
			},
			expectedErr: ErrPromotionProductsEmpty,
		},
		{
			name: "unknown product",
			input: PromotionInput{
				Name:         "Unknown product",
				Type:         models.PromotionTypeFixed,
				Value:        5,
				IsActive:     true,
				AppliesToAll: false,
				ProductIDs:   []uint{product.ID, 999999},
			},
			expectedErr: ErrPromotionProductsNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.CreatePromotion(context.Background(), tc.input)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestPromotionServiceUpdateAndDeleteNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewPromotionService(tx)

	_, err := service.UpdatePromotion(context.Background(), 999999, PromotionInput{
		Name:         "Missing",
		Type:         models.PromotionTypeFixed,
		Value:        5,
		IsActive:     true,
		AppliesToAll: true,
	})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	err = service.DeletePromotion(context.Background(), 999999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}
