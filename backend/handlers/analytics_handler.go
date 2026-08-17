package handlers

import (
	"net/http"
	"security-solution/models"
	"time"
	"github.com/gin-gonic/gin"
)

// GetCaseAnalytics returns case statistics
func GetCaseAnalytics(c *gin.Context) {
	// Count by status
	var statusCounts []struct {
		Status string
		Count  int
	}
	if err := DB.Model(&models.Case{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch case status counts"})
		return
	}

	// Count by priority
	var priorityCounts []struct {
		Priority string
		Count    int
	}
	if err := DB.Model(&models.Case{}).
		Select("priority, count(*) as count").
		Group("priority").
		Scan(&priorityCounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch case priority counts"})
		return
	}

	// Total cases
	var totalCases int64
	DB.Model(&models.Case{}).Count(&totalCases)

	c.JSON(http.StatusOK, gin.H{
		"statusCounts":   statusCounts,
		"priorityCounts": priorityCounts,
		"totalCases":     totalCases,
	})
}

// GetSOSAnalytics returns SOS statistics (last 7 days)
func GetSOSAnalytics(c *gin.Context) {
	// Daily SOS count for last 7 days
	var dailySOS []struct {
		Date  string
		Count int
	}
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := DB.Model(&models.SOSAlert{}).
		Select("DATE(created_at) as date, count(*) as count").
		Where("created_at >= ?", sevenDaysAgo).
		Group("DATE(created_at)").
		Order("date asc").
		Scan(&dailySOS).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SOS daily counts"})
		return
	}

	// Total SOS
	var totalSOS int64
	DB.Model(&models.SOSAlert{}).Count(&totalSOS)

	// SOS by status
	var statusCounts []struct {
		Status string
		Count  int
	}
	if err := DB.Model(&models.SOSAlert{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SOS status counts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dailySOS":     dailySOS,
		"totalSOS":     totalSOS,
		"statusCounts": statusCounts,
	})
}

// GetUnitAnalytics returns unit performance
func GetUnitAnalytics(c *gin.Context) {
	var units []models.Unit
	if err := DB.Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}

	type UnitStats struct {
		UnitID    string  `json:"unitId"`
		Name      string  `json:"name"`
		CaseCount int64   `json:"caseCount"`
		AvgRating float64 `json:"avgRating"`
	}
	var stats []UnitStats

	for _, u := range units {
		var count int64
		DB.Model(&models.Case{}).Where("unit_id = ?", u.ID).Count(&count)

		var avgRating float64
		DB.Model(&models.Rating{}).Where("unit_id = ?", u.ID).
			Select("COALESCE(AVG(rating), 0)").Scan(&avgRating)

		stats = append(stats, UnitStats{
			UnitID:    u.ID.String(),
			Name:      u.Name,
			CaseCount: count,
			AvgRating: avgRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{"unitStats": stats})
}