package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UnitAdminSeat struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UnitID                uuid.UUID      `gorm:"type:uuid;not null;index:idx_unit_admin_seat_unit;uniqueIndex:idx_unit_admin_seat_number" json:"unitId"`
	ElectionID            uuid.UUID      `gorm:"type:uuid;not null;index:idx_unit_admin_seat_election" json:"electionId"`
	SeatNumber            int            `gorm:"not null;index:idx_unit_admin_seat_number;uniqueIndex:idx_unit_admin_seat_number" json:"seatNumber"`
	RotationGroup         string         `gorm:"not null;index:idx_unit_admin_seat_rotation_group" json:"rotationGroup"`
	MemberID              *uuid.UUID     `gorm:"type:uuid;index:idx_unit_admin_seat_member" json:"memberId,omitempty"`
	TermStart             time.Time      `gorm:"not null;index:idx_unit_admin_seat_term_start" json:"termStart"`
	TermEnd               time.Time      `gorm:"not null;index:idx_unit_admin_seat_term_end" json:"termEnd"`
	Status                string         `gorm:"not null;default:active;index:idx_unit_admin_seat_status" json:"status"`
	ElectedAt             time.Time      `gorm:"not null;default:now()" json:"electedAt"`
	VacatedAt             *time.Time     `gorm:"index:idx_unit_admin_seat_vacated_at" json:"vacatedAt,omitempty"`
	VacancyOfSeatID       *uuid.UUID     `gorm:"type:uuid;index:idx_unit_admin_seat_vacancy_of" json:"vacancyOfSeatId,omitempty"`
	AlternateRank         *int           `gorm:"index:idx_unit_admin_seat_alternate_rank" json:"alternateRank,omitempty"`
	AlternateUserID       *uuid.UUID     `gorm:"type:uuid;index:idx_unit_admin_seat_alternate_user" json:"alternateUserId,omitempty"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}
