package services

import (
	"math"
	"time"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

type RankingService struct {
	Ratings *RatingService
}

func NewRankingService() *RankingService {
	return &RankingService{Ratings: NewRatingService()}
}

// RecomputeAllOfficers rebuilds OfficerScore for every officer that has ratings.
func (s *RankingService) RecomputeAllOfficers() error {
	var officerIDs []uuid.UUID
	if err := config.DB.Model(&models.Rating{}).
		Where("target_type = ? AND status = ?", "officer", "active").
		Distinct("target_id").
		Pluck("target_id", &officerIDs).Error; err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, officerID := range officerIDs {
		_, count, bayes := s.Ratings.AggregateForOfficer(officerID)
		tier := TierForOfficer(bayes, count)

		var officer models.Officer
		if err := config.DB.First(&officer, "id = ?", officerID).Error; err != nil {
			continue
		}

		var existing models.OfficerScore
		err := config.DB.Where("officer_id = ?", officerID).First(&existing).Error
		if err == nil {
			existing.UnitID = officer.UnitID
			existing.BayesianScore = bayes
			existing.RatingCount = int(count)
			existing.Tier = tier
			existing.LastComputedAt = now
			config.DB.Save(&existing)
		} else {
			score := models.OfficerScore{
				OfficerID:      officerID,
				UnitID:         officer.UnitID,
				BayesianScore:  bayes,
				RatingCount:    int(count),
				Tier:           tier,
				LastComputedAt: now,
			}
			config.DB.Create(&score)
		}
	}
	return nil
}

// RecomputeAllUnits rebuilds UnitScore for every active unit.
func (s *RankingService) RecomputeAllUnits() error {
	var units []models.SecurityUnit
	if err := config.DB.Where("status = ?", "active").Find(&units).Error; err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, unit := range units {
		s.recomputeOneUnit(unit.ID, now)
	}
	return nil
}

func (s *RankingService) recomputeOneUnit(unitID uuid.UUID, now time.Time) {
	// 1. Officer aggregate — average of all officer Bayesian scores in this unit
	var officerScores []models.OfficerScore
	config.DB.Where("unit_id = ?", unitID).Find(&officerScores)

	var officerSum float64
	var officerCount int
	for _, os := range officerScores {
		officerSum += os.BayesianScore
		officerCount++
	}
	var officerAvg float64
	if officerCount > 0 {
		officerAvg = officerSum / float64(officerCount)
	}

	// 2. Direct unit ratings
	dSum, directCount, directBayes := s.Ratings.AggregateForUnit(unitID)
	var directAvg float64
	if directCount > 0 {
		directAvg = dSum / float64(directCount)
	}

	// 3. Velocity — closed-via-platform / total cases
	var totalCases int64
	var closedViaPlatform int64
	config.DB.Model(&models.Case{}).Where("unit_id = ?", unitID).Count(&totalCases)
	config.DB.Model(&models.Case{}).
		Where("unit_id = ? AND closed_via_platform = ?", unitID, true).
		Count(&closedViaPlatform)
	var velocity float64
	if totalCases > 0 {
		velocity = float64(closedViaPlatform) / float64(totalCases)
	}

	// 4. Composite — 50% officer, 30% direct, 20% velocity (velocity scaled 0..5)
	composite := 0.5*officerAvg + 0.3*directBayes + 0.2*(velocity*5.0)
	composite = math.Round(composite*100) / 100

	tier := TierForUnit(composite, int64(directCount+int64(officerCount)))

	var existing models.UnitScore
	err := config.DB.Where("unit_id = ?", unitID).First(&existing).Error
	if err == nil {
		existing.OfficerAvg = officerAvg
		existing.OfficerCount = officerCount
		existing.BayesianOfficer = officerAvg
		existing.DirectAvg = directAvg
		existing.DirectCount = int(directCount)
		existing.BayesianDirect = directBayes
		existing.VelocityRatio = velocity
		existing.CompositeScore = composite
		existing.Tier = tier
		existing.LastComputedAt = now
		config.DB.Save(&existing)
	} else {
		score := models.UnitScore{
			UnitID:          unitID,
			OfficerAvg:      officerAvg,
			OfficerCount:    officerCount,
			BayesianOfficer: officerAvg,
			DirectAvg:       directAvg,
			DirectCount:     int(directCount),
			BayesianDirect:  directBayes,
			VelocityRatio:   velocity,
			CompositeScore:  composite,
			Tier:            tier,
			LastComputedAt:  now,
		}
		config.DB.Create(&score)
	}
}

// TierForUnit maps a composite score + combined count to a tier label.
func TierForUnit(score float64, count int64) string {
	return tierForScore(score, count)
}

// RecomputeAll rebuilds every officer score and every unit score.
// Called by the scheduler every hour.
func (s *RankingService) RecomputeAll() error {
	if err := s.RecomputeAllOfficers(); err != nil {
		return err
	}
	return s.RecomputeAllUnits()
}
