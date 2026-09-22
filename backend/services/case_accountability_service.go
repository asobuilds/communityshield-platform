package services

import (
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"security-solution/config"
	"security-solution/models"
)

const (
	AccountabilityOnTrack   = "on_track"
	AccountabilityDue       = "update_due"
	AccountabilityOverdue   = "update_overdue"
	AccountabilityEscalated = "escalated"
	AccountabilityExempt    = "deescalated"
)

type CaseAccountability struct {
	CaseID             uuid.UUID `json:"caseId"`
	OfficerID          uuid.UUID `json:"officerId"`
	Status             string    `json:"status"`
	WeekStart          time.Time `json:"weekStart"`
	WeekEnd            time.Time `json:"weekEnd"`
	UpdateSubmitted    bool      `json:"updateSubmitted"`
	AdminDeescalations int64     `json:"adminDeescalations"`
	DaysOverdue        int       `json:"daysOverdue"`
}

func currentReportingWeek(now time.Time) (time.Time, time.Time) {
	now = now.UTC()

	weekday := int(now.Weekday())
	daysSinceMonday := (weekday + 6) % 7

	start := time.Date(
		now.Year(),
		now.Month(),
		now.Day()-daysSinceMonday,
		0, 0, 0, 0,
		time.UTC,
	)

	end := start.AddDate(0, 0, 6)

	return start, end
}

func countCaseDeescalations(caseID uuid.UUID) int64 {
	var count int64

	config.DB.Model(&models.CaseReview{}).
		Where(
			"case_id = ? AND decision = ? AND TRIM(comment) <> ?",
			caseID,
			models.CaseReviewDecisionDeescalate,
			"",
		).
		Distinct("admin_id").
		Count(&count)

	return count
}

func caseHasThreeAdminDeescalations(caseID uuid.UUID) bool {
	return countCaseDeescalations(caseID) >= 3
}

func createCaseNotification(
	userID uuid.UUID,
	title string,
	message string,
	notificationType string,
) {
	notification := models.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notificationType,
		Status:  "unread",
	}

	if err := config.DB.Create(&notification).Error; err != nil {
		log.Printf(
			"failed to create accountability notification for %s: %v",
			userID,
			err,
		)
	}
}

func notifyAssignedOfficer(
	caseRecord models.Case,
	title string,
	message string,
	notificationType string,
) {
	if caseRecord.AssignedTo == nil {
		return
	}

	createCaseNotification(
		*caseRecord.AssignedTo,
		title,
		message,
		notificationType,
	)
}

func notifyCaseAdmins(
	caseRecord models.Case,
	title string,
	message string,
	notificationType string,
) {
	var admins []models.User

	config.DB.
		Where("role IN (?)", []string{"admin", "super_admin"}).
		Find(&admins)

	seen := make(map[uuid.UUID]bool)

	for _, admin := range admins {
		seen[admin.ID] = true

		createCaseNotification(
			admin.ID,
			title,
			message,
			notificationType,
		)
	}

	var unitAdmins []models.UnitMembership

	config.DB.
		Where(
			"unit_id = ? AND role = ? AND status = ?",
			caseRecord.UnitID,
			models.UnitRoleAdmin,
			models.MembershipActive,
		).
		Find(&unitAdmins)

	for _, membership := range unitAdmins {
		if seen[membership.UserID] {
			continue
		}

		createCaseNotification(
			membership.UserID,
			title,
			message,
			notificationType,
		)
	}
}

// recordAccountabilityEvent atomically reserves an accountability
// notification event for a specific case, event type, and reporting week.
// The unique database index prevents duplicate events when the worker
// runs repeatedly or multiple instances race.
func recordAccountabilityEvent(
	caseID uuid.UUID,
	officerID *uuid.UUID,
	eventType string,
	weekStart time.Time,
) bool {
	event := models.CaseAccountabilityEvent{
		CaseID:    caseID,
		OfficerID: officerID,
		EventType: eventType,
		WeekStart: weekStart,
		SentAt:    time.Now().UTC(),
	}

	err := config.DB.Create(&event).Error
	if err == nil {
		return true
	}

	// Another worker/run already recorded this event.
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return false
	}

	// PostgreSQL/GORM may expose the unique violation through the driver
	// rather than gorm.ErrDuplicatedKey, so check whether the event now exists.
	var existing models.CaseAccountabilityEvent

	lookupErr := config.DB.
		Where(
			"case_id = ? AND event_type = ? AND week_start = ?",
			caseID,
			eventType,
			weekStart,
		).
		First(&existing).Error

	if lookupErr == nil {
		return false
	}

	log.Printf(
		"failed to record accountability event case=%s type=%s: %v",
		caseID,
		eventType,
		err,
	)

	return false
}

func processCaseAccountability(caseRecord models.Case, now time.Time) {
	if caseRecord.AssignedTo == nil {
		return
	}

	if caseRecord.Status == models.CaseStatusClosed {
		return
	}

	weekStart, weekEnd := currentReportingWeek(now)

	if caseHasThreeAdminDeescalations(caseRecord.ID) {
		return
	}

	var update models.CaseWeeklyUpdate

	err := config.DB.
		Where(
			"case_id = ? AND officer_id = ? AND week_start = ?",
			caseRecord.ID,
			*caseRecord.AssignedTo,
			weekStart,
		).
		First(&update).Error

	if err == nil {
		return
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf(
			"weekly update lookup failed for case %s: %v",
			caseRecord.ID,
			err,
		)
		return
	}

	// The reporting week ends Sunday. The update becomes due Monday.
	dueAt := weekEnd.Add(24 * time.Hour)

	if now.Before(dueAt) {
		return
	}

	overdueDays := int(now.Sub(dueAt).Hours() / 24)

	switch overdueDays {
	case 0:
		if !recordAccountabilityEvent(
			caseRecord.ID,
			caseRecord.AssignedTo,
			models.AccountabilityEventWeeklyDue,
			weekStart,
		) {
			return
		}

		notifyAssignedOfficer(
			caseRecord,
			"Weekly Case Update Due",
			fmt.Sprintf(
				"Your weekly update for case %s is due. Please update the investigation file.",
				caseRecord.TrackingID,
			),
			"case_weekly_update_due",
		)

	case 1:
		if !recordAccountabilityEvent(
			caseRecord.ID,
			caseRecord.AssignedTo,
			models.AccountabilityEventWeeklyOverdue,
			weekStart,
		) {
			return
		}

		notifyAssignedOfficer(
			caseRecord,
			"Weekly Case Update Overdue",
			fmt.Sprintf(
				"Your weekly update for case %s is overdue. Please update the case file immediately.",
				caseRecord.TrackingID,
			),
			"case_weekly_update_overdue",
		)

	case 3:
		if !recordAccountabilityEvent(
			caseRecord.ID,
			caseRecord.AssignedTo,
			models.AccountabilityEventAdminEscalation,
			weekStart,
		) {
			return
		}

		notifyCaseAdmins(
			caseRecord,
			"Case Accountability Alert",
			fmt.Sprintf(
				"Case %s has an overdue weekly officer update for %s.",
				caseRecord.TrackingID,
				caseRecord.Title,
			),
			"case_accountability_escalation",
		)
	}
}

func RunCaseAccountabilityCheck() {
	now := time.Now().UTC()

	var cases []models.Case

	err := config.DB.
		Where(
			"assigned_to IS NOT NULL AND status NOT IN (?)",
			[]string{
				models.CaseStatusClosed,
			},
		).
		Find(&cases).Error

	if err != nil {
		log.Printf("case accountability scan failed: %v", err)
		return
	}

	for _, caseRecord := range cases {
		processCaseAccountability(caseRecord, now)
	}
}

func StartCaseAccountabilityWorker() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("case-accountability: worker panic: %v\n%s", r, debug.Stack())
			}
		}()
		RunCaseAccountabilityCheck()

		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			RunCaseAccountabilityCheck()
		}
	}()

	log.Println("Nativity Guard case accountability worker started")
}

func GetCaseAccountability(caseID uuid.UUID) (*CaseAccountability, error) {
	var caseRecord models.Case

	if err := config.DB.First(&caseRecord, "id = ?", caseID).Error; err != nil {
		return nil, err
	}

	if caseRecord.AssignedTo == nil {
		return &CaseAccountability{
			CaseID: caseID,
			Status: AccountabilityOnTrack,
		}, nil
	}

	weekStart, weekEnd := currentReportingWeek(time.Now().UTC())

	var update models.CaseWeeklyUpdate

	err := config.DB.
		Where(
			"case_id = ? AND officer_id = ? AND week_start = ?",
			caseID,
			*caseRecord.AssignedTo,
			weekStart,
		).
		First(&update).Error

	submitted := err == nil

	deescalations := countCaseDeescalations(caseID)

	status := AccountabilityOnTrack
	daysOverdue := 0

	if deescalations >= 3 {
		status = AccountabilityExempt
	} else if !submitted {
		now := time.Now().UTC()

		if now.After(weekEnd.Add(24 * time.Hour)) {
			status = AccountabilityOverdue
			daysOverdue = int(
				now.Sub(weekEnd.Add(24*time.Hour)).Hours() / 24,
			)
		} else {
			status = AccountabilityDue
		}
	}

	return &CaseAccountability{
		CaseID:             caseID,
		OfficerID:          *caseRecord.AssignedTo,
		Status:             status,
		WeekStart:          weekStart,
		WeekEnd:            weekEnd,
		UpdateSubmitted:    submitted,
		AdminDeescalations: deescalations,
		DaysOverdue:        daysOverdue,
	}, nil
}
