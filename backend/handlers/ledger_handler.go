package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// GetPublicUnitLedger returns the ledger for a unit — public read.
// Only confirmed/posted entries are returned. No sensitive fields.
func GetPublicUnitLedger(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	year := 0
	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}

	svc := services.NewLedgerService()
	entries, err := svc.ListForUnit(unitID, year, 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load ledger"})
		return
	}

	if year == 0 {
		year = int(entries[0].Year)
		for _, e := range entries {
			if e.Year > year {
				year = int(e.Year)
			}
		}
	}
	summary, _ := svc.SummaryForUnitYear(unitID, year)

	c.JSON(http.StatusOK, gin.H{
		"unitId":  unitID,
		"year":    year,
		"summary": summary,
		"entries": entries,
	})
}

// GetUnitLedger returns the ledger to an authenticated member of the unit.
// Citizens are blocked. Super admin sees any unit.
func GetUnitLedger(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("unitId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj, ok := userValue.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	if !canViewUnitLedger(userObj, unitID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this unit's ledger"})
		return
	}

	year := 0
	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}

	svc := services.NewLedgerService()
	entries, err := svc.ListForUnit(unitID, year, 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load ledger"})
		return
	}

	if year == 0 {
		year = int(entries[0].Year)
		for _, e := range entries {
			if e.Year > year {
				year = int(e.Year)
			}
		}
	}
	summary, _ := svc.SummaryForUnitYear(unitID, year)

	c.JSON(http.StatusOK, gin.H{
		"unitId":  unitID,
		"year":    year,
		"summary": summary,
		"entries": entries,
	})
}

// canViewUnitLedger returns true if the user may view the unit's ledger.
// Super admin: any unit. Officer / admin / head admin: only their own unit.
// Citizens: never.
func canViewUnitLedger(user *models.User, unitID uuid.UUID) bool {
	if user == nil {
		return false
	}
	if user.IsSuperAdmin || user.Role == "super_admin" {
		return true
	}
	if user.Role == "citizen" {
		return false
	}
	var membership models.UnitMembership
	if err := config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ?",
			unitID, user.ID, models.MembershipActive).
		First(&membership).Error; err != nil {
		return false
	}
	return true
}
