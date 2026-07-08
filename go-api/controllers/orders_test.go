package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"poc-gin/services"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func performOrderJSONRequest(
	handlerFunc gin.HandlerFunc,
	method,
	target string,
	body any,
	params []gin.Param,
	setup func(c *gin.Context),
) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	}

	req, _ := http.NewRequest(method, target, reader)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	ctx.Params = params

	if setup != nil {
		setup(ctx)
	}

	handlerFunc(ctx)
	return recorder
}

func TestOrderHandlerCreateOrder(t *testing.T) {
	t.Run("returns 401 when auth context is missing", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(handler.CreateOrder, http.MethodPost, "/orders", map[string]any{
			"items": []map[string]any{{"product_id": 2, "quantity": 1}},
		}, nil, nil)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
	})

	t.Run("returns 400 on invalid payload", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(handler.CreateOrder, http.MethodPost, "/orders", map[string]any{
			"items": []map[string]any{{"product_id": 0, "quantity": 0}},
		}, nil, func(c *gin.Context) {
			c.Set(constants.ContextKeyUserID, float64(5))
		})

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", recorder.Code)
		}
	})

	t.Run("returns 201 on success", func(t *testing.T) {
		var receivedUserID uint
		var receivedItems []services.OrderItemInput
		handler := NewOrderHandler(&mockOrderService{
			createFn: func(userID uint, items []services.OrderItemInput) (*models.Order, error) {
				receivedUserID = userID
				receivedItems = items
				return &models.Order{Status: models.OrderStatusAwaitingPayment, Total: 29.99}, nil
			},
		}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(handler.CreateOrder, http.MethodPost, "/orders", map[string]any{
			"items": []map[string]any{
				{"product_id": 4, "quantity": 2},
			},
		}, nil, func(c *gin.Context) {
			c.Set(constants.ContextKeyUserID, float64(8))
		})

		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", recorder.Code)
		}
		if receivedUserID != 8 {
			t.Fatalf("expected user id 8, got %d", receivedUserID)
		}
		if len(receivedItems) != 1 || receivedItems[0].ProductID != 4 || receivedItems[0].Quantity != 2 {
			t.Fatalf("unexpected items: %#v", receivedItems)
		}
	})

	t.Run("returns 403 when ordering own product", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{
			createFn: func(userID uint, items []services.OrderItemInput) (*models.Order, error) {
				return nil, services.ErrOrderOwnProduct
			},
		}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(handler.CreateOrder, http.MethodPost, "/orders", map[string]any{
			"items": []map[string]any{{"product_id": 4, "quantity": 1}},
		}, nil, func(c *gin.Context) {
			c.Set(constants.ContextKeyUserID, float64(8))
		})

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", recorder.Code)
		}
	})
}

func TestOrderHandlerFindOrder(t *testing.T) {
	handler := NewOrderHandler(&mockOrderService{
		listFn: func(userID uint) ([]*models.Order, error) {
			if userID != 12 {
				t.Fatalf("expected user id 12, got %d", userID)
			}
			return []*models.Order{{Status: models.OrderStatusAwaitingPayment}}, nil
		},
	}, &mockOrderPaymentService{})

	recorder := performOrderJSONRequest(handler.FindOrder, http.MethodGet, "/orders", nil, nil, func(c *gin.Context) {
		c.Set(constants.ContextKeyUserID, float64(12))
	})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestOrderHandlerFindOneOrder(t *testing.T) {
	t.Run("returns 404 when order is missing", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{
			getByIDFn: func(actorID, orderID uint, actorRole string) (*models.Order, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(
			handler.FindOneOrder,
			http.MethodGet,
			"/orders/9",
			nil,
			[]gin.Param{{Key: "id", Value: "9"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(4))
			},
		)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{
			getByIDFn: func(actorID, orderID uint, actorRole string) (*models.Order, error) {
				if actorID != 4 || orderID != 9 || actorRole != constants.RoleUser {
					t.Fatalf("unexpected lookup args: actorID=%d orderID=%d role=%s", actorID, orderID, actorRole)
				}
				return &models.Order{Status: models.OrderStatusAwaitingPayment}, nil
			},
		}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(
			handler.FindOneOrder,
			http.MethodGet,
			"/orders/9",
			nil,
			[]gin.Param{{Key: "id", Value: "9"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(4))
			},
		)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}

func TestOrderHandlerUpdateOrder(t *testing.T) {
	t.Run("returns 409 when transition is not allowed", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{
			updateFn: func(actorID, orderID uint, actorRole, status string) (*models.Order, error) {
				return nil, services.ErrOrderStatusTransitionNotAllowed
			},
		}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(
			handler.UpdateOrder,
			http.MethodPut,
			"/orders/3",
			map[string]any{"status": models.OrderStatusDelivered},
			[]gin.Param{{Key: "id", Value: "3"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(7))
			},
		)

		if recorder.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", recorder.Code)
		}
	})

	t.Run("returns 200 when payment validation succeeds", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{
			updateFn: func(actorID, orderID uint, actorRole, status string) (*models.Order, error) {
				if status != models.OrderStatusPreparation {
					t.Fatalf("expected preparation status, got %s", status)
				}
				return &models.Order{Status: status}, nil
			},
		}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(
			handler.UpdateOrder,
			http.MethodPut,
			"/orders/3",
			map[string]any{"status": models.OrderStatusPreparation},
			[]gin.Param{{Key: "id", Value: "3"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(7))
			},
		)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}

func TestOrderHandlerDeleteOrder(t *testing.T) {
	t.Run("returns 409 when deletion is not allowed", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{
			deleteFn: func(actorID, orderID uint, actorRole string) error {
				return services.ErrOrderDeletionNotAllowed
			},
		}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(
			handler.DeleteOrder,
			http.MethodDelete,
			"/orders/4",
			nil,
			[]gin.Param{{Key: "id", Value: "4"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(9))
			},
		)

		if recorder.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d", recorder.Code)
		}
	})

	t.Run("returns 500 on unexpected error", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{
			deleteFn: func(actorID, orderID uint, actorRole string) error {
				return errors.New("db error")
			},
		}, &mockOrderPaymentService{})

		recorder := performOrderJSONRequest(
			handler.DeleteOrder,
			http.MethodDelete,
			"/orders/4",
			nil,
			[]gin.Param{{Key: "id", Value: "4"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(9))
			},
		)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("returns 502 and keeps the order when checkout release fails", func(t *testing.T) {
		deleted := false
		handler := NewOrderHandler(&mockOrderService{
			deleteFn: func(actorID, orderID uint, actorRole string) error {
				deleted = true
				return nil
			},
		}, &mockOrderPaymentService{
			releaseCheckoutFn: func(actorID, orderID uint, actorRole string) error {
				return errors.New("stripe unavailable")
			},
		})

		recorder := performOrderJSONRequest(
			handler.DeleteOrder,
			http.MethodDelete,
			"/orders/4",
			nil,
			[]gin.Param{{Key: "id", Value: "4"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(9))
			},
		)

		if recorder.Code != http.StatusBadGateway {
			t.Fatalf("expected 502, got %d", recorder.Code)
		}
		if deleted {
			t.Fatal("expected order deletion to be skipped when release fails")
		}
	})

	t.Run("releases the checkout session before deleting", func(t *testing.T) {
		released := false
		handler := NewOrderHandler(&mockOrderService{
			deleteFn: func(actorID, orderID uint, actorRole string) error {
				if !released {
					t.Fatal("expected checkout release before deletion")
				}
				return nil
			},
		}, &mockOrderPaymentService{
			releaseCheckoutFn: func(actorID, orderID uint, actorRole string) error {
				released = true
				return nil
			},
		})

		recorder := performOrderJSONRequest(
			handler.DeleteOrder,
			http.MethodDelete,
			"/orders/4",
			nil,
			[]gin.Param{{Key: "id", Value: "4"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(9))
			},
		)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", recorder.Code)
		}
	})
}

func TestOrderHandlerCreateCheckoutSession(t *testing.T) {
	t.Run("returns 503 when stripe is not configured", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{}, &mockOrderPaymentService{
			createCheckoutFn: func(actorID, orderID uint, actorRole, successURL, cancelURL string) (*services.OrderCheckoutSessionResult, error) {
				return nil, services.ErrStripeNotEnabled
			},
		})

		recorder := performOrderJSONRequest(
			handler.CreateCheckoutSession,
			http.MethodPost,
			"/orders/3/checkout-session",
			map[string]any{"success_url": "http://localhost/success", "cancel_url": "http://localhost/cancel"},
			[]gin.Param{{Key: "id", Value: "3"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(7))
			},
		)

		if recorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d", recorder.Code)
		}
	})

	t.Run("returns 201 with stripe redirect url", func(t *testing.T) {
		handler := NewOrderHandler(&mockOrderService{}, &mockOrderPaymentService{
			createCheckoutFn: func(actorID, orderID uint, actorRole, successURL, cancelURL string) (*services.OrderCheckoutSessionResult, error) {
				if actorID != 7 || orderID != 3 {
					t.Fatalf("unexpected actor/order ids: %d %d", actorID, orderID)
				}
				if successURL == "" || cancelURL == "" {
					t.Fatalf("expected success and cancel urls")
				}
				return &services.OrderCheckoutSessionResult{
					SessionID: "cs_test_123",
					URL:       "https://checkout.stripe.com/c/pay/cs_test_123",
				}, nil
			},
		})

		recorder := performOrderJSONRequest(
			handler.CreateCheckoutSession,
			http.MethodPost,
			"/orders/3/checkout-session",
			map[string]any{"success_url": "http://localhost/success", "cancel_url": "http://localhost/cancel"},
			[]gin.Param{{Key: "id", Value: "3"}},
			func(c *gin.Context) {
				c.Set(constants.ContextKeyUserID, float64(7))
			},
		)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", recorder.Code)
		}
	})
}
