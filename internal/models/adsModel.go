package models

import (
	"time"
	"gorm.io/gorm"
)

// ENums for Conditions
const (
	ConditionUsedFair         = "Used - Fair"
	ConditionUsedGood         = "Used - Good"
	ConditionUsedExcellent    = "Used - Excellent"
	ConditionBrandNewUnboxed  = "Brand New - Unboxed"
	ConditionBrandNewSealed   = "Brand New - Used"
)

// Ads model
type Ad struct {
	gorm.Model
	Title        string    `gorm:"not null"`
	Description  string    `gorm:"type:text"`
	CategoryID   uint      `gorm:"not null"` 
	UserID       uint      `gorm:"not null"` 
	Condition    string    `gorm:"not null;check:condition IN ('Used - Fair','Used - Good','Used - Excellent','Brand New - Unboxed','Brand New - Sealed')"`
	City         string
	Postcode     string
	PhoneNumber  string
	EmailAddress string
	CreatedAt    time.Time
	Category     Category `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` 
	User         User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	
	// @Danny See how to best implement S3 storage for the ads images
}