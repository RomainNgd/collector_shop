package services

import (
	"poc-gin/models"
	"testing"

	"gorm.io/gorm"
)

func TestChooseBestPromotion(t *testing.T) {
	finalPrice, applied := chooseBestPromotion(100, []models.Promotion{
		{Model: gorm.Model{ID: 10}, Name: "Fixed", Type: models.PromotionTypeFixed, Value: 5, IsActive: true},
		{Model: gorm.Model{ID: 2}, Name: "Best percent", Type: models.PromotionTypePercentage, Value: 20, IsActive: true},
		{Model: gorm.Model{ID: 1}, Name: "Inactive", Type: models.PromotionTypeFixed, Value: 50, IsActive: false},
	})

	if finalPrice != 80 {
		t.Fatalf("expected best final price 80, got %f", finalPrice)
	}
	if applied == nil || applied.Name != "Best percent" || applied.DiscountAmount != 20 {
		t.Fatalf("unexpected applied promotion: %#v", applied)
	}
}

func TestApplyProductPricingFloorsAtZero(t *testing.T) {
	product := &models.Product{
		Price: 12,
		Promotions: []models.Promotion{
			{Model: gorm.Model{ID: 3}, Name: "Too large", Type: models.PromotionTypeFixed, Value: 20, IsActive: true},
		},
	}

	applyProductPricing(product, nil)

	if product.EffectivePrice != 0 {
		t.Fatalf("expected price floor at 0, got %f", product.EffectivePrice)
	}
	if product.AppliedPromotion == nil || product.AppliedPromotion.DiscountAmount != 12 {
		t.Fatalf("unexpected applied promotion: %#v", product.AppliedPromotion)
	}
}
