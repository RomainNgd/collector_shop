package controllers

import (
	"errors"
	"net/http"
	"poc-gin/pkg/logger"
	"poc-gin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	orderService  services.OrderServiceInterface
	orderPayments services.OrderPaymentServiceInterface
}

func NewOrderHandler(orderService services.OrderServiceInterface, orderPayments services.OrderPaymentServiceInterface) *OrderHandler {
	return &OrderHandler{
		orderService:  orderService,
		orderPayments: orderPayments,
	}
}

func (h *OrderHandler) FindOrder(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", "Invalid authentication context", nil)
		return
	}

	orders, err := h.orderService.GetOrdersForUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch orders for user %d: %v", userID, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch orders", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, orders)
}

func (h *OrderHandler) FindOneOrder(c *gin.Context) {
	ctx := c.Request.Context()

	userID, role, orderID, ok := h.readActorAndOrderID(c)
	if !ok {
		return
	}

	order, err := h.orderService.GetOrderByID(ctx, userID, orderID, role)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondError(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found", nil)
			return
		}
		logger.Error("Failed to fetch order %d for user %d: %v", orderID, userID, err)
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch order", nil)
		return
	}

	RespondSuccess(c, http.StatusOK, order)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", "Invalid authentication context", nil)
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload", err.Error())
		return
	}

	items := make([]services.OrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, services.OrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.orderService.CreateOrder(ctx, userID, items)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOrderEmpty),
			errors.Is(err, services.ErrOrderInvalidQuantity),
			errors.Is(err, services.ErrOrderProductNotFound):
			RespondError(c, http.StatusBadRequest, "ORDER_INVALID", "Invalid order payload", err.Error())
			return
		default:
			logger.Error("Failed to create order for user %d: %v", userID, err)
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create order", nil)
			return
		}
	}

	RespondSuccess(c, http.StatusCreated, order)
}

func (h *OrderHandler) CreateCheckoutSession(c *gin.Context) {
	ctx := c.Request.Context()

	userID, role, orderID, ok := h.readActorAndOrderID(c)
	if !ok {
		return
	}

	var req CreateOrderCheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload", err.Error())
		return
	}

	result, err := h.orderPayments.CreateStripeCheckoutSession(
		ctx,
		userID,
		orderID,
		role,
		req.SuccessURL,
		req.CancelURL,
	)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrStripeNotEnabled):
			RespondError(c, http.StatusServiceUnavailable, "STRIPE_NOT_CONFIGURED", "Stripe payment is not configured", nil)
			return
		case errors.Is(err, services.ErrOrderPaymentNotAvailable):
			RespondError(c, http.StatusConflict, "ORDER_PAYMENT_NOT_AVAILABLE", "Order cannot be paid in its current state", nil)
			return
		case errors.Is(err, services.ErrOrderCheckoutURLMissing):
			RespondError(c, http.StatusBadGateway, "CHECKOUT_URL_MISSING", "Stripe checkout session URL is missing", nil)
			return
		case errors.Is(err, services.ErrCheckoutReturnURLInvalid):
			RespondError(c, http.StatusBadRequest, "CHECKOUT_RETURN_URL_INVALID", "Checkout return URL is not allowed", nil)
			return
		case errors.Is(err, gorm.ErrRecordNotFound):
			RespondError(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found", nil)
			return
		default:
			logger.Error("Failed to create checkout session for order %d and user %d: %v", orderID, userID, err)
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start payment session", nil)
			return
		}
	}

	RespondSuccess(c, http.StatusCreated, gin.H{
		"session_id": result.SessionID,
		"url":        result.URL,
		"reused":     result.Reused,
	})
}

func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	ctx := c.Request.Context()

	userID, role, orderID, ok := h.readActorAndOrderID(c)
	if !ok {
		return
	}

	var req UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request payload", err.Error())
		return
	}

	order, err := h.orderService.UpdateOrderStatus(ctx, userID, orderID, role, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOrderInvalidStatus):
			RespondError(c, http.StatusBadRequest, "ORDER_STATUS_INVALID", "Invalid order status", nil)
			return
		case errors.Is(err, services.ErrOrderStatusTransitionNotAllowed):
			RespondError(c, http.StatusConflict, "ORDER_STATUS_TRANSITION_INVALID", "Order status transition not allowed", nil)
			return
		case errors.Is(err, gorm.ErrRecordNotFound):
			RespondError(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found", nil)
			return
		default:
			logger.Error("Failed to update order %d for user %d: %v", orderID, userID, err)
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update order", nil)
			return
		}
	}

	RespondSuccess(c, http.StatusOK, order)
}

func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	ctx := c.Request.Context()

	userID, role, orderID, ok := h.readActorAndOrderID(c)
	if !ok {
		return
	}

	if err := h.orderService.DeleteOrder(ctx, userID, orderID, role); err != nil {
		switch {
		case errors.Is(err, services.ErrOrderDeletionNotAllowed):
			RespondError(c, http.StatusConflict, "ORDER_DELETE_NOT_ALLOWED", "Order cannot be deleted in its current state", nil)
			return
		case errors.Is(err, gorm.ErrRecordNotFound):
			RespondError(c, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found", nil)
			return
		default:
			logger.Error("Failed to delete order %d for user %d: %v", orderID, userID, err)
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete order", nil)
			return
		}
	}

	RespondSuccess(c, http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

func (h *OrderHandler) readActorAndOrderID(c *gin.Context) (uint, string, uint, bool) {
	userID, err := userIDFromContext(c)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, "AUTH_CONTEXT_INVALID", "Invalid authentication context", nil)
		return 0, "", 0, false
	}

	orderID64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "INVALID_ID", "Invalid order ID", nil)
		return 0, "", 0, false
	}

	return userID, userRoleFromContext(c), uint(orderID64), true
}
