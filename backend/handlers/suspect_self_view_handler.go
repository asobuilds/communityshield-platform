package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
)

// SuspectCaseView is the public-safe projection of a case returned to
// a citizen who is linked as a suspect.
type SuspectCaseView struct {
	CaseID       string `json:"caseId"`
	TrackingID   string `json:"trackingId"`
	Title        string `json:"title"`
	Category     string `json:"category"`
	Status       string `json:"status"`
	RoleInCase   string `json:"roleInCase"`
	OpenedAt     string `json:"openedAt"`
	ClosedAt     string `json:"closedAt,omitempty"`
	LastProgress string `json:"lastProgress,omitempty"`
}

// GetMySuspectCases returns the cases linked to the authenticated
// user's suspect record(s). Only public-safe fields are exposed.
// Active cases are returned under "active". Closed cases are returned
// under "history" (resolved).
func GetMySuspectCases(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj := userValue.(*models.User)

	var suspects []models.Suspect
	if err := config.DB.Where("user_id = ?", userObj.ID).Find(&suspects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load suspect records"})
		return
	}

	if len(suspects) == 0 {
		c.JSON(http.StatusOK, gin.H{"active": []SuspectCaseView{}, "history": []SuspectCaseView{}})
		return
	}

	suspectIDs := make([]string, 0, len(suspects))
	for _, s := range suspects {
		suspectIDs = append(suspectIDs, s.ID.String())
	}

	var links []models.SuspectCase
	if err := config.DB.
		Preload("Case").
		Where("suspect_id IN ?", suspectIDs).
		Order("created_at DESC").
		Find(&links).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load linked cases"})
		return
	}

	active := []SuspectCaseView{}
	history := []SuspectCaseView{}

	for _, link := range links {
		view := SuspectCaseView{
			CaseID:     link.CaseID.String(),
			TrackingID: link.Case.TrackingID,
			Title:      link.Case.Title,
			Status:     link.Case.Status,
			RoleInCase: link.Role,
			OpenedAt:   link.Case.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if link.Case.ClosedAt != nil {
			view.ClosedAt = link.Case.ClosedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		var lastProgress models.CaseProgress
		if err := config.DB.
			Where("case_id = ?", link.CaseID).
			Order("created_at DESC").
			First(&lastProgress).Error; err == nil {
			view.LastProgress = lastProgress.Description
		}

		if link.Case.Status == "closed" {
			history = append(history, view)
		} else {
			active = append(active, view)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"active":  active,
		"history": history,
	})
}
