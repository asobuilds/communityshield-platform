package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

type CaseProgressRequest struct {
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

func AddCaseProgress(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid case id",
		})
		return
	}

	var req CaseProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "case not found",
		})
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authenticated user not found",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return
	}

	officerID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return
	}

	var officer models.Officer
	if err := config.DB.
		Where("id = ?", officerID).
		First(&officer).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "authenticated user is not an officer",
		})
		return
	}

	if officer.UnitID != caseRecord.UnitID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "officer is not assigned to this case unit",
		})
		return
	}

	progress := models.Progress{
		CaseID:      caseID,
		OfficerID:   officerID,
		Action:      req.Action,
		Description: req.Description,
	}

	if err := config.DB.Create(&progress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record case progress",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "case progress recorded successfully",
		"progress": progress,
	})
}

func GetCaseProgress(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid case id",
		})
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "case not found",
		})
		return
	}

	var progress []models.Progress

	if err := config.DB.
		Where("case_id = ?", caseID).
		Order("created_at ASC").
		Find(&progress).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve case progress",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"caseId":   caseID,
		"progress": progress,
		"count":    len(progress),
	})
}
