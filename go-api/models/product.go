package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name             string            `gorm:"not null;size:120" json:"name"`
	Description      string            `gorm:"not null;size:1000" json:"description"`
	Image            string            `gorm:"size:255" json:"image"`
	Price            float64           `gorm:"not null" json:"price"`
	Stock            int               `gorm:"not null;default:1" json:"stock"`
	IsActive         bool              `gorm:"not null;default:true;index" json:"is_active"`
	SellerID         *uint             `gorm:"index" json:"seller_id,omitempty"`
	Seller           User              `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	SellerEmail      string            `gorm:"-" json:"seller_email,omitempty"`
	PromotionType    string            `gorm:"size:20" json:"promotion_type,omitempty"`
	PromotionValue   float64           `json:"promotion_value,omitempty"`
	PromotionActive  bool              `gorm:"not null;default:false" json:"promotion_active"`
	CategoryID       uint              `gorm:"not null;index" json:"category_id"`
	Category         Category          `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"category"`
	Promotions       []Promotion       `gorm:"many2many:product_promotions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	EffectivePrice   float64           `gorm:"-" json:"effective_price"`
	AppliedPromotion *AppliedPromotion `gorm:"-" json:"applied_promotion,omitempty"`
}
