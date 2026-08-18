package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func SendOTP(c *gin.Context) {
	var input struct {
		UserID string `json:"userId" binding:"required"`
		Type   string `json:"type"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	code := generateOTP()
	log.Printf("🔑 OTP for user %s: %s", userID, code)

	expiresAt := time.Now().Add(10 * time.Minute)

	otp := models.OTP{
		UserID:    userID,
		Code:      code,
		Type:      input.Type,
		Attempts:  0,
		ExpiresAt: expiresAt,
		Verified:  false,
	}

	if err := config.DB.Create(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	log.Printf("✅ OTP generated successfully for %s", user.Email)
	log.Printf("📧 OTP Code: %s", code)

	c.JSON(http.StatusCreated, gin.H{
		"message": "OTP sent successfully",
		"otpId":   otp.ID,
	})
}

func VerifyOTP(c *gin.Context) {
	var input struct {
		UserID string `json:"userId" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var otp models.OTP
	if err := config.DB.Where("user_id = ? AND code = ? AND verified = ?", userID, input.Code, false).First(&otp).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP has expired"})
		return
	}

	otp.Attempts++
	if otp.Attempts >= 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many attempts"})
		return
	}

	otp.Verified = true
	if err := config.DB.Save(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err == nil {
		user.Status = "verified"
		config.DB.Save(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully",
	})
}

func ResendOTP(c *gin.Context) {
	var input struct {
		UserID string `json:"userId" binding:"required"`
		Type   string `json:"type"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := config.DB.Where("user_id = ? AND verified = ?", userID, false).Delete(&models.OTP{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear old OTPs"})
		return
	}

	code := generateOTP()
	log.Printf("🔑 Resent OTP for user %s: %s", userID, code)

	expiresAt := time.Now().Add(10 * time.Minute)

	otp := models.OTP{
		UserID:    userID,
		Code:      code,
		Type:      input.Type,
		Attempts:  0,
		ExpiresAt: expiresAt,
		Verified:  false,
	}

	if err := config.DB.Create(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend OTP"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "OTP resent successfully",
		"otpId":   otp.ID,
	})
}