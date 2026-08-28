package handlers

import (
	"net/http"
	"time"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}

	var req CaseProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}

	var progress models.Progress

	progress.CaseID = caseID
	progress.Action = req.Action
	progress.Description = req.Description

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authenticated user not found",
		})
		return
	}

	switch value := userIDValue.(type) {
	case uuid.UUID:
		progress.OfficerID = value
	case string:
		officerID, err := uuid.Parse(value)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authenticated user",
			})
			return
		}
		progress.OfficerID = officerID
	default:
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return
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

func DispatchCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}

	if caseRecord.Status != "assigned" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "case must be assigned before dispatch",
		})
		return
	}

	now := time.Now()
	caseRecord.Status = "dispatched"
	caseRecord.DispatchedAt = &now

	if err := config.DB.Save(&caseRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to dispatch case",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "case dispatched successfully",
		"case":    caseRecord,
	})
}

func ArriveAtCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}

	if caseRecord.Status != "dispatched" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "case must be dispatched before arrival",
		})
		return
	}

	now := time.Now()
	caseRecord.Status = "in_progress"
	caseRecord.ArrivedAt = &now

	if err := config.DB.Save(&caseRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record arrival",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "arrival recorded successfully",
		"case":    caseRecord,
	})
}
