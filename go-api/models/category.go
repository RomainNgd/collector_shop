package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name        string `gorm:"not null;size:120;uniqueIndex" json:"name"`
	Description string `gorm:"not null;size:1000" json:"description"`
}
