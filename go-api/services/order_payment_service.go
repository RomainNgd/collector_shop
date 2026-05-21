package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"poc-gin/models"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	stripeEventCheckoutCompleted           = "checkout.session.completed"
	stripeEventCheckoutExpired             = "checkout.session.expired"
	stripeEventCheckoutAsyncPaymentFailed  = "checkout.session.async_payment_failed"
	stripeEventCheckoutAsyncPaymentSuccess = "checkout.session.async_payment_succeeded"
)

var (
	ErrOrderPaymentNotAvailable = errors.New("order payment is not available")
	ErrOrderCheckoutURLMissing  = errors.New("stripe checkout session url is missing")
	ErrCheckoutReturnURLInvalid = errors.New("checkout return url is invalid")
)

type OrderCheckoutSessionResult struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
	Reused    bool   `json:"reused"`
}

type OrderPaymentServiceInterface interface {
	CreateStripeCheckoutSession(ctx context.Context, actorID, orderID uint, actorRole, successURL, cancelURL string) (*OrderCheckoutSessionResult, error)
	HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error
}

type OrderPaymentService struct {
	db                   *gorm.DB
	stripe               StripeServiceInterface
	orderService         OrderServiceInterface
	allowedReturnOrigins []string
}

func NewOrderPaymentService(db *gorm.DB, stripe StripeServiceInterface, orderService OrderServiceInterface, allowedReturnOrigins ...[]string) *OrderPaymentService {
	origins := []string(nil)
	if len(allowedReturnOrigins) > 0 {
		origins = normalizeAllowedOrigins(allowedReturnOrigins[0])
	}

	return &OrderPaymentService{
		db:                   db,
		stripe:               stripe,
		orderService:         orderService,
		allowedReturnOrigins: origins,
	}
}

func (s *OrderPaymentService) CreateStripeCheckoutSession(ctx context.Context, actorID, orderID uint, actorRole, successURL, cancelURL string) (*OrderCheckoutSessionResult, error) {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	if s.stripe == nil || !s.stripe.Enabled() {
		return nil, ErrStripeNotEnabled
	}

	if !s.checkoutReturnURLAllowed(successURL) || !s.checkoutReturnURLAllowed(cancelURL) {
		return nil, ErrCheckoutReturnURLInvalid
	}

	order, err := s.orderService.GetOrderByID(ctx, actorID, orderID, actorRole)
	if err != nil {
		return nil, err
	}

	if order.Status != models.OrderStatusAwaitingPayment {
		return nil, ErrOrderPaymentNotAvailable
	}

	if order.PaymentStatus == models.OrderPaymentStatusPaid {
		return nil, ErrOrderPaymentNotAvailable
	}

	if order.StripeCheckoutSessionID != "" {
		existingSession, err := s.stripe.GetCheckoutSession(ctx, order.StripeCheckoutSessionID)
		switch {
		case err == nil:
			if _, err := s.applyCheckoutSessionState(ctx, existingSession, ""); err != nil {
				return nil, err
			}

			refreshedOrder, refreshErr := s.orderService.GetOrderByID(ctx, actorID, orderID, actorRole)
			if refreshErr != nil {
				return nil, refreshErr
			}

			if refreshedOrder.Status != models.OrderStatusAwaitingPayment {
				return nil, ErrOrderPaymentNotAvailable
			}

			if existingSession.Status == "open" && existingSession.PaymentStatus != "paid" {
				if existingSession.URL == "" {
					return nil, ErrOrderCheckoutURLMissing
				}

				return &OrderCheckoutSessionResult{
					SessionID: existingSession.ID,
					URL:       existingSession.URL,
					Reused:    true,
				}, nil
			}
		case errors.Is(err, ErrStripeSessionNotFound):
		default:
			return nil, err
		}
	}

	lineItems := make([]StripeCheckoutLineItem, 0, len(order.Items))
	for _, item := range order.Items {
		lineItems = append(lineItems, StripeCheckoutLineItem{
			Name:        item.ProductName,
			Description: item.ProductDescription,
			Quantity:    int64(item.Quantity),
			Currency:    order.Currency,
			UnitAmount:  currencyToMinorUnits(item.UnitPrice),
		})
	}

	session, err := s.stripe.CreateCheckoutSession(ctx, StripeCheckoutSessionInput{
		SuccessURL:        successURL,
		CancelURL:         cancelURL,
		ClientReferenceID: strconv.FormatUint(uint64(order.ID), 10),
		Metadata: map[string]string{
			"order_id": strconv.FormatUint(uint64(order.ID), 10),
		},
		LineItems: lineItems,
	})
	if err != nil {
		return nil, err
	}

	if session.URL == "" {
		return nil, ErrOrderCheckoutURLMissing
	}

	if _, err := s.applyCheckoutSessionState(ctx, session, ""); err != nil {
		return nil, err
	}

	return &OrderCheckoutSessionResult{
		SessionID: session.ID,
		URL:       session.URL,
		Reused:    false,
	}, nil
}

func (s *OrderPaymentService) HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	ctx, cancel := withDBTimeout(ctx)
	defer cancel()

	if s.stripe == nil || !s.stripe.Enabled() {
		return ErrStripeNotEnabled
	}

	event, err := s.stripe.ConstructWebhookEvent(payload, signature)
	if err != nil {
		return err
	}

	switch event.Type {
	case stripeEventCheckoutCompleted,
		stripeEventCheckoutExpired,
		stripeEventCheckoutAsyncPaymentFailed,
		stripeEventCheckoutAsyncPaymentSuccess:
		_, err := s.applyCheckoutSessionState(ctx, &event.CheckoutSession, event.Type)
		return err
	default:
		return nil
	}
}

func (s *OrderPaymentService) applyCheckoutSessionState(ctx context.Context, session *StripeCheckoutSession, eventType string) (*models.Order, error) {
	if session == nil {
		return nil, fmt.Errorf("stripe checkout session is required")
	}

	var updatedOrder models.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := s.findOrderByCheckoutSession(tx, session)
		if err != nil {
			return err
		}

		updates := map[string]any{
			"payment_provider":               models.PaymentProviderStripe,
			"payment_status":                 mapStripePaymentStatus(session.Status, session.PaymentStatus, eventType),
			"stripe_checkout_session_id":     session.ID,
			"stripe_checkout_session_status": session.Status,
			"stripe_payment_intent_id":       session.PaymentIntentID,
			"stripe_checkout_expires_at":     session.ExpiresAt,
		}

		if session.PaymentStatus == "paid" && order.PaidAt == nil {
			now := time.Now().UTC()
			updates["paid_at"] = &now
		}

		if session.PaymentStatus == "paid" && order.Status == models.OrderStatusAwaitingPayment {
			updates["status"] = models.OrderStatusPreparation
		}

		if err := tx.Model(order).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to persist stripe checkout state: %w", err)
		}

		reloadedOrder, err := preloadOrder(tx.WithContext(ctx), order.ID)
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

func (s *OrderPaymentService) findOrderByCheckoutSession(tx *gorm.DB, session *StripeCheckoutSession) (*models.Order, error) {
	var order models.Order
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&models.Order{})

	if orderIDValue, exists := session.Metadata["order_id"]; exists && orderIDValue != "" {
		orderID, err := strconv.ParseUint(orderIDValue, 10, 32)
		if err == nil && orderID > 0 {
			if err := query.First(&order, uint(orderID)).Error; err == nil {
				return &order, nil
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
		}
	}

	if session.ID != "" {
		if err := query.Where("stripe_checkout_session_id = ?", session.ID).First(&order).Error; err == nil {
			return &order, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	return nil, gorm.ErrRecordNotFound
}

func mapStripePaymentStatus(sessionStatus, paymentStatus, eventType string) string {
	switch {
	case paymentStatus == "paid":
		return models.OrderPaymentStatusPaid
	case paymentStatus == "no_payment_required":
		return models.OrderPaymentStatusNoPaymentNeeded
	case eventType == stripeEventCheckoutAsyncPaymentFailed:
		return models.OrderPaymentStatusFailed
	case eventType == stripeEventCheckoutExpired || sessionStatus == "expired":
		return models.OrderPaymentStatusExpired
	case sessionStatus == "open":
		return models.OrderPaymentStatusCheckoutOpen
	default:
		return models.OrderPaymentStatusPending
	}
}

func currencyToMinorUnits(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

func normalizeAllowedOrigins(origins []string) []string {
	normalized := make([]string, 0, len(origins))
	for _, origin := range origins {
		trimmed := strings.TrimRight(strings.TrimSpace(origin), "/")
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	return normalized
}

func (s *OrderPaymentService) checkoutReturnURLAllowed(rawURL string) bool {
	if len(s.allowedReturnOrigins) == 0 {
		return true
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	origin := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
	for _, allowedOrigin := range s.allowedReturnOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}
