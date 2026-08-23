package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
)

// ApplyForSecurityUnit allows an authenticated user to apply to join
// an existing security unit. Approval is required before membership
// becomes active.
func ApplyForSecurityUnit(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return
	}

	var input struct {
		UnitID uuid.UUID `json:"unitId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unitId is required",
		})
		return
	}

	var unit models.SecurityUnit

	if err := config.DB.
		First(&unit, "id = ?", input.UnitID).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "security unit not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to find security unit",
		})
		return
	}

	var existing models.UnitMember

	err := config.DB.
		Where("user_id = ? AND unit_id = ?", userID, input.UnitID).
		First(&existing).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "membership already exists",
			"status": existing.Status,
			"role":   existing.Role,
		})
		return
	}

	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to check existing membership",
		})
		return
	}

	member := models.UnitMember{
		UserID: userID,
		UnitID: input.UnitID,
		Role:   "officer",
		Status: "pending",
	}

	if err := config.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to submit security unit application",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "security unit application submitted",
		"membership": gin.H{
			"id":     member.ID,
			"unitId": member.UnitID,
			"userId": member.UserID,
			"role":   member.Role,
			"status": member.Status,
		},
	})
}

// GetMyUnitMembership returns all security-unit memberships belonging
// to the authenticated user.
func GetMyUnitMembership(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return
	}

	var memberships []models.UnitMember

	if err := config.DB.
		Preload("Unit").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&memberships).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve memberships",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"memberships": memberships,
	})
}

// SubmitGovernmentID submits a government-issued identification document
// for verification.
//
// Important:
// This endpoint ONLY submits the identity verification request.
// It does NOT grant officer or administrator privileges.
func SubmitGovernmentID(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return
	}

	var input struct {
		IDType      string `json:"idType" binding:"required"`
		IDNumber    string `json:"idNumber" binding:"required"`
		DocumentURL string `json:"documentUrl" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "idType, idNumber and documentUrl are required",
		})
		return
	}

	input.IDType = strings.TrimSpace(input.IDType)
	input.IDNumber = strings.TrimSpace(input.IDNumber)
	input.DocumentURL = strings.TrimSpace(input.DocumentURL)

	if input.IDType == "" || input.IDNumber == "" || input.DocumentURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "idType, idNumber and documentUrl cannot be empty",
		})
		return
	}

	var existing models.GovernmentIDVerification

	err := config.DB.
		Where("user_id = ? AND status IN ?", userID, []string{
			"pending",
			"approved",
		}).
		Order("created_at DESC").
		First(&existing).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":  "an active government ID verification already exists",
			"status": existing.Status,
		})
		return
	}

	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to check existing government ID verification",
		})
		return
	}

	verification := models.GovernmentIDVerification{
		UserID:      userID,
		IDType:      input.IDType,
		IDNumber:    input.IDNumber,
		DocumentURL: input.DocumentURL,
		Status:      "pending",
	}

	if err := config.DB.Create(&verification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to submit government ID",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "government ID submitted for verification",
		"verification": gin.H{
			"id":     verification.ID,
			"status": verification.Status,
			"idType": verification.IDType,
		},
	})
}
