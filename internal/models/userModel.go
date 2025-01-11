package models

import (
	"time"
	"gorm.io/gorm"
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
	CreatedAt	time.Time
}