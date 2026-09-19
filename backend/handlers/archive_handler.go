package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// NOTE (security): `archived` is not one of the 8 canonical case lifecycle
// states (pending / assigned / dispatched / on_scene / investigating /
// pending_admin_review / admin_changes_requested / closed). These handlers
// can bypass the closure-approval workflow by flipping a case back to
// pending. Access is restricted to super admin and head admin of the
// affected unit. A future wave should either remove them or fold archiving
// into the canonical lifecycle.

// requireArchiveAccess returns the calling user if they may archive or read
// archived cases for the given unit. Super admin passes for any unit. A head
// admin passes only for their own unit.
func requireArchiveAccess(c *gin.Context, caseUnitID uuid.UUID) (*models.User, bool) {
	value, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil, false
	}
	user, ok := value.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return nil, false
	}

	if user.IsSuperAdmin || user.Role == "super_admin" {
		return user, true
	}

	var membership models.UnitMembership
	err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ? AND is_head_admin = ?",
			caseUnitID, user.ID, models.MembershipActive, true).
		First(&membership).Error
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super admin or head admin of this unit may archive"})
		return nil, false
	}

	return user, true
}

// requireArchiveAccessAnyUnit returns the calling user plus the set of unit IDs
// they may read archived cases for. Super admin gets nil (means "all units").
// Head admin gets a slice with their one unit. Anyone else → 403.
func requireArchiveAccessAnyUnit(c *gin.Context) (*models.User, []uuid.UUID, bool) {
	value, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil, nil, false
	}
	user, ok := value.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return nil, nil, false
	}

	if user.IsSuperAdmin || user.Role == "super_admin" {
		return user, nil, true
	}

	var memberships []models.UnitMembership
	if err := config.DB.
		Where("user_id = ? AND status = ? AND is_head_admin = ?",
			user.ID, models.MembershipActive, true).
		Find(&memberships).Error; err != nil || len(memberships) == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super admin or head admin may access archived cases"})
		return nil, nil, false
	}

	units := make([]uuid.UUID, 0, len(memberships))
	for _, m := range memberships {
		units = append(units, m.UnitID)
	}
	return user, units, true
}

// ArchiveCase archives a case. Super admin or head admin of the case's unit.
func ArchiveCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	if _, ok := requireArchiveAccess(c, caseObj.UnitID); !ok {
		return
	}

	if caseObj.Status == "archived" {
		c.JSON(http.StatusConflict, gin.H{"error": "Case is already archived"})
		return
	}

	caseObj.Status = "archived"
	if err := config.DB.Save(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case archived successfully",
		"case":    caseObj,
	})
}

// UnarchiveCase restores an archived case to pending.
// Super admin or head admin of the case's unit.
func UnarchiveCase(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	if _, ok := requireArchiveAccess(c, caseObj.UnitID); !ok {
		return
	}

	if caseObj.Status != "archived" {
		c.JSON(http.StatusConflict, gin.H{"error": "Only archived cases can be unarchived"})
		return
	}

	caseObj.Status = "pending"
	if err := config.DB.Save(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unarchive case"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case unarchived successfully",
		"case":    caseObj,
	})
}

// GetArchivedCases returns archived cases.
// Super admin sees all. Head admin sees only their unit's archived cases.
func GetArchivedCases(c *gin.Context) {
	_, unitIDs, ok := requireArchiveAccessAnyUnit(c)
	if !ok {
		return
	}

	query := config.DB.Where("status = ?", "archived").
		Preload("Evidence").Preload("Progress")

	if unitIDs != nil {
		query = query.Where("unit_id IN ?", unitIDs)
	}

	var cases []models.Case
	if err := query.Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch archived cases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}