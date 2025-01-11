package models

import (
	"time"
	"gorm.io/gorm"
)

// Ad Favorite model
type Favorite struct {
	gorm.Model
	UserID uint `gorm:"not null"` 
	AdID   uint `gorm:"not null"`
	Ad     Ad   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt time.Time
}