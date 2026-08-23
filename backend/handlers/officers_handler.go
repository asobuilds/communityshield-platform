package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// CreateOfficer creates a new officer
func CreateOfficer(c *gin.Context) {
	var input struct {
		UnitID      string `json:"unitId" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Rank        string `json:"rank" binding:"required"`
		BadgeNumber string `json:"badgeNumber" binding:"required"`
		Role        string `json:"role" binding:"required"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
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

	officer := models.Officer{
		UnitID:      unitID,
		Name:        input.Name,
		Rank:        input.Rank,
		BadgeNumber: input.BadgeNumber,
		Role:        input.Role,
		Phone:       input.Phone,
		Email:       input.Email,
		Status:      "active",
	}

	if err := config.DB.Create(&officer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create officer"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Officer created successfully",
		"officer": officer,
	})
}

// GetAllOfficers gets all officers
func GetAllOfficers(c *gin.Context) {
	var officers []models.Officer
	if err := config.DB.Find(&officers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch officers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"officers": officers,
	})
}

// GetOfficerByID gets a specific officer
func GetOfficerByID(c *gin.Context) {
	id := c.Param("id")
	officerID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	var officer models.Officer
	if err := config.DB.First(&officer, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"officer": officer,
	})
}

// UpdateOfficer updates an officer
func UpdateOfficer(c *gin.Context) {
	id := c.Param("id")
	officerID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	var input struct {
		Name   string `json:"name"`
		Rank   string `json:"rank"`
		Role   string `json:"role"`
		Phone  string `json:"phone"`
		Email  string `json:"email"`
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var officer models.Officer
	if err := config.DB.First(&officer, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Officer not found"})
		return
	}

	if input.Name != "" {
		officer.Name = input.Name
	}
	if input.Rank != "" {
		officer.Rank = input.Rank
	}
	if input.Role != "" {
		officer.Role = input.Role
	}
	if input.Phone != "" {
		officer.Phone = input.Phone
	}
	if input.Email != "" {
		officer.Email = input.Email
	}
	if input.Status != "" {
		officer.Status = input.Status
	}

	if err := config.DB.Save(&officer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update officer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Officer updated successfully",
		"officer": officer,
	})
}

// DeleteOfficer deletes an officer
func DeleteOfficer(c *gin.Context) {
	id := c.Param("id")
	officerID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid officer ID"})
		return
	}

	if err := config.DB.Delete(&models.Officer{}, "id = ?", officerID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete officer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Officer deleted successfully",
	})
}

// GetOfficersByUnit gets officers for a specific unit
func GetOfficersByUnit(c *gin.Context) {
	unitID := c.Param("unitId")
	id, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var officers []models.Officer
	if err := config.DB.Where("unit_id = ?", id).Find(&officers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch officers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"officers": officers,
	})
}
