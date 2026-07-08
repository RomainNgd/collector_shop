package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const orderCurrencyEUR = "EUR"

var (
	ErrOrderEmpty                      = errors.New("order requires at least one item")
	ErrOrderInvalidQuantity            = errors.New("order item quantity must be positive")
	ErrOrderProductNotFound            = errors.New("order product not found")
	ErrOrderInvalidStatus              = errors.New("order status is invalid")
	ErrOrderStatusTransitionNotAllowed = errors.New("order status transition is not allowed")
	ErrOrderDeletionNotAllowed         = errors.New("order deletion is not allowed")
	ErrOrderInsufficientStock          = errors.New("order product stock is insufficient")
	ErrOrderOwnProduct                 = errors.New("cannot order your own product")
)

type OrderItemInput struct {
	ProductID uint
	Quantity  int
}

type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, userID uint, items []OrderItemInput) (*models.Order, error)
	GetOrdersForUser(ctx context.Context, userID uint) ([]*models.Order, error)
	GetOrderByID(ctx context.Context, actorID, orderID uint, actorRole string) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, actorID, orderID uint, actorRole, status string) (*models.Order, error)
	DeleteOrder(ctx context.Context, actorID, orderID uint, actorRole string) error
}

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID uint, items []OrderItemInput) (*models.Order, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	normalizedItems, err := normalizeOrderItems(items)
	if err != nil {
		return nil, err
	}

	var createdOrder models.Order
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		productsByID, err := loadPricedOrderProducts(ctx, tx, normalizedItems)
		if err != nil {
			return err
		}

		order := models.Order{
			UserID:        userID,
			Status:        models.OrderStatusAwaitingPayment,
			Currency:      orderCurrencyEUR,
			PaymentStatus: models.OrderPaymentStatusPending,
		}

		for _, item := range normalizedItems {
			product, ok := productsByID[item.ProductID]
			if !ok {
				return ErrOrderProductNotFound
			}
			if product.SellerID != nil && *product.SellerID == userID {
				return ErrOrderOwnProduct
			}
			if product.Stock < item.Quantity {
				return ErrOrderInsufficientStock
			}

			addOrderItem(&order, createOrderItem(product, item))
		}

		if err := tx.Create(&order).Error; err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, item := range normalizedItems {
			result := tx.Model(&models.Product{}).
				Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity))
			if result.Error != nil {
				return fmt.Errorf("failed to decrement product stock: %w", result.Error)
			}
			if result.RowsAffected == 0 {
				return ErrOrderInsufficientStock
			}
		}

		reloadedOrder, err := preloadOrder(tx.WithContext(ctx), order.ID)
		if err != nil {
			return err
		}

		createdOrder = *reloadedOrder
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &createdOrder, nil
}

func loadPricedOrderProducts(ctx context.Context, tx *gorm.DB, items []OrderItemInput) (map[uint]*models.Product, error) {
	productIDs := make([]uint, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}

	var products []models.Product
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Category").
		Preload("Seller").
		Where("id IN ? AND is_active = ?", productIDs, true).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch order products: %w", err)
	}
	if len(products) != len(productIDs) {
		return nil, ErrOrderProductNotFound
	}

	productPointers := make([]*models.Product, 0, len(products))
	productsByID := make(map[uint]*models.Product, len(products))
	for i := range products {
		product := &products[i]
		productPointers = append(productPointers, product)
		productsByID[product.ID] = product
	}
	if err := applyCurrentPricing(ctx, tx, productPointers); err != nil {
		return nil, err
	}

	return productsByID, nil
}

func createOrderItem(product *models.Product, input OrderItemInput) models.OrderItem {
	basePrice := roundCurrency(product.Price)
	unitPrice := roundCurrency(product.EffectivePrice)
	unitDiscount := max(0, roundCurrency(basePrice-unitPrice))
	lineBaseTotal := roundCurrency(basePrice * float64(input.Quantity))
	lineTotal := roundCurrency(unitPrice * float64(input.Quantity))
	lineDiscountTotal := max(0, roundCurrency(lineBaseTotal-lineTotal))

	item := models.OrderItem{
		ProductID:          product.ID,
		SellerID:           productSellerID(product),
		SellerEmail:        product.Seller.Email,
		ProductName:        product.Name,
		ProductDescription: product.Description,
		ProductImage:       product.Image,
		CategoryName:       product.Category.Name,
		Quantity:           input.Quantity,
		UnitBasePrice:      basePrice,
		UnitPrice:          unitPrice,
		UnitDiscount:       unitDiscount,
		LineBaseTotal:      lineBaseTotal,
		LineDiscountTotal:  lineDiscountTotal,
		LineTotal:          lineTotal,
	}
	if product.AppliedPromotion != nil {
		promotionID := product.AppliedPromotion.ID
		item.PromotionID = &promotionID
		item.PromotionName = product.AppliedPromotion.Name
		item.PromotionType = product.AppliedPromotion.Type
		item.PromotionValue = product.AppliedPromotion.Value
		item.PromotionAppliesAll = product.AppliedPromotion.AppliesToAll
	}

	return item
}

func productSellerID(product *models.Product) uint {
	if product.SellerID == nil {
		return 0
	}
	return *product.SellerID
}

func addOrderItem(order *models.Order, item models.OrderItem) {
	order.Items = append(order.Items, item)
	order.ItemCount += item.Quantity
	order.Subtotal = roundCurrency(order.Subtotal + item.LineBaseTotal)
	order.Total = roundCurrency(order.Total + item.LineTotal)
	order.DiscountTotal = roundCurrency(order.DiscountTotal + item.LineDiscountTotal)
}

func (s *OrderService) GetOrdersForUser(ctx context.Context, userID uint) ([]*models.Order, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	var orders []*models.Order
	query := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		})

	if err := query.Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}

	return orders, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, actorID, orderID uint, actorRole string) (*models.Order, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	return s.findAccessibleOrder(ctx, actorID, orderID, actorRole)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, actorID, orderID uint, actorRole, status string) (*models.Order, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	if !models.IsValidOrderStatus(status) {
		return nil, ErrOrderInvalidStatus
	}

	var updatedOrder models.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := findAccessibleOrderForUpdate(tx, actorID, orderID, actorRole)
		if err != nil {
			return err
		}

		if !canUpdateOrderStatus(actorRole, order.Status, status) {
			return ErrOrderStatusTransitionNotAllowed
		}

		if order.Status != status {
			// Cancellation is terminal: items go back on sale exactly once.
			if status == models.OrderStatusCancelled {
				if err := restockOrderItems(tx, order.Items); err != nil {
					return err
				}
			}
			if err := tx.Model(&models.Order{}).Where("id = ?", order.ID).Update("status", status).Error; err != nil {
				return fmt.Errorf("failed to update order status: %w", err)
			}
		}

		reloadedOrder, err := preloadOrder(tx, order.ID)
		if err != nil {
			return err
		}

		updatedOrder = *reloadedOrder
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &updatedOrder, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, actorID, orderID uint, actorRole string) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := findAccessibleOrderForUpdate(tx, actorID, orderID, actorRole)
		if err != nil {
			return err
		}

		if actorRole != constants.RoleAdmin && order.Status != models.OrderStatusAwaitingPayment {
			return ErrOrderDeletionNotAllowed
		}

		// Unpaid orders still hold reserved stock; cancelled orders were
		// already restocked, and shipped/delivered ones are genuinely sold.
		if order.Status == models.OrderStatusAwaitingPayment {
			if err := restockOrderItems(tx, order.Items); err != nil {
				return err
			}
		}

		if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderItem{}).Error; err != nil {
			return fmt.Errorf("failed to delete order items: %w", err)
		}

		if err := tx.Delete(&models.Order{}, order.ID).Error; err != nil {
			return fmt.Errorf("failed to delete order: %w", err)
		}

		return nil
	})
}

func restockOrderItems(tx *gorm.DB, items []models.OrderItem) error {
	for _, item := range items {
		if item.Quantity <= 0 {
			continue
		}
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
			return fmt.Errorf("failed to restore product stock: %w", err)
		}
	}
	return nil
}

func (s *OrderService) findAccessibleOrder(ctx context.Context, actorID, orderID uint, actorRole string) (*models.Order, error) {
	query := s.db.WithContext(ctx).Model(&models.Order{})
	if actorRole != constants.RoleAdmin {
		query = query.Where("user_id = ?", actorID)
	}

	return preloadOrder(query, orderID)
}

// findAccessibleOrderForUpdate locks the order row for the duration of the
// surrounding transaction so status changes, deletions and payment webhooks
// cannot interleave on the same order.
func findAccessibleOrderForUpdate(tx *gorm.DB, actorID, orderID uint, actorRole string) (*models.Order, error) {
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&models.Order{})
	if actorRole != constants.RoleAdmin {
		query = query.Where("user_id = ?", actorID)
	}

	return preloadOrder(query, orderID)
}

func preloadOrder(query *gorm.DB, orderID uint) (*models.Order, error) {
	var order models.Order
	if err := query.
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		First(&order, orderID).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

func normalizeOrderItems(items []OrderItemInput) ([]OrderItemInput, error) {
	if len(items) == 0 {
		return nil, ErrOrderEmpty
	}

	quantitiesByProduct := make(map[uint]int, len(items))
	productOrder := make([]uint, 0, len(items))

	for _, item := range items {
		if item.ProductID == 0 {
			return nil, ErrOrderProductNotFound
		}
		if item.Quantity <= 0 {
			return nil, ErrOrderInvalidQuantity
		}

		if _, exists := quantitiesByProduct[item.ProductID]; !exists {
			productOrder = append(productOrder, item.ProductID)
		}
		quantitiesByProduct[item.ProductID] += item.Quantity
	}

	normalized := make([]OrderItemInput, 0, len(productOrder))
	for _, productID := range productOrder {
		normalized = append(normalized, OrderItemInput{
			ProductID: productID,
			Quantity:  quantitiesByProduct[productID],
		})
	}

	return normalized, nil
}

func canUpdateOrderStatus(actorRole, currentStatus, nextStatus string) bool {
	if !models.IsValidOrderStatus(nextStatus) {
		return false
	}

	if currentStatus == nextStatus {
		return true
	}

	if actorRole != constants.RoleAdmin {
		return false
	}

	switch currentStatus {
	case models.OrderStatusAwaitingPayment:
		return nextStatus == models.OrderStatusPreparation || nextStatus == models.OrderStatusCancelled
	case models.OrderStatusPreparation:
		return nextStatus == models.OrderStatusShipping || nextStatus == models.OrderStatusCancelled
	case models.OrderStatusShipping:
		return nextStatus == models.OrderStatusDelivered
	default:
		return false
	}
}
