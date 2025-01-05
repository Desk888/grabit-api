package models

import (
	"time"
	"gorm.io/gorm"
)

// User model
type User struct {
	gorm.Model
	ID           uint   `gorm:"primaryKey"`
	FirstName    string 
	LastName     string
	Username     string `gorm:"unique"`
	Email        string `gorm:"unique"`
	PasswordHash string
	PhoneNumber  string
	CreatedAt    time.Time
}
