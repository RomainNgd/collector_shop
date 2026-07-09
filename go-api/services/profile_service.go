package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"

	"gorm.io/gorm"
)

type ProfileStats struct {
	Email          string `json:"email"`
	ProductsBought int64  `json:"products_bought"`
	ListingsPosted int64  `json:"listings_posted"`
	ProductsSold   int64  `json:"products_sold"`
}

type ProfileServiceInterface interface {
	GetProfileStats(ctx context.Context, userID uint) (*ProfileStats, error)
}

type ProfileService struct {
	db *gorm.DB
}

func NewProfileService(db *gorm.DB) *ProfileService {
	return &ProfileService{db: db}
}

func (s *ProfileService) GetProfileStats(ctx context.Context, userID uint) (*ProfileStats, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	productsBought, err := s.sumPaidOrderItemQuantities(ctx, "orders.user_id", userID)
	if err != nil {
		return nil, err
	}

	productsSold, err := s.sumPaidOrderItemQuantities(ctx, "order_items.seller_id", userID)
	if err != nil {
		return nil, err
	}

	var listingsPosted int64
	if err := s.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("seller_id = ?", userID).
		Count(&listingsPosted).Error; err != nil {
		return nil, fmt.Errorf("failed to count seller products: %w", err)
	}

	return &ProfileStats{
		Email:          user.Email,
		ProductsBought: productsBought,
		ListingsPosted: listingsPosted,
		ProductsSold:   productsSold,
	}, nil
}

func (s *ProfileService) sumPaidOrderItemQuantities(ctx context.Context, userColumn string, userID uint) (int64, error) {
	var total int64
	if err := s.db.WithContext(ctx).
		Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id AND orders.deleted_at IS NULL").
		Where("orders.payment_status = ?", models.OrderPaymentStatusPaid).
		Where(userColumn+" = ?", userID).
		Select("COALESCE(SUM(order_items.quantity), 0)").
		Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to sum order item quantities: %w", err)
	}

	return total, nil
}
