package handlers

import (
	"net/http"
	"time"

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

	token, jti, user, err := h.authService.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	sessionSvc := services.NewSessionService()
	session, err := sessionSvc.Create(
		user.ID,
		jti,
		"",
		c.GetHeader("User-Agent"),
		c.ClientIP(),
		expiresAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	refreshSvc := services.NewRefreshTokenService()
	refreshRaw, _, err := refreshSvc.Issue(user.ID, &session.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        token,
		"refreshToken": refreshRaw,
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
	userInterface, userExists := c.Get("user")
	jtiVal, jtiExists := c.Get("jti")
	expVal, expExists := c.Get("token_exp")

	if userExists && jtiExists && expExists {
		userObj, userOk := userInterface.(*models.User)
		jti, jtiOk := jtiVal.(string)
		expFloat, expOk := expVal.(float64)

		if userOk && jtiOk && expOk && jti != "" && expFloat > 0 {
			tokenSvc := services.NewTokenService()
			expiresAt := time.Unix(int64(expFloat), 0).UTC()
			_ = tokenSvc.Revoke(jti, userObj.ID, expiresAt, "logout")
		}
	}

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

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	var input struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authService.ChangePassword(userObj.ID, input.OldPassword, input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed. All existing sessions have been revoked.",
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := services.NewRefreshTokenService()
	newRaw, record, err := svc.Rotate(input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.GetUserByID(record.UserID.String())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	newAccess, err := h.authService.GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        newAccess,
		"refreshToken": newRaw,
	})
}

func (h *AuthHandler) ListSessions(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	svc := services.NewSessionService()
	sessions, err := svc.ListActive(userObj.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	jti := c.Param("jti")
	if jti == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "jti is required"})
		return
	}

	var session models.UserSession
	if err := config.DB.Where("user_id = ? AND jti = ?", userObj.ID, jti).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	reason := "user_revoked_device"

	if err := tx.Model(&models.UserSession{}).
		Where("id = ?", session.ID).
		Updates(map[string]interface{}{
			"revoked_at":     now,
			"revoked_reason": reason,
		}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke session"})
		return
	}

	revokedToken := models.RevokedToken{
		JTI:       jti,
		UserID:    userObj.ID,
		ExpiresAt: session.ExpiresAt,
		Reason:    reason,
	}
	if err := tx.Create(&revokedToken).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke token"})
		return
	}

	if err := tx.Model(&models.RefreshToken{}).
		Where("session_id = ? AND revoked_at IS NULL", session.ID).
		Update("revoked_at", now).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke refresh token"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize session revocation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session revoked"})
}

func (h *AuthHandler) RevokeAllSessions(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userObj, ok := userInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	now := time.Now().UTC()
	reason := "user_revoked_all"

	tx := config.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var sessions []models.UserSession
	if err := tx.Where("user_id = ? AND revoked_at IS NULL", userObj.ID).Find(&sessions).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load sessions"})
		return
	}

	if err := tx.Model(&models.UserSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userObj.ID).
		Updates(map[string]interface{}{
			"revoked_at":     now,
			"revoked_reason": reason,
		}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke sessions"})
		return
	}

	for _, s := range sessions {
		revokedToken := models.RevokedToken{
			JTI:       s.JTI,
			UserID:    userObj.ID,
			ExpiresAt: s.ExpiresAt,
			Reason:    reason,
		}
		if err := tx.Create(&revokedToken).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke tokens"})
			return
		}
	}

	if err := tx.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userObj.ID).
		Update("revoked_at", now).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke refresh tokens"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize revocation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All sessions revoked"})
}