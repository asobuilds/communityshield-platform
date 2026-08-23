package services

import (
	"encoding/json"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

func (s *AuditService) LogAction(userID uuid.UUID, action, entityType, entityID string, oldValue, newValue interface{}, ipAddress, userAgent string) error {
	oldJSON := ""
	newJSON := ""

	if oldValue != nil {
		if oldBytes, err := json.Marshal(oldValue); err == nil {
			oldJSON = string(oldBytes)
		}
	}
	if newValue != nil {
		if newBytes, err := json.Marshal(newValue); err == nil {
			newJSON = string(newBytes)
		}
	}

	audit := models.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValue:   oldJSON,
		NewValue:   newJSON,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	return config.DB.Create(&audit).Error
}

func (s *AuditService) GetAuditTrail(entityType, entityID string) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := config.DB.Preload("User").Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("timestamp desc").Find(&logs).Error
	return logs, err
}
