package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// LogActivity logs user activity
func LogActivity(c *gin.Context) {
	var input struct {
		ActivityType string `json:"activityType" binding:"required"`
		Description  string `json:"description"`
		Device       string `json:"device"`
		Location     string `json:"location"`
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

	activity := models.ActivityLog{
		UserID:       userObj.ID,
		SessionID:    uuid.New().String(),
		ActivityType: input.ActivityType,
		Description:  input.Description,
		IPAddress:    c.ClientIP(),
		Device:       input.Device,
		Location:     input.Location,
	}

	if err := config.DB.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log activity"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Activity logged",
		"activity": activity,
	})
}

// GetActivityLogs gets user activity logs
func GetActivityLogs(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var activities []models.ActivityLog
	query := config.DB.Preload("User").Order("created_at desc")

	// Role-based filtering
	if userObj.Role == "citizen" {
		query = query.Where("user_id = ?", userObj.ID)
	} else if userObj.Role == "officer" && userObj.UnitID != nil {
		// Officers see activities from their unit
		query = query.Joins("JOIN users ON users.id = activity_logs.user_id").
			Where("users.unit_id = ?", userObj.UnitID)
	} else if userObj.Role == "unit_admin" && userObj.UnitID != nil {
		// Admins see activities from their unit
		query = query.Joins("JOIN users ON users.id = activity_logs.user_id").
			Where("users.unit_id = ?", userObj.UnitID)
	}
	// Super admin sees all

	if err := query.Limit(100).Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
	})
}

// GetAuditLogs gets audit logs (admin only)
func GetAuditLogs(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view audit logs"})
		return
	}

	var logs []models.AuditLog
	query := config.DB.Preload("User").Preload("Unit").Order("created_at desc")

	// Filter by unit for unit admins
	if userObj.Role == "unit_admin" && userObj.UnitID != nil {
		query = query.Where("unit_id = ? OR unit_id IS NULL", userObj.UnitID)
	}

	if err := query.Limit(200).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auditLogs": logs,
	})
}

// CreateAuditLog creates an audit log entry
func CreateAuditLog(c *gin.Context) {
	var input struct {
		Action   string `json:"action" binding:"required"`
		Resource string `json:"resource" binding:"required"`
		Details  string `json:"details"`
		Severity string `json:"severity"`
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

	auditLog := models.AuditLog{
		UserID:     &userObj.ID,
		UnitID:     userObj.UnitID,
		Action:     input.Action,
		Resource:   input.Resource,
		Details:    input.Details,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		Status:     "success",
		Severity:   input.Severity,
	}

	if auditLog.Severity == "" {
		auditLog.Severity = "info"
	}

	if err := config.DB.Create(&auditLog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create audit log"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Audit log created",
		"auditLog": auditLog,
	})
}

// GetSystemHealth gets system health metrics (admin only)
func GetSystemHealth(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view system health"})
		return
	}

	var health models.SystemHealth
	// Get latest health record
	if err := config.DB.Order("created_at desc").First(&health).Error; err != nil {
		// Create default if none exists
		health = models.SystemHealth{
			CPUUsage:       0,
			MemoryUsage:    0,
			DiskUsage:      0,
			ActiveUsers:    0,
			TotalRequests:  0,
			ResponseTime:   0,
			DatabaseStatus: "healthy",
			ServerStatus:   "healthy",
			Uptime:         0,
			LastCheck:      time.Now(),
		}
	}

	// Get active users count
	var activeCount int64
	config.DB.Model(&models.User{}).Where("status = ?", "active").Count(&activeCount)

	// Get total requests (from audit logs)
	var totalRequests int64
	config.DB.Model(&models.AuditLog{}).Count(&totalRequests)

	// Get response time (average from last 100 logs)
	var avgResponse float64
	config.DB.Model(&models.AuditLog{}).
		Select("COALESCE(AVG(CAST(details as float)), 0)").
		Scan(&avgResponse)

	health.ActiveUsers = int(activeCount)
	health.TotalRequests = int(totalRequests)
	health.ResponseTime = avgResponse
	health.LastCheck = time.Now()

	c.JSON(http.StatusOK, gin.H{
		"health": health,
	})
}

// UpdateSystemHealth updates system health metrics (system only)
func UpdateSystemHealth(c *gin.Context) {
	var input struct {
		CPUUsage       float64 `json:"cpuUsage"`
		MemoryUsage    float64 `json:"memoryUsage"`
		DiskUsage      float64 `json:"diskUsage"`
		ResponseTime   float64 `json:"responseTime"`
		DatabaseStatus string  `json:"databaseStatus"`
		ServerStatus   string  `json:"serverStatus"`
		Uptime         int64   `json:"uptime"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	health := models.SystemHealth{
		CPUUsage:       input.CPUUsage,
		MemoryUsage:    input.MemoryUsage,
		DiskUsage:      input.DiskUsage,
		ResponseTime:   input.ResponseTime,
		DatabaseStatus: input.DatabaseStatus,
		ServerStatus:   input.ServerStatus,
		Uptime:         input.Uptime,
		LastCheck:      time.Now(),
	}

	if health.DatabaseStatus == "" {
		health.DatabaseStatus = "healthy"
	}
	if health.ServerStatus == "" {
		health.ServerStatus = "healthy"
	}

	if err := config.DB.Create(&health).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update health"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Health updated",
		"health":  health,
	})
}

// GetNotificationLogs gets notification logs (admin only)
func GetNotificationLogs(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can view notification logs"})
		return
	}

	var logs []models.NotificationLog
	query := config.DB.Preload("User").Order("created_at desc")

	if userObj.Role == "unit_admin" && userObj.UnitID != nil {
		query = query.Joins("JOIN users ON users.id = notification_logs.user_id").
			Where("users.unit_id = ?", userObj.UnitID)
	}

	if err := query.Limit(100).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notification logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notificationLogs": logs,
	})
}