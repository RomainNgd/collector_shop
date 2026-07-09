package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"testing"
	"time"

	"gorm.io/gorm"
)

func seedSellerProduct(t *testing.T, tx *gorm.DB, categoryID, sellerID uint) *models.Product {
	t.Helper()

	product := &models.Product{
		Name:        fmt.Sprintf("Listing-%d", time.Now().UnixNano()),
		Description: "Seller listing",
		Image:       "listing.png",
		Price:       25,
		Stock:       5,
		IsActive:    true,
		SellerID:    &sellerID,
		CategoryID:  categoryID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to seed seller product: %v", err)
	}
	return product
}

func seedOrderWithItems(t *testing.T, tx *gorm.DB, buyerID, sellerID uint, paymentStatus string, quantities ...int) *models.Order {
	t.Helper()

	order := &models.Order{
		UserID:        buyerID,
		Status:        models.OrderStatusPreparation,
		Currency:      "EUR",
		PaymentStatus: paymentStatus,
	}
	for index, quantity := range quantities {
		order.Items = append(order.Items, models.OrderItem{
			ProductID:          uint(index + 1),
			SellerID:           sellerID,
			ProductName:        "Snapshot item",
			ProductDescription: "Snapshot description",
			Quantity:           quantity,
		})
		order.ItemCount += quantity
	}
	if err := tx.Create(order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}
	return order
}

func TestProfileServiceGetProfileStats(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProfileService(tx)
	buyer := seedUser(t, tx, constants.RoleUser)
	seller := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)

	seedSellerProduct(t, tx, category.ID, seller.ID)
	seedSellerProduct(t, tx, category.ID, seller.ID)

	seedOrderWithItems(t, tx, buyer.ID, seller.ID, models.OrderPaymentStatusPaid, 2, 1)
	seedOrderWithItems(t, tx, buyer.ID, seller.ID, models.OrderPaymentStatusPending, 5)

	buyerStats, err := service.GetProfileStats(context.Background(), buyer.ID)
	if err != nil {
		t.Fatalf("expected buyer stats, got error %v", err)
	}
	expectedBuyer := ProfileStats{Email: buyer.Email, ProductsBought: 3, ListingsPosted: 0, ProductsSold: 0}
	if *buyerStats != expectedBuyer {
		t.Fatalf("unexpected buyer stats: %#v", buyerStats)
	}

	sellerStats, err := service.GetProfileStats(context.Background(), seller.ID)
	if err != nil {
		t.Fatalf("expected seller stats, got error %v", err)
	}
	expectedSeller := ProfileStats{Email: seller.Email, ProductsBought: 0, ListingsPosted: 2, ProductsSold: 3}
	if *sellerStats != expectedSeller {
		t.Fatalf("unexpected seller stats: %#v", sellerStats)
	}
}

func TestProfileServiceGetProfileStatsPropagatesUserQueryError(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProfileService(tx)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetProfileStats(ctx, 1)
	if err == nil || errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected wrapped query error, got %v", err)
	}
}

func TestProfileServiceGetProfileStatsPropagatesStatsQueryErrors(t *testing.T) {
	for _, table := range []string{"order_items", "products"} {
		t.Run(table, func(t *testing.T) {
			tx := openIntegrationTx(t)
			service := NewProfileService(tx)
			user := seedUser(t, tx, constants.RoleUser)

			if err := tx.Exec("ALTER TABLE " + table + " RENAME TO " + table + "_hidden").Error; err != nil {
				t.Fatalf("failed to hide table %s: %v", table, err)
			}

			if _, err := service.GetProfileStats(context.Background(), user.ID); err == nil {
				t.Fatalf("expected error when %s table is unavailable", table)
			}
		})
	}
}

func TestProfileServiceGetProfileStatsUserNotFound(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewProfileService(tx)

	_, err := service.GetProfileStats(context.Background(), 999999999)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
