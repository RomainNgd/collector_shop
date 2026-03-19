package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"

	"gorm.io/gorm"
)

var ErrCategoryInUse = errors.New("category is used by products")

type CategoryServiceInterface interface {
	GetAllCategories() ([]*models.Category, error)
	GetCategoryByID(id uint) (*models.Category, error)
	CreateCategory(category *models.Category) error
	UpdateCategory(id uint, updates map[string]interface{}) (*models.Category, error)
	DeleteCategory(id uint) error
}

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db}
}

func (s *CategoryService) GetAllCategories() ([]*models.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	var categories []*models.Category
	result := s.db.WithContext(ctx).Find(&categories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", result.Error)
	}

	return categories, nil
}

func (s *CategoryService) GetCategoryByID(id uint) (*models.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	var category models.Category
	result := s.db.WithContext(ctx).First(&category, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &category, nil
}

func (s *CategoryService) CreateCategory(category *models.Category) error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	result := s.db.WithContext(ctx).Create(category)
	if result.Error != nil {
		return fmt.Errorf("failed to create category: %w", result.Error)
	}

	return nil
}

func (s *CategoryService) UpdateCategory(id uint, updates map[string]interface{}) (*models.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	var category models.Category
	if err := s.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&category).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	if err := s.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

func (s *CategoryService) DeleteCategory(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DBTimeout)
	defer cancel()

	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Product{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}
	if count > 0 {
		return ErrCategoryInUse
	}

	result := s.db.WithContext(ctx).Delete(&models.Category{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete category: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
