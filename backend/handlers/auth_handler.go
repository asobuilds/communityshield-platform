package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"security-solution/config"
	"security-solution/models"
	"security-solution/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Email     string `json:"email" binding:"required,email"`
		Phone     string `json:"phone"`
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
		Password  string `json:"password" binding:"required,min=6"`
		Role      string `json:"role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Public registration must always create citizens.
	// Privileged roles should be assigned by an administrator.
	user := &models.User{
		Email:     input.Email,
		Phone:     input.Phone,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Password:  input.Password,
		Role:      "citizen",
	}

	createdUser, err := h.authService.Register(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":        createdUser.ID,
			"email":     createdUser.Email,
			"firstName": createdUser.FirstName,
			"lastName":  createdUser.LastName,
			"role":      createdUser.Role,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := h.authService.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"role":      user.Role,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	var freshUser models.User
	if err := config.DB.First(&freshUser, "id = ?", user.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":        freshUser.ID,
			"email":     freshUser.Email,
			"phone":     freshUser.Phone,
			"firstName": freshUser.FirstName,
			"lastName":  freshUser.LastName,
			"role":      freshUser.Role,
			"status":    freshUser.Status,
			"createdAt": freshUser.CreatedAt,
			"updatedAt": freshUser.UpdatedAt,
		},
	})
}