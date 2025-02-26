package controllers

import (
	"net/http"
	"strconv"
	"github.com/Desk888/api/internal/initializers"
	"github.com/Desk888/api/internal/models"
	"github.com/gin-gonic/gin"
)

func ViewProfile(c *gin.Context) {
	userIDParam := c.Param("userID")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
		return
	}

	var user models.User

	if err := initializers.DB.Select("profile_picture_url, first_name, last_name, username, city, country, bio, created_at").
    Where("id = ?", userID).
    First(&user).Error; err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
    return
}

	c.JSON(http.StatusOK, gin.H{
		"profile_image": user.ProfilePictureURL,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"username":      user.Username,
		"city":          user.City,
		"country":       user.Country,
		"bio":           user.Bio,
		"created_at":    user.CreatedAt,
	})
}


func EditProfile(c *gin.Context) {
	userIDParam := c.Param("userID")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
		return
	}

	var user models.User

	if err := initializers.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input struct {
		ProfilePictureURL string `json:"profile_picture_url"`
		FirstName         string `json:"first_name"`
		LastName          string `json:"last_name"`
		City              string `json:"city"`
		Country           string `json:"country"`
		Bio               string `json:"bio"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	updateData := map[string]interface{}{
		"profile_picture_url": input.ProfilePictureURL,
		"first_name":          input.FirstName,
		"last_name":           input.LastName,
		"city":                input.City,
		"country":             input.Country,
		"bio":                 input.Bio,
	}

	if err := initializers.DB.Model(&user).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}


func DeleteProfile(c *gin.Context) {
	userIDParam := c.Param("userID")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID"})
		return
	}

	var user models.User

	if err := initializers.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := initializers.DB.Unscoped().Delete(&user).Error; err != nil {
	    c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to permanently delete user"})
	    return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}
