package handlers

import (
	"net/http"
	"security-solution/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestCaseTransfer allows a citizen to request transfer of a case to a new officer or unit
func RequestCaseTransfer(c *gin.Context) {
	caseID := c.Param("id")
	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var input struct {
		ToOfficerID string `json:"toOfficerId"`
		ToUnitID    string `json:"toUnitId"`
		Reason      string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context (we'll use header for now)
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check case exists and belongs to this user
	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ?", parsedCaseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}
	if caseItem.ReportedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the reporter of this case"})
		return
	}
	if caseItem.Status == "closed" || caseItem.Status == "resolved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Case is already closed or resolved"})
		return
	}

	// Check if transfer already pending
	var existingTransfer models.CaseTransfer
	if err := DB.Where("case_id = ? AND status = ?", parsedCaseID, "pending").First(&existingTransfer).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer already requested and pending"})
		return
	}

	// Parse ToOfficerID (optional)
	var toOfficerID *uuid.UUID
	if input.ToOfficerID != "" {
		parsed, err := uuid.Parse(input.ToOfficerID)
		if err == nil {
			toOfficerID = &parsed
		}
	}

	// Parse ToUnitID (optional)
	var toUnitID *uuid.UUID
	if input.ToUnitID != "" {
		parsed, err := uuid.Parse(input.ToUnitID)
		if err == nil {
			toUnitID = &parsed
		}
	}

	if toOfficerID == nil && toUnitID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Must specify either a new officer or a new unit"})
		return
	}

	// Create transfer request
	transfer := models.CaseTransfer{
		CaseID:        parsedCaseID,
		RequestedBy:   userID,
		FromOfficerID: caseItem.AssignedTo,
		FromUnitID:    caseItem.UnitID,
		ToOfficerID:   toOfficerID,
		ToUnitID:      toUnitID,
		Reason:        input.Reason,
		Status:        "pending",
	}

	if err := DB.Create(&transfer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to request transfer"})
		return
	}

	// Notify admin/officers (we'll just create a notification)
	if caseItem.UnitID != nil {
		// Notify unit admin
		var admin models.User
		if err := DB.Where("unit_id = ? AND role = ?", caseItem.UnitID, "unit_admin").First(&admin).Error; err == nil {
			title := "Case Transfer Request"
			message := fmt.Sprintf("Case '%s' has been requested for transfer.", caseItem.Title)
			CreateNotification(admin.ID, parsedCaseID, title, message)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Transfer requested successfully",
		"transfer": transfer,
	})
}

// GetCaseTransferStatus returns the transfer status for a case
func GetCaseTransferStatus(c *gin.Context) {
	caseID := c.Param("id")
	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var transfers []models.CaseTransfer
	if err := DB.Where("case_id = ?", parsedCaseID).Order("created_at desc").Find(&transfers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transfer status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transfers": transfers})
}

// ApproveCaseTransfer allows an admin to approve a transfer request
func ApproveCaseTransfer(c *gin.Context) {
	transferID := c.Param("transferId")
	parsedTransferID, err := uuid.Parse(transferID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer ID"})
		return
	}

	// Get admin user from header
	adminIDStr := c.GetHeader("X-Admin-ID")
	if adminIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin ID required"})
		return
	}
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin ID"})
		return
	}

	var transfer models.CaseTransfer
	if err := DB.First(&transfer, "id = ?", parsedTransferID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}

	if transfer.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer is no longer pending"})
		return
	}

	// Update transfer status
	transfer.Status = "approved"
	transfer.ApprovedBy = &adminID
	if err := DB.Save(&transfer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve transfer"})
		return
	}

	// Update the case with new officer/unit
	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ?", transfer.CaseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	if transfer.ToOfficerID != nil {
		caseItem.AssignedTo = transfer.ToOfficerID
	}
	if transfer.ToUnitID != nil {
		caseItem.UnitID = transfer.ToUnitID
	}

	if err := DB.Save(&caseItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update case"})
		return
	}

	// Update transfer status to completed
	transfer.Status = "completed"
	if err := DB.Save(&transfer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transfer status"})
		return
	}

	// Notify new officer and reporter
	var reporter models.User
	if err := DB.First(&reporter, "id = ?", caseItem.ReportedBy).Error; err == nil {
		title := "Case Transfer Approved"
		message := fmt.Sprintf("Your case '%s' has been transferred.", caseItem.Title)
		CreateNotification(reporter.ID, caseItem.ID, title, message)
	}

	if transfer.ToOfficerID != nil {
		title := "Case Assigned"
		message := fmt.Sprintf("You have been assigned to case '%s'.", caseItem.Title)
		CreateNotification(*transfer.ToOfficerID, caseItem.ID, title, message)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Transfer approved and case updated",
		"transfer": transfer,
	})
}

// RejectCaseTransfer allows an admin to reject a transfer request
func RejectCaseTransfer(c *gin.Context) {
	transferID := c.Param("transferId")
	parsedTransferID, err := uuid.Parse(transferID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer ID"})
		return
	}

	adminIDStr := c.GetHeader("X-Admin-ID")
	if adminIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin ID required"})
		return
	}
	adminID, err := uuid.Parse(adminIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid admin ID"})
		return
	}

	var transfer models.CaseTransfer
	if err := DB.First(&transfer, "id = ?", parsedTransferID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}

	if transfer.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transfer is no longer pending"})
		return
	}

	transfer.Status = "rejected"
	transfer.ApprovedBy = &adminID
	if err := DB.Save(&transfer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject transfer"})
		return
	}

	// Notify reporter
	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ?", transfer.CaseID).Error; err == nil {
		var reporter models.User
		if err := DB.First(&reporter, "id = ?", caseItem.ReportedBy).Error; err == nil {
			title := "Case Transfer Rejected"
			message := fmt.Sprintf("Your request to transfer case '%s' has been rejected.", caseItem.Title)
			CreateNotification(reporter.ID, caseItem.ID, title, message)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Transfer rejected",
		"transfer": transfer,
	})
}

// GetPendingTransfers returns all pending transfer requests for a unit admin
func GetPendingTransfers(c *gin.Context) {
	unitIDStr := c.Query("unitId")
	if unitIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unitId is required"})
		return
	}
	unitID, err := uuid.Parse(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var transfers []models.CaseTransfer
	if err := DB.Where("to_unit_id = ? AND status = ?", unitID, "pending").Find(&transfers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending transfers"})
		return
	}

	// Prefetch case details
	type TransferWithCase struct {
		Transfer models.CaseTransfer `json:"transfer"`
		Case     models.Case         `json:"case"`
	}
	var result []TransferWithCase
	for _, t := range transfers {
		var caseItem models.Case
		if err := DB.First(&caseItem, "id = ?", t.CaseID).Error; err == nil {
			result = append(result, TransferWithCase{Transfer: t, Case: caseItem})
		}
	}

	c.JSON(http.StatusOK, gin.H{"transfers": result})
}