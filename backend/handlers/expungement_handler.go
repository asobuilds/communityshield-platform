package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/models"
	"security-solution/services"
)

// RequestExpungement allows a citizen linked to a suspect to request that
// the suspect record be expunged. The service enforces all eligibility rules.
func RequestExpungement(c *gin.Context) {
	suspectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suspect ID"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj := userValue.(*models.User)

	var input struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := services.NewExpungementService()
	reqRecord, err := svc.RequestExpungement(services.ExpungementRequestInput{
		SuspectID:   suspectID,
		RequesterID: userObj.ID,
		Reason:      input.Reason,
	})
	if err != nil {
		if isNotFound := strings.Contains(err.Error(), "not found"); isNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         reqRecord.ID,
		"suspectId":  reqRecord.SuspectID,
		"status":     reqRecord.Status,
		"requestedAt": reqRecord.RequestedAt,
	})
}

// ListMyExpungementRequests returns the expungement requests a citizen has
// filed for a suspect they are linked to.
func ListMyExpungementRequests(c *gin.Context) {
	suspectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suspect ID"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj := userValue.(*models.User)

	svc := services.NewExpungementService()
	reqs, err := svc.ListRequests(suspectID, userObj.ID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": reqs})
}

// DecideExpungement lets a head admin of the suspect's unit (or super admin)
// approve or deny a pending expungement request.
func DecideExpungement(c *gin.Context) {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj := userValue.(*models.User)

	var input struct {
		Decision string `json:"decision" binding:"required"`
		Reason   string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := services.NewExpungementService()
	reqRecord, err := svc.DecideExpungement(requestID, userObj.ID, input.Decision, input.Reason)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "only a head admin") || strings.Contains(err.Error(), "super admin") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            reqRecord.ID,
		"suspectId":     reqRecord.SuspectID,
		"status":        reqRecord.Status,
		"reviewedBy":    reqRecord.ReviewedBy,
		"reviewedAt":    reqRecord.ReviewedAt,
		"decisionReason": reqRecord.DecisionReason,
	})
}
