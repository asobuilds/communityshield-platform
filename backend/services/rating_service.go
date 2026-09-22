package services

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

// Bayesian constants. Prior Mean of 3.0 is neutral; Prior Count of 20 means a
// new ratee needs ~20 real ratings before the prior stops dominating.
const (
	BayesianPriorMean  = 3.0
	BayesianPriorCount = 20.0
)

type RatingService struct{}

func NewRatingService() *RatingService {
	return &RatingService{}
}

// BayesianScore computes a smoothed average from a sum and count.
// Score = (sum + priorMean*priorCount) / (count + priorCount)
func BayesianScore(sum float64, count int64) float64 {
	denom := float64(count) + BayesianPriorCount
	if denom <= 0 {
		return BayesianPriorMean
	}
	return (sum + BayesianPriorMean*BayesianPriorCount) / denom
}

// tierForScore maps a Bayesian score + count to a tier label.
func tierForScore(score float64, count int64) string {
	switch {
	case count >= 500 && score >= 4.8:
		return "diamond"
	case count >= 150 && score >= 4.6:
		return "platinum"
	case count >= 50 && score >= 4.3:
		return "gold"
	case count >= 20 && score >= 4.0:
		return "silver"
	case count >= 5 && score >= 3.5:
		return "bronze"
	default:
		return "unranked"
	}
}

// TierForOfficer maps a Bayesian score + count to a tier label.
func TierForOfficer(score float64, count int64) string {
	return tierForScore(score, count)
}

// SubmitRatingRequest is the input to SubmitRating.
type SubmitRatingRequest struct {
	RaterUserID uuid.UUID
	CaseID      uuid.UUID
	TargetID    uuid.UUID
	TargetType  string // "officer" or "unit"
	Rating      int    // 1..5
	Comment     string
}

// SubmitRating validates and stores a new rating.
// Rules enforced server-side:
//   - targetType must be "officer" or "unit"
//   - rating must be 1..5
//   - case must exist, be closed, and have ClosedViaPlatform = true
//   - only the case reporter can rate
//   - only within 30 days of case closure
//   - target must match the case (officer must be assigned, unit must own the case)
//   - one rating per (user, case, target, targetType)
func (s *RatingService) SubmitRating(req SubmitRatingRequest) (*models.Rating, error) {
	if req.TargetType != "officer" && req.TargetType != "unit" {
		return nil, errors.New("targetType must be 'officer' or 'unit'")
	}
	if req.Rating < 1 || req.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	var caseRecord models.Case
	if err := config.DB.First(&caseRecord, "id = ?", req.CaseID).Error; err != nil {
		return nil, errors.New("case not found")
	}

	if caseRecord.ReportedBy != req.RaterUserID {
		return nil, errors.New("only the case reporter may submit a rating")
	}
	if caseRecord.Status != "closed" || caseRecord.ClosedAt == nil {
		return nil, errors.New("case must be closed before rating")
	}
	if !caseRecord.ClosedViaPlatform {
		return nil, errors.New("only cases resolved via the platform may be rated")
	}
	if time.Now().UTC().Sub(*caseRecord.ClosedAt) > 30*24*time.Hour {
		return nil, errors.New("rating window has closed (30 days after case closure)")
	}

	switch req.TargetType {
	case "officer":
		if caseRecord.AssignedTo == nil || *caseRecord.AssignedTo != req.TargetID {
			return nil, errors.New("officer was not assigned to this case")
		}
	case "unit":
		if caseRecord.UnitID != req.TargetID {
			return nil, errors.New("unit does not own this case")
		}
	}

	var existing models.Rating
	err := config.DB.
		Where("user_id = ? AND case_id = ? AND target_id = ? AND target_type = ?",
			req.RaterUserID, req.CaseID, req.TargetID, req.TargetType).
		First(&existing).Error
	if err == nil {
		return nil, errors.New("you have already rated this for this case")
	}

	rating := models.Rating{
		UserID:     req.RaterUserID,
		CaseID:     &req.CaseID,
		TargetID:   req.TargetID,
		TargetType: req.TargetType,
		Rating:     req.Rating,
		Comment:    req.Comment,
		Status:     "active",
	}

	if err := config.DB.Create(&rating).Error; err != nil {
		return nil, err
	}
	return &rating, nil
}

// AggregateForOfficer returns the raw sum, count and Bayesian score for an officer.
func (s *RatingService) AggregateForOfficer(officerID uuid.UUID) (sum float64, count int64, bayes float64) {
	config.DB.Model(&models.Rating{}).
		Where("target_id = ? AND target_type = ? AND status = ?", officerID, "officer", "active").
		Select("COALESCE(SUM(rating), 0)").
		Scan(&sum)
	config.DB.Model(&models.Rating{}).
		Where("target_id = ? AND target_type = ? AND status = ?", officerID, "officer", "active").
		Count(&count)
	bayes = BayesianScore(sum, count)
	bayes = math.Round(bayes*100) / 100
	return
}

// AggregateForUnit returns the raw sum, count and Bayesian score for a unit's
// direct ratings.
func (s *RatingService) AggregateForUnit(unitID uuid.UUID) (sum float64, count int64, bayes float64) {
	config.DB.Model(&models.Rating{}).
		Where("target_id = ? AND target_type = ? AND status = ?", unitID, "unit", "active").
		Select("COALESCE(SUM(rating), 0)").
		Scan(&sum)
	config.DB.Model(&models.Rating{}).
		Where("target_id = ? AND target_type = ? AND status = ?", unitID, "unit", "active").
		Count(&count)
	bayes = BayesianScore(sum, count)
	bayes = math.Round(bayes*100) / 100
	return
}

// FlagRating lets an officer (or their head admin) flag a rating for review.
func (s *RatingService) FlagRating(ratingID uuid.UUID, flaggedByID uuid.UUID, reason string) error {
	var rating models.Rating
	if err := config.DB.First(&rating, "id = ?", ratingID).Error; err != nil {
		return errors.New("rating not found")
	}
	if rating.Status != "active" {
		return errors.New("rating is not active")
	}

	// Authorization: only the rated officer, the unit head admin of that
	// officer's (or unit)'s unit, or a super admin may flag a rating.
	if err := s.canFlag(flaggedByID, &rating); err != nil {
		return err
	}

	now := time.Now().UTC()
	rating.Status = "flagged"
	rating.FlaggedByID = &flaggedByID
	rating.FlaggedReason = reason
	rating.FlaggedAt = &now
	return config.DB.Save(&rating).Error
}

// canFlag enforces who may flag a rating.
func (s *RatingService) canFlag(actorID uuid.UUID, r *models.Rating) error {
	var actor models.User
	if config.DB.First(&actor, "id = ?", actorID).Error == nil {
		if actor.IsSuperAdmin || actor.Role == "super_admin" {
			return nil
		}
	}

	// The rated officer may flag their own rating.
	if r.TargetType == "officer" && r.TargetID == actorID {
		return nil
	}

	// Resolve the unit that owns the rating's target.
	var unitID uuid.UUID
	if r.TargetType == "officer" {
		var officer models.Officer
		if err := config.DB.First(&officer, "id = ?", r.TargetID).Error; err != nil {
			return errors.New("rating target officer not found")
		}
		unitID = officer.UnitID
	} else {
		unitID = r.TargetID
	}

	// Head admin of that unit may flag.
	var membership models.UnitMembership
	if config.DB.
		Where("unit_id = ? AND user_id = ? AND status = ?", unitID, actorID, models.MembershipActive).
		First(&membership).Error == nil && membership.IsHeadAdmin {
		return nil
	}

	return errors.New("only the rated officer, their unit head admin, or a super admin may flag a rating")
}
