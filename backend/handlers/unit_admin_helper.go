package handlers

import (
	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

func isUnitAdmin(userID uuid.UUID, unitID uuid.UUID) bool {
	var member models.UnitMembership

	err := config.DB.
		Where(
			"user_id = ? AND unit_id = ? AND role = ? AND status = ?",
			userID,
			unitID,
			models.UnitRoleAdmin,
			models.MembershipActive,
		).
		First(&member).Error

	return err == nil
}
