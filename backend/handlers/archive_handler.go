package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// ArchiveCase archives a case
func ArchiveCase(c *gin.Context) {
	id := c.Param("id")
	caseID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Set status to archived
	caseObj.Status = "archived"
	if err := config.DB.Save(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case archived successfully",
		"case":    caseObj,
	})
}

// UnarchiveCase unarchives a case
func UnarchiveCase(c *gin.Context) {
	id := c.Param("id")
	caseID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Set status back to pending
	caseObj.Status = "pending"
	if err := config.DB.Save(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unarchive case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case unarchived successfully",
		"case":    caseObj,
	})
}

// GetArchivedCases returns all archived cases
func GetArchivedCases(c *gin.Context) {
	var cases []models.Case
	if err := config.DB.Where("status = ?", "archived").Preload("Evidence").Preload("Progress").Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch archived cases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cases": cases,
	})
}