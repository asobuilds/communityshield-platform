package handlers

import (
	"net/http"
	"security-solution/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateCase(c *gin.Context) {
	var input struct {
		Title       string  `json:"title" binding:"required"`
		Description string  `json:"description" binding:"required"`
		Location    string  `json:"location"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		UnitID      string  `json:"unitId"`
		UserID      string  `json:"reportedBy"`
		IsAnonymous bool    `json:"isAnonymous"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var reportedBy uuid.UUID
	if input.UserID != "" {
		parsed, err := uuid.Parse(input.UserID)
		if err == nil {
			reportedBy = parsed
		}
	}
	if reportedBy == uuid.Nil {
		var user models.User
		if err := DB.First(&user).Error; err == nil {
			reportedBy = user.ID
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
	}

	var unitID *uuid.UUID
	if input.UnitID != "" {
		parsed, err := uuid.Parse(input.UnitID)
		if err == nil {
			unitID = &parsed
		}
	}

	newCase := models.Case{
		Title:       input.Title,
		Description: input.Description,
		Location:    input.Location,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		ReportedBy:  reportedBy,
		UnitID:      unitID,
		Status:      "pending",
		Priority:    "medium",
		IsAnonymous: input.IsAnonymous,
	}

	if err := DB.Create(&newCase).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create case"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Case reported successfully",
		"case":    newCase,
	})
}

func GetCase(c *gin.Context) {
	id := c.Param("id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ?", parsed).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Get assigned officers
	var assignedOfficers []models.CaseOfficer
	DB.Where("case_id = ?", parsed).Find(&assignedOfficers)

	c.JSON(http.StatusOK, gin.H{
		"case":             caseItem,
		"assignedOfficers": assignedOfficers,
	})
}

func GetCasesByUser(c *gin.Context) {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userId"})
		return
	}

	var cases []models.Case
	if err := DB.Where("reported_by = ?", userID).Order("created_at desc").Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// 🔥 NEW: Assign officer to case
func AssignOfficerToCase(c *gin.Context) {
	caseID := c.Param("id")
	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var input struct {
		OfficerID string `json:"officerId" binding:"required"`
		Role      string `json:"role"` // primary, investigator, support
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	officerID, err := uuid.Parse(input.OfficerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	// Check case exists
	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ?", parsedCaseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Check officer exists and is an admin
	var officer models.User
	if err := DB.First(&officer, "id = ? AND role = ?", officerID, "unit_admin").Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}

	// Check if already assigned
	var existing models.CaseOfficer
	if err := DB.Where("case_id = ? AND officer_id = ?", parsedCaseID, officerID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Officer already assigned to this case"})
		return
	}

	role := input.Role
	if role == "" {
		role = "investigator"
	}

	assignment := models.CaseOfficer{
		CaseID:    parsedCaseID,
		OfficerID: officerID,
		Role:      role,
	}
	if err := DB.Create(&assignment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign officer"})
		return
	}

	// Send notification to officer
	title := "Case Assigned"
	message := "You have been assigned to case: " + caseItem.Title
	CreateNotification(officerID, parsedCaseID, title, message)

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Officer assigned to case",
		"assignment":   assignment,
	})
}

// 🔥 NEW: Remove officer from case
func RemoveOfficerFromCase(c *gin.Context) {
	caseID := c.Param("id")
	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	officerIDStr := c.Query("officerId")
	if officerIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "officerId is required"})
		return
	}
	officerID, err := uuid.Parse(officerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	if err := DB.Delete(&models.CaseOfficer{}, "case_id = ? AND officer_id = ?", parsedCaseID, officerID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove officer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Officer removed from case"})
}

// 🔥 NEW: Get assigned officers for a case
func GetAssignedOfficers(c *gin.Context) {
	caseID := c.Param("id")
	parsedCaseID, err := uuid.Parse(caseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var assignments []models.CaseOfficer
	if err := DB.Where("case_id = ?", parsedCaseID).Find(&assignments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}

	// Fetch officer details
	var officers []models.User
	for _, a := range assignments {
		var user models.User
		if err := DB.First(&user, "id = ?", a.OfficerID).Error; err == nil {
			officers = append(officers, user)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"assignments": assignments,
		"officers":    officers,
	})
}