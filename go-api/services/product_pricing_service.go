package services

import (
	"context"
	"poc-gin/models"

	"gorm.io/gorm"
)

func applyCurrentPricing(_ context.Context, _ *gorm.DB, products []*models.Product) error {
	if len(products) == 0 {
		return nil
	}

	for _, product := range products {
		if product == nil {
			continue
		}
		applyProductPricing(product, nil)
	}

	return nil
}
