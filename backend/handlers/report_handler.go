package handlers

import (
	"net/http"
	"security-solution/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateReport(c *gin.Context) {
	var input struct {
		TargetType string `json:"targetType" binding:"required"`
		TargetID   string `json:"targetId" binding:"required"`
		Reason     string `json:"reason" binding:"required"`
		UserID     string `json:"userId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}

	report := models.Report{
		ReportedBy: userID,
		TargetType: input.TargetType,
		TargetID:   input.TargetID,
		Reason:     input.Reason,
		Status:     "pending",
	}
	if err := DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Report submitted",
		"report":  report,
	})
}

func AdminGetReports(c *gin.Context) {
	var reports []models.Report
	if err := DB.Order("created_at desc").Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reports": reports})
}

func AdminUpdateReportStatus(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	var input struct {
		Status   string `json:"status" binding:"required"`
		Reviewer string `json:"reviewer" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reviewerID, err := uuid.Parse(input.Reviewer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reviewer ID"})
		return
	}

	var report models.Report
	if err := DB.First(&report, "id = ?", parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	report.Status = input.Status
	report.ReviewedBy = &reviewerID
	if err := DB.Save(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Report updated",
		"report":  report,
	})
}