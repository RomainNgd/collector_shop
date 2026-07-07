package services

import (
	"context"
	"poc-gin/models"

	"gorm.io/gorm"
)

func applyCurrentPricing(ctx context.Context, db *gorm.DB, products []*models.Product) error {
	if len(products) == 0 {
		return nil
	}

	promotionsByProductID, err := loadApplicablePromotions(ctx, db, products)
	if err != nil {
		return err
	}

	for _, product := range products {
		if product == nil {
			continue
		}
		applyProductPricing(product, promotionsByProductID[product.ID])
	}

	return nil
}

type targetedPromotion struct {
	models.Promotion
	ProductID uint
}

func loadApplicablePromotions(ctx context.Context, db *gorm.DB, products []*models.Product) (map[uint][]models.Promotion, error) {
	productIDs := make([]uint, 0, len(products))
	for _, product := range products {
		if product != nil {
			productIDs = append(productIDs, product.ID)
		}
	}

	byProduct := make(map[uint][]models.Promotion, len(productIDs))
	if len(productIDs) == 0 {
		return byProduct, nil
	}

	var globalPromotions []models.Promotion
	if err := db.WithContext(ctx).
		Where("is_active = ? AND applies_to_all = ?", true, true).
		Find(&globalPromotions).Error; err != nil {
		return nil, err
	}

	for _, id := range productIDs {
		byProduct[id] = append([]models.Promotion{}, globalPromotions...)
	}

	var targeted []targetedPromotion
	if err := db.WithContext(ctx).
		Table("promotions").
		Select("promotions.*, product_promotions.product_id as product_id").
		Joins("JOIN product_promotions ON product_promotions.promotion_id = promotions.id").
		Where("promotions.is_active = ? AND product_promotions.product_id IN ?", true, productIDs).
		Scan(&targeted).Error; err != nil {
		return nil, err
	}

	for _, t := range targeted {
		byProduct[t.ProductID] = append(byProduct[t.ProductID], t.Promotion)
	}

	return byProduct, nil
}
