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
	seller := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)

	product := &models.Product{
		Name:            fmt.Sprintf("Console-%d", time.Now().UnixNano()),
		Description:     "Collector item",
		Image:           "console.png",
		Price:           100,
		Stock:           3,
		IsActive:        true,
		SellerID:        &seller.ID,
		CategoryID:      category.ID,
		PromotionType:   models.PromotionTypePercentage,
		PromotionValue:  10,
		PromotionActive: true,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to create product: %v", err)
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
	if line.PromotionID == nil || *line.PromotionID != product.ID {
		t.Fatalf("expected promotion snapshot, got %#v", line.PromotionID)
	}
	if line.SellerID != seller.ID || line.SellerEmail != seller.Email {
		t.Fatalf("expected seller snapshot, got seller=%d email=%q", line.SellerID, line.SellerEmail)
	}

	var reloadedProduct models.Product
	if err := tx.First(&reloadedProduct, product.ID).Error; err != nil {
		t.Fatalf("failed to reload product: %v", err)
	}
	if reloadedProduct.Stock != 1 {
		t.Fatalf("expected stock decremented to 1, got %d", reloadedProduct.Stock)
	}
}

func TestOrderServiceGetSellerStats(t *testing.T) {
	tx := openIntegrationTx(t)
	if err := tx.Where("applies_to_all = ?", true).Delete(&models.Promotion{}).Error; err != nil {
		t.Fatalf("failed to clear global promotions: %v", err)
	}
	service := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	seller := seedUser(t, tx, constants.RoleUser)
	otherSeller := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)

	product := &models.Product{
		Name:        fmt.Sprintf("Stats-%d", time.Now().UnixNano()),
		Description: "Collector item",
		Image:       "item.png",
		Price:       50,
		Stock:       10,
		IsActive:    true,
		SellerID:    &seller.ID,
		CategoryID:  category.ID,
	}
	if err := tx.Create(product).Error; err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	otherProduct := &models.Product{
		Name:        fmt.Sprintf("OtherStats-%d", time.Now().UnixNano()),
		Description: "Other seller item",
		Image:       "other.png",
		Price:       30,
		Stock:       10,
		IsActive:    true,
		SellerID:    &otherSeller.ID,
		CategoryID:  category.ID,
	}
	if err := tx.Create(otherProduct).Error; err != nil {
		t.Fatalf("failed to create other product: %v", err)
	}

	paidOrder, err := service.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 3},
	})
	if err != nil {
		t.Fatalf("expected paid order creation success, got %v", err)
	}
	if err := tx.Model(&models.Order{}).Where("id = ?", paidOrder.ID).Update("payment_status", models.OrderPaymentStatusPaid).Error; err != nil {
		t.Fatalf("failed to mark order paid: %v", err)
	}

	unpaidOrder, err := service.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 2},
	})
	if err != nil {
		t.Fatalf("expected unpaid order creation success, got %v", err)
	}
	_ = unpaidOrder

	if _, err := service.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: otherProduct.ID, Quantity: 5},
	}); err != nil {
		t.Fatalf("expected other seller order creation success, got %v", err)
	}
	if err := tx.Model(&models.Order{}).Where("user_id = ? AND id NOT IN ?", user.ID, []uint{paidOrder.ID, unpaidOrder.ID}).
		Update("payment_status", models.OrderPaymentStatusPaid).Error; err != nil {
		t.Fatalf("failed to mark other seller order paid: %v", err)
	}

	stats, err := service.GetSellerStats(context.Background(), seller.ID)
	if err != nil {
		t.Fatalf("expected stats success, got %v", err)
	}
	if stats.TotalRevenue != 150 {
		t.Fatalf("expected revenue 150 (only paid order counted), got %f", stats.TotalRevenue)
	}
	if stats.TotalSales != 3 {
		t.Fatalf("expected 3 units sold (only paid order counted), got %d", stats.TotalSales)
	}
	if stats.ProductCount != 1 {
		t.Fatalf("expected 1 product for seller, got %d", stats.ProductCount)
	}

	otherStats, err := service.GetSellerStats(context.Background(), otherSeller.ID)
	if err != nil {
		t.Fatalf("expected other seller stats success, got %v", err)
	}
	if otherStats.TotalRevenue != 150 || otherStats.TotalSales != 5 || otherStats.ProductCount != 1 {
		t.Fatalf("unexpected other seller stats: %#v", otherStats)
	}
}

func TestOrderServiceCreateOrderRejectsInsufficientStock(t *testing.T) {
	tx := openIntegrationTx(t)
	service := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID) // stock: 10

	_, err := service.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 999},
	})
	if !errors.Is(err, ErrOrderInsufficientStock) {
		t.Fatalf("expected ErrOrderInsufficientStock, got %v", err)
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

func TestCanUpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name          string
		actorRole     string
		currentStatus string
		nextStatus    string
		want          bool
	}{
		{"invalid next status rejected", constants.RoleAdmin, models.OrderStatusAwaitingPayment, "unknown", false},
		{"same status always allowed", constants.RoleUser, models.OrderStatusAwaitingPayment, models.OrderStatusAwaitingPayment, true},
		{"user cannot transition", constants.RoleUser, models.OrderStatusAwaitingPayment, models.OrderStatusPreparation, false},
		{"admin awaiting to preparation", constants.RoleAdmin, models.OrderStatusAwaitingPayment, models.OrderStatusPreparation, true},
		{"admin awaiting to cancelled", constants.RoleAdmin, models.OrderStatusAwaitingPayment, models.OrderStatusCancelled, true},
		{"admin awaiting to shipping rejected", constants.RoleAdmin, models.OrderStatusAwaitingPayment, models.OrderStatusShipping, false},
		{"admin preparation to shipping", constants.RoleAdmin, models.OrderStatusPreparation, models.OrderStatusShipping, true},
		{"admin preparation to cancelled", constants.RoleAdmin, models.OrderStatusPreparation, models.OrderStatusCancelled, true},
		{"admin preparation to delivered rejected", constants.RoleAdmin, models.OrderStatusPreparation, models.OrderStatusDelivered, false},
		{"admin shipping to delivered", constants.RoleAdmin, models.OrderStatusShipping, models.OrderStatusDelivered, true},
		{"admin shipping to cancelled rejected", constants.RoleAdmin, models.OrderStatusShipping, models.OrderStatusCancelled, false},
		{"admin delivered to any rejected", constants.RoleAdmin, models.OrderStatusDelivered, models.OrderStatusCancelled, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canUpdateOrderStatus(tt.actorRole, tt.currentStatus, tt.nextStatus); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestNormalizeOrderItemsMergesDuplicateProducts(t *testing.T) {
	normalized, err := normalizeOrderItems([]OrderItemInput{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 1},
		{ProductID: 1, Quantity: 3},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(normalized) != 2 {
		t.Fatalf("expected 2 normalized items, got %d", len(normalized))
	}
	if normalized[0].ProductID != 1 || normalized[0].Quantity != 5 {
		t.Fatalf("expected merged quantity 5 for product 1, got %#v", normalized[0])
	}
}

func TestNormalizeOrderItemsRejectsInvalidInput(t *testing.T) {
	if _, err := normalizeOrderItems(nil); !errors.Is(err, ErrOrderEmpty) {
		t.Fatalf("expected ErrOrderEmpty, got %v", err)
	}
	if _, err := normalizeOrderItems([]OrderItemInput{{ProductID: 0, Quantity: 1}}); !errors.Is(err, ErrOrderProductNotFound) {
		t.Fatalf("expected ErrOrderProductNotFound, got %v", err)
	}
	if _, err := normalizeOrderItems([]OrderItemInput{{ProductID: 1, Quantity: 0}}); !errors.Is(err, ErrOrderInvalidQuantity) {
		t.Fatalf("expected ErrOrderInvalidQuantity, got %v", err)
	}
}

func TestProductSellerID(t *testing.T) {
	if got := productSellerID(&models.Product{}); got != 0 {
		t.Fatalf("expected 0 for nil seller id, got %d", got)
	}
	sellerID := uint(9)
	if got := productSellerID(&models.Product{SellerID: &sellerID}); got != sellerID {
		t.Fatalf("expected %d, got %d", sellerID, got)
	}
}
