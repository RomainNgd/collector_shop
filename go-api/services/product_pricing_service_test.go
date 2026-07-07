package services

import (
	"context"
	"fmt"
	"poc-gin/models"
	"testing"
	"time"
)

func TestLoadApplicablePromotionsReturnsEmptyForNoProducts(t *testing.T) {
	tx := openIntegrationTx(t)

	byProduct, err := loadApplicablePromotions(context.Background(), tx, nil)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(byProduct) != 0 {
		t.Fatalf("expected empty map, got %#v", byProduct)
	}
}

func TestLoadApplicablePromotionsGlobalAppliesToAllProducts(t *testing.T) {
	tx := openIntegrationTx(t)
	if err := tx.Where("applies_to_all = ?", true).Delete(&models.Promotion{}).Error; err != nil {
		t.Fatalf("failed to clear existing global promotions: %v", err)
	}

	category := seedCategory(t, tx)
	productA := seedProduct(t, tx, category.ID)
	productB := seedProduct(t, tx, category.ID)

	global := &models.Promotion{
		Name:         fmt.Sprintf("Global-%d", time.Now().UnixNano()),
		Type:         models.PromotionTypePercentage,
		Value:        15,
		IsActive:     true,
		AppliesToAll: true,
	}
	if err := tx.Create(global).Error; err != nil {
		t.Fatalf("failed to create global promotion: %v", err)
	}

	byProduct, err := loadApplicablePromotions(context.Background(), tx, []*models.Product{productA, productB})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	for _, product := range []*models.Product{productA, productB} {
		promotions := byProduct[product.ID]
		if len(promotions) != 1 || promotions[0].ID != global.ID {
			t.Fatalf("expected global promotion applied to product %d, got %#v", product.ID, promotions)
		}
	}
}

func TestLoadApplicablePromotionsTargetedAppliesOnlyToLinkedProduct(t *testing.T) {
	tx := openIntegrationTx(t)
	if err := tx.Where("applies_to_all = ?", true).Delete(&models.Promotion{}).Error; err != nil {
		t.Fatalf("failed to clear existing global promotions: %v", err)
	}

	category := seedCategory(t, tx)
	targetedProduct := seedProduct(t, tx, category.ID)
	untouchedProduct := seedProduct(t, tx, category.ID)

	targeted := &models.Promotion{
		Name:         fmt.Sprintf("Targeted-%d", time.Now().UnixNano()),
		Type:         models.PromotionTypeFixed,
		Value:        5,
		IsActive:     true,
		AppliesToAll: false,
	}
	if err := tx.Create(targeted).Error; err != nil {
		t.Fatalf("failed to create targeted promotion: %v", err)
	}
	if err := tx.Model(targeted).Association("Products").Append(targetedProduct); err != nil {
		t.Fatalf("failed to link promotion to product: %v", err)
	}

	byProduct, err := loadApplicablePromotions(context.Background(), tx, []*models.Product{targetedProduct, untouchedProduct})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	targetedPromotions := byProduct[targetedProduct.ID]
	if len(targetedPromotions) != 1 || targetedPromotions[0].ID != targeted.ID {
		t.Fatalf("expected targeted promotion linked to product, got %#v", targetedPromotions)
	}

	if len(byProduct[untouchedProduct.ID]) != 0 {
		t.Fatalf("expected no promotions for untouched product, got %#v", byProduct[untouchedProduct.ID])
	}
}

func TestLoadApplicablePromotionsExcludesInactivePromotions(t *testing.T) {
	tx := openIntegrationTx(t)
	if err := tx.Where("applies_to_all = ?", true).Delete(&models.Promotion{}).Error; err != nil {
		t.Fatalf("failed to clear existing global promotions: %v", err)
	}

	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	inactiveGlobal := &models.Promotion{
		Name:         fmt.Sprintf("InactiveGlobal-%d", time.Now().UnixNano()),
		Type:         models.PromotionTypePercentage,
		Value:        50,
		IsActive:     true,
		AppliesToAll: true,
	}
	if err := tx.Create(inactiveGlobal).Error; err != nil {
		t.Fatalf("failed to create inactive global promotion: %v", err)
	}
	// gorm skips zero-value fields for columns with a `default` tag on create,
	// so IsActive:false must be applied via a follow-up update.
	if err := tx.Model(inactiveGlobal).Update("is_active", false).Error; err != nil {
		t.Fatalf("failed to deactivate global promotion: %v", err)
	}

	inactiveTargeted := &models.Promotion{
		Name:         fmt.Sprintf("InactiveTargeted-%d", time.Now().UnixNano()),
		Type:         models.PromotionTypeFixed,
		Value:        5,
		IsActive:     true,
		AppliesToAll: false,
	}
	if err := tx.Create(inactiveTargeted).Error; err != nil {
		t.Fatalf("failed to create inactive targeted promotion: %v", err)
	}
	if err := tx.Model(inactiveTargeted).Update("is_active", false).Error; err != nil {
		t.Fatalf("failed to deactivate targeted promotion: %v", err)
	}
	if err := tx.Model(inactiveTargeted).Association("Products").Append(product); err != nil {
		t.Fatalf("failed to link inactive promotion to product: %v", err)
	}

	byProduct, err := loadApplicablePromotions(context.Background(), tx, []*models.Product{product})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(byProduct[product.ID]) != 0 {
		t.Fatalf("expected inactive promotions to be excluded, got %#v", byProduct[product.ID])
	}
}

func TestLoadApplicablePromotionsCombinesGlobalAndTargeted(t *testing.T) {
	tx := openIntegrationTx(t)
	if err := tx.Where("applies_to_all = ?", true).Delete(&models.Promotion{}).Error; err != nil {
		t.Fatalf("failed to clear existing global promotions: %v", err)
	}

	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	global := &models.Promotion{
		Name:         fmt.Sprintf("Global-%d", time.Now().UnixNano()),
		Type:         models.PromotionTypePercentage,
		Value:        10,
		IsActive:     true,
		AppliesToAll: true,
	}
	if err := tx.Create(global).Error; err != nil {
		t.Fatalf("failed to create global promotion: %v", err)
	}

	targeted := &models.Promotion{
		Name:         fmt.Sprintf("Targeted-%d", time.Now().UnixNano()),
		Type:         models.PromotionTypeFixed,
		Value:        3,
		IsActive:     true,
		AppliesToAll: false,
	}
	if err := tx.Create(targeted).Error; err != nil {
		t.Fatalf("failed to create targeted promotion: %v", err)
	}
	if err := tx.Model(targeted).Association("Products").Append(product); err != nil {
		t.Fatalf("failed to link promotion to product: %v", err)
	}

	byProduct, err := loadApplicablePromotions(context.Background(), tx, []*models.Product{product})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(byProduct[product.ID]) != 2 {
		t.Fatalf("expected both global and targeted promotions applied, got %#v", byProduct[product.ID])
	}
}

func TestApplyCurrentPricingCombinesWithLegacyPerProductPromotion(t *testing.T) {
	tx := openIntegrationTx(t)
	if err := tx.Where("applies_to_all = ?", true).Delete(&models.Promotion{}).Error; err != nil {
		t.Fatalf("failed to clear existing global promotions: %v", err)
	}

	category := seedCategory(t, tx)
	seller := seedUser(t, tx, "ROLE_USER")

	product := &models.Product{
		Name:            fmt.Sprintf("Legacy-Promo-%d", time.Now().UnixNano()),
		Description:     "Test",
		Image:           "image.png",
		Price:           100,
		Stock:           5,
		IsActive:        true,
		SellerID:        &seller.ID,
		CategoryID:      category.ID,
		PromotionType:   models.PromotionTypeFixed,
		PromotionValue:  10,
		PromotionActive: true,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	global := &models.Promotion{
		Name:         fmt.Sprintf("Global-%d", time.Now().UnixNano()),
		Type:         models.PromotionTypePercentage,
		Value:        50,
		IsActive:     true,
		AppliesToAll: true,
	}
	if err := tx.Create(global).Error; err != nil {
		t.Fatalf("failed to create global promotion: %v", err)
	}

	if err := applyCurrentPricing(context.Background(), tx, []*models.Product{product}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	// The global 50% promotion (discount 50) should win over the legacy fixed
	// per-product promotion (discount 10), since chooseBestPromotion picks the
	// highest discount amount.
	if product.EffectivePrice != 50 {
		t.Fatalf("expected effective price 50, got %f", product.EffectivePrice)
	}
	if product.AppliedPromotion == nil || product.AppliedPromotion.ID != global.ID {
		t.Fatalf("expected global promotion to win, got %#v", product.AppliedPromotion)
	}
}

func TestApplyCurrentPricingNoProductsIsNoop(t *testing.T) {
	tx := openIntegrationTx(t)
	if err := applyCurrentPricing(context.Background(), tx, nil); err != nil {
		t.Fatalf("expected no-op success, got %v", err)
	}
}

func TestApplyCurrentPricingSkipsNilProducts(t *testing.T) {
	tx := openIntegrationTx(t)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	if err := applyCurrentPricing(context.Background(), tx, []*models.Product{nil, product}); err != nil {
		t.Fatalf("expected success skipping nil product, got %v", err)
	}
}
