package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// AddCaseProgress adds progress to a case
func AddCaseProgress(c *gin.Context) {
	var input struct {
		CaseID      string `json:"caseId" binding:"required"`
		Action      string `json:"action" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	caseID, err := uuid.Parse(input.CaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	// Get user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	// Check if case exists
	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	progress := models.Progress{
		CaseID:      caseID,
		OfficerID:   userObj.ID,
		Action:      input.Action,
		Description: input.Description,
	}

	if err := config.DB.Create(&progress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add progress"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Progress added successfully",
		"progress": progress,
	})
}

// GetCaseProgress gets all progress for a case
func GetCaseProgress(c *gin.Context) {
	caseID := c.Param("caseId")
	id, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var progress []models.Progress
	if err := config.DB.Where("case_id = ?", id).Order("created_at desc").Find(&progress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"progress": progress,
	})
}