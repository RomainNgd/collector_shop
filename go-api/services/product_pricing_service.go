package services

import (
	"context"
	"fmt"
	"poc-gin/models"

	"gorm.io/gorm"
)

func applyCurrentPricing(ctx context.Context, db *gorm.DB, products []*models.Product) error {
	if len(products) == 0 {
		return nil
	}

	var globalPromotions []models.Promotion
	if err := db.WithContext(ctx).
		Where("is_active = ? AND applies_to_all = ?", true, true).
		Find(&globalPromotions).Error; err != nil {
		return fmt.Errorf("failed to fetch global promotions: %w", err)
	}

	for _, product := range products {
		if product == nil {
			continue
		}
		applyProductPricing(product, globalPromotions)
	}

	return nil
}
