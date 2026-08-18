package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// RegisterPushToken registers a push notification token
func RegisterPushToken(c *gin.Context) {
	var input struct {
		UserID string `json:"userId" binding:"required"`
		Token  string `json:"token" binding:"required"`
		Device string `json:"device"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if token already exists
	var existing models.PushToken
	if err := config.DB.Where("user_id = ? AND token = ?", userID, input.Token).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Token already registered",
		})
		return
	}

	pushToken := models.PushToken{
		UserID: userID,
		Token:  input.Token,
		Device: input.Device,
		Active: true,
	}

	if err := config.DB.Create(&pushToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Push token registered successfully",
	})
}

// SendPushNotification sends a push notification
func SendPushNotification(c *gin.Context) {
	var input struct {
		UserID  string `json:"userId" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Body    string `json:"body" binding:"required"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user's push tokens
	var tokens []models.PushToken
	if err := config.DB.Where("user_id = ? AND active = ?", userID, true).Find(&tokens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get push tokens"})
		return
	}

	if len(tokens) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No push tokens found for user"})
		return
	}

	// Here you would integrate with a push notification service like Firebase Cloud Messaging
	// For now, just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Push notification sent successfully",
		"tokens":  len(tokens),
	})
}

// UnregisterPushToken unregisters a push token
func UnregisterPushToken(c *gin.Context) {
	var input struct {
		UserID string `json:"userId" binding:"required"`
		Token  string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := config.DB.Where("user_id = ? AND token = ?", userID, input.Token).Delete(&models.PushToken{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unregister token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Push token unregistered successfully",
	})
}

// GetUserPushTokens gets all push tokens for a user
func GetUserPushTokens(c *gin.Context) {
	userID := c.Param("userId")
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var tokens []models.PushToken
	if err := config.DB.Where("user_id = ?", id).Find(&tokens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch push tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
	})
}