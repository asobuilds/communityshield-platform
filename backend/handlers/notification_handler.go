package handlers

import (
	"net/http"
	"security-solution/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateNotification saves a notification for a user
func CreateNotification(userID uuid.UUID, caseID uuid.UUID, title, message string) error {
	notif := models.Notification{
		UserID:  userID,
		CaseID:  caseID,
		Title:   title,
		Message: message,
		Read:    false,
	}
	return DB.Create(&notif).Error
}

// GetUserNotifications returns all notifications for a user
func GetUserNotifications(c *gin.Context) {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}
	var notifs []models.Notification
	if err := DB.Where("user_id = ?", userID).Order("created_at desc").Find(&notifs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"notifications": notifs})
}

// MarkNotificationRead marks a single notification as read
func MarkNotificationRead(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}
	var notif models.Notification
	if err := DB.First(&notif, "id = ?", parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}
	notif.Read = true
	if err := DB.Save(&notif).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Marked as read"})
}

// MarkAllNotificationsRead marks all notifications for a user as read
func MarkAllNotificationsRead(c *gin.Context) {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}
	if err := DB.Model(&models.Notification{}).Where("user_id = ?", userID).Update("read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}