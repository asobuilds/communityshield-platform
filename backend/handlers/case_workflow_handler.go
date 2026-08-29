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
