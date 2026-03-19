package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name        string   `gorm:"not null;size:120" json:"name"`
	Description string   `gorm:"not null;size:1000" json:"description"`
	Image       string   `gorm:"size:255" json:"image"`
	Price       float64  `gorm:"not null" json:"price"`
	CategoryID  uint     `gorm:"not null;index" json:"category_id"`
	Category    Category `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"category"`
}
