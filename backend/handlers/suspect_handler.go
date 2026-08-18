package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// checkSuspectAccess checks if user has access to this suspect
func checkSuspectAccess(user *models.User, suspect *models.Suspect) bool {
	// Super admin has full access to everything
	if user.Role == "super_admin" {
		return true
	}

	// Unit admin can only access suspects in their own unit
	if user.Role == "unit_admin" {
		if user.UnitID == nil {
			return false
		}
		// Can access suspects in their unit OR suspects transferred to them
		if suspect.UnitID != nil && *suspect.UnitID == *user.UnitID {
			return true
		}
		// Also can access suspects that have a pending transfer to their unit
		if suspect.TransferStatus == "pending" && suspect.TransferToUnit != nil && *suspect.TransferToUnit == *user.UnitID {
			return true
		}
		return false
	}

	// Officer can only access suspects in their unit AND cases they're assigned to
	if user.Role == "officer" {
		if user.UnitID == nil {
			return false
		}
		// Must be in same unit
		if suspect.UnitID == nil || *suspect.UnitID != *user.UnitID {
			return false
		}
		// Check if officer is assigned to any case with this suspect
		var count int64
		config.DB.Model(&models.SuspectCase{}).
			Joins("JOIN cases ON cases.id = suspect_cases.case_id").
			Where("suspect_cases.suspect_id = ? AND cases.assigned_to = ?", suspect.ID, user.ID).
			Count(&count)
		if count > 0 {
			return true
		}
		// Unit admins in same unit can also view (officers can view all suspects in their unit if they have cases)
		return false
	}

	// Citizens have NO access to suspects
	return false
}

// checkAdminAccess checks if user is admin (super_admin or unit_admin)
func checkAdminAccess(user *models.User) bool {
	return user.Role == "super_admin" || user.Role == "unit_admin"
}

// CreateSuspect creates a new suspect (Admin only)
func CreateSuspect(c *gin.Context) {
	var input struct {
		FirstName   string `json:"firstName" binding:"required"`
		LastName    string `json:"lastName" binding:"required"`
		Alias       string `json:"alias"`
		Gender      string `json:"gender"`
		DateOfBirth string `json:"dateOfBirth"`
		Nationality string `json:"nationality"`
		IDNumber    string `json:"idNumber"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		Address     string `json:"address"`
		Description string `json:"description"`
		DangerLevel string `json:"dangerLevel"`
		Wanted      bool   `json:"wanted"`
		UnitID      string `json:"unitId"`
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

	// Only admins can create suspects
	if !checkAdminAccess(userObj) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create suspects"})
		return
	}

	var dob time.Time
	if input.DateOfBirth != "" {
		parsed, _ := time.Parse("2006-01-02", input.DateOfBirth)
		dob = parsed
	}

	var unitID *uuid.UUID
	if input.UnitID != "" {
		parsed, err := uuid.Parse(input.UnitID)
		if err == nil {
			unitID = &parsed
		}
	} else if userObj.UnitID != nil {
		// Default to admin's unit if not specified
		unitID = userObj.UnitID
	}

	suspect := models.Suspect{
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		Alias:       input.Alias,
		Gender:      input.Gender,
		DateOfBirth: dob,
		Nationality: input.Nationality,
		IDNumber:    input.IDNumber,
		Phone:       input.Phone,
		Email:       input.Email,
		Address:     input.Address,
		Description: input.Description,
		Status:      "active",
		DangerLevel: input.DangerLevel,
		Wanted:      input.Wanted,
		CreatedBy:   userObj.ID,
		UnitID:      unitID,
	}

	if suspect.DangerLevel == "" {
		suspect.DangerLevel = "medium"
	}

	if err := config.DB.Create(&suspect).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create suspect"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Suspect created successfully",
		"suspect": suspect,
	})
}

// GetAllSuspects gets all suspects (filtered by user access)
func GetAllSuspects(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var suspects []models.Suspect
	query := config.DB.Preload("CreatedByUser").Preload("Unit").Preload("TransferToUnitObj")

	// Role-based filtering
	if userObj.Role == "super_admin" {
		// Super admin sees ALL suspects across all units
		query = query.Order("created_at desc")
	} else if userObj.Role == "unit_admin" && userObj.UnitID != nil {
		// Unit admin sees suspects in their unit AND suspects transferred to them
		query = query.Where(
			"unit_id = ? OR (transfer_status = 'pending' AND transfer_to_unit = ?)",
			userObj.UnitID, userObj.UnitID,
		).Order("created_at desc")
	} else if userObj.Role == "officer" && userObj.UnitID != nil {
		// Officer sees suspects in their unit that are linked to cases assigned to them
		var caseSuspectIDs []string
		config.DB.Table("suspect_cases").
			Joins("JOIN cases ON cases.id = suspect_cases.case_id").
			Where("cases.assigned_to = ?", userObj.ID).
			Pluck("suspect_cases.suspect_id", &caseSuspectIDs)

		query = query.Where(
			"unit_id = ? AND id IN (?)",
			userObj.UnitID, caseSuspectIDs,
		).Order("created_at desc")
	} else {
		// Citizens and others have NO access
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view suspects"})
		return
	}

	if err := query.Find(&suspects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suspects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suspects": suspects,
	})
}

// GetSuspectByID gets a specific suspect (with access check)
func GetSuspectByID(c *gin.Context) {
	id := c.Param("id")
	suspectID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suspect ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var suspect models.Suspect
	if err := config.DB.Preload("CreatedByUser").Preload("Unit").Preload("TransferToUnitObj").First(&suspect, "id = ?", suspectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suspect not found"})
		return
	}

	// Check access
	if !checkSuspectAccess(userObj, &suspect) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this suspect"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suspect": suspect,
	})
}

// UpdateSuspect updates a suspect (with access check)
func UpdateSuspect(c *gin.Context) {
	id := c.Param("id")
	suspectID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suspect ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	// Only admins can update suspects
	if !checkAdminAccess(userObj) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can update suspects"})
		return
	}

	var input struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Alias       string `json:"alias"`
		Gender      string `json:"gender"`
		Nationality string `json:"nationality"`
		IDNumber    string `json:"idNumber"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		Address     string `json:"address"`
		Description string `json:"description"`
		Status      string `json:"status"`
		DangerLevel string `json:"dangerLevel"`
		Wanted      *bool  `json:"wanted"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var suspect models.Suspect
	if err := config.DB.First(&suspect, "id = ?", suspectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suspect not found"})
		return
	}

	// Check if admin has access to this suspect
	if !checkSuspectAccess(userObj, &suspect) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this suspect"})
		return
	}

	if input.FirstName != "" {
		suspect.FirstName = input.FirstName
	}
	if input.LastName != "" {
		suspect.LastName = input.LastName
	}
	if input.Alias != "" {
		suspect.Alias = input.Alias
	}
	if input.Gender != "" {
		suspect.Gender = input.Gender
	}
	if input.Nationality != "" {
		suspect.Nationality = input.Nationality
	}
	if input.IDNumber != "" {
		suspect.IDNumber = input.IDNumber
	}
	if input.Phone != "" {
		suspect.Phone = input.Phone
	}
	if input.Email != "" {
		suspect.Email = input.Email
	}
	if input.Address != "" {
		suspect.Address = input.Address
	}
	if input.Description != "" {
		suspect.Description = input.Description
	}
	if input.Status != "" {
		suspect.Status = input.Status
	}
	if input.DangerLevel != "" {
		suspect.DangerLevel = input.DangerLevel
	}
	if input.Wanted != nil {
		suspect.Wanted = *input.Wanted
	}

	if err := config.DB.Save(&suspect).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suspect"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Suspect updated successfully",
		"suspect": suspect,
	})
}

// DeleteSuspect deletes a suspect (admin only)
func DeleteSuspect(c *gin.Context) {
	id := c.Param("id")
	suspectID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suspect ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	// Only super admin can delete
	if userObj.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super admin can delete suspects"})
		return
	}

	if err := config.DB.Delete(&models.Suspect{}, "id = ?", suspectID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete suspect"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Suspect deleted successfully",
	})
}

// ReportSighting reports a suspect sighting (officers and admins only)
func ReportSighting(c *gin.Context) {
	var input struct {
		SuspectID   string  `json:"suspectId" binding:"required"`
		Latitude    float64 `json:"latitude" binding:"required"`
		Longitude   float64 `json:"longitude" binding:"required"`
		Location    string  `json:"location"`
		Description string  `json:"description"`
		UnitID      string  `json:"unitId"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	suspectID, err := uuid.Parse(input.SuspectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suspect ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	// Only officers and admins can report sightings
	if userObj.Role == "citizen" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Citizens cannot report suspect sightings"})
		return
	}

	// Verify suspect exists and user has access
	var suspect models.Suspect
	if err := config.DB.First(&suspect, "id = ?", suspectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suspect not found"})
		return
	}

	if !checkSuspectAccess(userObj, &suspect) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to report sightings for this suspect"})
		return
	}

	// Determine unit ID
	var unitID uuid.UUID
	if input.UnitID != "" {
		parsed, err := uuid.Parse(input.UnitID)
		if err == nil {
			unitID = parsed
		}
	} else if userObj.UnitID != nil {
		unitID = *userObj.UnitID
	}

	sighting := models.SuspectSighting{
		SuspectID:   suspectID,
		ReportedBy:  userObj.ID,
		UnitID:      unitID,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		Location:    input.Location,
		Description: input.Description,
		Timestamp:   time.Now(),
	}

	if err := config.DB.Create(&sighting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to report sighting"})
		return
	}

	// Mark suspect as active
	suspect.Status = "active"
	config.DB.Save(&suspect)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Sighting reported successfully",
		"sighting": sighting,
	})
}

// GetSuspectSightings gets all sightings for a suspect (with access check)
func GetSuspectSightings(c *gin.Context) {
	suspectID := c.Param("id")
	id, err := uuid.Parse(suspectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suspect ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var suspect models.Suspect
	if err := config.DB.First(&suspect, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suspect not found"})
		return
	}

	if !checkSuspectAccess(userObj, &suspect) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view sightings for this suspect"})
		return
	}

	var sightings []models.SuspectSighting
	if err := config.DB.Preload("Reporter").Preload("Unit").Where("suspect_id = ?", id).Order("timestamp desc").Find(&sightings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sightings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sightings": sightings,
	})
}

// GetSuspectCases gets all cases for a suspect (with access check)
func GetSuspectCases(c *gin.Context) {
	suspectID := c.Param("id")
	id, err := uuid.Parse(suspectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suspect ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj := user.(*models.User)

	var suspect models.Suspect
	if err := config.DB.First(&suspect, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suspect not found"})
		return
	}

	if !checkSuspectAccess(userObj, &suspect) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view cases for this suspect"})
		return
	}

	var suspectCases []models.SuspectCase
	if err := config.DB.Preload("Case").Where("suspect_id = ?", id).Find(&suspectCases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suspect cases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suspectCases": suspectCases,
	})
}