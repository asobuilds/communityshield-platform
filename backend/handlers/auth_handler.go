package handlers

import (
	"fmt"
	"net/http"
	"security-solution/models"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Register(c *gin.Context) {
	var input struct {
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=6"`
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
		Phone     string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:     input.Email,
		Password:  string(hashedPassword),
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Phone:     input.Phone,
		Role:      "citizen",
	}

	if err := DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists or invalid data"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"role":      user.Role,
		},
	})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed: " + err.Error()})
		return
	}

	var user models.User
	if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"role":      user.Role,
		},
	})
}

func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "If your email is registered, you will receive a password reset link."})
		return
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(1 * time.Hour)
	user.ResetToken = token
	user.ResetTokenExpires = &expiresAt
	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
		return
	}

	resetLink := "http://localhost:5173/reset-password?token=" + token

	// Send email
	if err := SendPasswordResetEmail(input.Email, resetLink); err != nil {
		fmt.Printf("Failed to send password reset email: %v\n", err)
	} else {
		fmt.Println("✅ Password reset email sent to", input.Email)
	}

	// 🔥 Send SMS with reset link
	if user.Phone != "" {
		if err := SendPasswordResetSMS(user.Phone, resetLink); err != nil {
			fmt.Printf("Failed to send password reset SMS: %v\n", err)
		} else {
			fmt.Println("✅ Password reset SMS sent to", user.Phone)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "If your email is registered, you will receive a password reset link.",
		"token":     token,
		"resetLink": resetLink,
	})
}

func ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := DB.Where("reset_token = ? AND reset_token_expires > ?", input.Token, time.Now()).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = string(hashedPassword)
	user.ResetToken = ""
	user.ResetTokenExpires = nil
	if err := DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful. Please login with your new password."})
}

// RegisterUnit creates a new unit and an admin user
func RegisterUnit(c *gin.Context) {
	var input struct {
		Unit struct {
			Name               string `json:"name" binding:"required"`
			Type               string `json:"type" binding:"required"`
			CoverageArea       string `json:"coverageArea"`
			ContactPerson      string `json:"contactPerson" binding:"required"`
			ContactPhone       string `json:"contactPhone" binding:"required"`
			ContactEmail       string `json:"contactEmail"`
			Motto              string `json:"motto"`
			RegistrationNumber string `json:"registrationNumber"`
		} `json:"unit" binding:"required"`
		Admin struct {
			Email     string `json:"email" binding:"required,email"`
			Password  string `json:"password" binding:"required,min=6"`
			FirstName string `json:"firstName" binding:"required"`
			LastName  string `json:"lastName" binding:"required"`
			Phone     string `json:"phone"`
		} `json:"admin" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if admin email already exists
	var existingUser models.User
	if err := DB.Where("email = ?", input.Admin.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
		return
	}

	// Create unit
	unit := models.Unit{
		Name:               input.Unit.Name,
		Type:               input.Unit.Type,
		CoverageArea:       input.Unit.CoverageArea,
		ContactPerson:      input.Unit.ContactPerson,
		ContactPhone:       input.Unit.ContactPhone,
		ContactEmail:       input.Unit.ContactEmail,
		Motto:              input.Unit.Motto,
		RegistrationNumber: input.Unit.RegistrationNumber,
		Status:             "active",
	}
	if err := DB.Create(&unit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create unit"})
		return
	}

	// Hash admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Admin.Password), bcrypt.DefaultCost)
	if err != nil {
		DB.Delete(&unit)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create admin user
	admin := models.User{
		Email:     input.Admin.Email,
		Password:  string(hashedPassword),
		FirstName: input.Admin.FirstName,
		LastName:  input.Admin.LastName,
		Phone:     input.Admin.Phone,
		Role:      "unit_admin",
		UnitID:    &unit.ID,
		Status:    "active",
	}
	if err := DB.Create(&admin).Error; err != nil {
		DB.Delete(&unit)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin"})
		return
	}

	// 🔥 Send welcome SMS
	if admin.Phone != "" {
		if err := SendUnitRegistrationSMS(admin.Phone, unit.Name); err != nil {
			fmt.Printf("Failed to send welcome SMS: %v\n", err)
		} else {
			fmt.Println("✅ Welcome SMS sent to", admin.Phone)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Unit and admin registered successfully",
		"unit":    unit,
		"admin": gin.H{
			"id":        admin.ID,
			"email":     admin.Email,
			"firstName": admin.FirstName,
			"lastName":  admin.LastName,
			"role":      admin.Role,
		},
	})
}