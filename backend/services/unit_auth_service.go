package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
)

// UnitAuthPolicy is the parsed governance policy for a unit.
// Values match the locked Stage A rules.
type UnitAuthPolicy struct {
	RegularAdminRemovalVotes int            `json:"regularAdminRemovalVotes"`
	HeadAdminRemovalMajority float64        `json:"headAdminRemovalMajority"`
	QuorumPercent            float64        `json:"quorumPercent"`
	CoolingOffMonths         int            `json:"coolingOffMonths"`
	TermMonths               int            `json:"termMonths"`
	StaggerMonths            int            `json:"staggerMonths"`
	ConsecutiveTermLimit     int            `json:"consecutiveTermLimit"`
	SeatBands                map[string]int `json:"seatBands"`
}

func DefaultUnitAuthPolicy() UnitAuthPolicy {
	return UnitAuthPolicy{
		RegularAdminRemovalVotes: 10,
		HeadAdminRemovalMajority: 0.5,
		QuorumPercent:            0.5,
		CoolingOffMonths:         6,
		TermMonths:               12,
		StaggerMonths:            6,
		ConsecutiveTermLimit:     2,
		SeatBands: map[string]int{
			"12-20":  5,
			"21-40":  7,
			"41-70":  9,
			"71-100": 10,
		},
	}
}

type UnitAuthService struct{}

func NewUnitAuthService() *UnitAuthService {
	return &UnitAuthService{}
}

// GetPolicy returns the parsed policy for a unit plus its version string.
// If no record exists, the default policy and version "default-v1" are returned.
func (s *UnitAuthService) GetPolicy(unitID uuid.UUID) (UnitAuthPolicy, string, error) {
	var record models.UnitAuth
	err := config.DB.Where("unit_id = ?", unitID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultUnitAuthPolicy(), "default-v1", nil
	}
	if err != nil {
		return UnitAuthPolicy{}, "", err
	}

	var policy UnitAuthPolicy
	if err := json.Unmarshal([]byte(record.Rules), &policy); err != nil {
		return UnitAuthPolicy{}, "", fmt.Errorf("invalid unit auth policy: %w", err)
	}

	return policy, fmt.Sprintf("v%d", record.Version), nil
}

// UpsertPolicy creates or updates the UnitAuth record and bumps the version.
func (s *UnitAuthService) UpsertPolicy(unitID uuid.UUID, policy UnitAuthPolicy) (string, error) {
	raw, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	var record models.UnitAuth
	err = config.DB.Where("unit_id = ?", unitID).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		record = models.UnitAuth{
			UnitID:  unitID,
			Rules:   string(raw),
			Version: 1,
		}
		if err := config.DB.Create(&record).Error; err != nil {
			return "", err
		}
		return fmt.Sprintf("v%d", record.Version), nil
	}
	if err != nil {
		return "", err
	}

	record.Rules = string(raw)
	record.Version = record.Version + 1
	record.UpdatedAt = time.Now().UTC()

	if err := config.DB.Save(&record).Error; err != nil {
		return "", err
	}

	return fmt.Sprintf("v%d", record.Version), nil
}

// GetPolicyVersion returns only the version string, for snapshotting into
// elections and revocation cycles.
func (s *UnitAuthService) GetPolicyVersion(unitID uuid.UUID) string {
	_, version, err := s.GetPolicy(unitID)
	if err != nil {
		return "unknown"
	}
	return version
}
