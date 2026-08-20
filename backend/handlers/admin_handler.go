package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
)

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