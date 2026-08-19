package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// GetAllCases - Get all cases with optional filtering
func GetAllCases(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var cases []models.Case
	query := config.DB.Preload("Evidence").Preload("Progress").Order("created_at desc")

	// Filter based on role
	if userObj.Role == "citizen" {
		// Citizens see only their own cases
		query = query.Where("reported_by = ?", userObj.ID)
	} else if userObj.Role == "officer" && userObj.UnitID != nil {
		// Officers see cases in their unit
		query = query.Where("unit_id = ?", userObj.UnitID)
	} else if userObj.Role == "unit_admin" && userObj.UnitID != nil {
		// Unit admins see cases in their unit
		query = query.Where("unit_id = ?", userObj.UnitID)
	}
	// Super admin sees all cases

	if err := query.Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cases": cases,
	})
}

// CreateCase - Enhanced with automatic timeline and notifications
func CreateCase(c *gin.Context) {
	var input struct {
		UnitID      string  `json:"unitId"`
		Title       string  `json:"title" binding:"required"`
		Description string  `json:"description" binding:"required"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Location    string  `json:"location"`
		Priority    string  `json:"priority"`
		TemplateID  string  `json:"templateId"`
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

	priority := input.Priority
	if priority == "" {
		priority = "medium"
	}

	caseObj := models.Case{
		Title:       input.Title,
		Description: input.Description,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		Location:    input.Location,
		Status:      "pending",
		Priority:    priority,
		IsPublic:    true,
		ReportedBy:  userObj.ID,
	}

	if input.UnitID != "" {
		unitID, _ := uuid.Parse(input.UnitID)
		caseObj.UnitID = unitID
	}

	if err := config.DB.Create(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create case"})
		return
	}

	// Create timeline entry
	timeline := models.CaseTimeline{
		CaseID:      caseObj.ID,
		UserID:      userObj.ID,
		Action:      "created",
		Description: "Case reported by citizen",
		Status:      "pending",
	}
	config.DB.Create(&timeline)

	// Send notification to citizen (async)
	go notifyCitizen(userObj.ID, "Case Reported", "Your case has been received. We'll update you shortly.")

	// Send notification to unit admins (async)
	if caseObj.UnitID != uuid.Nil {
		go notifyUnitAdmins(caseObj.UnitID, "New Case Assigned", "A new case has been assigned to your unit.")
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Case reported successfully",
		"case":    caseObj,
	})
}

// GetCaseByID - Enhanced with timeline and feedback
func GetCaseByID(c *gin.Context) {
	id := c.Param("id")
	caseID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var caseObj models.Case
	if err := config.DB.Preload("Evidence").Preload("Progress").First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	// Check permissions
	if userObj.Role == "citizen" && caseObj.ReportedBy != userObj.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this case"})
		return
	}

	// Get timeline
	var timeline []models.CaseTimeline
	config.DB.Where("case_id = ?", caseID).Order("created_at asc").Find(&timeline)

	// Get feedback
	var feedback []models.CaseFeedback
	config.DB.Where("case_id = ?", caseID).Find(&feedback)

	c.JSON(http.StatusOK, gin.H{
		"case":      caseObj,
		"timeline":  timeline,
		"feedback":  feedback,
	})
}

// UpdateCaseStatus - Enhanced with timeline and notifications
func UpdateCaseStatus(c *gin.Context) {
	id := c.Param("id")
	caseID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var input struct {
		Status      string `json:"status" binding:"required"`
		Description string `json:"description"`
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

	// Only officers and admins can update status
	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" && userObj.Role != "officer" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update case status"})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	oldStatus := caseObj.Status
	caseObj.Status = input.Status
	if err := config.DB.Save(&caseObj).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update case"})
		return
	}

	// Create timeline entry
	desc := input.Description
	if desc == "" {
		desc = "Status updated from " + oldStatus + " to " + input.Status
	}
	timeline := models.CaseTimeline{
		CaseID:      caseID,
		UserID:      userObj.ID,
		Action:      "status_update",
		Description: desc,
		Status:      input.Status,
	}
	config.DB.Create(&timeline)

	// Notify citizen
	if caseObj.ReportedBy != uuid.Nil {
		go notifyCitizen(caseObj.ReportedBy, "Case Status Updated", "Your case has been updated to: "+input.Status)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case status updated successfully",
		"case":    caseObj,
	})
}

// AddCaseTimeline - Add manual timeline entry
func AddCaseTimeline(c *gin.Context) {
	id := c.Param("id")
	caseID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var input struct {
		Action      string `json:"action" binding:"required"`
		Description string `json:"description" binding:"required"`
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

	// Only officers and admins can add timeline entries
	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" && userObj.Role != "officer" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add timeline entries"})
		return
	}

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	timeline := models.CaseTimeline{
		CaseID:      caseID,
		UserID:      userObj.ID,
		Action:      input.Action,
		Description: input.Description,
		Status:      caseObj.Status,
	}
	config.DB.Create(&timeline)

	// Notify citizen
	if caseObj.ReportedBy != uuid.Nil {
		go notifyCitizen(caseObj.ReportedBy, "Case Progress Updated", input.Description)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Timeline entry added",
		"timeline": timeline,
	})
}

// GetCaseTimeline - Get case timeline
func GetCaseTimeline(c *gin.Context) {
	id := c.Param("id")
	caseID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var timeline []models.CaseTimeline
	if err := config.DB.Preload("User").Where("case_id = ?", caseID).Order("created_at asc").Find(&timeline).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch timeline"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"timeline": timeline,
	})
}

// SubmitCaseFeedback - Citizen feedback on case resolution
func SubmitCaseFeedback(c *gin.Context) {
	id := c.Param("id")
	caseID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var input struct {
		Rating     int    `json:"rating" binding:"required,min=1,max=5"`
		Comment    string `json:"comment"`
		Categories string `json:"categories"`
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

	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	if caseObj.ReportedBy != userObj.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to provide feedback for this case"})
		return
	}

	var existing models.CaseFeedback
	if err := config.DB.Where("case_id = ? AND user_id = ?", caseID, userObj.ID).First(&existing).Error; err == nil {
		existing.Rating = input.Rating
		existing.Comment = input.Comment
		existing.Categories = input.Categories
		config.DB.Save(&existing)
		c.JSON(http.StatusOK, gin.H{
			"message":  "Feedback updated",
			"feedback": existing,
		})
		return
	}

	feedback := models.CaseFeedback{
		CaseID:     caseID,
		UserID:     userObj.ID,
		Rating:     input.Rating,
		Comment:    input.Comment,
		Categories: input.Categories,
		IsPublic:   true,
	}
	config.DB.Create(&feedback)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Feedback submitted",
		"feedback": feedback,
	})
}

// GetCaseAnalytics - Get case performance metrics
func GetCaseAnalytics(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	if userObj.Role != "super_admin" && userObj.Role != "unit_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view analytics"})
		return
	}

	query := config.DB.Model(&models.Case{})
	if userObj.Role == "unit_admin" && userObj.UnitID != nil {
		query = query.Where("unit_id = ?", userObj.UnitID)
	}

	var totalCases int64
	var resolvedCases int64
	var pendingCases int64

	query.Count(&totalCases)
	query.Where("status = ?", "resolved").Count(&resolvedCases)
	query.Where("status = ?", "pending").Count(&pendingCases)

	c.JSON(http.StatusOK, gin.H{
		"totalCases":     totalCases,
		"resolvedCases":  resolvedCases,
		"pendingCases":   pendingCases,
		"resolutionRate": float64(resolvedCases) / float64(totalCases) * 100,
	})
}

// Helper: Notify citizen
func notifyCitizen(userID uuid.UUID, title, message string) {
	notification := models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    "case_update",
		Status:  "unread",
	}
	config.DB.Create(&notification)
}

// Helper: Notify unit admins
func notifyUnitAdmins(unitID uuid.UUID, title, message string) {
	var admins []models.User
	config.DB.Where("unit_id = ? AND role IN (?)", unitID, []string{"unit_admin", "super_admin"}).Find(&admins)

	for _, admin := range admins {
		notification := models.Notification{
			UserID:  admin.ID,
			Title:   title,
			Message: message,
			Type:    "case_assignment",
			Status:  "unread",
		}
		config.DB.Create(&notification)
	}
}