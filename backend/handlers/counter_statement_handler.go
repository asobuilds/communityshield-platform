package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
)

const maxCounterStatementLength = 2000

// isCaseSuspect checks whether the given user is linked to the case as a
// suspect (or person of interest) via Suspect.UserID -> SuspectCase.
func isCaseSuspect(userID, caseID uuid.UUID) bool {
	var suspect models.Suspect
	if config.DB.First(&suspect, "user_id = ?", userID).Error != nil {
		return false
	}
	var link models.SuspectCase
	return config.DB.
		Where("suspect_id = ? AND case_id = ? AND role IN ?", suspect.ID, caseID, []string{"suspect", "person_of_interest"}).
		First(&link).Error == nil
}

// isAuthorizedForCase returns true if the authenticated user holds a
// personnel-level case access level granted by CanAccessCase (i.e. everyone
// except the reporter-only gate).
func isAuthorizedForCase(c *gin.Context) bool {
	level, exists := c.Get("case_access_level")
	if !exists {
		return false
	}
	switch level {
	case "super_admin", "head_admin", "assigned_admin",
		"officer_primary", "officer_paired", "officer_support":
		return true
	}
	return false
}

// blockIfReporter enforces that the case reporter can never read or write a
// counter-statement, even though CanAccessCase grants them a "reporter" level.
func blockIfReporter(c *gin.Context) (models.Case, bool) {
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID"})
		return models.Case{}, true
	}
	var caseObj models.Case
	if err := config.DB.First(&caseObj, "id = ?", caseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Case not found"})
		return models.Case{}, true
	}
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return models.Case{}, true
	}
	userObj := userValue.(*models.User)
	if caseObj.ReportedBy == userObj.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Reporter has no access to counter-statements"})
		return models.Case{}, true
	}
	return caseObj, false
}

// PutCounterStatement upserts the citizen's counter-statement. Empty content
// withdraws (deletes) the existing statement. Blocked when the case is closed.
func PutCounterStatement(c *gin.Context) {
	caseObj, abort := blockIfReporter(c)
	if abort {
		return
	}
	caseID := caseObj.ID
	userValue, _ := c.Get("user")
	userObj := userValue.(*models.User)

	if caseObj.Status == "closed" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Counter-statement is read-only after case closure"})
		return
	}

	if !isCaseSuspect(userObj.ID, caseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only a citizen linked to this case as a suspect may submit a counter-statement"})
		return
	}

	var input struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len([]rune(input.Content)) > maxCounterStatementLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content exceeds 2000 characters"})
		return
	}

	if input.Content == "" {
		config.DB.Unscoped().Where("case_id = ? AND user_id = ?", caseID, userObj.ID).Delete(&models.CounterStatement{})
		c.JSON(http.StatusOK, gin.H{"message": "Counter-statement withdrawn"})
		return
	}

	var stmt models.CounterStatement
	result := config.DB.
		Where("case_id = ? AND user_id = ?", caseID, userObj.ID).
		FirstOrCreate(&stmt, models.CounterStatement{
			CaseID: caseID,
			UserID: userObj.ID,
		})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	stmt.Content = input.Content
	if err := config.DB.Save(&stmt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        stmt.ID,
		"caseId":    stmt.CaseID,
		"content":   stmt.Content,
		"createdAt": stmt.CreatedAt,
		"updatedAt": stmt.UpdatedAt,
		"readOnly":  false,
	})
}

// GetCounterStatement returns the single counter-statement on a case (if any)
// to authorized personnel or the linked suspect.
func GetCounterStatement(c *gin.Context) {
	caseObj, abort := blockIfReporter(c)
	if abort {
		return
	}
	caseID := caseObj.ID
	userValue, _ := c.Get("user")
	userObj := userValue.(*models.User)

	if !isAuthorizedForCase(c) && !isCaseSuspect(userObj.ID, caseID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to view counter-statement"})
		return
	}

	var stmt models.CounterStatement
	if err := config.DB.Where("case_id = ?", caseID).First(&stmt).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("counter-statement: lookup failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"caseId":    caseID,
		"content":   stmt.Content,
		"createdAt": stmt.CreatedAt,
		"updatedAt": stmt.UpdatedAt,
		"readOnly":  caseObj.Status == "closed",
	})
}
