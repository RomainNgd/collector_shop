package services

import (
	"context"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"

	"gorm.io/gorm"
)

type ProductServiceInterface interface {
	GetAllProducts() ([]*models.Product, error)
	GetProductByID(id uint) (*models.Product, error)
	CreateProduct(product *models.Product) error
	UpdateProduct(id uint, updates map[string]interface{}) (*models.Product, error)
	DeleteProduct(id uint) error
}

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

func (s *ProductService) GetAllProducts() ([]*models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	var products []*models.Product
	result := s.db.WithContext(ctx).Preload("Category").Find(&products)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", result.Error)
	}

	return products, nil
}

func (s *ProductService) GetProductByID(id uint) (*models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	var product models.Product
	result := s.db.WithContext(ctx).Preload("Category").First(&product, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &product, nil
}

func (s *ProductService) CreateProduct(product *models.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	result := s.db.WithContext(ctx).Create(product)
	if result.Error != nil {
		return fmt.Errorf("failed to create product: %w", result.Error)
	}

	if err := s.db.WithContext(ctx).Preload("Category").First(product, product.ID).Error; err != nil {
		return fmt.Errorf("failed to reload product: %w", err)
	}

	return nil
}

func (s *ProductService) UpdateProduct(id uint, updates map[string]interface{}) (*models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	var product models.Product
	if err := s.db.WithContext(ctx).First(&product, id).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&product).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	if err := s.db.WithContext(ctx).Preload("Category").First(&product, id).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *ProductService) DeleteProduct(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
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
