package handlers

import (
	"net/http"
	"security-solution/models"
	"time"
	"github.com/gin-gonic/gin"
)

// GetUnitLeaderboard returns units sorted by resolved cases, ratings, or response time
func GetUnitLeaderboard(c *gin.Context) {
	sortBy := c.DefaultQuery("sortBy", "cases")
	period := c.DefaultQuery("period", "monthly")

	var units []models.Unit
	if err := DB.Where("status = ?", "active").Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch units"})
		return
	}

	type UnitStats struct {
		UnitID        string  `json:"unitId"`
		Name          string  `json:"name"`
		Type          string  `json:"type"`
		Motto         string  `json:"motto"`
		ProfileImage  string  `json:"profileImage"`
		ContactPerson string  `json:"contactPerson"`
		ContactPhone  string  `json:"contactPhone"`
		TotalCases    int64   `json:"totalCases"`
		ResolvedCases int64   `json:"resolvedCases"`
		AvgRating     float64 `json:"avgRating"`
		RatingCount   int64   `json:"ratingCount"`
		Score         float64 `json:"score"`
	}

	var stats []UnitStats

	var startDate time.Time
	switch period {
	case "weekly":
		startDate = time.Now().AddDate(0, 0, -7)
	case "monthly":
		startDate = time.Now().AddDate(0, -1, 0)
	default:
		startDate = time.Time{}
	}

	for _, u := range units {
		var totalCases int64
		query := DB.Model(&models.Case{}).Where("unit_id = ?", u.ID)
		if !startDate.IsZero() {
			query = query.Where("created_at >= ?", startDate)
		}
		query.Count(&totalCases)

		var resolvedCases int64
		queryResolved := DB.Model(&models.Case{}).Where("unit_id = ? AND status = ?", u.ID, "resolved")
		if !startDate.IsZero() {
			queryResolved = queryResolved.Where("updated_at >= ?", startDate)
		}
		queryResolved.Count(&resolvedCases)

		// 🔥 FIX: Use a struct for Scan
		type RatingResult struct {
			Avg   float64
			Count int64
		}
		var result RatingResult
		DB.Model(&models.Rating{}).Where("unit_id = ?", u.ID).
			Select("COALESCE(AVG(rating), 0) as avg, COUNT(*) as count").
			Scan(&result)

		avgRating := result.Avg
		ratingCount := result.Count

		var score float64
		switch sortBy {
		case "rating":
			score = avgRating * 2
		case "response":
			score = float64(resolvedCases)
		default:
			score = float64(resolvedCases)*0.6 + float64(totalCases)*0.4
		}

		stats = append(stats, UnitStats{
			UnitID:        u.ID.String(),
			Name:          u.Name,
			Type:          u.Type,
			Motto:         u.Motto,
			ProfileImage:  u.ProfileImage,
			ContactPerson: u.ContactPerson,
			ContactPhone:  u.ContactPhone,
			TotalCases:    totalCases,
			ResolvedCases: resolvedCases,
			AvgRating:     avgRating,
			RatingCount:   ratingCount,
			Score:         score,
		})
	}

	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[i].Score < stats[j].Score {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": stats,
		"sortBy":      sortBy,
		"period":      period,
	})
}