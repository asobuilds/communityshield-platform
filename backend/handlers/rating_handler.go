package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// GetUnitRatings returns ratings for a unit
func GetUnitRatings(c *gin.Context) {
	unitID := c.Param("unitId")
	id, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var ratings []models.Rating
	if err := config.DB.Where("target_id = ? AND target_type = ?", id, "unit").Find(&ratings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ratings"})
		return
	}

	var average float64
	var count int64
	if len(ratings) > 0 {
		var sum int
		for _, r := range ratings {
			sum += r.Rating
		}
		average = float64(sum) / float64(len(ratings))
		count = int64(len(ratings))
	}

	c.JSON(http.StatusOK, gin.H{
		"average": average,
		"count":   count,
		"ratings": ratings,
	})
}

// SubmitRating submits a rating for a unit
func SubmitRating(c *gin.Context) {
	var input struct {
		TargetID   string `json:"targetId" binding:"required"`
		TargetType string `json:"targetType" binding:"required"`
		Rating     int    `json:"rating" binding:"required,min=1,max=5"`
		Comment    string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetID, err := uuid.Parse(input.TargetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	// Check if rating already exists
	var existing models.Rating
	if err := config.DB.Where("user_id = ? AND target_id = ? AND target_type = ?", userObj.ID, targetID, input.TargetType).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already rated this"})
		return
	}

	rating := models.Rating{
		UserID:     userObj.ID,
		TargetID:   targetID,
		TargetType: input.TargetType,
		Rating:     input.Rating,
		Comment:    input.Comment,
	}

	if err := config.DB.Create(&rating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit rating"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rating submitted successfully",
		"rating":  rating,
	})
}