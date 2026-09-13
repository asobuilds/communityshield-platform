package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

type LocationUpdateRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Accuracy  float64 `json:"accuracy"`
}

func UpdateMyLocation(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	userID, err := uuid.Parse(userIDValue.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user identity",
		})
		return
	}

	var request LocationUpdateRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "latitude and longitude are required",
		})
		return
	}

	if request.Latitude < -90 || request.Latitude > 90 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid latitude",
		})
		return
	}

	if request.Longitude < -180 || request.Longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid longitude",
		})
		return
	}

	if request.Accuracy < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid GPS accuracy",
		})
		return
	}

	location := models.UserLocation{
		UserID:     userID,
		Latitude:   request.Latitude,
		Longitude:  request.Longitude,
		Accuracy:   request.Accuracy,
		RecordedAt: time.Now(),
	}

	if err := config.DB.Create(&location).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save location",
		})
		return
	}

	arrivalDetected := false

	var cases []models.Case

	if err := config.DB.
		Where("assigned_to = ? AND status = ?", userID, "dispatched").
		Find(&cases).Error; err == nil {

		for _, caseRecord := range cases {
			reached, checkErr := services.CheckOfficerReachedCase(userID, caseRecord.ID)
			if checkErr == nil && reached {
				arrivalDetected = true
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "location updated",
		"location":        location,
		"arrivalDetected": arrivalDetected,
	})
}
