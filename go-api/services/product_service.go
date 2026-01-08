package services

import (
	"poc-gin/models"

	"gorm.io/gorm"
)

type ProductService struct {
	DB *gorm.DB
}

func NewProductService() *ProductService {
	return &ProductService{}
}

func (s *ProductService) GetAllProducts() ([]models.Product, error) {
	var products []models.Product
	var result = s.DB.Find(&products)
	return products, result.Error
}

func (s *ProductService) GetProduct(id string) (models.Product, error) {
	var product models.Product

	result := s.DB.First(&product, id)

	return product, result.Error
}

func (s *ProductService) CreateProduct(product models.Product) (models.Product, error) {
	result := s.DB.Create(&product)
	return product, result.Error
}

func (s *ProductService) DeleteProduct(id string) error {
	var product models.Product
	result := s.DB.Delete(&product, id)

	return result.Error
}

func (s *ProductService) UpdateProduct(product models.Product, input models.Product) (models.Product, error) {
	result := s.DB.Model(&product).Updates(input)
	return product, result.Error
}
