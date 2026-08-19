package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// CreatePeaceCommittee creates a peace committee
func CreatePeaceCommittee(c *gin.Context) {
	var input struct {
		UnitID      string `json:"unitId" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Location    string `json:"location"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create peace committees"})
		return
	}

	unitID, err := uuid.Parse(input.UnitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	committee := models.PeaceCommittee{
		UnitID:      unitID,
		Name:        input.Name,
		Description: input.Description,
		Location:    input.Location,
		Status:      "active",
		CreatedBy:   userObj.ID,
	}

	if err := config.DB.Create(&committee).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create peace committee"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Peace committee created",
		"committee": committee,
	})
}

// GetPeaceCommittees gets all peace committees
func GetPeaceCommittees(c *gin.Context) {
	unitID := c.Param("unitId")
	id, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission"})
		return
	}

	var committees []models.PeaceCommittee
	if err := config.DB.Preload("Members").Where("unit_id = ?", id).Find(&committees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch committees"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"committees": committees,
	})
}

// AddCommitteeMember adds a member to a peace committee
func AddCommitteeMember(c *gin.Context) {
	committeeID := c.Param("id")
	id, err := uuid.Parse(committeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid committee ID"})
		return
	}

	var input struct {
		UserID string `json:"userId" binding:"required"`
		Role   string `json:"role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can add committee members"})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if input.Role == "" {
		input.Role = "member"
	}

	member := models.CommitteeMember{
		CommitteeID: id,
		UserID:      userID,
		Role:        input.Role,
		Status:      "active",
		JoinedAt:    time.Now(),
	}

	if err := config.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Member added successfully",
		"member":  member,
	})
}

// CreateConflictResolution creates a conflict resolution case
func CreateConflictResolution(c *gin.Context) {
	var input struct {
		CaseID      string `json:"caseId"`
		UnitID      string `json:"unitId" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description" binding:"required"`
		Parties     string `json:"parties"`
		Priority    string `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create conflict resolutions"})
		return
	}

	unitID, err := uuid.Parse(input.UnitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	if input.Priority == "" {
		input.Priority = "medium"
	}

	conflict := models.ConflictResolution{
		UnitID:      unitID,
		Title:       input.Title,
		Description: input.Description,
		Parties:     input.Parties,
		Status:      "open",
		Priority:    input.Priority,
		CreatedBy:   userObj.ID,
	}

	if input.CaseID != "" {
		caseID, err := uuid.Parse(input.CaseID)
		if err == nil {
			conflict.CaseID = &caseID
		}
	}

	if err := config.DB.Create(&conflict).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conflict resolution"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Conflict resolution created",
		"conflict": conflict,
	})
}

// GetConflictResolutions gets all conflict resolutions
func GetConflictResolutions(c *gin.Context) {
	unitID := c.Param("unitId")
	id, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission"})
		return
	}

	var conflicts []models.ConflictResolution
	if err := config.DB.Preload("Mediator").Where("unit_id = ?", id).Order("created_at desc").Find(&conflicts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conflicts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conflicts": conflicts,
	})
}

// UpdateConflictResolution updates a conflict resolution
func UpdateConflictResolution(c *gin.Context) {
	id := c.Param("id")
	conflictID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conflict ID"})
		return
	}

	var input struct {
		Status     string `json:"status"`
		MediatorID string `json:"mediatorId"`
		Resolution string `json:"resolution"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update conflicts"})
		return
	}

	var conflict models.ConflictResolution
	if err := config.DB.First(&conflict, "id = ?", conflictID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conflict not found"})
		return
	}

	if input.Status != "" {
		conflict.Status = input.Status
		if input.Status == "resolved" || input.Status == "closed" {
			now := time.Now()
			conflict.ResolvedAt = &now
		}
	}

	if input.MediatorID != "" {
		mediatorID, err := uuid.Parse(input.MediatorID)
		if err == nil {
			conflict.MediatorID = &mediatorID
		}
	}

	if input.Resolution != "" {
		conflict.Resolution = input.Resolution
	}

	if err := config.DB.Save(&conflict).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update conflict"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Conflict updated",
		"conflict": conflict,
	})
}

// GetPeaceMetrics gets peace metrics for a unit
func GetPeaceMetrics(c *gin.Context) {
	unitID := c.Param("unitId")
	id, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission"})
		return
	}

	var metrics []models.PeaceMetric
	if err := config.DB.Where("unit_id = ?", id).Order("month desc").Limit(12).Find(&metrics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch peace metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

// GetTrustScores gets community trust scores
func GetTrustScores(c *gin.Context) {
	unitID := c.Param("unitId")
	id, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission"})
		return
	}

	var scores []models.CommunityTrustScore
	if err := config.DB.Where("unit_id = ?", id).Order("score desc").Find(&scores).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trust scores"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"trustScores": scores,
	})
}

// UpdateTrustScore updates a community trust score
func UpdateTrustScore(c *gin.Context) {
	var input struct {
		UnitID     string  `json:"unitId"`
		OfficerID  string  `json:"officerId"`
		Score      float64 `json:"score" binding:"required,min=0,max=5"`
		Category   string  `json:"category"`
		RatingType string  `json:"ratingType"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update trust scores"})
		return
	}

	if input.Category == "" {
		input.Category = "effectiveness"
	}

	var trustScore models.CommunityTrustScore

	if input.RatingType == "unit" && input.UnitID != "" {
		unitID, err := uuid.Parse(input.UnitID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
			return
		}
		trustScore.UnitID = &unitID
	} else if input.RatingType == "officer" && input.OfficerID != "" {
		officerID, err := uuid.Parse(input.OfficerID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
			return
		}
		trustScore.OfficerID = &officerID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Must specify either unitId or officerId"})
		return
	}

	trustScore.Score = input.Score
	trustScore.RatingCount = 1
	trustScore.Category = input.Category
	trustScore.Trend = "stable"
	trustScore.LastUpdated = time.Now()

	if err := config.DB.Create(&trustScore).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update trust score"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Trust score updated",
		"trustScore": trustScore,
	})
}