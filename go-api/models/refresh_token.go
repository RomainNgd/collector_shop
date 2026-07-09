package models

import "time"

type RefreshToken struct {
	ID           uint `gorm:"primaryKey"`
	CreatedAt    time.Time
	UserID       uint   `gorm:"not null;index"`
	TokenHash    string `gorm:"not null;uniqueIndex"`
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	ReplacedByID *uint
}
