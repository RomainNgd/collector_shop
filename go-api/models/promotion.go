package models

import "gorm.io/gorm"

const (
	PromotionTypePercentage = "percentage"
	PromotionTypeFixed      = "fixed"
)

type Promotion struct {
	gorm.Model
	Name         string    `gorm:"not null;size:120" json:"name"`
	Description  string    `gorm:"size:1000" json:"description"`
	Type         string    `gorm:"not null;size:20;index" json:"type"`
	Value        float64   `gorm:"not null" json:"value"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	AppliesToAll bool      `gorm:"not null;default:false" json:"applies_to_all"`
	SellerID     *uint     `gorm:"index" json:"seller_id,omitempty"`
	Products     []Product `gorm:"many2many:product_promotions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	ProductIDs   []uint    `gorm:"-" json:"product_ids"`
	ProductCount int       `gorm:"-" json:"product_count"`
}

type AppliedPromotion struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Value          float64 `json:"value"`
	DiscountAmount float64 `json:"discount_amount"`
	AppliesToAll   bool    `json:"applies_to_all"`
}
