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

func seedUser(t *testing.T, tx *gorm.DB, role string) *models.User {
	t.Helper()

	user := &models.User{
		Email:    fmt.Sprintf("%s-%d@example.com", role, time.Now().UnixNano()),
		Password: "hash",
		Role:     role,
	}
	if err := tx.Create(user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return user
}

func TestOrderServiceCreateOrderLocksPricing(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)

	product := &models.Product{
		Name:        fmt.Sprintf("Console-%d", time.Now().UnixNano()),
		Description: "Collector item",
		Image:       "console.png",
		Price:       100,
		CategoryID:  category.ID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	promotion := &models.Promotion{
		Name:         "Launch",
		Type:         models.PromotionTypePercentage,
		Value:        10,
		IsActive:     true,
		AppliesToAll: false,
	}
	if err := tx.Create(promotion).Error; err != nil {
		t.Fatalf("failed to create promotion: %v", err)
	}
	if err := tx.Model(promotion).Association("Products").Append(product); err != nil {
		t.Fatalf("failed to link promotion: %v", err)
	}

	order, err := service.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 2},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	if order.Status != models.OrderStatusAwaitingPayment {
		t.Fatalf("expected awaiting payment status, got %s", order.Status)
	}
	if order.ItemCount != 2 {
		t.Fatalf("expected item count 2, got %d", order.ItemCount)
	}
	if order.Subtotal != 200 {
		t.Fatalf("expected subtotal 200, got %f", order.Subtotal)
	}
	if order.DiscountTotal != 20 {
		t.Fatalf("expected discount total 20, got %f", order.DiscountTotal)
	}
	if order.Total != 180 {
		t.Fatalf("expected total 180, got %f", order.Total)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected a single order item, got %d", len(order.Items))
	}

	line := order.Items[0]
	if line.UnitBasePrice != 100 || line.UnitPrice != 90 || line.LineTotal != 180 {
		t.Fatalf("unexpected line snapshot: %#v", line)
	}
	if line.PromotionID == nil || *line.PromotionID != promotion.ID {
		t.Fatalf("expected promotion snapshot, got %#v", line.PromotionID)
	}
}

func TestOrderServiceCreateOrderRejectsMissingProduct(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)

	_, err := service.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: 999999, Quantity: 1},
	})
	if !errors.Is(err, ErrOrderProductNotFound) {
		t.Fatalf("expected ErrOrderProductNotFound, got %v", err)
	}
}

func TestOrderServiceUserCannotDirectlyValidatePayment(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := service.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	_, err = service.UpdateOrderStatus(
		context.Background(),
		user.ID,
		order.ID,
		constants.RoleUser,
		models.OrderStatusPreparation,
	)
	if !errors.Is(err, ErrOrderStatusTransitionNotAllowed) {
		t.Fatalf("expected ErrOrderStatusTransitionNotAllowed, got %v", err)
	}
}

func TestOrderServiceDeleteRules(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	admin := seedUser(t, tx, constants.RoleAdmin)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := service.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	if _, err := service.UpdateOrderStatus(
		context.Background(),
		admin.ID,
		order.ID,
		constants.RoleAdmin,
		models.OrderStatusPreparation,
	); err != nil {
		t.Fatalf("expected admin transition success, got %v", err)
	}

	err = service.DeleteOrder(context.Background(), user.ID, order.ID, constants.RoleUser)
	if !errors.Is(err, ErrOrderDeletionNotAllowed) {
		t.Fatalf("expected ErrOrderDeletionNotAllowed, got %v", err)
	}

	if err := service.DeleteOrder(context.Background(), admin.ID, order.ID, constants.RoleAdmin); err != nil {
		t.Fatalf("expected admin delete success, got %v", err)
	}

	_, err = service.GetOrderByID(context.Background(), admin.ID, order.ID, constants.RoleAdmin)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected deleted order to be missing, got %v", err)
	}
}

func TestOrderServiceListOrdersForUser(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	otherUser := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	if _, err := service.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	}); err != nil {
		t.Fatalf("expected first order creation success, got %v", err)
	}
	if _, err := service.CreateOrder(context.Background(), otherUser.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 2},
	}); err != nil {
		t.Fatalf("expected second order creation success, got %v", err)
	}

	orders, err := service.GetOrdersForUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected one order for user, got %d", len(orders))
	}
	if orders[0].UserID != user.ID {
		t.Fatalf("expected user id %d, got %d", user.ID, orders[0].UserID)
	}
}
