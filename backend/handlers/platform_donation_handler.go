package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// GetPlatformDonationInfo returns the platform's own bank account details.
// Public endpoint — no auth required.
func GetPlatformDonationInfo(c *gin.Context) {
	bankName := os.Getenv("PLATFORM_BANK_NAME")
	accountNumber := os.Getenv("PLATFORM_ACCOUNT_NUMBER")
	accountName := os.Getenv("PLATFORM_ACCOUNT_NAME")

	if bankName == "" || accountNumber == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "platform donation account is not configured",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bankName":      bankName,
		"accountNumber": accountNumber,
		"accountName":   accountName,
		"purpose":       "Maintenance of WardGuard platform — servers, storage, development",
	})
}

// CreatePlatformDonation records a pending donation to WardGuard.
// Public endpoint — any visitor may submit. A super admin confirms later.
func CreatePlatformDonation(c *gin.Context) {
	var input struct {
		DonorName   string  `json:"donorName" binding:"required"`
		DonorEmail  string  `json:"donorEmail"`
		DonorPhone  string  `json:"donorPhone"`
		Amount      float64 `json:"amount" binding:"required"`
		Method      string  `json:"method"`
		ReferenceID string  `json:"referenceId"`
		Message     string  `json:"message"`
		IsPublic    *bool   `json:"isPublic"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	var donorID *uuid.UUID
	if userValue, exists := c.Get("user"); exists {
		if userObj, ok := userValue.(*models.User); ok && userObj != nil {
			donorID = &userObj.ID
		}
	}

	method := input.Method
	if method == "" {
		method = "bank_transfer"
	}

	isPublic := true
	if input.IsPublic != nil {
		isPublic = *input.IsPublic
	}

	donation := models.PlatformDonation{
		DonorID:     donorID,
		DonorName:   input.DonorName,
		DonorEmail:  input.DonorEmail,
		DonorPhone:  input.DonorPhone,
		Amount:      input.Amount,
		Currency:    "NGN",
		Method:      method,
		ReferenceID: input.ReferenceID,
		Message:     input.Message,
		IsPublic:    isPublic,
		Status:      "pending",
	}

	if err := config.DB.Create(&donation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record donation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Thank you. Your donation is recorded and awaiting confirmation.",
		"donation": gin.H{
			"id":     donation.ID,
			"amount": donation.Amount,
			"status": donation.Status,
		},
	})
}

// ListPlatformDonations — super admin only. Full list with donor detail.
func ListPlatformDonations(c *gin.Context) {
	var donations []models.PlatformDonation
	if err := config.DB.Order("created_at DESC").Limit(500).Find(&donations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch donations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"donations": donations})
}

// ConfirmPlatformDonation — super admin only. Marks a donation as confirmed.
func ConfirmPlatformDonation(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid donation ID"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj, ok := userValue.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	var donation models.PlatformDonation
	if err := config.DB.First(&donation, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Donation not found"})
		return
	}

	if donation.Status == "confirmed" {
		c.JSON(http.StatusConflict, gin.H{"error": "Donation is already confirmed"})
		return
	}

	now := time.Now().UTC()
	donation.Status = "confirmed"
	donation.ConfirmedBy = &userObj.ID
	donation.ConfirmedAt = &now

	if err := config.DB.Save(&donation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm donation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Donation confirmed",
		"donation": donation,
	})
}

// ListPublicSupporters — public. Returns confirmed donations where IsPublic=true.
// Donor email and phone are never exposed. Names may be shown.
func ListPublicSupporters(c *gin.Context) {
	var donations []models.PlatformDonation
	if err := config.DB.
		Where("status = ? AND is_public = ?", "confirmed", true).
		Order("confirmed_at DESC").
		Limit(200).
		Find(&donations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch supporters"})
		return
	}

	type Supporter struct {
		DonorName string    `json:"donorName"`
		Amount    float64   `json:"amount"`
		Message   string    `json:"message,omitempty"`
		CreatedAt time.Time `json:"createdAt"`
	}

	out := make([]Supporter, 0, len(donations))
	for _, d := range donations {
		out = append(out, Supporter{
			DonorName: d.DonorName,
			Amount:    d.Amount,
			Message:   d.Message,
			CreatedAt: d.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"supporters": out})
}
