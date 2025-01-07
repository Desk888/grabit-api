package initializers

import "github.com/Desk888/api/internal/models"

func MigrateTables() {
	// Migrate all models
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Ad{})
	DB.AutoMigrate(&models.Favorite{})
	
}
