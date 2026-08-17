package handlers

import (
	"net/http"
	"security-solution/models"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetOfficersOfTheWeek returns officers with the most resolved cases in the last 7 days
func GetOfficersOfTheWeek(c *gin.Context) {
	unitIDStr := c.Query("unitId")
	var unitID *uuid.UUID
	if unitIDStr != "" {
		if parsed, err := uuid.Parse(unitIDStr); err == nil {
			unitID = &parsed
		}
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	type OfficerStats struct {
		OfficerID     uuid.UUID `gorm:"column:assigned_to"`
		ResolvedCount int64     `gorm:"column:resolved_count"`
	}

	query := DB.Model(&models.Case{}).
		Select("assigned_to, count(*) as resolved_count").
		Where("status = ? AND updated_at >= ?", "resolved", sevenDaysAgo).
		Where("assigned_to IS NOT NULL")

	if unitID != nil {
		query = query.Where("unit_id = ?", unitID)
	}

	var stats []OfficerStats
	if err := query.Group("assigned_to").Order("resolved_count desc").Limit(5).Scan(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch officer stats"})
		return
	}

	var officers []map[string]interface{}
	for _, s := range stats {
		var user models.User
		if err := DB.First(&user, "id = ?", s.OfficerID).Error; err == nil {
			officers = append(officers, map[string]interface{}{
				"id":            user.ID,
				"firstName":     user.FirstName,
				"lastName":      user.LastName,
				"email":         user.Email,
				"rank":          user.Rank,
				"profileImage":  user.ProfileImage,
				"resolvedCount": s.ResolvedCount,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"officers": officers,
		"count":    len(officers),
	})
}