package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
)

// GetLeaderboard returns the leaderboard for units
func GetLeaderboard(c *gin.Context) {
	// Get case counts by unit
	var unitStats []struct {
		UnitID    string
		UnitName  string
		CaseCount int64
		ResolvedCount int64
	}

	config.DB.Table("cases").
		Select("security_units.id as unit_id, security_units.name as unit_name, count(cases.id) as case_count, sum(case when cases.status = 'resolved' then 1 else 0 end) as resolved_count").
		Joins("left join security_units on security_units.id = cases.unit_id").
		Group("security_units.id, security_units.name").
		Order("case_count desc").
		Limit(10).
		Scan(&unitStats)

	// Calculate ratings
	var leaderboard []gin.H
	for _, stat := range unitStats {
		resolutionRate := 0.0
		if stat.CaseCount > 0 {
			resolutionRate = float64(stat.ResolvedCount) / float64(stat.CaseCount) * 100
		}
		leaderboard = append(leaderboard, gin.H{
			"unitId":        stat.UnitID,
			"unitName":      stat.UnitName,
			"caseCount":     stat.CaseCount,
			"resolvedCount": stat.ResolvedCount,
			"resolutionRate": resolutionRate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
	})
}

// GetOfficerLeaderboard returns the leaderboard for officers
func GetOfficerLeaderboard(c *gin.Context) {
	var officerStats []struct {
		OfficerID   string
		OfficerName string
		CaseCount   int64
		ResolvedCount int64
	}

	config.DB.Table("cases").
		Select("officers.id as officer_id, officers.name as officer_name, count(cases.id) as case_count, sum(case when cases.status = 'resolved' then 1 else 0 end) as resolved_count").
		Joins("left join officers on officers.id = cases.assigned_to").
		Group("officers.id, officers.name").
		Order("case_count desc").
		Limit(10).
		Scan(&officerStats)

	var leaderboard []gin.H
	for _, stat := range officerStats {
		resolutionRate := 0.0
		if stat.CaseCount > 0 {
			resolutionRate = float64(stat.ResolvedCount) / float64(stat.CaseCount) * 100
		}
		leaderboard = append(leaderboard, gin.H{
			"officerId":     stat.OfficerID,
			"officerName":   stat.OfficerName,
			"caseCount":     stat.CaseCount,
			"resolvedCount": stat.ResolvedCount,
			"resolutionRate": resolutionRate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
	})
}

// GetUnitRanking returns ranking for a specific unit
func GetUnitRanking(c *gin.Context) {
	unitID := c.Param("unitId")

	var caseCount int64
	var resolvedCount int64

	config.DB.Model(&models.Case{}).Where("unit_id = ?", unitID).Count(&caseCount)
	config.DB.Model(&models.Case{}).Where("unit_id = ? AND status = ?", unitID, "resolved").Count(&resolvedCount)

	resolutionRate := 0.0
	if caseCount > 0 {
		resolutionRate = float64(resolvedCount) / float64(caseCount) * 100
	}

	// Get rank
	var rank int64
	config.DB.Raw(`
		SELECT COUNT(*) + 1 as rank 
		FROM (
			SELECT unit_id, COUNT(*) as case_count 
			FROM cases 
			GROUP BY unit_id 
			HAVING COUNT(*) > (
				SELECT COUNT(*) 
				FROM cases 
				WHERE unit_id = ? 
				GROUP BY unit_id
			)
		) as ranked_units
	`, unitID).Scan(&rank)

	c.JSON(http.StatusOK, gin.H{
		"unitId":        unitID,
		"caseCount":     caseCount,
		"resolvedCount": resolvedCount,
		"resolutionRate": resolutionRate,
		"rank":          rank,
	})
}