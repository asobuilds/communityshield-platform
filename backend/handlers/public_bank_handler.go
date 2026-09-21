package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// GetPublicBankAccounts returns the public-facing bank accounts for a unit.
// Only accounts with IsPublic=true are returned. Sensitive fields
// (Creator, internal notes) are not exposed.
func GetPublicBankAccounts(c *gin.Context) {
	unitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid unit id"})
		return
	}

	var unit models.SecurityUnit
	if err := config.DB.First(&unit, "id = ?", unitID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
		return
	}

	var accounts []models.BankAccount
	if err := config.DB.
		Where("unit_id = ? AND is_public = ? AND status = ?", unitID, true, "active").
		Order("display_order ASC, created_at ASC").
		Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load accounts"})
		return
	}

	// Project only public-safe fields
	type PublicAccount struct {
		ID            uuid.UUID `json:"id"`
		BankName      string    `json:"bankName"`
		AccountNumber string    `json:"accountNumber"`
		AccountName   string    `json:"accountName"`
		BankCode      string    `json:"bankCode,omitempty"`
		Branch        string    `json:"branch,omitempty"`
		IsDefault     bool      `json:"isDefault"`
		DisplayOrder  int       `json:"displayOrder"`
	}

	out := make([]PublicAccount, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, PublicAccount{
			ID:            a.ID,
			BankName:      a.BankName,
			AccountNumber: a.AccountNumber,
			AccountName:   a.AccountName,
			BankCode:      a.BankCode,
			Branch:        a.Branch,
			IsDefault:     a.IsDefault,
			DisplayOrder:  a.DisplayOrder,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"unitId":       unit.ID,
		"unitName":     unit.Name,
		"bankAccounts": out,
	})
}

// ToggleBankAccountPublic flips IsPublic for a bank account.
// Only admins of the owning unit (or super admin) may toggle.
func ToggleBankAccountPublic(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	userObj, ok := userValue.(*models.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	var account models.BankAccount
	if err := config.DB.First(&account, "id = ?", accountID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bank account not found"})
		return
	}

	// Authorization: super admin OR admin of the same unit
	if !userObj.IsSuperAdmin && userObj.Role != "super_admin" {
		if userObj.Role != "unit_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins may change account visibility"})
			return
		}
		var membership models.UnitMembership
		if err := config.DB.
			Where("unit_id = ? AND user_id = ? AND status = ?", account.UnitID, userObj.ID, models.MembershipActive).
			First(&membership).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this unit"})
			return
		}
		if !membership.IsHeadAdmin && membership.Role != models.UnitRoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins of this unit may change visibility"})
			return
		}
	}

	var input struct {
		IsPublic     *bool `json:"isPublic"`
		DisplayOrder *int  `json:"displayOrder"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if input.IsPublic != nil {
		account.IsPublic = *input.IsPublic
	}
	if input.DisplayOrder != nil {
		account.DisplayOrder = *input.DisplayOrder
	}

	if err := config.DB.Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "bank account visibility updated",
		"account": account,
	})
}
