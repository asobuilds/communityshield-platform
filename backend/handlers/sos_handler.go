package handlers

import (
	"fmt"
	"log"
	"net/http"
	"security-solution/models"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SendSOS(c *gin.Context) {
	log.Println("🔴 SOS request received")

	var input struct {
		Latitude    float64 `json:"latitude" binding:"required"`
		Longitude   float64 `json:"longitude" binding:"required"`
		Description string  `json:"description"`
		UserID      string  `json:"userId" binding:"required"`
		UnitID      string  `json:"unitId"`
		Anonymous   bool    `json:"anonymous"`
		DeviceID    string  `json:"deviceId"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		log.Println("❌ SOS bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.DeviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "DeviceID is required"})
		return
	}

	var userID *uuid.UUID
	if input.Anonymous {
		userID = nil
	} else {
		if input.UserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "UserID is required when not anonymous"})
			return
		}
		parsed, err := uuid.Parse(input.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UserID"})
			return
		}
		userID = &parsed
	}

	var unitID *uuid.UUID
	if input.UnitID != "" {
		parsed, err := uuid.Parse(input.UnitID)
		if err == nil {
			unitID = &parsed
		}
	}

	// Anti-abuse: cooldown and daily limit
	var count int64
	now := time.Now()
	cooldownThreshold := now.Add(-5 * time.Minute)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if err := DB.Model(&models.SOSAlert{}).
		Where("device_id = ? AND created_at > ?", input.DeviceID, cooldownThreshold).
		Count(&count).Error; err == nil && count > 0 {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":       "Please wait 5 minutes before sending another SOS.",
			"retry_after": 300,
		})
		return
	}

	if input.Anonymous {
		if err := DB.Model(&models.SOSAlert{}).
			Where("device_id = ? AND anonymous = true AND created_at >= ?", input.DeviceID, startOfDay).
			Count(&count).Error; err == nil && count >= 5 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Daily limit (5 anonymous SOS) reached. Please try again tomorrow or login to continue.",
			})
			return
		}
	} else {
		if err := DB.Model(&models.SOSAlert{}).
			Where("device_id = ? AND anonymous = false AND created_at >= ?", input.DeviceID, startOfDay).
			Count(&count).Error; err == nil && count >= 10 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Daily limit (10 SOS) reached. Please try again tomorrow.",
			})
			return
		}
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	sos := models.SOSAlert{
		UserID:      userID,
		UnitID:      unitID,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		Description: input.Description,
		Status:      "pending",
		Anonymous:   input.Anonymous,
		DeviceID:    input.DeviceID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}

	if err := DB.Create(&sos).Error; err != nil {
		log.Println("❌ Failed to save SOS:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send SOS"})
		return
	}

	log.Println("✅ SOS saved successfully")

	// 🔥 Send SMS to unit admin if unit is assigned
	if sos.UnitID != nil {
		var unit models.Unit
		if err := DB.First(&unit, "id = ?", sos.UnitID).Error; err == nil {
			var admin models.User
			if err := DB.Where("unit_id = ? AND role = ?", sos.UnitID, "unit_admin").First(&admin).Error; err == nil {
				if admin.Phone != "" {
					if err := SendSOSAlertSMS(admin.Phone, unit.Name, sos.Description); err != nil {
						fmt.Printf("Failed to send SOS SMS: %v\n", err)
					} else {
						fmt.Println("✅ SOS SMS sent to", admin.Phone)
					}
				}
				// 🔥 Push notification
				if admin.ID != uuid.Nil {
					title := "🚨 SOS Alert"
					message := fmt.Sprintf("SOS received from %s", unit.Name)
					SendPushNotification(admin.ID, title, message, "/vite.svg", "/admin")
				}
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "SOS alert sent successfully",
		"sos":     sos,
	})
}

func GetSOSHistory(c *gin.Context) {
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

	var sosList []models.SOSAlert
	if err := DB.Where("user_id = ?", userID).Order("created_at desc").Find(&sosList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SOS history"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sos": sosList})
}

func GetAllSOSAlerts(c *gin.Context) {
	var sosList []models.SOSAlert
	if err := DB.Order("created_at desc").Find(&sosList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SOS alerts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sos": sosList})
}

func UpdateSOSStatus(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SOS ID"})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sos models.SOSAlert
	if err := DB.First(&sos, "id = ?", parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SOS alert not found"})
		return
	}

	sos.Status = input.Status
	if err := DB.Save(&sos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SOS status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SOS status updated",
		"sos":     sos,
	})
}