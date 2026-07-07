package services

import (
	"math"
	"poc-gin/models"

	"gorm.io/gorm"
)

func roundCurrency(value float64) float64 {
	return math.Round(value*100) / 100
}

func promotionDiscount(basePrice float64, promotion models.Promotion) float64 {
	switch promotion.Type {
	case models.PromotionTypePercentage:
		return roundCurrency(basePrice * (promotion.Value / 100))
	case models.PromotionTypeFixed:
		return roundCurrency(promotion.Value)
	default:
		return 0
	}
}

func promotionFinalPrice(basePrice float64, promotion models.Promotion) (float64, float64) {
	discount := promotionDiscount(basePrice, promotion)
	if discount <= 0 {
		return roundCurrency(basePrice), 0
	}

	finalPrice := roundCurrency(basePrice - discount)
	if finalPrice < 0 {
		finalPrice = 0
		discount = roundCurrency(basePrice)
	}

	return finalPrice, discount
}

func chooseBestPromotion(basePrice float64, promotions []models.Promotion) (float64, *models.AppliedPromotion) {
	bestPrice := roundCurrency(basePrice)
	var applied *models.AppliedPromotion

	for _, promotion := range promotions {
		if !promotion.IsActive {
			continue
		}

		finalPrice, discountAmount := promotionFinalPrice(basePrice, promotion)
		if discountAmount <= 0 {
			continue
		}

		if applied == nil || finalPrice < bestPrice || (finalPrice == bestPrice && promotion.ID < applied.ID) {
			bestPrice = finalPrice
			applied = &models.AppliedPromotion{
				ID:             promotion.ID,
				Name:           promotion.Name,
				Type:           promotion.Type,
				Value:          promotion.Value,
				DiscountAmount: discountAmount,
				AppliesToAll:   promotion.AppliesToAll,
			}
		}
	}

	return bestPrice, applied
}

func applyProductPricing(product *models.Product, _ []models.Promotion) {
	product.EffectivePrice = roundCurrency(product.Price)
	product.AppliedPromotion = nil

	if product.PromotionActive {
		bestPrice, appliedPromotion := chooseBestPromotion(product.Price, []models.Promotion{
			{
				Model:        gormModel(product.ID),
				Name:         "Promotion vendeur",
				Type:         product.PromotionType,
				Value:        product.PromotionValue,
				IsActive:     product.PromotionActive,
				AppliesToAll: false,
			},
		})
		product.EffectivePrice = bestPrice
		product.AppliedPromotion = appliedPromotion
		return
	}
}

func gormModel(id uint) gorm.Model {
	return gorm.Model{ID: id}
}
