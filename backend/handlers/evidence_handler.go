package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

func getAuthenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authenticated user not found",
		})
		return uuid.Nil, false
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return uuid.Nil, false
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authenticated user",
		})
		return uuid.Nil, false
	}

	return id, true
}

// canAccessCase enforces the core rule: unit access is NOT case access.
// A user can see a case if they are one of:
//   - super_admin (any case)
//   - the reporter
//   - primary officer via Case.AssignedTo
//   - Head Admin of the case's unit
//   - Regular admin of the unit AND explicitly submitted (CaseAdminAssignment)
//   - primary or paired officer on the case (CaseOfficer)
// Support-tier officers do NOT get evidence access.
func canAccessCase(user *models.User, caseRecord *models.Case) bool {
	if user == nil || caseRecord == nil {
		return false
	}
	if user.IsSuperAdmin || user.Role == "super_admin" {
		return true
	}
	if caseRecord.ReportedBy == user.ID {
		return true
	}
	if caseRecord.AssignedTo != nil && *caseRecord.AssignedTo == user.ID {
		return true
	}

	var membership models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ?",
			caseRecord.UnitID, user.ID, models.MembershipActive).
		First(&membership).Error; err != nil {
		return false
	}

	if membership.IsHeadAdmin {
		return true
	}

	if membership.Role == models.UnitRoleAdmin {
		var assignment models.CaseAdminAssignment
		if config.DB.
			Where("case_id = ? AND admin_id = ? AND status IN ?",
				caseRecord.ID, user.ID, []string{"pending", "approved"}).
			First(&assignment).Error == nil {
			return true
		}
	}

	if membership.Role == models.UnitRoleOfficer {
		var caseOfficer models.CaseOfficer
		if config.DB.
			Where("case_id = ? AND officer_id = ?", caseRecord.ID, user.ID).
			First(&caseOfficer).Error == nil {
			role := strings.ToLower(caseOfficer.Role)
			if role == "primary" || role == "paired" {
				return true
			}
		}
	}

	return false
}

// canManageEvidence is like canAccessCase but excludes the reporter.
// Only officers/admins with investigative authority may verify or delete evidence.
func canManageEvidence(user *models.User, caseRecord *models.Case) bool {
	if user == nil || caseRecord == nil {
		return false
	}
	if user.IsSuperAdmin || user.Role == "super_admin" {
		return true
	}
	if caseRecord.AssignedTo != nil && *caseRecord.AssignedTo == user.ID {
		return true
	}

	var membership models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ?",
			caseRecord.UnitID, user.ID, models.MembershipActive).
		First(&membership).Error; err != nil {
		return false
	}

	if membership.IsHeadAdmin {
		return true
	}

	if membership.Role == models.UnitRoleAdmin {
		var assignment models.CaseAdminAssignment
		if config.DB.
			Where("case_id = ? AND admin_id = ? AND status IN ?",
				caseRecord.ID, user.ID, []string{"pending", "approved"}).
			First(&assignment).Error == nil {
			return true
		}
	}

	if membership.Role == models.UnitRoleOfficer {
		var caseOfficer models.CaseOfficer
		if config.DB.
			Where("case_id = ? AND officer_id = ?", caseRecord.ID, user.ID).
			First(&caseOfficer).Error == nil {
			role := strings.ToLower(caseOfficer.Role)
			if role == "primary" || role == "paired" {
				return true
			}
		}
	}

	return false
}

// userCanAccessCase is a legacy helper kept for existing callers.
// It resolves the user record and delegates to canAccessCase.
func userCanAccessCase(userID uuid.UUID, caseRecord *models.Case) bool {
	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		return false
	}
	return canAccessCase(&user, caseRecord)
}

// officerCanManageCaseEvidence is a legacy helper kept for existing callers.
// It resolves the user record and delegates to canManageEvidence.
func officerCanManageCaseEvidence(userID uuid.UUID, caseRecord *models.Case) bool {
	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		return false
	}
	return canManageEvidence(&user, caseRecord)
}

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
			"error": "invalid case id",
		})
		return
	}

	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "case not found",
		})
		return
	}

	if !userCanAccessCase(userID, &caseRecord) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you are not authorized to add evidence to this case",
		})
		return
	}

	if caseRecord.Status == "closed" {
		c.JSON(http.StatusConflict, gin.H{
			"error": "evidence cannot be uploaded to a closed case",
		})
		return
	}

	evidence := models.Evidence{
		CaseID:      caseID,
		UploadedBy:  userID,
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

// GetEvidenceByCase returns evidence for a case.
func GetEvidenceByCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("caseId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid case id",
		})
		return
	}

	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "case not found",
		})
		return
	}

	if !userCanAccessCase(userID, &caseRecord) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you are not authorized to view evidence for this case",
		})
		return
	}

	var evidence []models.Evidence
	if err := config.DB.
		Where("case_id = ?", caseID).
		Order("created_at ASC").
		Find(&evidence).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch evidence",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"evidence": evidence,
	})
}

// VerifyEvidence verifies evidence belonging to an authorized case.
func VerifyEvidence(c *gin.Context) {
	evidenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid evidence id",
		})
		return
	}

	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var evidence models.Evidence
	if err := config.DB.First(&evidence, "id = ?", evidenceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "evidence not found",
		})
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", evidence.CaseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "case not found",
		})
		return
	}

	if !officerCanManageCaseEvidence(userID, &caseRecord) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you are not authorized to verify evidence for this case",
		})
		return
	}

	if evidence.IsVerified {
		c.JSON(http.StatusConflict, gin.H{
			"error": "evidence is already verified",
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

// DeleteEvidence deletes evidence belonging to an authorized case.
func DeleteEvidence(c *gin.Context) {
	evidenceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid evidence id",
		})
		return
	}

	userID, ok := getAuthenticatedUserID(c)
	if !ok {
		return
	}

	var evidence models.Evidence
	if err := config.DB.First(&evidence, "id = ?", evidenceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "evidence not found",
		})
		return
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", evidence.CaseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "case not found",
		})
		return
	}

	if !officerCanManageCaseEvidence(userID, &caseRecord) {
		if evidence.UploadedBy != userID {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "you are not authorized to delete this evidence",
			})
			return
		}
	}

	if caseRecord.Status == "closed" {
		c.JSON(http.StatusConflict, gin.H{
			"error": "evidence cannot be deleted from a closed case",
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
