package services

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
)

const (
	DefaultCaseArrivalRadiusMeters = 100.0
	MaxAcceptedGPSAccuracyMeters   = 100.0
	MaxLocationAge                 = 2 * time.Minute
)

func CheckOfficerReachedCase(officerID uuid.UUID, caseID uuid.UUID) (bool, error) {
	var caseRecord models.Case

	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		return false, err
	}

	if caseRecord.AssignedTo == nil || *caseRecord.AssignedTo != officerID {
		return false, nil
	}

	if caseRecord.Status != models.CaseStatusDispatched {
		return false, nil
	}

	var location models.UserLocation

	if err := config.DB.
		Where("user_id = ?", officerID).
		Order("recorded_at DESC").
		First(&location).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	now := time.Now().UTC()

	if location.RecordedAt.IsZero() ||
		now.Sub(location.RecordedAt.UTC()) > MaxLocationAge ||
		location.RecordedAt.UTC().After(now.Add(10*time.Second)) {
		return false, nil
	}

	if location.Accuracy > MaxAcceptedGPSAccuracyMeters {
		return false, nil
	}

	caseLatitude := caseRecord.Latitude
	caseLongitude := caseRecord.Longitude

	if caseRecord.GISLatitude != 0 || caseRecord.GISLongitude != 0 {
		caseLatitude = caseRecord.GISLatitude
		caseLongitude = caseRecord.GISLongitude
	}

	if caseLatitude < -90 || caseLatitude > 90 ||
		caseLongitude < -180 || caseLongitude > 180 {
		return false, nil
	}

	distance := calculateDistanceMeters(
		caseLatitude,
		caseLongitude,
		location.Latitude,
		location.Longitude,
	)

	if distance > DefaultCaseArrivalRadiusMeters {
		return false, nil
	}

	// Atomic state transition prevents concurrent location requests
	// from transitioning the same case more than once.
	result := config.DB.Model(&models.Case{}).
		Where("id = ? AND assigned_to = ? AND status = ?",
			caseID,
			officerID,
			models.CaseStatusDispatched,
		).
		Updates(map[string]interface{}{
			"status":     models.CaseStatusOnScene,
			"arrived_at": now,
			"updated_at": now,
		})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected == 1, nil
}

func calculateDistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMeters = 6371000.0

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	sinLat := math.Sin(dLat / 2)
	sinLon := math.Sin(dLon / 2)

	a := sinLat*sinLat +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*sinLon*sinLon

	return earthRadiusMeters * 2 * math.Atan2(
		math.Sqrt(a),
		math.Sqrt(1-a),
	)
}
