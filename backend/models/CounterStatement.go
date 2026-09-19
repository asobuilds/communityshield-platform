package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CounterStatement is a citizen's (a suspect's) one-time counter-statement
// on a case — the "right to respond". One per (case, user) is enforced by a
// composite unique index; in practice a case has at most one suspect linked
// to a given citizen user.
type CounterStatement struct {
	ID      uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CaseID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_counter_case_user" json:"caseId"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_counter_case_user" json:"userId"`
	Content string    `gorm:"type:text;not null" json:"content"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CounterStatement) TableName() string {
	return "counter_statements"
}
