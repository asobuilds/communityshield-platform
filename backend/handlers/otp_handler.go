package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"security-solution/models"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	if err := DB.First(&user, "id = ?", userID).Error; err != nil {
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

	if err := DB.Create(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	if input.Type == "email" || input.Type == "both" {
		subject := "Your OTP Code"
		body := fmt.Sprintf("Your OTP code is: <b>%s</b><br>It expires in 10 minutes.", code)
		if err := SendEmail(user.Email, subject, body); err != nil {
			fmt.Printf("Failed to send OTP email: %v\n", err)
		} else {
			fmt.Println("✅ OTP email sent to", user.Email)
		}
	}

	if input.Type == "phone" || input.Type == "both" {
		if user.Phone != "" {
			message := fmt.Sprintf("Your CommunityShield OTP code is: %s. It expires in 10 minutes.", code)
			if err := SendSMS(user.Phone, message); err != nil {
				fmt.Printf("Failed to send OTP SMS: %v\n", err)
			} else {
				fmt.Println("✅ OTP SMS sent to", user.Phone)
			}
		}
	}

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
	if err := DB.Where("user_id = ? AND code = ? AND verified = ?", userID, input.Code, false).First(&otp).Error; err != nil {
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
	if err := DB.Save(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP"})
		return
	}

	var user models.User
	if err := DB.First(&user, "id = ?", userID).Error; err == nil {
		user.Status = "verified"
		DB.Save(&user)
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

	// Delete old unverified OTPs for this user
	if err := DB.Where("user_id = ? AND verified = ?", userID, false).Delete(&models.OTP{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear old OTPs"})
		return
	}

	code := generateOTP()
	expiresAt := time.Now().Add(10 * time.Minute)

	otp := models.OTP{
		UserID:    userID,
		Code:      code,
		Type:      input.Type,
		Attempts:  0,
		ExpiresAt: expiresAt,
		Verified:  false,
	}

	if err := DB.Create(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend OTP"})
		return
	}

	var user models.User
	if err := DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if input.Type == "email" || input.Type == "both" {
		subject := "Your New OTP Code"
		body := fmt.Sprintf("Your new OTP code is: <b>%s</b><br>It expires in 10 minutes.", code)
		if err := SendEmail(user.Email, subject, body); err != nil {
			fmt.Printf("Failed to send OTP email: %v\n", err)
		} else {
			fmt.Println("✅ OTP email resent to", user.Email)
		}
	}

	if input.Type == "phone" || input.Type == "both" {
		if user.Phone != "" {
			message := fmt.Sprintf("Your new CommunityShield OTP code is: %s. It expires in 10 minutes.", code)
			if err := SendSMS(user.Phone, message); err != nil {
				fmt.Printf("Failed to send OTP SMS: %v\n", err)
			} else {
				fmt.Println("✅ OTP SMS resent to", user.Phone)
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "OTP resent successfully",
		"otpId":   otp.ID,
	})
}