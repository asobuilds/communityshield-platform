package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

// SubmitRating — citizen rates an officer or unit after a closed case.
func SubmitRating(c *gin.Context) {
	if !requireMinorApproved(c) {
		return
	}
	var input struct {
		CaseID     string `json:"caseId" binding:"required"`
		TargetID   string `json:"targetId" binding:"required"`
		TargetType string `json:"targetType" binding:"required"`
		Rating     int    `json:"rating" binding:"required,min=1,max=5"`
		Comment    string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	caseID, err := uuid.Parse(input.CaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}
	targetID, err := uuid.Parse(input.TargetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target ID"})
		return
	}

	svc := services.NewRatingService()
	rating, err := svc.SubmitRating(services.SubmitRatingRequest{
		RaterUserID: userObj.ID,
		CaseID:      caseID,
		TargetID:    targetID,
		TargetType:  input.TargetType,
		Rating:      input.Rating,
		Comment:     input.Comment,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rating submitted. Thank you.",
		"rating":  rating,
	})
}

// GetOfficerRating — public aggregate rating for an officer.
func GetOfficerRating(c *gin.Context) {
	officerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	var score models.OfficerScore
	if err := config.DB.Where("officer_id = ?", officerID).First(&score).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"officerId":     officerID,
			"bayesianScore": services.BayesianPriorMean,
			"ratingCount":   0,
			"tier":          "unranked",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"officerId":     score.OfficerID,
		"unitId":        score.UnitID,
		"bayesianScore": score.BayesianScore,
		"ratingCount":   score.RatingCount,
		"tier":          score.Tier,
		"lastUpdated":   score.LastComputedAt,
	})
}

// GetUnitRating — public aggregate rating for a unit.
func GetUnitRating(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var score models.UnitScore
	if err := config.DB.Where("unit_id = ?", unitID).First(&score).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"unitId":         unitID,
			"compositeScore": services.BayesianPriorMean,
			"directCount":    0,
			"officerCount":   0,
			"tier":           "unranked",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"unitId":         score.UnitID,
		"compositeScore": score.CompositeScore,
		"officerAvg":     score.OfficerAvg,
		"officerCount":   score.OfficerCount,
		"directAvg":      score.DirectAvg,
		"directCount":    score.DirectCount,
		"velocityRatio":  score.VelocityRatio,
		"tier":           score.Tier,
		"lastUpdated":    score.LastComputedAt,
	})
}

// GetPublicLeaderboard — public. Top 50 units by composite score.
func GetPublicLeaderboard(c *gin.Context) {
	var scores []models.UnitScore
	config.DB.
		Where("composite_score > ?", 0).
		Order("composite_score DESC, direct_count DESC").
		Limit(50).
		Find(&scores)

	type Row struct {
		UnitID         uuid.UUID `json:"unitId"`
		CompositeScore float64   `json:"compositeScore"`
		Tier           string    `json:"tier"`
		OfficerCount   int       `json:"officerCount"`
		DirectCount    int       `json:"directCount"`
		UnitName       string    `json:"unitName,omitempty"`
	}

	rows := make([]Row, 0, len(scores))
	for _, s := range scores {
		var unit models.SecurityUnit
		config.DB.First(&unit, "id = ?", s.UnitID)
		rows = append(rows, Row{
			UnitID:         s.UnitID,
			CompositeScore: s.CompositeScore,
			Tier:           s.Tier,
			OfficerCount:   s.OfficerCount,
			DirectCount:    s.DirectCount,
			UnitName:       unit.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": rows})
}

// GetOfficersInUnitRanking — members-only view of a unit's officer leaderboard.
// Citizens blocked; only members of the unit (and super admin) can view.
func GetOfficersInUnitRanking(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
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

	if !userObj.IsSuperAdmin && userObj.Role != "super_admin" {
		if userObj.Role == "citizen" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Members only"})
			return
		}
		var membership models.UnitMembership
		if err := config.DB.
			Where("unit_id = ? AND user_id = ? AND status = ?", unitID, userObj.ID, models.MembershipActive).
			First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this unit"})
			return
		}
	}

	var scores []models.OfficerScore
	config.DB.
		Where("unit_id = ?", unitID).
		Order("bayesian_score DESC, rating_count DESC").
		Find(&scores)

	type Row struct {
		OfficerID      uuid.UUID `json:"officerId"`
		BayesianScore  float64   `json:"bayesianScore"`
		RatingCount    int       `json:"ratingCount"`
		Tier           string    `json:"tier"`
		OfficerName    string    `json:"officerName,omitempty"`
	}
	rows := make([]Row, 0, len(scores))
	for _, s := range scores {
		var officer models.Officer
		config.DB.First(&officer, "id = ?", s.OfficerID)
		rows = append(rows, Row{
			OfficerID:     s.OfficerID,
			BayesianScore: s.BayesianScore,
			RatingCount:   s.RatingCount,
			Tier:          s.Tier,
			OfficerName:   officer.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{"unitId": unitID, "officers": rows})
}

// FlagRating — officer or admin flags a rating for review.
func FlagRating(c *gin.Context) {
	ratingID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating ID"})
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

	var input struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		return
	}

	svc := services.NewRatingService()
	if err := svc.FlagRating(ratingID, userObj.ID, input.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating flagged for review"})
}

// SuggestUnits — public. Ranks units by composite score, then by proximity
// if lat/lng are provided. Returns top 10.
func SuggestUnits(c *gin.Context) {
	var scores []models.UnitScore
	config.DB.
		Order("composite_score DESC, direct_count DESC").
		Limit(10).
		Find(&scores)

	type Row struct {
		UnitID         uuid.UUID `json:"unitId"`
		UnitName       string    `json:"unitName"`
		CompositeScore float64   `json:"compositeScore"`
		Tier           string    `json:"tier"`
	}

	rows := make([]Row, 0, len(scores))
	for _, s := range scores {
		var unit models.SecurityUnit
		if err := config.DB.First(&unit, "id = ?", s.UnitID).Error; err != nil {
			continue
		}
		if unit.Status != "active" {
			continue
		}
		rows = append(rows, Row{
			UnitID:         s.UnitID,
			UnitName:       unit.Name,
			CompositeScore: s.CompositeScore,
			Tier:           s.Tier,
		})
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": rows})
}
