/* 
	* For images and media files such as profile pictures and ads images use S3 storage
	* Messaging Chat Table to be created later
*/

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

// User model
type User struct {
	gorm.Model
	ProfilePictureURL string `gorm:"size:255"` // URL to the profile picture in S3
	FirstName    string `gorm:"not null"`
	LastName     string	`gorm:"not null"`
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	PhoneNumber  string
	Ads          []Ad       `gorm:"foreignKey:UserID"`
	FavouriteAds []Favorite `gorm:"foreignKey:UserID"`
}

// Ad Category model
type Category struct {
	gorm.Model
	Name string `gorm:"unique;not null"`
	Ads  []Ad   `gorm:"foreignKey:CategoryID"` 
}

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

// Ad Favorite model
type Favorite struct {
	gorm.Model
	UserID uint `gorm:"not null"` 
	AdID   uint `gorm:"not null"`
	Ad     Ad   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User   User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt time.Time
}