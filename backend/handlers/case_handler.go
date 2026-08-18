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

// CreateCase handles creating a new case (Now with Async AI)
func CreateCase(c *gin.Context) {
	var input struct {
		UnitID      string  `json:"unitId"`
		Title       string  `json:"title" binding:"required"`
		Description string  `json:"description" binding:"required"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Location    string  `json:"location"`
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

	caseObj := models.Case{
		Title:       input.Title,
		Description: input.Description,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		Location:    input.Location,
		Status:      "pending",
		Priority:    "medium",
		IsPublic:    true,
		ReportedBy:  userObj.ID,
	}

	if input.UnitID != "" {
		unitID, _ := uuid.Parse(input.UnitID)
		caseObj.UnitID = unitID
	}

	if err := config.DB.Create(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create case"})
		return
	}

	// ---- ASYNCHRONOUS AI PROCESSING ----
	// This runs in a separate goroutine so it doesn't block the response
	go func() {
		// Use a new DB connection for the background task
		db := config.DB
		aiResult := services.GenerateLocationRisk(db, input.Latitude, input.Longitude, input.Location)
		
		// Update the case with AI analysis
		config.DB.Model(&models.Case{}).
			Where("id = ?", caseObj.ID).
			Update("description", caseObj.Description+"\n\n🤖 AI Analysis: "+aiResult)
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Case reported successfully! AI analysis is being processed.",
		"case":    caseObj,
	})
}