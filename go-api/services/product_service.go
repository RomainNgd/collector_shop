package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"

	"gorm.io/gorm"
)

var (
	ErrProductAccessDenied          = errors.New("product access denied")
	ErrProductSellerRequired        = errors.New("product seller is required")
	ErrProductInvalidStock          = errors.New("product stock must be positive")
	ErrProductInvalidPromotionType  = errors.New("product promotion type is invalid")
	ErrProductInvalidPromotionValue = errors.New("product promotion value is invalid")
)

type ProductServiceInterface interface {
	GetAllProducts(ctx context.Context, excludeSellerID *uint, page Pagination) ([]*models.Product, int64, error)
	GetProductsForSeller(ctx context.Context, sellerID uint) ([]*models.Product, error)
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
	GetProductForManagement(ctx context.Context, actorID uint, actorRole string, id uint) (*models.Product, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, actorID uint, actorRole string, id uint, updates map[string]interface{}) (*models.Product, error)
	DeleteProduct(ctx context.Context, actorID uint, actorRole string, id uint) error
}

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

func (s *ProductService) GetAllProducts(ctx context.Context, excludeSellerID *uint, page Pagination) ([]*models.Product, int64, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	limit, offset := page.normalized()

	baseQuery := s.db.WithContext(ctx).Model(&models.Product{}).
		Where("is_active = ? AND stock > ?", true, 0)
	if excludeSellerID != nil {
		baseQuery = baseQuery.Where("seller_id IS NULL OR seller_id != ?", *excludeSellerID)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	var products []*models.Product
	if err := baseQuery.
		Preload("Category").
		Preload("Seller").
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch products: %w", err)
	}

	prepared, err := s.prepareProducts(ctx, products)
	if err != nil {
		return nil, 0, err
	}

	return prepared, total, nil
}

func (s *ProductService) GetProductsForSeller(ctx context.Context, sellerID uint) ([]*models.Product, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var products []*models.Product
	result := s.db.WithContext(ctx).
		Preload("Category").
		Preload("Seller").
		Where("seller_id = ?", sellerID).
		Order("created_at DESC").
		Find(&products)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch seller products: %w", result.Error)
	}

	return s.prepareProducts(ctx, products)
}

func (s *ProductService) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var product models.Product
	result := s.db.WithContext(ctx).
		Preload("Category").
		Preload("Seller").
		Where("is_active = ? AND stock > ?", true, 0).
		First(&product, id)
	if result.Error != nil {
		return nil, result.Error
	}

	if err := s.prepareProduct(ctx, &product); err != nil {
		return nil, err
	}

	return &product, nil
}

// GetProductForManagement fetches a product without the public catalog filters
// (is_active, stock) and enforces that the actor owns it or is an admin. It is
// meant for seller/admin management flows such as image upload or removal.
func (s *ProductService) GetProductForManagement(ctx context.Context, actorID uint, actorRole string, id uint) (*models.Product, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var product models.Product
	if err := s.db.WithContext(ctx).Preload("Category").Preload("Seller").First(&product, id).Error; err != nil {
		return nil, err
	}
	if !canManageProduct(actorID, actorRole, product.SellerID) {
		return nil, ErrProductAccessDenied
	}

	if err := s.prepareProduct(ctx, &product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	if err := validateProductForSale(product); err != nil {
		return err
	}

	result := s.db.WithContext(ctx).Create(product)
	if result.Error != nil {
		return fmt.Errorf("failed to create product: %w", result.Error)
	}

	if err := s.db.WithContext(ctx).Preload("Category").Preload("Seller").First(product, product.ID).Error; err != nil {
		return fmt.Errorf("failed to reload product: %w", err)
	}

	if err := s.prepareProduct(ctx, product); err != nil {
		return err
	}

	return nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, actorID uint, actorRole string, id uint, updates map[string]interface{}) (*models.Product, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var product models.Product
	if err := s.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}
	if !canManageProduct(actorID, actorRole, product.SellerID) {
		return nil, ErrProductAccessDenied
	}
	if err := validateProductUpdates(&product, updates); err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&product).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	if err := s.db.WithContext(ctx).Preload("Category").Preload("Seller").First(&product, id).Error; err != nil {
		return nil, err
	}

	if err := s.prepareProduct(ctx, &product); err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, actorID uint, actorRole string, id uint) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var product models.Product
	if err := s.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return err
	}
	if !canManageProduct(actorID, actorRole, product.SellerID) {
		return ErrProductAccessDenied
	}

	if err := s.db.WithContext(ctx).Delete(&product).Error; err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (s *ProductService) prepareProducts(ctx context.Context, products []*models.Product) ([]*models.Product, error) {
	if err := applyCurrentPricing(ctx, s.db, products); err != nil {
		return nil, err
	}
	for _, product := range products {
		populateProductSellerEmail(product)
	}
	return products, nil
}

func (s *ProductService) prepareProduct(ctx context.Context, product *models.Product) error {
	if _, err := s.prepareProducts(ctx, []*models.Product{product}); err != nil {
		return err
	}
	return nil
}

func canManageProduct(actorID uint, actorRole string, sellerID *uint) bool {
	return actorRole == constants.RoleAdmin || (actorID != 0 && sellerID != nil && actorID == *sellerID)
}

func populateProductSellerEmail(product *models.Product) {
	if product == nil {
		return
	}
	product.SellerEmail = product.Seller.Email
}

func validateProductForSale(product *models.Product) error {
	if product.SellerID == nil || *product.SellerID == 0 {
		return ErrProductSellerRequired
	}
	if product.Stock <= 0 {
		return ErrProductInvalidStock
	}
	return validateProductPromotion(product.PromotionActive, product.PromotionType, product.PromotionValue)
}

func validateProductUpdates(product *models.Product, updates map[string]interface{}) error {
	stock := product.Stock
	if rawStock, ok := updates["stock"]; ok {
		value, valid := intFromUpdate(rawStock)
		if !valid {
			return ErrProductInvalidStock
		}
		stock = value
	}
	if stock <= 0 {
		return ErrProductInvalidStock
	}

	promotionActive := product.PromotionActive
	if rawActive, ok := updates["promotion_active"]; ok {
		value, valid := boolFromUpdate(rawActive)
		if !valid {
			return ErrProductInvalidPromotionValue
		}
		promotionActive = value
	}

	promotionType := product.PromotionType
	if rawType, ok := updates["promotion_type"]; ok {
		if value, valid := rawType.(string); valid {
			promotionType = value
		}
	}

	promotionValue := product.PromotionValue
	if rawValue, ok := updates["promotion_value"]; ok {
		value, valid := floatFromUpdate(rawValue)
		if !valid {
			return ErrProductInvalidPromotionValue
		}
		promotionValue = value
	}

	return validateProductPromotion(promotionActive, promotionType, promotionValue)
}

func validateProductPromotion(active bool, promotionType string, value float64) error {
	if !active {
		return nil
	}

	switch promotionType {
	case models.PromotionTypePercentage:
		if value <= 0 || value > 100 {
			return ErrProductInvalidPromotionValue
		}
	case models.PromotionTypeFixed:
		if value <= 0 {
			return ErrProductInvalidPromotionValue
		}
	default:
		return ErrProductInvalidPromotionType
	}

	return nil
}

func intFromUpdate(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case uint:
		if v > uint(^uint(0)>>1) {
			return 0, false
		}
		return int(v), true
	default:
		return 0, false
	}
}

func floatFromUpdate(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	default:
		return 0, false
	}
}

func boolFromUpdate(value interface{}) (bool, bool) {
	v, ok := value.(bool)
	return v, ok
}
