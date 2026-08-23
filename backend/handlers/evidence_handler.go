package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// UploadEvidence uploads evidence for a case
func UploadEvidence(c *gin.Context) {
	var input struct {
		CaseID      string  `json:"caseId" binding:"required"`
		Type        string  `json:"type" binding:"required"`
		FileURL     string  `json:"fileUrl" binding:"required"`
		Description string  `json:"description"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
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

	evidence := models.Evidence{
		CaseID:      caseID,
		UploadedBy:  userObj.ID,
		Type:        input.Type,
		FileURL:     input.FileURL,
		Description: input.Description,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		IsVerified:  false,
		UploadedAt:  time.Now(),
	}

	if err := config.DB.Create(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload evidence"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Evidence uploaded successfully",
		"evidence": evidence,
	})
}

// GetEvidenceByCase returns all evidence for a case
func GetEvidenceByCase(c *gin.Context) {
	caseID := c.Param("caseId")
	id, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var evidence []models.Evidence
	if err := config.DB.Where("case_id = ?", id).Order("created_at desc").Find(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch evidence"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"evidence": evidence,
	})
}

// VerifyEvidence verifies evidence
func VerifyEvidence(c *gin.Context) {
	id := c.Param("id")
	evidenceID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidence ID"})
		return
	}

	var evidence models.Evidence
	if err := config.DB.First(&evidence, "id = ?", evidenceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evidence not found"})
		return
	}

	evidence.IsVerified = true
	if err := config.DB.Save(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify evidence"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Evidence verified successfully",
		"evidence": evidence,
	})
}

// DeleteEvidence deletes evidence
func DeleteEvidence(c *gin.Context) {
	id := c.Param("id")
	evidenceID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidence ID"})
		return
	}

	if err := config.DB.Delete(&models.Evidence{}, "id = ?", evidenceID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete evidence"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Evidence deleted successfully",
	})
}
