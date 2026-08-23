package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
)

// RegisterDevice registers a device for push notifications
func RegisterDevice(c *gin.Context) {
	var input struct {
		DeviceToken string `json:"deviceToken" binding:"required"`
		DeviceType  string `json:"deviceType" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	// Check if device exists
	var existing models.PushToken
	if err := config.DB.Where("user_id = ? AND token = ?", userObj.ID, input.DeviceToken).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Device already registered"})
		return
	}

	pushToken := models.PushToken{
		UserID: userObj.ID,
		Token:  input.DeviceToken,
		Device: input.DeviceType,
		Active: true,
	}

	if err := config.DB.Create(&pushToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register device"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Device registered successfully",
	})
}

// UnregisterDevice unregisters a device
func UnregisterDevice(c *gin.Context) {
	var input struct {
		DeviceToken string `json:"deviceToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if err := config.DB.Where("user_id = ? AND token = ?", userObj.ID, input.DeviceToken).Delete(&models.PushToken{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unregister device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Device unregistered",
	})
}

// TestNotification sends test push notification
func TestNotification(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var tokens []models.PushToken
	if err := config.DB.Where("user_id = ? AND active = ?", userObj.ID, true).Find(&tokens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No push tokens found"})
		return
	}

	if len(tokens) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No devices registered"})
		return
	}

	// Create notification record
	for range tokens {
		notification := models.Notification{
			UserID:  userObj.ID,
			Title:   "🔔 Test Notification",
			Message: "Your device is connected to CommunityShield!",
			Type:    "test",
			Status:  "sent",
		}
		config.DB.Create(&notification)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test notifications sent",
		"count":   len(tokens),
	})
}
