package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

type caseProgressRequest struct {
	Action      string `json:"action" binding:"required"`
	Description string `json:"description"`
}

type closeCaseRequest struct {
	FinalReport string `json:"finalReport" binding:"required"`
}

// DispatchCase moves an assigned case into the dispatched state.
func DispatchCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}

	if caseObj.Status != "assigned" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "case must be assigned before dispatch",
		})
		return
	}

	now := time.Now()
	caseObj.Status = "dispatched"
	caseObj.DispatchedAt = &now

	if err := config.DB.Save(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to dispatch case",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "case dispatched successfully",
		"case":    caseObj,
	})
}

// ArriveAtCase moves a dispatched case into the in-progress state.
func ArriveAtCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}

	if caseObj.Status != "dispatched" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "case must be dispatched before arrival",
		})
		return
	}

	now := time.Now()
	caseObj.Status = "in_progress"
	caseObj.ArrivedAt = &now

	if err := config.DB.Save(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record arrival",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "arrival recorded successfully",
		"case":    caseObj,
	})
}

// AddCaseProgress records operational progress for an active case.
func AddCaseProgress(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	var req caseProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "action is required",
		})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}

	if caseObj.Status != "in_progress" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "case must be in progress before recording progress",
		})
		return
	}

	officerIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	officerID, ok := officerIDValue.(uuid.UUID)
	if !ok {
		parsedOfficerID, err := uuid.Parse(officerIDValue.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
			return
		}
		officerID = parsedOfficerID
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

// CloseCase closes an active case and records the final report.
func CloseCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	var req closeCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "finalReport is required",
		})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "case not found"})
		return
	}

	if caseObj.Status != "in_progress" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "case must be in progress before closure",
		})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		parsedUserID, err := uuid.Parse(userIDValue.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
			return
		}
		userID = parsedUserID
	}

	now := time.Now()
	caseObj.Status = "closed"
	caseObj.FinalReport = req.FinalReport
	caseObj.ClosedAt = &now
	caseObj.ClosedBy = &userID

	if err := config.DB.Save(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to close case",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "case closed successfully",
		"case":    caseObj,
	})
}
