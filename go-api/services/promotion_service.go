package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"sort"

	"gorm.io/gorm"
)

var (
	ErrInvalidPromotionType      = errors.New("invalid promotion type")
	ErrInvalidPromotionValue     = errors.New("invalid promotion value")
	ErrPromotionProductsEmpty    = errors.New("promotion must target at least one product")
	ErrPromotionProductsNotFound = errors.New("promotion references unknown products")
	ErrPromotionProductsNotOwned = errors.New("promotion references products you do not own")
	ErrPromotionAppliesAllDenied = errors.New("only admins can create promotions that apply to all products")
	ErrPromotionAccessDenied     = errors.New("promotion access denied")
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
	GetAllPromotions(ctx context.Context, actorID uint, actorRole string) ([]*models.Promotion, error)
	GetPromotionByID(ctx context.Context, actorID uint, actorRole string, id uint) (*models.Promotion, error)
	CreatePromotion(ctx context.Context, actorID uint, actorRole string, input PromotionInput) (*models.Promotion, error)
	UpdatePromotion(ctx context.Context, actorID uint, actorRole string, id uint, input PromotionInput) (*models.Promotion, error)
	DeletePromotion(ctx context.Context, actorID uint, actorRole string, id uint) error
}

type PromotionService struct {
	db *gorm.DB
}

func NewPromotionService(db *gorm.DB) *PromotionService {
	return &PromotionService{db: db}
}

func (s *PromotionService) GetAllPromotions(ctx context.Context, actorID uint, actorRole string) ([]*models.Promotion, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	query := s.db.WithContext(ctx).
		Preload("Products", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("products.id").Order("products.id ASC")
		}).
		Order("id DESC")
	if actorRole != constants.RoleAdmin {
		query = query.Where("seller_id = ?", actorID)
	}

	var promotions []*models.Promotion
	if err := query.Find(&promotions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch promotions: %w", err)
	}

	for _, promotion := range promotions {
		populatePromotionSelection(promotion)
	}

	return promotions, nil
}

func (s *PromotionService) GetPromotionByID(ctx context.Context, actorID uint, actorRole string, id uint) (*models.Promotion, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	promotion, err := s.loadPromotion(ctx, s.db, id)
	if err != nil {
		return nil, err
	}

	if !canManagePromotion(actorID, actorRole, promotion.SellerID) {
		return nil, ErrPromotionAccessDenied
	}

	return promotion, nil
}

func (s *PromotionService) CreatePromotion(ctx context.Context, actorID uint, actorRole string, input PromotionInput) (*models.Promotion, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	isAdmin := actorRole == constants.RoleAdmin
	if !isAdmin && input.AppliesToAll {
		return nil, ErrPromotionAppliesAllDenied
	}

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
		if !isAdmin {
			if err := ensureProductsOwnedBySeller(products, actorID); err != nil {
				return err
			}
		}

		promotion := &models.Promotion{
			Name:         input.Name,
			Description:  input.Description,
			Type:         input.Type,
			Value:        roundCurrency(input.Value),
			IsActive:     input.IsActive,
			AppliesToAll: input.AppliesToAll,
		}
		if !isAdmin {
			promotion.SellerID = &actorID
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

func (s *PromotionService) UpdatePromotion(ctx context.Context, actorID uint, actorRole string, id uint, input PromotionInput) (*models.Promotion, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	isAdmin := actorRole == constants.RoleAdmin
	if !isAdmin && input.AppliesToAll {
		return nil, ErrPromotionAppliesAllDenied
	}

	if err := validatePromotionInput(input); err != nil {
		return nil, err
	}

	productIDs := normalizePromotionProductIDs(input.ProductIDs)

	var updated *models.Promotion
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var transactionErr error
		updated, transactionErr = s.updatePromotion(ctx, tx, actorID, actorRole, id, input, productIDs)
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
	actorID uint,
	actorRole string,
	id uint,
	input PromotionInput,
	productIDs []uint,
) (*models.Promotion, error) {
	var promotion models.Promotion
	if err := tx.WithContext(ctx).First(&promotion, id).Error; err != nil {
		return nil, err
	}
	if !canManagePromotion(actorID, actorRole, promotion.SellerID) {
		return nil, ErrPromotionAccessDenied
	}

	products, err := s.loadPromotionProducts(ctx, tx, input.AppliesToAll, productIDs)
	if err != nil {
		return nil, err
	}
	if actorRole != constants.RoleAdmin {
		if err := ensureProductsOwnedBySeller(products, actorID); err != nil {
			return nil, err
		}
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

func canManagePromotion(actorID uint, actorRole string, sellerID *uint) bool {
	return actorRole == constants.RoleAdmin || (actorID != 0 && sellerID != nil && actorID == *sellerID)
}

func ensureProductsOwnedBySeller(products []models.Product, sellerID uint) error {
	for _, product := range products {
		if product.SellerID == nil || *product.SellerID != sellerID {
			return ErrPromotionProductsNotOwned
		}
	}
	return nil
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

func (s *PromotionService) DeletePromotion(ctx context.Context, actorID uint, actorRole string, id uint) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var promotion models.Promotion
	if err := s.db.WithContext(ctx).First(&promotion, id).Error; err != nil {
		return err
	}
	if !canManagePromotion(actorID, actorRole, promotion.SellerID) {
		return ErrPromotionAccessDenied
	}

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
