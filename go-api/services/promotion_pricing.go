package services

import (
	"math"
	"poc-gin/models"
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

func applyProductPricing(product *models.Product, globalPromotions []models.Promotion) {
	product.EffectivePrice = roundCurrency(product.Price)
	product.AppliedPromotion = nil

	applicablePromotions := make([]models.Promotion, 0, len(product.Promotions)+len(globalPromotions))

	for _, promotion := range product.Promotions {
		if promotion.IsActive {
			applicablePromotions = append(applicablePromotions, promotion)
		}
	}

	applicablePromotions = append(applicablePromotions, globalPromotions...)

	bestPrice, appliedPromotion := chooseBestPromotion(product.Price, applicablePromotions)
	product.EffectivePrice = bestPrice
	product.AppliedPromotion = appliedPromotion
}
