package handlers

import (
	"fmt"
	"net/http"
	"security-solution/models"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AdminGetSOSAlerts(c *gin.Context) {
	status := c.Query("status")
	userId := c.Query("userId")
	unitId := c.Query("unitId")
	search := c.Query("search")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	query := DB.Model(&models.SOSAlert{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if userId != "" {
		if parsed, err := uuid.Parse(userId); err == nil {
			query = query.Where("user_id = ?", parsed)
		}
	}
	if unitId != "" {
		if parsed, err := uuid.Parse(unitId); err == nil {
			query = query.Where("unit_id = ?", parsed)
		}
	}
	if search != "" {
		query = query.Where("description ILIKE ?", "%"+search+"%")
	}
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var sosList []models.SOSAlert
	if err := query.Order("created_at desc").Limit(100).Find(&sosList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SOS alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sos":   sosList,
		"count": len(sosList),
	})
}

func AdminGetCases(c *gin.Context) {
	status := c.Query("status")
	priority := c.Query("priority")
	reportedBy := c.Query("reportedBy")
	unitId := c.Query("unitId")
	assignedTo := c.Query("assignedTo")
	search := c.Query("search")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	query := DB.Model(&models.Case{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if reportedBy != "" {
		if parsed, err := uuid.Parse(reportedBy); err == nil {
			query = query.Where("reported_by = ?", parsed)
		}
	}
	if unitId != "" {
		if parsed, err := uuid.Parse(unitId); err == nil {
			query = query.Where("unit_id = ?", parsed)
		}
	}
	if assignedTo != "" {
		if parsed, err := uuid.Parse(assignedTo); err == nil {
			query = query.Where("assigned_to = ?", parsed)
		}
	}
	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var cases []models.Case
	if err := query.Order("created_at desc").Limit(100).Find(&cases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cases": cases,
		"count": len(cases),
	})
}

func AdminUpdateCaseStatus(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return
	}

	var input struct {
		Status     string     `json:"status" binding:"required"`
		AssignedTo *uuid.UUID `json:"assignedTo"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var caseItem models.Case
	if err := DB.First(&caseItem, "id = ?", parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return
	}

	caseItem.Status = input.Status
	if input.AssignedTo != nil {
		caseItem.AssignedTo = input.AssignedTo
	}
	if err := DB.Save(&caseItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update case"})
		return
	}

	if caseItem.ReportedBy != uuid.Nil {
		title := "Case Status Updated"
		message := fmt.Sprintf("Your case '%s' has been updated to %s.", caseItem.Title, caseItem.Status)
		if err := CreateNotification(caseItem.ReportedBy, caseItem.ID, title, message); err != nil {
			fmt.Printf("Failed to create notification: %v\n", err)
		}
	}

	if caseItem.ReportedBy != uuid.Nil {
		var reporter models.User
		if err := DB.First(&reporter, "id = ?", caseItem.ReportedBy).Error; err == nil {
			if reporter.Phone != "" {
				if err := SendCaseStatusSMS(reporter.Phone, caseItem.Title, caseItem.Status); err != nil {
					fmt.Printf("Failed to send case status SMS: %v\n", err)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Case updated",
		"case":    caseItem,
	})
}

func AdminUpdateSOSStatus(c *gin.Context) {
	id := c.Param("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SOS ID"})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sos models.SOSAlert
	if err := DB.First(&sos, "id = ?", parsedID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SOS alert not found"})
		return
	}

	sos.Status = input.Status
	if err := DB.Save(&sos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SOS status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SOS status updated",
		"sos":     sos,
	})
}

func GetUnitOfficers(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedUnitID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var users []models.User
	if err := DB.Where("unit_id = ? AND role = ?", parsedUnitID, "unit_admin").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch officers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"officers": users})
}

func AddOfficer(c *gin.Context) {
	unitID := c.Param("unitId")
	parsedUnitID, err := uuid.Parse(unitID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	var input struct {
		Email     string `json:"email" binding:"required"`
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
		Phone     string `json:"phone"`
		Rank      string `json:"rank"`
		Role      string `json:"role"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := input.Role
	if role == "" {
		role = "officer"
	}

	tempPassword := "password123"

	user := models.User{
		Email:     input.Email,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Phone:     input.Phone,
		Rank:      input.Rank,
		Role:      role,
		UnitID:    &parsedUnitID,
		Status:    "active",
	}
	user.Password = tempPassword

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists or invalid data"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Officer added successfully",
		"officer": user,
	})
}

func RemoveOfficer(c *gin.Context) {
	userID := c.Param("userId")
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := DB.Delete(&models.User{}, "id = ?", parsedUserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove officer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Officer removed successfully"})
}