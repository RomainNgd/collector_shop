package services

import (
	"context"
	"errors"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"testing"

	"gorm.io/gorm"
)

func TestPromotionServiceCRUD(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)

	productA := seedProduct(t, tx, category.ID)
	productB := seedProduct(t, tx, category.ID)

	service := NewPromotionService(tx)

	created, err := service.CreatePromotion(context.Background(), 1, constants.RoleAdmin, PromotionInput{
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

	found, err := service.GetPromotionByID(context.Background(), 1, constants.RoleAdmin, created.ID)
	if err != nil {
		t.Fatalf("expected get success, got %v", err)
	}
	if found.Name != "Spring sale" || !found.IsActive {
		t.Fatalf("unexpected promotion after get: %#v", found)
	}

	all, err := service.GetAllPromotions(context.Background(), 1, constants.RoleAdmin)
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least one promotion")
	}

	updated, err := service.UpdatePromotion(context.Background(), 1, constants.RoleAdmin, created.ID, PromotionInput{
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

	if err := service.DeletePromotion(context.Background(), 1, constants.RoleAdmin, created.ID); err != nil {
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
			_, err := service.CreatePromotion(context.Background(), 1, constants.RoleAdmin, tc.input)
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected %v, got %v", tc.expectedErr, err)
			}
		})
	}
}

func TestPromotionServiceUpdateAndDeleteNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewPromotionService(tx)

	_, err := service.UpdatePromotion(context.Background(), 1, constants.RoleAdmin, 999999, PromotionInput{
		Name:         "Missing",
		Type:         models.PromotionTypeFixed,
		Value:        5,
		IsActive:     true,
		AppliesToAll: true,
	})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}

	err = service.DeletePromotion(context.Background(), 1, constants.RoleAdmin, 999999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestPromotionServiceSellerScoping(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)

	sellerAProduct := seedProduct(t, tx, category.ID)
	sellerBProduct := seedProduct(t, tx, category.ID)
	sellerAID := *sellerAProduct.SellerID
	sellerBID := *sellerBProduct.SellerID

	service := NewPromotionService(tx)

	t.Run("seller cannot create an applies-to-all promotion", func(t *testing.T) {
		_, err := service.CreatePromotion(context.Background(), sellerAID, constants.RoleUser, PromotionInput{
			Name:         "Global",
			Type:         models.PromotionTypeFixed,
			Value:        5,
			IsActive:     true,
			AppliesToAll: true,
		})
		if !errors.Is(err, ErrPromotionAppliesAllDenied) {
			t.Fatalf("expected ErrPromotionAppliesAllDenied, got %v", err)
		}
	})

	t.Run("seller cannot target another seller's product", func(t *testing.T) {
		_, err := service.CreatePromotion(context.Background(), sellerAID, constants.RoleUser, PromotionInput{
			Name:         "Cross seller",
			Type:         models.PromotionTypeFixed,
			Value:        5,
			IsActive:     true,
			AppliesToAll: false,
			ProductIDs:   []uint{sellerBProduct.ID},
		})
		if !errors.Is(err, ErrPromotionProductsNotOwned) {
			t.Fatalf("expected ErrPromotionProductsNotOwned, got %v", err)
		}
	})

	sellerAPromotion, err := service.CreatePromotion(context.Background(), sellerAID, constants.RoleUser, PromotionInput{
		Name:         "Seller A promo",
		Type:         models.PromotionTypeFixed,
		Value:        5,
		IsActive:     true,
		AppliesToAll: false,
		ProductIDs:   []uint{sellerAProduct.ID},
	})
	if err != nil {
		t.Fatalf("expected seller to create own promotion, got %v", err)
	}
	if sellerAPromotion.SellerID == nil || *sellerAPromotion.SellerID != sellerAID {
		t.Fatalf("expected promotion seller_id to be set, got %#v", sellerAPromotion.SellerID)
	}

	t.Run("other seller cannot read it", func(t *testing.T) {
		_, err := service.GetPromotionByID(context.Background(), sellerBID, constants.RoleUser, sellerAPromotion.ID)
		if !errors.Is(err, ErrPromotionAccessDenied) {
			t.Fatalf("expected ErrPromotionAccessDenied, got %v", err)
		}
	})

	t.Run("other seller cannot update it", func(t *testing.T) {
		_, err := service.UpdatePromotion(context.Background(), sellerBID, constants.RoleUser, sellerAPromotion.ID, PromotionInput{
			Name:         "Hijacked",
			Type:         models.PromotionTypeFixed,
			Value:        5,
			IsActive:     true,
			AppliesToAll: false,
			ProductIDs:   []uint{sellerAProduct.ID},
		})
		if !errors.Is(err, ErrPromotionAccessDenied) {
			t.Fatalf("expected ErrPromotionAccessDenied, got %v", err)
		}
	})

	t.Run("other seller cannot delete it", func(t *testing.T) {
		err := service.DeletePromotion(context.Background(), sellerBID, constants.RoleUser, sellerAPromotion.ID)
		if !errors.Is(err, ErrPromotionAccessDenied) {
			t.Fatalf("expected ErrPromotionAccessDenied, got %v", err)
		}
	})

	t.Run("list is scoped to own promotions", func(t *testing.T) {
		all, err := service.GetAllPromotions(context.Background(), sellerBID, constants.RoleUser)
		if err != nil {
			t.Fatalf("expected list success, got %v", err)
		}
		for _, promotion := range all {
			if promotion.ID == sellerAPromotion.ID {
				t.Fatalf("seller B should not see seller A's promotion")
			}
		}
	})

	t.Run("owner can update and delete", func(t *testing.T) {
		updated, err := service.UpdatePromotion(context.Background(), sellerAID, constants.RoleUser, sellerAPromotion.ID, PromotionInput{
			Name:         "Seller A promo updated",
			Type:         models.PromotionTypeFixed,
			Value:        7,
			IsActive:     true,
			AppliesToAll: false,
			ProductIDs:   []uint{sellerAProduct.ID},
		})
		if err != nil {
			t.Fatalf("expected owner update success, got %v", err)
		}
		if updated.Name != "Seller A promo updated" {
			t.Fatalf("unexpected updated promotion: %#v", updated)
		}

		if err := service.DeletePromotion(context.Background(), sellerAID, constants.RoleUser, sellerAPromotion.ID); err != nil {
			t.Fatalf("expected owner delete success, got %v", err)
		}
	})
}
