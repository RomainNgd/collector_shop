package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"sort"

	"gorm.io/gorm"
)

var (
	ErrInvalidPromotionType      = errors.New("invalid promotion type")
	ErrInvalidPromotionValue     = errors.New("invalid promotion value")
	ErrPromotionProductsEmpty    = errors.New("promotion must target at least one product")
	ErrPromotionProductsNotFound = errors.New("promotion references unknown products")
)

type PromotionInput struct {
	Name         string
	Description  string
	Type         string
	Value        float64
	IsActive     bool
	AppliesToAll bool
	ProductIDs   []uint
}

type PromotionServiceInterface interface {
	GetAllPromotions(ctx context.Context) ([]*models.Promotion, error)
	GetPromotionByID(ctx context.Context, id uint) (*models.Promotion, error)
	CreatePromotion(ctx context.Context, input PromotionInput) (*models.Promotion, error)
	UpdatePromotion(ctx context.Context, id uint, input PromotionInput) (*models.Promotion, error)
	DeletePromotion(ctx context.Context, id uint) error
}

type PromotionService struct {
	db *gorm.DB
}

func NewPromotionService(db *gorm.DB) *PromotionService {
	return &PromotionService{db: db}
}

func (s *PromotionService) GetAllPromotions(ctx context.Context) ([]*models.Promotion, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var promotions []*models.Promotion
	result := s.db.WithContext(ctx).
		Preload("Products", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("products.id").Order("products.id ASC")
		}).
		Order("id DESC").
		Find(&promotions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch promotions: %w", result.Error)
	}

	for _, promotion := range promotions {
		populatePromotionSelection(promotion)
	}

	return promotions, nil
}

func (s *PromotionService) GetPromotionByID(ctx context.Context, id uint) (*models.Promotion, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	promotion, err := s.loadPromotion(ctx, s.db, id)
	if err != nil {
		return nil, err
	}

	return promotion, nil
}

func (s *PromotionService) CreatePromotion(ctx context.Context, input PromotionInput) (*models.Promotion, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	if err := validatePromotionInput(input); err != nil {
		return nil, err
	}

	productIDs := normalizePromotionProductIDs(input.ProductIDs)

	var created *models.Promotion
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		products, err := s.loadPromotionProducts(ctx, tx, input.AppliesToAll, productIDs)
		if err != nil {
			return err
		}

		promotion := &models.Promotion{
			Name:         input.Name,
			Description:  input.Description,
			Type:         input.Type,
			Value:        roundCurrency(input.Value),
			IsActive:     input.IsActive,
			AppliesToAll: input.AppliesToAll,
		}

		if err := tx.WithContext(ctx).Create(promotion).Error; err != nil {
			return fmt.Errorf("failed to create promotion: %w", err)
		}

		if !input.AppliesToAll {
			if err := tx.WithContext(ctx).Model(promotion).Association("Products").Replace(products); err != nil {
				return fmt.Errorf("failed to link promotion products: %w", err)
			}
		}

		created, err = s.loadPromotion(ctx, tx, promotion.ID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *PromotionService) UpdatePromotion(ctx context.Context, id uint, input PromotionInput) (*models.Promotion, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	if err := validatePromotionInput(input); err != nil {
		return nil, err
	}

	productIDs := normalizePromotionProductIDs(input.ProductIDs)

	var updated *models.Promotion
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var transactionErr error
		updated, transactionErr = s.updatePromotion(ctx, tx, id, input, productIDs)
		return transactionErr
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *PromotionService) updatePromotion(
	ctx context.Context,
	tx *gorm.DB,
	id uint,
	input PromotionInput,
	productIDs []uint,
) (*models.Promotion, error) {
	var promotion models.Promotion
	if err := tx.WithContext(ctx).First(&promotion, id).Error; err != nil {
		return nil, err
	}

	products, err := s.loadPromotionProducts(ctx, tx, input.AppliesToAll, productIDs)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"name":           input.Name,
		"description":    input.Description,
		"type":           input.Type,
		"value":          roundCurrency(input.Value),
		"is_active":      input.IsActive,
		"applies_to_all": input.AppliesToAll,
	}
	if err := tx.WithContext(ctx).Model(&promotion).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update promotion: %w", err)
	}

	if err := replacePromotionProducts(ctx, tx, &promotion, products, input.AppliesToAll); err != nil {
		return nil, err
	}

	return s.loadPromotion(ctx, tx, id)
}

func replacePromotionProducts(ctx context.Context, tx *gorm.DB, promotion *models.Promotion, products []models.Product, appliesToAll bool) error {
	association := tx.WithContext(ctx).Model(promotion).Association("Products")
	if appliesToAll {
		if err := association.Clear(); err != nil {
			return fmt.Errorf("failed to clear promotion products: %w", err)
		}
		return nil
	}

	if err := association.Replace(products); err != nil {
		return fmt.Errorf("failed to update promotion products: %w", err)
	}
	return nil
}

func (s *PromotionService) DeletePromotion(ctx context.Context, id uint) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	result := s.db.WithContext(ctx).Delete(&models.Promotion{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete promotion: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func validatePromotionInput(input PromotionInput) error {
	switch input.Type {
	case models.PromotionTypePercentage:
		if input.Value <= 0 || input.Value > 100 {
			return ErrInvalidPromotionValue
		}
	case models.PromotionTypeFixed:
		if input.Value <= 0 {
			return ErrInvalidPromotionValue
		}
	default:
		return ErrInvalidPromotionType
	}

	if !input.AppliesToAll && len(normalizePromotionProductIDs(input.ProductIDs)) == 0 {
		return ErrPromotionProductsEmpty
	}

	return nil
}

func normalizePromotionProductIDs(productIDs []uint) []uint {
	if len(productIDs) == 0 {
		return nil
	}

	deduped := make(map[uint]struct{}, len(productIDs))
	normalized := make([]uint, 0, len(productIDs))

	for _, productID := range productIDs {
		if productID == 0 {
			continue
		}
		if _, exists := deduped[productID]; exists {
			continue
		}

		deduped[productID] = struct{}{}
		normalized = append(normalized, productID)
	}

	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i] < normalized[j]
	})

	return normalized
}

func populatePromotionSelection(promotion *models.Promotion) {
	productIDs := make([]uint, 0, len(promotion.Products))
	for _, product := range promotion.Products {
		productIDs = append(productIDs, product.ID)
	}

	sort.Slice(productIDs, func(i, j int) bool {
		return productIDs[i] < productIDs[j]
	})

	promotion.ProductIDs = productIDs
	promotion.ProductCount = len(productIDs)
}

func (s *PromotionService) loadPromotionProducts(
	ctx context.Context,
	db *gorm.DB,
	appliesToAll bool,
	productIDs []uint,
) ([]models.Product, error) {
	if appliesToAll {
		return nil, nil
	}

	var products []models.Product
	if err := db.WithContext(ctx).
		Where("id IN ?", productIDs).
		Order("id ASC").
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch promotion products: %w", err)
	}

	if len(products) != len(productIDs) {
		return nil, ErrPromotionProductsNotFound
	}

	return products, nil
}

func (s *PromotionService) loadPromotion(ctx context.Context, db *gorm.DB, id uint) (*models.Promotion, error) {
	var promotion models.Promotion
	result := db.WithContext(ctx).
		Preload("Products", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("products.id").Order("products.id ASC")
		}).
		First(&promotion, id)
	if result.Error != nil {
		return nil, result.Error
	}

	populatePromotionSelection(&promotion)
	return &promotion, nil
}
