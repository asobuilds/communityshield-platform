package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// CreateCase handles creating a new case
func CreateCase(c *gin.Context) {
	var input struct {
		UnitID      string  `json:"unitId"`
		Title       string  `json:"title" binding:"required"`
		Description string  `json:"description" binding:"required"`
		IncidentDate string `json:"incidentDate"`
		Location    string  `json:"location"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Priority    string  `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	incidentDate := time.Now()
	if input.IncidentDate != "" {
		parsed, err := time.Parse(time.RFC3339, input.IncidentDate)
		if err == nil {
			incidentDate = parsed
		}
	}

	caseObj := models.Case{
		ReportedBy:   userObj.ID,
		Title:        input.Title,
		Description:  input.Description,
		IncidentDate: incidentDate,
		Location:     input.Location,
		Latitude:     input.Latitude,
		Longitude:    input.Longitude,
		Priority:     input.Priority,
		Status:       "pending",
		IsPublic:     true,
	}

	if input.UnitID != "" {
		unitID, err := uuid.Parse(input.UnitID)
		if err == nil {
			caseObj.UnitID = unitID
		}
	}

	if caseObj.Priority == "" {
		caseObj.Priority = "medium"
	}

	if err := config.DB.Create(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create case"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Case created successfully",
		"case":    caseObj,
	})
}

// GetAllCases returns all cases
func GetAllCases(c *gin.Context) {
	var cases []models.Case
	if err := config.DB.Preload("Evidence").Preload("Progress").Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cases": cases,
	})
}

// GetCaseByID returns a specific case
func GetCaseByID(c *gin.Context) {
	id := c.Param("id")
	caseID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var caseObj models.Case
	if err := config.DB.Preload("Evidence").Preload("Progress").First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"case": caseObj,
	})
}

// UpdateCaseStatus updates a case status
func UpdateCaseStatus(c *gin.Context) {
	id := c.Param("id")
	caseID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	caseObj.Status = input.Status
	if err := config.DB.Save(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case status updated successfully",
		"case":    caseObj,
	})
}

// Dashboard stats
func GetDashboardStats(c *gin.Context) {
	var totalCases int64
	var pendingCases int64
	var totalUnits int64

	config.DB.Model(&models.Case{}).Count(&totalCases)
	config.DB.Model(&models.Case{}).Where("status = ?", "pending").Count(&pendingCases)
	config.DB.Model(&models.SecurityUnit{}).Count(&totalUnits)

	c.JSON(http.StatusOK, gin.H{
		"totalCases":   totalCases,
		"pendingCases": pendingCases,
		"totalUnits":   totalUnits,
	})
}