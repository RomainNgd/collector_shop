package services

import (
	"context"
	"errors"
	"fmt"
	"poc-gin/models"
	"poc-gin/pkg/constants"
	"strconv"
	"testing"
	"time"
)

type fakeStripeService struct {
	createFn    func(input StripeCheckoutSessionInput) (*StripeCheckoutSession, error)
	getFn       func(sessionID string) (*StripeCheckoutSession, error)
	expireFn    func(sessionID string) (*StripeCheckoutSession, error)
	webhookFn   func(payload []byte, signature string) (*StripeWebhookEvent, error)
	createCalls int
	getCalls    int
	expireCalls int
}

func (f *fakeStripeService) Enabled() bool {
	return true
}

func (f *fakeStripeService) CreateCheckoutSession(_ context.Context, input StripeCheckoutSessionInput) (*StripeCheckoutSession, error) {
	f.createCalls++
	if f.createFn != nil {
		return f.createFn(input)
	}
	return nil, fmt.Errorf("unexpected CreateCheckoutSession call")
}

func (f *fakeStripeService) GetCheckoutSession(_ context.Context, sessionID string) (*StripeCheckoutSession, error) {
	f.getCalls++
	if f.getFn != nil {
		return f.getFn(sessionID)
	}
	return nil, fmt.Errorf("unexpected GetCheckoutSession call")
}

func (f *fakeStripeService) ExpireCheckoutSession(_ context.Context, sessionID string) (*StripeCheckoutSession, error) {
	f.expireCalls++
	if f.expireFn != nil {
		return f.expireFn(sessionID)
	}
	return nil, fmt.Errorf("unexpected ExpireCheckoutSession call")
}

func (f *fakeStripeService) ConstructWebhookEvent(_ []byte, _ string) (*StripeWebhookEvent, error) {
	if f.webhookFn != nil {
		return f.webhookFn(nil, "")
	}
	return nil, fmt.Errorf("unexpected ConstructWebhookEvent call")
}

func TestOrderPaymentServiceCreateStripeCheckoutSession(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := orderService.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 2},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	stripeService := &fakeStripeService{
		createFn: func(input StripeCheckoutSessionInput) (*StripeCheckoutSession, error) {
			if input.Metadata["order_id"] != strconv.FormatUint(uint64(order.ID), 10) {
				t.Fatalf("expected order metadata, got %#v", input.Metadata)
			}
			if len(input.LineItems) != 1 {
				t.Fatalf("expected one line item, got %d", len(input.LineItems))
			}

			return &StripeCheckoutSession{
				ID:            "cs_test_order_1",
				URL:           "https://checkout.stripe.com/c/pay/cs_test_order_1",
				Status:        "open",
				PaymentStatus: "unpaid",
				ExpiresAt:     &expiresAt,
				Metadata:      input.Metadata,
			}, nil
		},
	}

	service := NewOrderPaymentService(tx, stripeService, orderService)
	result, err := service.CreateStripeCheckoutSession(
		context.Background(),
		user.ID,
		order.ID,
		constants.RoleUser,
		"http://localhost:5173/mes-commandes/1?payment=processing",
		"http://localhost:5173/mes-commandes/1?payment=cancelled",
	)
	if err != nil {
		t.Fatalf("expected checkout session creation success, got %v", err)
	}

	if result.URL == "" || result.Reused {
		t.Fatalf("unexpected checkout result: %#v", result)
	}

	updatedOrder, err := orderService.GetOrderByID(context.Background(), user.ID, order.ID, constants.RoleUser)
	if err != nil {
		t.Fatalf("expected updated order fetch success, got %v", err)
	}
	if updatedOrder.PaymentProvider != models.PaymentProviderStripe {
		t.Fatalf("expected stripe provider, got %s", updatedOrder.PaymentProvider)
	}
	if updatedOrder.PaymentStatus != models.OrderPaymentStatusCheckoutOpen {
		t.Fatalf("expected checkout_open payment status, got %s", updatedOrder.PaymentStatus)
	}
	if updatedOrder.StripeCheckoutSessionID != "cs_test_order_1" {
		t.Fatalf("expected stored checkout session id, got %s", updatedOrder.StripeCheckoutSessionID)
	}
}

func TestOrderPaymentServiceRejectsDisallowedReturnURL(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := orderService.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	stripeService := &fakeStripeService{
		createFn: func(input StripeCheckoutSessionInput) (*StripeCheckoutSession, error) {
			return nil, fmt.Errorf("create should not be called for disallowed return URLs")
		},
	}

	service := NewOrderPaymentService(
		tx,
		stripeService,
		orderService,
		[]string{"https://collector-app.example.test"},
	)
	_, err = service.CreateStripeCheckoutSession(
		context.Background(),
		user.ID,
		order.ID,
		constants.RoleUser,
		"https://evil.example.test/mes-commandes/1?payment=processing",
		"https://collector-app.example.test/mes-commandes/1?payment=cancelled",
	)
	if err != ErrCheckoutReturnURLInvalid {
		t.Fatalf("expected ErrCheckoutReturnURLInvalid, got %v", err)
	}
	if stripeService.createCalls != 0 {
		t.Fatalf("expected no stripe create call, got %d", stripeService.createCalls)
	}
}

func TestOrderPaymentServiceReusesOpenStripeCheckoutSession(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := orderService.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	if err := tx.Model(order).Updates(map[string]any{
		"payment_provider":               models.PaymentProviderStripe,
		"payment_status":                 models.OrderPaymentStatusCheckoutOpen,
		"stripe_checkout_session_id":     "cs_test_existing",
		"stripe_checkout_session_status": "open",
	}).Error; err != nil {
		t.Fatalf("failed to seed order checkout session: %v", err)
	}

	stripeService := &fakeStripeService{
		createFn: func(input StripeCheckoutSessionInput) (*StripeCheckoutSession, error) {
			return nil, fmt.Errorf("create should not be called when checkout session is reusable")
		},
		getFn: func(sessionID string) (*StripeCheckoutSession, error) {
			return &StripeCheckoutSession{
				ID:            sessionID,
				URL:           "https://checkout.stripe.com/c/pay/cs_test_existing",
				Status:        "open",
				PaymentStatus: "unpaid",
				Metadata: map[string]string{
					"order_id": strconv.FormatUint(uint64(order.ID), 10),
				},
			}, nil
		},
	}

	service := NewOrderPaymentService(tx, stripeService, orderService)
	result, err := service.CreateStripeCheckoutSession(
		context.Background(),
		user.ID,
		order.ID,
		constants.RoleUser,
		"http://localhost:5173/mes-commandes/1?payment=processing",
		"http://localhost:5173/mes-commandes/1?payment=cancelled",
	)
	if err != nil {
		t.Fatalf("expected checkout session reuse success, got %v", err)
	}
	if !result.Reused || stripeService.createCalls != 0 {
		t.Fatalf("expected existing checkout session to be reused, got %#v createCalls=%d", result, stripeService.createCalls)
	}
}

func TestOrderPaymentServiceWebhookMarksOrderAsPaid(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := orderService.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	stripeService := &fakeStripeService{
		webhookFn: func(payload []byte, signature string) (*StripeWebhookEvent, error) {
			return &StripeWebhookEvent{
				Type: stripeEventCheckoutCompleted,
				CheckoutSession: StripeCheckoutSession{
					ID:              "cs_test_paid",
					Status:          "complete",
					PaymentStatus:   "paid",
					PaymentIntentID: "pi_test_123",
					Metadata: map[string]string{
						"order_id": strconv.FormatUint(uint64(order.ID), 10),
					},
				},
			}, nil
		},
	}

	service := NewOrderPaymentService(tx, stripeService, orderService)
	if err := service.HandleStripeWebhook(context.Background(), []byte(`{}`), "t=1,v1=test"); err != nil {
		t.Fatalf("expected webhook processing success, got %v", err)
	}

	updatedOrder, err := orderService.GetOrderByID(context.Background(), user.ID, order.ID, constants.RoleUser)
	if err != nil {
		t.Fatalf("expected updated order fetch success, got %v", err)
	}
	if updatedOrder.Status != models.OrderStatusPreparation {
		t.Fatalf("expected preparation status, got %s", updatedOrder.Status)
	}
	if updatedOrder.PaymentStatus != models.OrderPaymentStatusPaid {
		t.Fatalf("expected paid payment status, got %s", updatedOrder.PaymentStatus)
	}
	if updatedOrder.PaidAt == nil {
		t.Fatal("expected paid_at to be set")
	}
	if updatedOrder.StripePaymentIntentID != "pi_test_123" {
		t.Fatalf("expected payment intent id to be stored, got %s", updatedOrder.StripePaymentIntentID)
	}
}

func TestOrderPaymentServiceWebhookRejectsInvalidStripeSignature(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)

	stripeService := &fakeStripeService{
		webhookFn: func(payload []byte, signature string) (*StripeWebhookEvent, error) {
			return nil, ErrStripeInvalidWebhook
		},
	}

	service := NewOrderPaymentService(tx, stripeService, orderService)
	err := service.HandleStripeWebhook(context.Background(), []byte(`{"id":"evt_invalid"}`), "bad-signature")
	if !errors.Is(err, ErrStripeInvalidWebhook) {
		t.Fatalf("expected ErrStripeInvalidWebhook, got %v", err)
	}
}

func TestOrderPaymentServiceWebhookMarksOrderAsExpired(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := orderService.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	stripeService := &fakeStripeService{
		webhookFn: func(payload []byte, signature string) (*StripeWebhookEvent, error) {
			return &StripeWebhookEvent{
				Type: stripeEventCheckoutExpired,
				CheckoutSession: StripeCheckoutSession{
					ID:            "cs_test_expired",
					Status:        "expired",
					PaymentStatus: "unpaid",
					Metadata: map[string]string{
						"order_id": strconv.FormatUint(uint64(order.ID), 10),
					},
				},
			}, nil
		},
	}

	service := NewOrderPaymentService(tx, stripeService, orderService)
	if err := service.HandleStripeWebhook(context.Background(), []byte(`{}`), "t=1,v1=test"); err != nil {
		t.Fatalf("expected expired webhook processing success, got %v", err)
	}

	updatedOrder, err := orderService.GetOrderByID(context.Background(), user.ID, order.ID, constants.RoleUser)
	if err != nil {
		t.Fatalf("expected updated order fetch success, got %v", err)
	}
	if updatedOrder.Status != models.OrderStatusAwaitingPayment {
		t.Fatalf("expected order to stay awaiting payment, got %s", updatedOrder.Status)
	}
	if updatedOrder.PaymentStatus != models.OrderPaymentStatusExpired {
		t.Fatalf("expected expired payment status, got %s", updatedOrder.PaymentStatus)
	}
	if updatedOrder.PaidAt != nil {
		t.Fatal("expected paid_at to stay empty")
	}
}

func TestOrderPaymentServiceWebhookMarksAsyncPaymentFailed(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := orderService.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	stripeService := &fakeStripeService{
		webhookFn: func(payload []byte, signature string) (*StripeWebhookEvent, error) {
			return &StripeWebhookEvent{
				Type: stripeEventCheckoutAsyncPaymentFailed,
				CheckoutSession: StripeCheckoutSession{
					ID:            "cs_test_async_failed",
					Status:        "complete",
					PaymentStatus: "unpaid",
					Metadata: map[string]string{
						"order_id": strconv.FormatUint(uint64(order.ID), 10),
					},
				},
			}, nil
		},
	}

	service := NewOrderPaymentService(tx, stripeService, orderService)
	if err := service.HandleStripeWebhook(context.Background(), []byte(`{}`), "t=1,v1=test"); err != nil {
		t.Fatalf("expected async failed webhook processing success, got %v", err)
	}

	updatedOrder, err := orderService.GetOrderByID(context.Background(), user.ID, order.ID, constants.RoleUser)
	if err != nil {
		t.Fatalf("expected updated order fetch success, got %v", err)
	}
	if updatedOrder.Status != models.OrderStatusAwaitingPayment {
		t.Fatalf("expected order to stay awaiting payment, got %s", updatedOrder.Status)
	}
	if updatedOrder.PaymentStatus != models.OrderPaymentStatusFailed {
		t.Fatalf("expected failed payment status, got %s", updatedOrder.PaymentStatus)
	}
	if updatedOrder.PaidAt != nil {
		t.Fatal("expected paid_at to stay empty")
	}
}

func TestOrderPaymentServiceWebhookMarksAsyncPaymentSucceeded(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := orderService.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	stripeService := &fakeStripeService{
		webhookFn: func(payload []byte, signature string) (*StripeWebhookEvent, error) {
			return &StripeWebhookEvent{
				Type: stripeEventCheckoutAsyncPaymentSuccess,
				CheckoutSession: StripeCheckoutSession{
					ID:              "cs_test_async_success",
					Status:          "complete",
					PaymentStatus:   "paid",
					PaymentIntentID: "pi_test_async_success",
					Metadata: map[string]string{
						"order_id": strconv.FormatUint(uint64(order.ID), 10),
					},
				},
			}, nil
		},
	}

	service := NewOrderPaymentService(tx, stripeService, orderService)
	if err := service.HandleStripeWebhook(context.Background(), []byte(`{}`), "t=1,v1=test"); err != nil {
		t.Fatalf("expected async success webhook processing success, got %v", err)
	}

	updatedOrder, err := orderService.GetOrderByID(context.Background(), user.ID, order.ID, constants.RoleUser)
	if err != nil {
		t.Fatalf("expected updated order fetch success, got %v", err)
	}
	if updatedOrder.Status != models.OrderStatusPreparation {
		t.Fatalf("expected preparation status, got %s", updatedOrder.Status)
	}
	if updatedOrder.PaymentStatus != models.OrderPaymentStatusPaid {
		t.Fatalf("expected paid payment status, got %s", updatedOrder.PaymentStatus)
	}
	if updatedOrder.PaidAt == nil {
		t.Fatal("expected paid_at to be set")
	}
	if updatedOrder.StripePaymentIntentID != "pi_test_async_success" {
		t.Fatalf("expected payment intent id to be stored, got %s", updatedOrder.StripePaymentIntentID)
	}
}

func TestOrderPaymentServiceReleaseCheckoutSessionExpiresOpenSession(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := orderService.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	stripeService := &fakeStripeService{
		createFn: func(input StripeCheckoutSessionInput) (*StripeCheckoutSession, error) {
			return &StripeCheckoutSession{
				ID:            "cs_test_release",
				URL:           "https://checkout.stripe.com/c/pay/cs_test_release",
				Status:        "open",
				PaymentStatus: "unpaid",
				Metadata:      input.Metadata,
			}, nil
		},
		getFn: func(sessionID string) (*StripeCheckoutSession, error) {
			return &StripeCheckoutSession{
				ID:            sessionID,
				URL:           "https://checkout.stripe.com/c/pay/" + sessionID,
				Status:        "open",
				PaymentStatus: "unpaid",
			}, nil
		},
		expireFn: func(sessionID string) (*StripeCheckoutSession, error) {
			return &StripeCheckoutSession{
				ID:            sessionID,
				Status:        "expired",
				PaymentStatus: "unpaid",
			}, nil
		},
	}

	service := NewOrderPaymentService(tx, stripeService, orderService)
	if _, err := service.CreateStripeCheckoutSession(
		context.Background(),
		user.ID,
		order.ID,
		constants.RoleUser,
		"https://shop.example/success",
		"https://shop.example/cancel",
	); err != nil {
		t.Fatalf("expected checkout session creation success, got %v", err)
	}

	if err := service.ReleaseCheckoutSession(context.Background(), user.ID, order.ID, constants.RoleUser); err != nil {
		t.Fatalf("expected release success, got %v", err)
	}
	if stripeService.expireCalls != 1 {
		t.Fatalf("expected one expire call, got %d", stripeService.expireCalls)
	}

	updatedOrder, err := orderService.GetOrderByID(context.Background(), user.ID, order.ID, constants.RoleUser)
	if err != nil {
		t.Fatalf("expected order fetch success, got %v", err)
	}
	if updatedOrder.PaymentStatus != models.OrderPaymentStatusExpired {
		t.Fatalf("expected expired payment status, got %s", updatedOrder.PaymentStatus)
	}
}

func TestOrderPaymentServiceReleaseCheckoutSessionWithoutSessionIsNoop(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)
	user := seedUser(t, tx, constants.RoleUser)
	category := seedCategory(t, tx)
	product := seedProduct(t, tx, category.ID)

	order, err := orderService.CreateOrder(context.Background(), user.ID, []OrderItemInput{
		{ProductID: product.ID, Quantity: 1},
	})
	if err != nil {
		t.Fatalf("expected order creation success, got %v", err)
	}

	stripeService := &fakeStripeService{}
	service := NewOrderPaymentService(tx, stripeService, orderService)

	if err := service.ReleaseCheckoutSession(context.Background(), user.ID, order.ID, constants.RoleUser); err != nil {
		t.Fatalf("expected noop release success, got %v", err)
	}
	if stripeService.getCalls != 0 || stripeService.expireCalls != 0 {
		t.Fatalf("expected no stripe calls, got get=%d expire=%d", stripeService.getCalls, stripeService.expireCalls)
	}
}

func TestOrderPaymentServiceWebhookAcknowledgesUnknownSession(t *testing.T) {
	tx := openIntegrationTx(t)
	orderService := NewOrderService(tx)

	stripeService := &fakeStripeService{
		webhookFn: func(payload []byte, signature string) (*StripeWebhookEvent, error) {
			return &StripeWebhookEvent{
				Type: stripeEventCheckoutCompleted,
				CheckoutSession: StripeCheckoutSession{
					ID:            "cs_test_orphan",
					Status:        "complete",
					PaymentStatus: "paid",
					Metadata: map[string]string{
						"order_id": "999999",
					},
				},
			}, nil
		},
	}

	service := NewOrderPaymentService(tx, stripeService, orderService)
	if err := service.HandleStripeWebhook(context.Background(), []byte(`{}`), "t=1,v1=test"); err != nil {
		t.Fatalf("expected orphan webhook to be acknowledged, got %v", err)
	}
}
