package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	OrderStatusAwaitingPayment = "awaiting_payment"
	OrderStatusPreparation     = "preparation"
	OrderStatusShipping        = "shipping"
	OrderStatusDelivered       = "delivered"
	OrderStatusCancelled       = "cancelled"

	PaymentProviderStripe = "stripe"

	OrderPaymentStatusPending         = "pending"
	OrderPaymentStatusCheckoutOpen    = "checkout_open"
	OrderPaymentStatusPaid            = "paid"
	OrderPaymentStatusFailed          = "failed"
	OrderPaymentStatusExpired         = "expired"
	OrderPaymentStatusNoPaymentNeeded = "no_payment_needed"
)

type Order struct {
	gorm.Model
	UserID                      uint        `gorm:"not null;index" json:"user_id"`
	User                        User        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Status                      string      `gorm:"not null;size:40;index;default:'awaiting_payment'" json:"status"`
	Currency                    string      `gorm:"not null;size:3;default:'EUR'" json:"currency"`
	ItemCount                   int         `gorm:"not null" json:"item_count"`
	Subtotal                    float64     `gorm:"not null" json:"subtotal"`
	DiscountTotal               float64     `gorm:"not null" json:"discount_total"`
	Total                       float64     `gorm:"not null" json:"total"`
	PaymentProvider             string      `gorm:"size:40" json:"payment_provider,omitempty"`
	PaymentStatus               string      `gorm:"size:40;index;default:'pending'" json:"payment_status,omitempty"`
	PaidAt                      *time.Time  `json:"paid_at,omitempty"`
	StripeCheckoutSessionID     string      `gorm:"size:255;index" json:"-"`
	StripeCheckoutSessionStatus string      `gorm:"size:40" json:"-"`
	StripePaymentIntentID       string      `gorm:"size:255" json:"-"`
	StripeCheckoutExpiresAt     *time.Time  `json:"stripe_checkout_expires_at,omitempty"`
	Items                       []OrderItem `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"items"`
}

type OrderItem struct {
	gorm.Model
	OrderID             uint    `gorm:"not null;index" json:"order_id"`
	ProductID           uint    `gorm:"not null;index" json:"product_id"`
	SellerID            uint    `gorm:"index" json:"seller_id"`
	SellerEmail         string  `gorm:"size:255" json:"seller_email,omitempty"`
	ProductName         string  `gorm:"not null;size:120" json:"product_name"`
	ProductDescription  string  `gorm:"not null;size:1000" json:"product_description"`
	ProductImage        string  `gorm:"size:255" json:"product_image"`
	CategoryName        string  `gorm:"size:120" json:"category_name"`
	Quantity            int     `gorm:"not null" json:"quantity"`
	UnitBasePrice       float64 `gorm:"not null" json:"unit_base_price"`
	UnitPrice           float64 `gorm:"not null" json:"unit_price"`
	UnitDiscount        float64 `gorm:"not null" json:"unit_discount"`
	LineBaseTotal       float64 `gorm:"not null" json:"line_base_total"`
	LineDiscountTotal   float64 `gorm:"not null" json:"line_discount_total"`
	LineTotal           float64 `gorm:"not null" json:"line_total"`
	PromotionID         *uint   `json:"promotion_id,omitempty"`
	PromotionName       string  `gorm:"size:120" json:"promotion_name,omitempty"`
	PromotionType       string  `gorm:"size:20" json:"promotion_type,omitempty"`
	PromotionValue      float64 `json:"promotion_value,omitempty"`
	PromotionAppliesAll bool    `json:"promotion_applies_to_all,omitempty"`
}

func IsValidOrderStatus(status string) bool {
	switch status {
	case OrderStatusAwaitingPayment, OrderStatusPreparation, OrderStatusShipping, OrderStatusDelivered, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func IsValidOrderPaymentStatus(status string) bool {
	switch status {
	case OrderPaymentStatusPending,
		OrderPaymentStatusCheckoutOpen,
		OrderPaymentStatusPaid,
		OrderPaymentStatusFailed,
		OrderPaymentStatusExpired,
		OrderPaymentStatusNoPaymentNeeded:
		return true
	default:
		return false
	}
}
