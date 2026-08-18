package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

func SendNotification(c *gin.Context) {
	var input struct {
		UserID  string `json:"userId" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Message string `json:"message" binding:"required"`
		Type    string `json:"type"`
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

	notification := models.Notification{
		UserID:  userID,
		Title:   input.Title,
		Message: input.Message,
		Type:    input.Type,
		Status:  "unread",
	}

	if err := config.DB.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Notification sent successfully",
		"notification": notification,
	})
}

func GetUserNotifications(c *gin.Context) {
	userID := c.Param("userId")
	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var notifications []models.Notification
	if err := config.DB.Where("user_id = ?", id).Order("created_at desc").Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
	})
}

func MarkNotificationAsRead(c *gin.Context) {
	id := c.Param("id")
	notificationID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	var notification models.Notification
	if err := config.DB.First(&notification, "id = ?", notificationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	notification.Status = "read"
	if err := config.DB.Save(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

func DeleteNotification(c *gin.Context) {
	id := c.Param("id")
	notificationID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := config.DB.Delete(&models.Notification{}, "id = ?", notificationID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification deleted successfully",
	})
}