package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// MobileAppConfig returns mobile app configuration
func MobileAppConfig(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	config := gin.H{
		"appName":    "CommunityShield",
		"version":    "1.0.0",
		"build":      "100",
		"features": gin.H{
			"sos":           true,
			"caseReporting": true,
			"unitMap":       true,
			"walkieTalkie":  true,
			"newsFeed":      true,
			"alerts":        true,
			"suspectTracking": userObj.Role != "citizen",
			"finance":       userObj.Role == "unit_admin" || userObj.Role == "super_admin",
			"adminPanel":    userObj.Role == "unit_admin" || userObj.Role == "super_admin",
		},
		"userRole":   userObj.Role,
		"unitId":     userObj.UnitID,
		"userId":     userObj.ID,
		"apiBaseUrl": "/api/v1",
	}

	c.JSON(http.StatusOK, gin.H{
		"config": config,
	})
}

// MobileDashboard returns mobile-specific dashboard data
func MobileDashboard(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var totalCases int64
	var pendingCases int64
	var sosCount int64
	var recentCases []models.Case

	config.DB.Model(&models.Case{}).Count(&totalCases)
	config.DB.Model(&models.Case{}).Where("status = ?", "pending").Count(&pendingCases)
	config.DB.Model(&models.SOSAlert{}).Where("user_id = ?", userObj.ID).Count(&sosCount)

	query := config.DB.Model(&models.Case{}).Order("created_at desc").Limit(5)
	if userObj.UnitID != nil {
		query = query.Where("unit_id = ? OR unit_id IS NULL", userObj.UnitID)
	}
	query.Find(&recentCases)

	c.JSON(http.StatusOK, gin.H{
		"totalCases":   totalCases,
		"pendingCases": pendingCases,
		"sosCount":     sosCount,
		"recentCases":  recentCases,
		"user": gin.H{
			"id":        userObj.ID,
			"firstName": userObj.FirstName,
			"lastName":  userObj.LastName,
			"email":     userObj.Email,
			"role":      userObj.Role,
		},
	})
}

// MobileNotifications returns mobile notifications
func MobileNotifications(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var notifications []models.Notification
	if err := config.DB.Where("user_id = ?", userObj.ID).Order("created_at desc").Limit(20).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	// Count unread
	var unreadCount int64
	config.DB.Model(&models.Notification{}).Where("user_id = ? AND status = ?", userObj.ID, "unread").Count(&unreadCount)

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"unreadCount":   unreadCount,
	})
}

// MobileMarkNotificationRead marks notification as read
func MobileMarkNotificationRead(c *gin.Context) {
	id := c.Param("id")
	notificationID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var notification models.Notification
	if err := config.DB.Where("id = ? AND user_id = ?", notificationID, userObj.ID).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	notification.Status = "read"
	config.DB.Save(&notification)

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

// MobileMarkAllRead marks all notifications as read
func MobileMarkAllRead(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	config.DB.Model(&models.Notification{}).Where("user_id = ? AND status = ?", userObj.ID, "unread").Update("status", "read")

	c.JSON(http.StatusOK, gin.H{
		"message": "All notifications marked as read",
	})
}

// MobileSync syncs mobile data
func MobileSync(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	lastSync := c.Query("lastSync")

	var cases []models.Case
	query := config.DB.Model(&models.Case{}).Order("updated_at desc")

	if userObj.UnitID != nil {
		query = query.Where("unit_id = ? OR unit_id IS NULL", userObj.UnitID)
	}

	if lastSync != "" {
		lastSyncTime, err := time.Parse(time.RFC3339, lastSync)
		if err == nil {
			query = query.Where("updated_at > ?", lastSyncTime)
		}
	}

	query.Limit(50).Find(&cases)

	// Get units if admin
	var units []models.SecurityUnit
	if userObj.Role == "unit_admin" || userObj.Role == "super_admin" {
		config.DB.Where("status = ?", "active").Find(&units)
	}

	// Get SOS alerts
	var sosAlerts []models.SOSAlert
	if userObj.Role != "citizen" {
		config.DB.Where("user_id = ? OR unit_id = ?", userObj.ID, userObj.UnitID).Order("created_at desc").Limit(20).Find(&sosAlerts)
	}

	c.JSON(http.StatusOK, gin.H{
		"cases":      cases,
		"units":      units,
		"sosAlerts":  sosAlerts,
		"syncTime":   time.Now().Format(time.RFC3339),
		"user": gin.H{
			"id":        userObj.ID,
			"firstName": userObj.FirstName,
			"lastName":  userObj.LastName,
			"email":     userObj.Email,
			"role":      userObj.Role,
		},
	})
}

// MobileCrashReport submits crash report
func MobileCrashReport(c *gin.Context) {
	var input struct {
		Message   string `json:"message" binding:"required"`
		Stack     string `json:"stack"`
		Device    string `json:"device"`
		OS        string `json:"os"`
		AppVersion string `json:"appVersion"`
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

	// Log crash
	auditLog := models.AuditLog{
		UserID:   &userObj.ID,
		Action:   "mobile_crash",
		Resource: "mobile_app",
		Details:  "Message: " + input.Message + "\nDevice: " + input.Device + "\nOS: " + input.OS + "\nVersion: " + input.AppVersion + "\nStack: " + input.Stack,
		Severity: "error",
	}
	config.DB.Create(&auditLog)

	c.JSON(http.StatusOK, gin.H{
		"message": "Crash report submitted",
	})
}