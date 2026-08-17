package handlers

import (
	"net/http"
	"security-solution/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AddProgress(c *gin.Context) {
	caseID := c.Param("id")
	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var input struct {
		OfficerID   string `json:"officerId" binding:"required"`
		Action      string `json:"action" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	officerID, err := uuid.Parse(input.OfficerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	// Check if case exists
	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ?", parsedCaseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Check if officer exists
	var officer models.User
	if err := DB.First(&officer, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}

	progress := models.CaseProgress{
		CaseID:      parsedCaseID,
		OfficerID:   officerID,
		Action:      input.Action,
		Description: input.Description,
	}

	if err := DB.Create(&progress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add progress"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Progress added",
		"progress": progress,
	})
}

func GetProgress(c *gin.Context) {
	caseID := c.Param("id")
	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var progressList []models.CaseProgress
	if err := DB.Where("case_id = ?", parsedCaseID).Order("created_at desc").Find(&progressList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"progress": progressList})
}