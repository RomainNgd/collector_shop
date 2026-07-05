package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"poc-gin/config"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v84"
	checkoutsession "github.com/stripe/stripe-go/v84/checkout/session"
	stripewebhook "github.com/stripe/stripe-go/v84/webhook"
)

var (
	ErrStripeNotEnabled      = errors.New("stripe is not enabled")
	ErrStripeInvalidWebhook  = errors.New("stripe webhook is invalid")
	ErrStripeSessionNotFound = errors.New("stripe checkout session not found")
)

type StripeCheckoutLineItem struct {
	Name        string
	Description string
	ImageURL    string
	Quantity    int64
	Currency    string
	UnitAmount  int64
}

type StripeCheckoutSessionInput struct {
	SuccessURL        string
	CancelURL         string
	CustomerEmail     string
	ClientReferenceID string
	Metadata          map[string]string
	LineItems         []StripeCheckoutLineItem
}

type StripeCheckoutSession struct {
	ID              string
	URL             string
	Status          string
	PaymentStatus   string
	PaymentIntentID string
	ExpiresAt       *time.Time
	Metadata        map[string]string
}

type StripeWebhookEvent struct {
	Type            string
	CheckoutSession StripeCheckoutSession
}

type StripeServiceInterface interface {
	Enabled() bool
	CreateCheckoutSession(ctx context.Context, input StripeCheckoutSessionInput) (*StripeCheckoutSession, error)
	GetCheckoutSession(ctx context.Context, sessionID string) (*StripeCheckoutSession, error)
	ConstructWebhookEvent(payload []byte, signature string) (*StripeWebhookEvent, error)
}

type StripeService struct {
	enabled       bool
	secretKey     string
	webhookSecret string
}

func NewStripeService(cfg *config.StripeConfig) *StripeService {
	if cfg == nil {
		return &StripeService{}
	}

	return &StripeService{
		enabled:       cfg.Enabled,
		secretKey:     cfg.SecretKey,
		webhookSecret: cfg.WebhookSecret,
	}
}

func (s *StripeService) Enabled() bool {
	return s != nil && s.enabled
}

func (s *StripeService) CreateCheckoutSession(_ context.Context, input StripeCheckoutSessionInput) (*StripeCheckoutSession, error) {
	if !s.Enabled() {
		return nil, ErrStripeNotEnabled
	}

	params := &stripe.CheckoutSessionParams{
		SuccessURL:        stripe.String(input.SuccessURL),
		CancelURL:         stripe.String(input.CancelURL),
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		ClientReferenceID: stripe.String(input.ClientReferenceID),
	}

	if input.CustomerEmail != "" {
		params.CustomerEmail = stripe.String(input.CustomerEmail)
	}

	for key, value := range input.Metadata {
		params.AddMetadata(key, value)
	}

	for _, item := range input.LineItems {
		lineItem := &stripe.CheckoutSessionLineItemParams{
			Quantity: stripe.Int64(item.Quantity),
			PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
				Currency: stripe.String(strings.ToLower(item.Currency)),
				UnitAmount: stripe.Int64(
					item.UnitAmount,
				),
				ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
					Name: stripe.String(item.Name),
				},
			},
		}

		if item.Description != "" {
			lineItem.PriceData.ProductData.Description = stripe.String(item.Description)
		}

		if item.ImageURL != "" {
			lineItem.PriceData.ProductData.Images = []*string{stripe.String(item.ImageURL)}
		}

		params.LineItems = append(params.LineItems, lineItem)
	}

	stripe.Key = s.secretKey
	session, err := checkoutsession.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe checkout session: %w", err)
	}

	return mapStripeCheckoutSession(session), nil
}

func (s *StripeService) GetCheckoutSession(_ context.Context, sessionID string) (*StripeCheckoutSession, error) {
	if !s.Enabled() {
		return nil, ErrStripeNotEnabled
	}

	stripe.Key = s.secretKey
	session, err := checkoutsession.Get(sessionID, nil)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Code == stripe.ErrorCodeResourceMissing {
			return nil, ErrStripeSessionNotFound
		}
		return nil, fmt.Errorf("failed to fetch stripe checkout session: %w", err)
	}

	return mapStripeCheckoutSession(session), nil
}

func (s *StripeService) ConstructWebhookEvent(payload []byte, signature string) (*StripeWebhookEvent, error) {
	if !s.Enabled() {
		return nil, ErrStripeNotEnabled
	}

	event, err := stripewebhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf(errorWrapFormat, ErrStripeInvalidWebhook, err)
	}

	result := &StripeWebhookEvent{
		Type: string(event.Type),
	}

	if !strings.HasPrefix(string(event.Type), "checkout.session.") {
		return result, nil
	}

	var checkoutSession stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &checkoutSession); err != nil {
		return nil, fmt.Errorf(errorWrapFormat, ErrStripeInvalidWebhook, err)
	}

	result.CheckoutSession = *mapStripeCheckoutSession(&checkoutSession)
	return result, nil
}

func mapStripeCheckoutSession(session *stripe.CheckoutSession) *StripeCheckoutSession {
	if session == nil {
		return nil
	}

	var expiresAt *time.Time
	if session.ExpiresAt > 0 {
		timestamp := time.Unix(session.ExpiresAt, 0).UTC()
		expiresAt = &timestamp
	}

	metadata := make(map[string]string, len(session.Metadata))
	for key, value := range session.Metadata {
		metadata[key] = value
	}

	paymentIntentID := ""
	if session.PaymentIntent != nil {
		paymentIntentID = session.PaymentIntent.ID
	}

	return &StripeCheckoutSession{
		ID:              session.ID,
		URL:             session.URL,
		Status:          string(session.Status),
		PaymentStatus:   string(session.PaymentStatus),
		PaymentIntentID: paymentIntentID,
		ExpiresAt:       expiresAt,
		Metadata:        metadata,
	}
}
