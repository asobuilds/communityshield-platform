package handlers

import (
	"net/http"
	"security-solution/models"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AutoArchiveCases archives cases that have been resolved for more than 30 days
func AutoArchiveCases() error {
	cutoffDate := time.Now().AddDate(0, 0, -30)
	var cases []models.Case
	if err := DB.Where("status = ? AND updated_at < ? AND archived = ?", "resolved", cutoffDate, false).Find(&cases).Error; err != nil {
		return err
	}
	for _, c := range cases {
		c.Archived = true
		if err := DB.Save(&c).Error; err != nil {
			return err
		}
	}
	return nil
}

// AdminGetArchivedCases returns all archived cases
func AdminGetArchivedCases(c *gin.Context) {
	var cases []models.Case
	if err := DB.Where("archived = ?", true).Order("updated_at desc").Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch archived cases"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// RestoreArchivedCase restores a case from archive
func RestoreArchivedCase(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}
	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ? AND archived = ?", parsedID, true).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Archived case not found"})
		return
	}
	caseItem.Archived = false
	if err := DB.Save(&caseItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore case"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Case restored successfully", "case": caseItem})
}

// AdminDeleteArchivedCase permanently deletes an archived case
func AdminDeleteArchivedCase(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}
	if err := DB.Unscoped().Delete(&models.Case{}, "id = ? AND archived = ?", parsedID, true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete archived case"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Archived case deleted permanently"})
}