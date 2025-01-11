package models

import (
	"gorm.io/gorm"
)

// Ad Category model
type Category struct {
	gorm.Model
	Name string `gorm:"unique;not null"`
	Ads  []Ad   `gorm:"foreignKey:CategoryID"` 
}