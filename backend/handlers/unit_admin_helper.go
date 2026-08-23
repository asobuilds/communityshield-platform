package handlers

import (
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

func isUnitAdmin(userID uuid.UUID, unitID uuid.UUID) bool {
	var member models.UnitMember

	err := config.DB.
		Where(
			"user_id = ? AND unit_id = ? AND role = ? AND status = ?",
			userID,
			unitID,
			"admin",
			"active",
		).
		First(&member).Error

	return err == nil
}
