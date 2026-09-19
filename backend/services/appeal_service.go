package services

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"security-solution/config"
	"security-solution/models"
)

const appealWindowDays = 14
const appealStaleDays = 30

type AppealService struct{}

func NewAppealService() *AppealService {
	return &AppealService{}
}

type FileAppealInput struct {
	RevocationCycleID uuid.UUID
	AppellantUserID   uuid.UUID
	Reason            string
}

// FileAppeal validates eligibility and creates a pending appeal.
func (s *AppealService) FileAppeal(in FileAppealInput) (*models.Appeal, error) {
	var cycle models.RevocationCycle
	if err := config.DB.First(&cycle, "id = ?", in.RevocationCycleID).Error; err != nil {
		return nil, errors.New("revocation cycle not found")
	}

	// Must be a completed cycle.
	if cycle.CompletedAt == nil {
		return nil, errors.New("revocation cycle has not completed")
	}

	// 14-day filing window.
	if time.Since(*cycle.CompletedAt) > appealWindowDays*24*time.Hour {
		return nil, errors.New("appeal window closed (14 days after cycle completion)")
	}

	// Load the removed membership; must be revoked with RevokedAt set.
	var membership models.UnitMembership
	if err := config.DB.First(&membership, "id = ?", cycle.TargetMembershipID).Error; err != nil {
		return nil, errors.New("target membership not found")
	}
	if membership.Status != models.MembershipRevoked {
		return nil, errors.New("target membership is not revoked")
	}
	if membership.RevokedAt == nil {
		return nil, errors.New("target membership has no revocation timestamp")
	}

	// Only the removed member may file.
	if membership.UserID != in.AppellantUserID {
		return nil, errors.New("only the removed member may file an appeal")
	}

	if len([]rune(in.Reason)) > 2000 {
		return nil, errors.New("reason must be 2000 characters or fewer")
	}

	// One appeal per revocation cycle.
	var existing int64
	if err := config.DB.Model(&models.Appeal{}).
		Where("revocation_cycle_id = ?", in.RevocationCycleID).
		Count(&existing).Error; err == nil && existing > 0 {
		return nil, errors.New("an appeal for this revocation cycle already exists")
	}

	appeal := models.Appeal{
		RevocationCycleID: in.RevocationCycleID,
		AppellantUserID:   in.AppellantUserID,
		FiledAt:           time.Now().UTC(),
		Status:            "pending",
		Reason:            in.Reason,
	}
	if err := config.DB.Create(&appeal).Error; err != nil {
		return nil, err
	}
	return &appeal, nil
}

type DecideAppealInput struct {
	Decision string
	Reason   string
}

// DecideAppeal is restricted to super admin. On overturn, restores the
// target membership; on uphold, leaves it revoked. Notifies the appellant.
func (s *AppealService) DecideAppeal(requestID, deciderID uuid.UUID, in DecideAppealInput) (*models.Appeal, error) {
	var appeal models.Appeal
	if err := config.DB.First(&appeal, "id = ?", requestID).Error; err != nil {
		return nil, errors.New("appeal not found")
	}
	if appeal.Status != "pending" {
		return nil, errors.New("only a pending appeal can be decided")
	}

	var decider models.User
	if err := config.DB.First(&decider, "id = ?", deciderID).Error; err != nil {
		return nil, errors.New("decider not found")
	}
	if !decider.IsSuperAdmin && decider.Role != "super_admin" {
		return nil, errors.New("only a super admin may decide an appeal")
	}

	if in.Decision != "uphold" && in.Decision != "overturn" {
		return nil, errors.New("decision must be 'uphold' or 'overturn'")
	}
	if in.Reason == "" {
		return nil, errors.New("reason is required")
	}

	now := time.Now().UTC()
	if in.Decision == "uphold" {
		appeal.Status = "upheld"
	} else {
		appeal.Status = "overturned"
	}
	appeal.DecisionReason = in.Reason
	appeal.DecidedBy = &deciderID
	appeal.DecidedAt = &now

	if in.Decision == "overturn" {
		// Restore the revocation target membership to active.
		var cycle models.RevocationCycle
		config.DB.First(&cycle, "id = ?", appeal.RevocationCycleID)
		var membership models.UnitMembership
		if err := config.DB.First(&membership, "id = ?", cycle.TargetMembershipID).Error; err == nil {
			membership.Status = models.MembershipActive
			membership.RevokedAt = nil
			membership.CoolingOffUntil = nil
			config.DB.Save(&membership)
		}
	}

	if err := config.DB.Save(&appeal).Error; err != nil {
		return nil, err
	}

	// Notify the appellant of the decision.
	var title, message string
	if in.Decision == "uphold" {
		title = "Your appeal was upheld"
		message = "Your appeal was reviewed and upheld. The removal stands."
	} else {
		title = "Your appeal was overturned"
		message = "Your appeal was reviewed and overturned. Your membership has been restored."
	}
	notification := models.Notification{
		UserID:  appeal.AppellantUserID,
		Title:   title,
		Message: message,
		Type:    "appeal_decision",
		Status:  "unread",
	}
	config.DB.Create(&notification)

	return &appeal, nil
}

// GetAppeal returns a single appeal.
func (s *AppealService) GetAppeal(requestID, actorID uuid.UUID) (*models.Appeal, error) {
	var appeal models.Appeal
	if err := config.DB.First(&appeal, "id = ?", requestID).Error; err != nil {
		return nil, errors.New("appeal not found")
	}

	var actor models.User
	if err := config.DB.First(&actor, "id = ?", actorID).Error; err != nil {
		return nil, errors.New("actor not found")
	}
	if !actor.IsSuperAdmin && actor.Role != "super_admin" && appeal.AppellantUserID != actorID {
		return nil, errors.New("not authorized to view this appeal")
	}
	return &appeal, nil
}

// ListAppeals returns appeals for super admins, optionally filtered by status.
func (s *AppealService) ListAppeals(status string) ([]models.Appeal, error) {
	q := config.DB.Order("filed_at DESC").Limit(100)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var appeals []models.Appeal
	if err := q.Find(&appeals).Error; err != nil {
		return nil, err
	}
	return appeals, nil
}

// EscalateStaleAppeals flags unresolved appeals past the 30-day threshold and
// notifies all super admins of the backlog. Called from the scheduler.
func (s *AppealService) EscalateStaleAppeals() {
	cutoff := time.Now().UTC().AddDate(0, 0, -appealStaleDays)

	type row struct {
		ID uuid.UUID
	}
	var rows []row
	if err := config.DB.Model(&models.Appeal{}).
		Select("id").
		Where("status = ? AND escalated_at IS NULL AND filed_at < ?", "pending", cutoff).
		Find(&rows).Error; err != nil {
		return
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	if len(ids) == 0 {
		return
	}

	now := time.Now().UTC()
	config.DB.Model(&models.Appeal{}).
		Where("status = ? AND escalated_at IS NULL AND filed_at < ?", "pending", cutoff).
		Update("escalated_at", now)

	var supers []models.User
	if err := config.DB.Where("is_super_admin = ? OR role = ?", true, "super_admin").Find(&supers).Error; err != nil {
		return
	}
	title := "Stale appeal backlog"
	message := "One or more appeals are awaiting a decision beyond the 30-day threshold."
	for _, u := range supers {
		notification := models.Notification{
			UserID:  u.ID,
			Title:   title,
			Message: message,
			Type:    "appeal_backlog",
			Status:  "unread",
		}
		config.DB.Create(&notification)
	}
}
