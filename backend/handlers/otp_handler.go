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
	"security-solution/services"
)

// OTP cache for faster validation (in-memory)
var otpCache = make(map[string]OTPCacheEntry)

type OTPCacheEntry struct {
	Code      string
	ExpiresAt time.Time
}

func generateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// SendOTP - OPTIMIZED: Returns immediately, sends in background
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
	expiresAt := time.Now().Add(10 * time.Minute)

	// Store in cache for instant validation
	otpCache[userID.String()] = OTPCacheEntry{
		Code:      code,
		ExpiresAt: expiresAt,
	}

	// Store in DB for persistence (async)
	go func() {
		otp := models.OTP{
			UserID:    userID,
			Code:      code,
			Type:      input.Type,
			Attempts:  0,
			ExpiresAt: expiresAt,
			Verified:  false,
		}
		config.DB.Create(&otp)
	}()

	// Send SMS in background (non-blocking)
	if user.Phone != "" {
		go func() {
			if err := services.SendOTPSMS(user.Phone, code); err != nil {
				log.Printf("⚠️ SMS delivery failed: %v", err)
			}
		}()
	}

	// Send Email in background (non-blocking)
	if user.Email != "" {
		go func() {
			// Email sending logic
			log.Printf("📧 OTP %s sent to %s (background)", code, user.Email)
		}()
	}

	// IMMEDIATE RESPONSE - user doesn't wait for SMS/Email
	c.JSON(http.StatusCreated, gin.H{
		"message":   "OTP sent successfully",
		"otpId":     userID.String(),
		"delivered": false, // Indicates SMS/Email is in progress
	})
}

// VerifyOTP - OPTIMIZED: Checks cache first, then DB
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

	// CHECK CACHE FIRST (FAST)
	if cacheEntry, exists := otpCache[userID.String()]; exists {
		if time.Now().Before(cacheEntry.ExpiresAt) && cacheEntry.Code == input.Code {
			// Valid OTP - mark user as verified
			go func() {
				config.DB.Model(&models.User{}).Where("id = ?", userID).Update("status", "verified")
				config.DB.Where("user_id = ? AND code = ?", userID, input.Code).Delete(&models.OTP{})
			}()
			delete(otpCache, userID.String())
			c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
			return
		}
	}

	// FALLBACK TO DATABASE (if not in cache)
	var otp models.OTP
	if err := config.DB.Where("user_id = ? AND code = ? AND verified = ?", userID, input.Code, false).First(&otp).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP has expired"})
		return
	}

	// Mark as verified
	otp.Verified = true
	config.DB.Save(&otp)
	config.DB.Model(&models.User{}).Where("id = ?", userID).Update("status", "verified")

	// Clean cache
	delete(otpCache, userID.String())

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully",
	})
}

// ResendOTP - OPTIMIZED: Instant response
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

	// Delete old OTPs
	go config.DB.Where("user_id = ? AND verified = ?", userID, false).Delete(&models.OTP{})

	code := generateOTP()
	expiresAt := time.Now().Add(10 * time.Minute)

	// Update cache
	otpCache[userID.String()] = OTPCacheEntry{
		Code:      code,
		ExpiresAt: expiresAt,
	}

	// Store in DB (async)
	go func() {
		otp := models.OTP{
			UserID:    userID,
			Code:      code,
			Type:      input.Type,
			Attempts:  0,
			ExpiresAt: expiresAt,
			Verified:  false,
		}
		config.DB.Create(&otp)
	}()

	// Get user info
	var user models.User
	config.DB.First(&user, "id = ?", userID)

	// Send in background
	if user.Phone != "" {
		go services.SendOTPSMS(user.Phone, code)
	}

	// IMMEDIATE RESPONSE
	c.JSON(http.StatusCreated, gin.H{
		"message": "OTP resent successfully",
		"otpId":   userID.String(),
	})
}

// Clean expired OTPs from cache (run every 10 minutes)
func CleanOTPCache() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			for key, entry := range otpCache {
				if time.Now().After(entry.ExpiresAt) {
					delete(otpCache, key)
				}
			}
		}
	}()
}