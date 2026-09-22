package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

const maxAgeOverrideReason = 1000

type AgeService struct{}

func NewAgeService() *AgeService {
	return &AgeService{}
}

type AgeOverrideInput struct {
	UserID       uuid.UUID
	ActorID      uuid.UUID
	Decision     string
	DateOfBirth  string
	Reason       string
}

// ApplyAgeOverride applies a super-admin age-gating change and writes an
// append-only AgeAudit row. DOB may optionally be corrected here.
func (s *AgeService) ApplyAgeOverride(in AgeOverrideInput) (*models.User, *models.AgeAudit, error) {
	var actor models.User
	if err := config.DB.First(&actor, "id = ?", in.ActorID).Error; err != nil {
		return nil, nil, errors.New("actor not found")
	}
	if !actor.IsSuperAdmin && actor.Role != "super_admin" {
		return nil, nil, errors.New("only a super admin may override age-gating")
	}

	if in.Decision != "verify_dob" && in.Decision != "grant_minor_exception" &&
		in.Decision != "revoke_minor_exception" && in.Decision != "dob_corrected" {
		return nil, nil, errors.New("invalid decision")
	}
	if len([]rune(in.Reason)) > maxAgeOverrideReason {
		return nil, nil, errors.New("reason must be 1000 characters or fewer")
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", in.UserID).Error; err != nil {
		return nil, nil, errors.New("user not found")
	}

	audit := models.AgeAudit{
		UserID:            user.ID,
		ActorID:           actor.ID,
		Action:            in.Decision,
		Reason:            in.Reason,
		OldMinorStatus:    user.MinorStatus,
		OldMinorException: user.MinorExceptionGranted,
	}

	now := time.Now().UTC()

	switch in.Decision {
	case "verify_dob":
		user.DOBVerified = true
	case "grant_minor_exception":
		if user.MinorStatus != "minor" {
			return nil, nil, errors.New("minor exception can only be granted to a minor")
		}
		user.MinorExceptionGranted = true
	case "revoke_minor_exception":
		user.MinorExceptionGranted = false
	case "dob_corrected":
		parsed, err := time.Parse("2006-01-02", in.DateOfBirth)
		if err != nil {
			return nil, nil, errors.New("Invalid date of birth")
		}
		if parsed.After(now) {
			return nil, nil, errors.New("Date of birth cannot be in the future")
		}
		age := YearsBetween(parsed, now)
		if age > 120 {
			return nil, nil, errors.New("Date of birth is not valid")
		}
		if age < 16 {
			return nil, nil, errors.New("Registration refused: users under 16 are not permitted")
		}
		user.DateOfBirth = &parsed
		if age >= 18 {
			user.MinorStatus = "adult"
		} else {
			user.MinorStatus = "minor"
		}
	}

	audit.NewMinorStatus = user.MinorStatus
	audit.NewMinorException = user.MinorExceptionGranted
	audit.CreatedAt = now
	audit.UpdatedAt = now

	if err := config.DB.Save(&user).Error; err != nil {
		return nil, nil, err
	}
	if err := config.DB.Create(&audit).Error; err != nil {
		return nil, nil, err
	}
	return &user, &audit, nil
}

// YearsBetween returns the number of whole years between dob and now.
func YearsBetween(dob, now time.Time) int {
	age := now.Year() - dob.Year()
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		age--
	}
	return age
}

// RecomputeMinorStatuses rolls over minors who have aged out (and adults who
// somehow drifted into "minor") based on the stored DateOfBirth. Users with a
// super-admin-granted MinorExceptionGranted are skipped.
func (s *AgeService) RecomputeMinorStatuses() {
	now := time.Now().UTC()

	type uow struct {
		ID          uuid.UUID
		DateOfBirth *time.Time
		MinorStatus string
		Exception   bool
	}
	var rows []uow
	if err := config.DB.Model(&models.User{}).
		Select("id, date_of_birth, minor_status, minor_exception_granted").
		Where("date_of_birth IS NOT NULL").
		Find(&rows).Error; err != nil {
		return
	}

	for _, r := range rows {
		if r.Exception || r.DateOfBirth == nil {
			continue
		}
		age := YearsBetween(*r.DateOfBirth, now)
		var target string
		if age >= 18 {
			target = "adult"
		} else {
			target = "minor"
		}
		if target == r.MinorStatus {
			continue
		}
		config.DB.Model(&models.User{}).
			Where("id = ?", r.ID).
			Update("minor_status", target)
	}
}
