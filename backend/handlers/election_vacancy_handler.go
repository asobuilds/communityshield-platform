package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// FillSeatVacancy — head admin of the unit (or super admin) fills a vacant
// seat with the next-highest vote-getter from the original election.
func FillSeatVacancy(c *gin.Context) {
	seatID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seat id"})
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

	var seat models.UnitAdminSeat
	if err := config.DB.First(&seat, "id = ?", seatID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "seat not found"})
		return
	}

	// Auth: super admin, or head admin of the seat's unit.
	if !userObj.IsSuperAdmin && userObj.Role != "super_admin" {
		var membership models.UnitMembership
		if err := config.DB.
			Where("unit_id = ? AND user_id = ? AND status = ? AND is_head_admin = ?",
				seat.UnitID, userObj.ID, models.MembershipActive, true).
			First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the head admin of this unit may fill vacancies"})
			return
		}
	}

	svc := services.NewElectionService()
	if err := svc.FillVacancy(seatID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updated models.UnitAdminSeat
	config.DB.First(&updated, "id = ?", seatID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Vacancy filled",
		"seat":    updated,
	})
}
