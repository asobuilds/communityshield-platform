package handlers

import (
	"net/http"
	"security-solution/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Submit a rating for a unit
func SubmitRating(c *gin.Context) {
	var input struct {
		UnitID  string `json:"unitId" binding:"required"`
		CaseID  string `json:"caseId" binding:"required"`
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Comment string `json:"comment"`
		UserID  string `json:"userId" binding:"required"`
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
	caseID, err := uuid.Parse(input.CaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if the case exists and belongs to this user
	var caseItem models.Case
	if err := DB.Where("id = ? AND reported_by = ?", caseID, userID).First(&caseItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found or you are not the reporter"})
		return
	}

	// Only allow rating if the case is resolved or closed
	if caseItem.Status != "resolved" && caseItem.Status != "closed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You can only rate after the case is resolved or closed"})
		return
	}

	// Check if this case already has a rating
	var existingRating models.Rating
	if err := DB.Where("case_id = ?", caseID).First(&existingRating).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already rated this case"})
		return
	}

	rating := models.Rating{
		UnitID:  unitID,
		UserID:  userID,
		CaseID:  caseID,
		Rating:  input.Rating,
		Comment: input.Comment,
	}

	if err := DB.Create(&rating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit rating"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rating submitted successfully",
		"rating":  rating,
	})
}

// Get average rating for a unit
func GetUnitRating(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedUnitID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var ratings []models.Rating
	if err := DB.Where("unit_id = ?", parsedUnitID).Find(&ratings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ratings"})
		return
	}

	if len(ratings) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"unitId":    parsedUnitID,
			"average":   0.0,
			"count":     0,
			"ratings":   []models.Rating{},
		})
		return
	}

	var total int
	for _, r := range ratings {
		total += r.Rating
	}
	average := float64(total) / float64(len(ratings))

	c.JSON(http.StatusOK, gin.H{
		"unitId":  parsedUnitID,
		"average": average,
		"count":   len(ratings),
		"ratings": ratings,
	})
}

// Get all ratings for a unit (with user info)
func GetUnitReviews(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedUnitID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var ratings []models.Rating
	if err := DB.Where("unit_id = ?", parsedUnitID).Order("created_at desc").Find(&ratings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": ratings})
}

// Check if user has rated a specific case
func CheckUserRating(c *gin.Context) {
	caseID := c.Param("caseId")
	userID := c.Query("userId")
	if caseID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "caseId and userId are required"})
		return
	}

	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var rating models.Rating
	if err := DB.Where("case_id = ? AND user_id = ?", parsedCaseID, parsedUserID).First(&rating).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"rated": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rated": true, "rating": rating})
}