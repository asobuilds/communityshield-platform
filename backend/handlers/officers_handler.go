package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// canManageOfficersInUnit returns true if the caller may create, update,
// or delete officers in the given unit.
func canManageOfficersInUnit(user *models.User, unitID uuid.UUID) bool {
	if user == nil {
		return false
	}
	if user.IsSuperAdmin || user.Role == "super_admin" {
		return true
	}
	var membership models.UnitMembership
	err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ?",
			unitID, user.ID, models.MembershipActive).
		First(&membership).Error
	if err != nil {
		return false
	}
	return membership.IsHeadAdmin || membership.Role == models.UnitRoleAdmin
}

// canViewOfficersInUnit returns true if the caller may read the officer
// roster of the given unit.
func canViewOfficersInUnit(user *models.User, unitID uuid.UUID) bool {
	if user == nil {
		return false
	}
	if user.IsSuperAdmin || user.Role == "super_admin" {
		return true
	}
	var membership models.UnitMembership
	err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ?",
			unitID, user.ID, models.MembershipActive).
		First(&membership).Error
	return err == nil
}

func callerUser(c *gin.Context) (*models.User, bool) {
	value, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil, false
	}
	u, ok := value.(*models.User)
	if !ok || u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return nil, false
	}
	return u, true
}

// CreateOfficer creates a new officer.
func CreateOfficer(c *gin.Context) {
	user, ok := callerUser(c)
	if !ok {
		return
	}

	var input struct {
		UnitID      string `json:"unitId" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Rank        string `json:"rank" binding:"required"`
		BadgeNumber string `json:"badgeNumber" binding:"required"`
		Role        string `json:"role" binding:"required"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	unitID, err := uuid.Parse(input.UnitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	if !canManageOfficersInUnit(user, unitID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins of this unit may create officers"})
		return
	}

	officer := models.Officer{
		UnitID:      unitID,
		Name:        input.Name,
		Rank:        input.Rank,
		BadgeNumber: input.BadgeNumber,
		Role:        input.Role,
		Phone:       input.Phone,
		Email:       input.Email,
		Status:      "active",
	}

	if err := config.DB.Create(&officer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create officer"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Officer created successfully",
		"officer": officer,
	})
}

// GetAllOfficers gets all officers (super admin only).
func GetAllOfficers(c *gin.Context) {
	user, ok := callerUser(c)
	if !ok {
		return
	}
	if !user.IsSuperAdmin && user.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
		return
	}

	var officers []models.Officer
	if err := config.DB.Find(&officers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch officers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"officers": officers})
}

// GetOfficerByID gets a specific officer.
func GetOfficerByID(c *gin.Context) {
	user, ok := callerUser(c)
	if !ok {
		return
	}

	officerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	var officer models.Officer
	if err := config.DB.First(&officer, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}

	if !canViewOfficersInUnit(user, officer.UnitID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this officer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"officer": officer})
}

// UpdateOfficer updates an officer.
func UpdateOfficer(c *gin.Context) {
	user, ok := callerUser(c)
	if !ok {
		return
	}

	officerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	var input struct {
		Name   string `json:"name"`
		Rank   string `json:"rank"`
		Role   string `json:"role"`
		Phone  string `json:"phone"`
		Email  string `json:"email"`
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var officer models.Officer
	if err := config.DB.First(&officer, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}

	// Snapshot pre-update values for the audit trail.
	oldOfficer := officer

	if !canManageOfficersInUnit(user, officer.UnitID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins of this unit may update officers"})
		return
	}

	if input.Name != "" {
		officer.Name = input.Name
	}
	if input.Rank != "" {
		officer.Rank = input.Rank
	}
	if input.Role != "" {
		officer.Role = input.Role
	}
	if input.Phone != "" {
		officer.Phone = input.Phone
	}
	if input.Email != "" {
		officer.Email = input.Email
	}
	if input.Status != "" {
		officer.Status = input.Status
	}

	if err := config.DB.Save(&officer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update officer"})
		return
	}

	auditSvc := services.NewAuditService()
	_ = auditSvc.LogAction(
		user.ID,
		"officer.update",
		"officer",
		officerID.String(),
		oldOfficer,
		officer,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Officer updated successfully",
		"officer": officer,
	})
}

// DeleteOfficer deletes an officer.
func DeleteOfficer(c *gin.Context) {
	user, ok := callerUser(c)
	if !ok {
		return
	}

	officerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	var officer models.Officer
	if err := config.DB.First(&officer, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}

	if !canManageOfficersInUnit(user, officer.UnitID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins of this unit may delete officers"})
		return
	}

	if err := config.DB.Delete(&officer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete officer"})
		return
	}

	auditSvc := services.NewAuditService()
	_ = auditSvc.LogAction(
		user.ID,
		"officer.delete",
		"officer",
		officerID.String(),
		officer,
		nil,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Officer deleted successfully"})
}

// GetOfficersByUnit gets officers for a specific unit.
func GetOfficersByUnit(c *gin.Context) {
	user, ok := callerUser(c)
	if !ok {
		return
	}

	unitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	if !canViewOfficersInUnit(user, unitID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this unit's roster"})
		return
	}

	var officers []models.Officer
	if err := config.DB.Where("unit_id = ?", unitID).Find(&officers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch officers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"officers": officers})
}
