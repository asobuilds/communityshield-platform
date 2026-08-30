package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// UploadEvidence uploads evidence for a case.
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	caseID, err := uuid.Parse(input.CaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid case ID",
		})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	user, ok := userValue.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "case not found",
		})
		return
	}

	evidence := models.Evidence{
		CaseID:      caseID,
		UploadedBy:  user.ID,
		Type:        input.Type,
		FileURL:     input.FileURL,
		Description: input.Description,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		IsVerified:  false,
		UploadedAt:  time.Now(),
	}

	if err := config.DB.Create(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to upload evidence",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "evidence uploaded successfully",
		"evidence": evidence,
	})
}

// GetEvidenceByCase returns all evidence belonging to a case.
func GetEvidenceByCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("caseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid case ID",
		})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "case not found",
		})
		return
	}

	var evidence []models.Evidence

	if err := config.DB.
		Where("case_id = ?", caseID).
		Order("created_at DESC").
		Find(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch evidence",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"caseId":   caseID,
		"evidence": evidence,
		"count":    len(evidence),
	})
}

// VerifyEvidence marks evidence as verified.
func VerifyEvidence(c *gin.Context) {
	evidenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid evidence ID",
		})
		return
	}

	var evidence models.Evidence

	if err := config.DB.First(&evidence, "id = ?", evidenceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "evidence not found",
		})
		return
	}

	evidence.IsVerified = true

	if err := config.DB.Save(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to verify evidence",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "evidence verified successfully",
		"evidence": evidence,
	})
}

// DeleteEvidence soft-deletes evidence.
func DeleteEvidence(c *gin.Context) {
	evidenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid evidence ID",
		})
		return
	}

	var evidence models.Evidence

	if err := config.DB.First(&evidence, "id = ?", evidenceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "evidence not found",
		})
		return
	}

	if err := config.DB.Delete(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete evidence",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "evidence deleted successfully",
	})
}
