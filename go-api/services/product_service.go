package services

import (
	"context"
	"fmt"
	"poc-gin/models"

	"gorm.io/gorm"
)

type ProductServiceInterface interface {
	GetAllProducts(ctx context.Context) ([]*models.Product, error)
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}) (*models.Product, error)
	DeleteProduct(ctx context.Context, id uint) error
}

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

func (s *ProductService) GetAllProducts(ctx context.Context) ([]*models.Product, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var products []*models.Product
	result := s.db.WithContext(ctx).Preload("Category").Preload("Promotions").Find(&products)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", result.Error)
	}

	if err := s.applyPricing(ctx, products); err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var product models.Product
	result := s.db.WithContext(ctx).Preload("Category").Preload("Promotions").First(&product, id)
	if result.Error != nil {
		return nil, result.Error
	}

	if err := s.applyPricing(ctx, []*models.Product{&product}); err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.Product) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	result := s.db.WithContext(ctx).Create(product)
	if result.Error != nil {
		return fmt.Errorf("failed to create product: %w", result.Error)
	}

	if err := s.db.WithContext(ctx).Preload("Category").Preload("Promotions").First(product, product.ID).Error; err != nil {
		return fmt.Errorf("failed to reload product: %w", err)
	}

	if err := s.applyPricing(ctx, []*models.Product{product}); err != nil {
		return err
	}

	return nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}) (*models.Product, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var product models.Product
	if err := s.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&product).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	if err := s.db.WithContext(ctx).Preload("Category").Preload("Promotions").First(&product, id).Error; err != nil {
		return nil, err
	}

	if err := s.applyPricing(ctx, []*models.Product{&product}); err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uint) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	result := s.db.WithContext(ctx).Delete(&models.Product{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete product: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (s *ProductService) applyPricing(ctx context.Context, products []*models.Product) error {
	return applyCurrentPricing(ctx, s.db, products)
}
