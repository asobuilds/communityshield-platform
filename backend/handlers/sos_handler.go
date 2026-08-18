package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// SendSOSAlert sends an SOS alert
func SendSOSAlert(c *gin.Context) {
	var input struct {
		Latitude    float64 `json:"latitude" binding:"required"`
		Longitude   float64 `json:"longitude" binding:"required"`
		Description string  `json:"description"`
		UnitID      string  `json:"unitId"`
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

	sos := models.SOSAlert{
		UserID:      userObj.ID,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		Description: input.Description,
		Status:      "pending",
		Priority:    "high",
	}

	if input.UnitID != "" {
		unitID, err := uuid.Parse(input.UnitID)
		if err == nil {
			sos.UnitID = &unitID
		}
	}

	if err := config.DB.Create(&sos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send SOS alert"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "SOS alert sent successfully",
		"sos":     sos,
	})
}

// GetSOSAlerts gets all SOS alerts
func GetSOSAlerts(c *gin.Context) {
	var alerts []models.SOSAlert
	if err := config.DB.Preload("User").Preload("Unit").Order("created_at desc").Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SOS alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
	})
}

// GetSOSAlertByID gets a specific SOS alert
func GetSOSAlertByID(c *gin.Context) {
	id := c.Param("id")
	alertID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var alert models.SOSAlert
	if err := config.DB.Preload("User").Preload("Unit").First(&alert, "id = ?", alertID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SOS alert not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alert": alert,
	})
}

// UpdateSOSAlertStatus updates an SOS alert status
func UpdateSOSAlertStatus(c *gin.Context) {
	id := c.Param("id")
	alertID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var alert models.SOSAlert
	if err := config.DB.First(&alert, "id = ?", alertID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SOS alert not found"})
		return
	}

	alert.Status = input.Status
	if err := config.DB.Save(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update alert status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Alert status updated successfully",
		"alert":   alert,
	})
}

// GetUserSOSAlerts gets SOS alerts for a specific user
func GetUserSOSAlerts(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var alerts []models.SOSAlert
	if err := config.DB.Where("user_id = ?", userObj.ID).Order("created_at desc").Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SOS alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
	})
}