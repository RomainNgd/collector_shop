package controllers

import (
	"net/http"
	"net/http/httptest"
	"poc-gin/services"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPaymentHandlerHandleStripeWebhook(t *testing.T) {
	t.Run("returns 400 when signature is missing", func(t *testing.T) {
		handler := NewPaymentHandler(&mockOrderPaymentService{})

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest(http.MethodPost, "/payments/stripe/webhook", strings.NewReader(`{}`))
		ctx.Request = req

		handler.HandleStripeWebhook(ctx)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 400 for invalid stripe webhook", func(t *testing.T) {
		handler := NewPaymentHandler(&mockOrderPaymentService{
			handleWebhookFn: func(payload []byte, signature string) error {
				return services.ErrStripeInvalidWebhook
			},
		})

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest(http.MethodPost, "/payments/stripe/webhook", strings.NewReader(`{}`))
		req.Header.Set("Stripe-Signature", "t=1,v1=test")
		ctx.Request = req

		handler.HandleStripeWebhook(ctx)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		handler := NewPaymentHandler(&mockOrderPaymentService{
			handleWebhookFn: func(payload []byte, signature string) error {
				return nil
			},
		})

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest(http.MethodPost, "/payments/stripe/webhook", strings.NewReader(`{}`))
		req.Header.Set("Stripe-Signature", "t=1,v1=test")
		ctx.Request = req

		handler.HandleStripeWebhook(ctx)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}
