package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// GenerateCaseReport generates a report for a case
func GenerateCaseReport(c *gin.Context) {
	caseID := c.Param("caseId")
	id, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var caseObj models.Case
	if err := config.DB.Preload("Evidence").Preload("Progress").First(&caseObj, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	var evidenceCount int64
	config.DB.Model(&models.Evidence{}).Where("case_id = ?", id).Count(&evidenceCount)

	var progressCount int64
	config.DB.Model(&models.Progress{}).Where("case_id = ?", id).Count(&progressCount)

	report := gin.H{
		"case":           caseObj,
		"evidenceCount":  evidenceCount,
		"progressCount":  progressCount,
		"generatedAt":    time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}

// GenerateUnitReport generates a report for a unit
func GenerateUnitReport(c *gin.Context) {
	unitID := c.Param("unitId")
	id, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var totalCases int64
	var resolvedCases int64
	var pendingCases int64
	var transferredCases int64

	config.DB.Model(&models.Case{}).Where("unit_id = ?", id).Count(&totalCases)
	config.DB.Model(&models.Case{}).Where("unit_id = ? AND status = ?", id, "resolved").Count(&resolvedCases)
	config.DB.Model(&models.Case{}).Where("unit_id = ? AND status = ?", id, "pending").Count(&pendingCases)
	config.DB.Model(&models.Case{}).Where("unit_id = ? AND status = ?", id, "transferred").Count(&transferredCases)

	var officers []models.Officer
	config.DB.Where("unit_id = ?", id).Find(&officers)

	var recentCases []models.Case
	config.DB.Where("unit_id = ?", id).Order("created_at desc").Limit(5).Find(&recentCases)

	report := gin.H{
		"unitId":          id,
		"totalCases":      totalCases,
		"resolvedCases":   resolvedCases,
		"pendingCases":    pendingCases,
		"transferredCases": transferredCases,
		"officerCount":    len(officers),
		"recentCases":     recentCases,
		"generatedAt":     time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}

// GenerateDailyReport generates a daily report
func GenerateDailyReport(c *gin.Context) {
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	var newCases int64
	var resolvedCases int64
	var totalCases int64

	config.DB.Model(&models.Case{}).Where("created_at >= ?", startOfDay).Count(&newCases)
	config.DB.Model(&models.Case{}).Where("updated_at >= ? AND status = ?", startOfDay, "resolved").Count(&resolvedCases)
	config.DB.Model(&models.Case{}).Count(&totalCases)

	var topUnits []struct {
		UnitID string
		Count  int64
	}
	config.DB.Table("cases").
		Select("unit_id, COUNT(*) as count").
		Where("created_at >= ?", startOfDay).
		Group("unit_id").
		Order("count desc").
		Limit(5).
		Scan(&topUnits)

	report := gin.H{
		"date":           today.Format("2006-01-02"),
		"newCases":       newCases,
		"resolvedCases":  resolvedCases,
		"totalCases":     totalCases,
		"topUnits":       topUnits,
		"generatedAt":    time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"report": report,
	})
}

// ExportCaseReport exports a case report as PDF/CSV
func ExportCaseReport(c *gin.Context) {
	caseID := c.Param("caseId")
	id, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	format := c.Query("format")
	if format == "" {
		format = "json"
	}

	var caseObj models.Case
	if err := config.DB.Preload("Evidence").Preload("Progress").First(&caseObj, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// For now, return JSON. In production, generate PDF/CSV
	c.JSON(http.StatusOK, gin.H{
		"format":  format,
		"case":    caseObj,
		"exportedAt": time.Now(),
	})
}