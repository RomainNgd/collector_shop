package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"poc-gin/config"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v84"
	stripewebhook "github.com/stripe/stripe-go/v84/webhook"
)

type stripeTestBackend struct {
	session *stripe.CheckoutSession
	err     error
}

func (b *stripeTestBackend) Call(_ string, _ string, _ string, _ stripe.ParamsContainer, target stripe.LastResponseSetter) error {
	if b.err != nil {
		return b.err
	}
	if session, ok := target.(*stripe.CheckoutSession); ok && b.session != nil {
		*session = *b.session
	}
	return nil
}

func (b *stripeTestBackend) CallStreaming(_ string, _ string, _ string, _ stripe.ParamsContainer, _ stripe.StreamingLastResponseSetter) error {
	return b.err
}

func (b *stripeTestBackend) CallRaw(_ string, _ string, _ string, _ []byte, _ *stripe.Params, _ stripe.LastResponseSetter) error {
	return b.err
}

func (b *stripeTestBackend) CallMultipart(_ string, _ string, _ string, _ string, _ *bytes.Buffer, _ *stripe.Params, _ stripe.LastResponseSetter) error {
	return b.err
}

func (b *stripeTestBackend) SetMaxNetworkRetries(_ int64) {}

func useStripeTestBackend(t *testing.T, backend stripe.Backend) {
	t.Helper()
	original := stripe.GetBackend(stripe.APIBackend)
	stripe.SetBackend(stripe.APIBackend, backend)
	t.Cleanup(func() { stripe.SetBackend(stripe.APIBackend, original) })
}

func TestStripeServiceDisabled(t *testing.T) {
	service := NewStripeService(nil)
	if service.Enabled() {
		t.Fatal("expected nil Stripe configuration to disable the service")
	}

	if _, err := service.CreateCheckoutSession(context.Background(), StripeCheckoutSessionInput{}); !errors.Is(err, ErrStripeNotEnabled) {
		t.Fatalf("expected disabled create error, got %v", err)
	}
	if _, err := service.GetCheckoutSession(context.Background(), "cs_test"); !errors.Is(err, ErrStripeNotEnabled) {
		t.Fatalf("expected disabled get error, got %v", err)
	}
	if _, err := service.ConstructWebhookEvent(nil, ""); !errors.Is(err, ErrStripeNotEnabled) {
		t.Fatalf("expected disabled webhook error, got %v", err)
	}
	if _, err := service.ExpireCheckoutSession(context.Background(), "cs_test"); !errors.Is(err, ErrStripeNotEnabled) {
		t.Fatalf("expected disabled expire error, got %v", err)
	}
}

func TestStripeServiceConfigurationAndInvalidWebhook(t *testing.T) {
	service := NewStripeService(&config.StripeConfig{
		Enabled:       true,
		SecretKey:     "sk_test_placeholder",
		WebhookSecret: "whsec_test_placeholder",
	})
	if !service.Enabled() {
		t.Fatal("expected Stripe service to be enabled")
	}

	if _, err := service.ConstructWebhookEvent([]byte("invalid"), "invalid"); !errors.Is(err, ErrStripeInvalidWebhook) {
		t.Fatalf("expected invalid webhook error, got %v", err)
	}
}

func TestMapStripeCheckoutSession(t *testing.T) {
	if mapStripeCheckoutSession(nil) != nil {
		t.Fatal("expected nil Stripe session to stay nil")
	}

	expiresAt := time.Now().UTC().Truncate(time.Second)
	mapped := mapStripeCheckoutSession(&stripe.CheckoutSession{
		ID:            "cs_test",
		URL:           "https://checkout.stripe.test/cs_test",
		Status:        stripe.CheckoutSessionStatusOpen,
		PaymentStatus: stripe.CheckoutSessionPaymentStatusUnpaid,
		PaymentIntent: &stripe.PaymentIntent{ID: "pi_test"},
		ExpiresAt:     expiresAt.Unix(),
		Metadata:      map[string]string{"order_id": "42"},
	})

	if mapped.ID != "cs_test" || mapped.PaymentIntentID != "pi_test" {
		t.Fatalf("unexpected mapped Stripe session: %#v", mapped)
	}
	if mapped.ExpiresAt == nil || !mapped.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected expiration date: %#v", mapped.ExpiresAt)
	}
	if mapped.Metadata["order_id"] != "42" {
		t.Fatalf("unexpected metadata: %#v", mapped.Metadata)
	}
}

func TestStripeCheckoutSessionCalls(t *testing.T) {
	backend := &stripeTestBackend{session: &stripe.CheckoutSession{
		ID:            "cs_test",
		URL:           "https://checkout.stripe.test/cs_test",
		Status:        stripe.CheckoutSessionStatusOpen,
		PaymentStatus: stripe.CheckoutSessionPaymentStatusUnpaid,
	}}
	useStripeTestBackend(t, backend)
	service := NewStripeService(&config.StripeConfig{Enabled: true, SecretKey: "sk_test"})

	created, err := service.CreateCheckoutSession(context.Background(), StripeCheckoutSessionInput{
		SuccessURL:        "https://collector.test/success",
		CancelURL:         "https://collector.test/cancel",
		CustomerEmail:     "user@collector.test",
		ClientReferenceID: "42",
		Metadata:          map[string]string{"order_id": "42"},
		LineItems: []StripeCheckoutLineItem{{
			Name:        "Console",
			Description: "Retro console",
			ImageURL:    "https://collector.test/console.png",
			Quantity:    2,
			Currency:    "EUR",
			UnitAmount:  4990,
		}},
	})
	if err != nil || created.ID != "cs_test" {
		t.Fatalf("expected checkout creation success, got %#v, %v", created, err)
	}

	loaded, err := service.GetCheckoutSession(context.Background(), "cs_test")
	if err != nil || loaded.ID != "cs_test" {
		t.Fatalf("expected checkout lookup success, got %#v, %v", loaded, err)
	}

	expired, err := service.ExpireCheckoutSession(context.Background(), "cs_test")
	if err != nil || expired.ID != "cs_test" {
		t.Fatalf("expected checkout expiration success, got %#v, %v", expired, err)
	}
}

func TestStripeCheckoutSessionErrors(t *testing.T) {
	backend := &stripeTestBackend{err: errors.New("stripe unavailable")}
	useStripeTestBackend(t, backend)
	service := NewStripeService(&config.StripeConfig{Enabled: true, SecretKey: "sk_test"})

	if _, err := service.CreateCheckoutSession(context.Background(), StripeCheckoutSessionInput{}); err == nil {
		t.Fatal("expected checkout creation error")
	}
	if _, err := service.GetCheckoutSession(context.Background(), "cs_test"); err == nil {
		t.Fatal("expected checkout lookup error")
	}
	if _, err := service.ExpireCheckoutSession(context.Background(), "cs_test"); err == nil {
		t.Fatal("expected checkout expiration error")
	}

	backend.err = &stripe.Error{Code: stripe.ErrorCodeResourceMissing, Msg: "missing"}
	if _, err := service.GetCheckoutSession(context.Background(), "missing"); !errors.Is(err, ErrStripeSessionNotFound) {
		t.Fatalf("expected missing session error, got %v", err)
	}
	if _, err := service.ExpireCheckoutSession(context.Background(), "missing"); !errors.Is(err, ErrStripeSessionNotFound) {
		t.Fatalf("expected missing session error, got %v", err)
	}
}

func signedWebhookPayload(t *testing.T, secret string, payload []byte) (string, string) {
	t.Helper()

	signed := stripewebhook.GenerateTestSignedPayload(&stripewebhook.UnsignedPayload{
		Payload: payload,
		Secret:  secret,
	})

	return string(signed.Payload), signed.Header
}

func TestConstructWebhookEventIgnoresNonCheckoutEvents(t *testing.T) {
	const secret = "whsec_test_construct"
	service := NewStripeService(&config.StripeConfig{
		Enabled:       true,
		SecretKey:     "sk_test",
		WebhookSecret: secret,
	})

	rawPayload := []byte(fmt.Sprintf(`{
		"id": "evt_test_charge",
		"object": "event",
		"api_version": %q,
		"type": "charge.succeeded",
		"data": {"object": {}}
	}`, stripe.APIVersion))

	payload, signature := signedWebhookPayload(t, secret, rawPayload)

	event, err := service.ConstructWebhookEvent([]byte(payload), signature)
	if err != nil {
		t.Fatalf("expected event construction success, got %v", err)
	}
	if event.Type != "charge.succeeded" {
		t.Fatalf("expected charge.succeeded type, got %s", event.Type)
	}
	if event.CheckoutSession.ID != "" {
		t.Fatalf("expected empty checkout session for non-checkout event, got %#v", event.CheckoutSession)
	}
}

func TestConstructWebhookEventParsesCheckoutSession(t *testing.T) {
	const secret = "whsec_test_construct"
	service := NewStripeService(&config.StripeConfig{
		Enabled:       true,
		SecretKey:     "sk_test",
		WebhookSecret: secret,
	})

	rawPayload := []byte(fmt.Sprintf(`{
		"id": "evt_test_checkout",
		"object": "event",
		"api_version": %q,
		"type": "checkout.session.completed",
		"data": {
			"object": {
				"id": "cs_test_webhook",
				"url": "https://checkout.stripe.test/cs_test_webhook",
				"status": "complete",
				"payment_status": "paid",
				"metadata": {"order_id": "7"}
			}
		}
	}`, stripe.APIVersion))

	payload, signature := signedWebhookPayload(t, secret, rawPayload)

	event, err := service.ConstructWebhookEvent([]byte(payload), signature)
	if err != nil {
		t.Fatalf("expected event construction success, got %v", err)
	}
	if event.Type != "checkout.session.completed" {
		t.Fatalf("expected checkout.session.completed type, got %s", event.Type)
	}
	if event.CheckoutSession.ID != "cs_test_webhook" {
		t.Fatalf("expected mapped checkout session, got %#v", event.CheckoutSession)
	}
	if event.CheckoutSession.PaymentStatus != "paid" {
		t.Fatalf("expected paid payment status, got %s", event.CheckoutSession.PaymentStatus)
	}
	if event.CheckoutSession.Metadata["order_id"] != "7" {
		t.Fatalf("expected order metadata, got %#v", event.CheckoutSession.Metadata)
	}
}
