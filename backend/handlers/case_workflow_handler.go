package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

type CaseCloseRequest struct {
	FinalReport string `json:"finalReport" binding:"required"`
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

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated user not found"})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	officerID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	var officer models.Officer
	if err := config.DB.First(&officer, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "authenticated user is not an officer"})
		return
	}

	if officer.UnitID != caseRecord.UnitID {
		c.JSON(http.StatusForbidden, gin.H{"error": "officer is not assigned to this case unit"})
		return
	}

	now := time.Now()
	caseRecord.Status = "dispatched"
	caseRecord.DispatchedAt = &now

	if err := config.DB.Save(&caseRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to dispatch case"})
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

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated user not found"})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	officerID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	var officer models.Officer
	if err := config.DB.First(&officer, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "authenticated user is not an officer"})
		return
	}

	if officer.UnitID != caseRecord.UnitID {
		c.JSON(http.StatusForbidden, gin.H{"error": "officer is not assigned to this case unit"})
		return
	}

	now := time.Now()
	caseRecord.Status = "on_scene"
	caseRecord.ArrivedAt = &now

	if err := config.DB.Save(&caseRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record arrival"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "arrival recorded successfully",
		"case":    caseRecord,
	})
}

func CloseCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}

	var req CaseCloseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "finalReport is required",
		})
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}

	if caseRecord.Status == "closed" {
		c.JSON(http.StatusConflict, gin.H{"error": "case is already closed"})
		return
	}

	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authenticated user not found"})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	officerID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	var officer models.Officer
	if err := config.DB.First(&officer, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "authenticated user is not an officer"})
		return
	}

	if officer.UnitID != caseRecord.UnitID {
		c.JSON(http.StatusForbidden, gin.H{"error": "officer is not assigned to this case unit"})
		return
	}

	now := time.Now()

	caseRecord.Status = "closed"
	caseRecord.ClosedAt = &now
	caseRecord.ClosedBy = &officerID
	caseRecord.FinalReport = req.FinalReport

	if err := config.DB.Save(&caseRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "case closed successfully",
		"case":    caseRecord,
	})
}
