package controllers

import (
	"errors"
	"io"
	"net/http"
	"poc-gin/pkg/logger"
	"poc-gin/services"

	"github.com/gin-gonic/gin"
)

// stripeWebhookMaxBodyBytes follows Stripe's own guidance: webhook payloads
// are small JSON events, so anything past a generous margin is abuse.
const stripeWebhookMaxBodyBytes = 64 * 1024

type PaymentHandler struct {
	orderPayments services.OrderPaymentServiceInterface
}

func NewPaymentHandler(orderPayments services.OrderPaymentServiceInterface) *PaymentHandler {
	return &PaymentHandler{orderPayments: orderPayments}
}

func (h *PaymentHandler) HandleStripeWebhook(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, stripeWebhookMaxBodyBytes)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "WEBHOOK_BODY_INVALID", "Invalid webhook payload", nil)
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		RespondError(c, http.StatusBadRequest, "WEBHOOK_SIGNATURE_MISSING", "Missing Stripe signature", nil)
		return
	}

	if err := h.orderPayments.HandleStripeWebhook(c.Request.Context(), payload, signature); err != nil {
		switch {
		case errors.Is(err, services.ErrStripeNotEnabled):
			RespondError(c, http.StatusServiceUnavailable, "STRIPE_NOT_CONFIGURED", "Stripe payment is not configured", nil)
			return
		case errors.Is(err, services.ErrStripeInvalidWebhook):
			RespondError(c, http.StatusBadRequest, "STRIPE_WEBHOOK_INVALID", "Invalid Stripe webhook", nil)
			return
		default:
			logger.Error("Failed to process Stripe webhook: %v", err)
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to process Stripe webhook", nil)
			return
		}
	}

	c.Status(http.StatusOK)
}
