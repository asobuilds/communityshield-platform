package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
)

// Admin-only functions
func CreateUnit(c *gin.Context) {
	var input struct {
		Name             string  `json:"name" binding:"required"`
		Type             string  `json:"type" binding:"required"`
		Latitude         float64 `json:"latitude"`
		Longitude        float64 `json:"longitude"`
		CoverageArea     string  `json:"coverageArea"`
		ContactPerson    string  `json:"contactPerson"`
		ContactPhone     string  `json:"contactPhone"`
		ContactEmail     string  `json:"contactEmail"`
		RegistrationNumber string `json:"registrationNumber"`
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

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create units"})
		return
	}

	unit := models.SecurityUnit{
		Name:             input.Name,
		Type:             input.Type,
		Latitude:         input.Latitude,
		Longitude:        input.Longitude,
		CoverageArea:     input.CoverageArea,
		ContactPerson:    input.ContactPerson,
		ContactPhone:     input.ContactPhone,
		ContactEmail:     input.ContactEmail,
		RegistrationNumber: input.RegistrationNumber,
		Status:           "active",
	}

	if err := config.DB.Create(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unit"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Unit created successfully",
		"unit":    unit,
	})
}

// Admin dashboard stats
func GetAdminDashboardStats(c *gin.Context) {
	var totalCases int64
	var pendingCases int64
	var totalUnits int64
	var totalOfficers int64

	config.DB.Model(&models.Case{}).Count(&totalCases)
	config.DB.Model(&models.Case{}).Where("status = ?", "pending").Count(&pendingCases)
	config.DB.Model(&models.SecurityUnit{}).Count(&totalUnits)
	config.DB.Model(&models.Officer{}).Count(&totalOfficers)

	c.JSON(http.StatusOK, gin.H{
		"totalCases":    totalCases,
		"pendingCases":  pendingCases,
		"totalUnits":    totalUnits,
		"totalOfficers": totalOfficers,
	})
}