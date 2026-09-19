package models

import (
	"time"

	"github.com/google/uuid"
)

// FinancialLedger is an append-only record of every financial event in a unit.
// Entries are never edited or deleted. Corrections are added as new reversal
// entries that reference the original. No UpdatedAt or DeletedAt fields by
// design — the schema enforces immutability.
type FinancialLedger struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Reference      string     `gorm:"type:varchar(32);not null;uniqueIndex:idx_ledger_reference" json:"reference"`
	UnitID         uuid.UUID  `gorm:"type:uuid;not null;index:idx_ledger_unit" json:"unitId"`
	Year           int        `gorm:"not null;index:idx_ledger_year" json:"year"`
	SequenceNumber int64      `gorm:"not null" json:"sequenceNumber"`
	Direction      string     `gorm:"type:varchar(8);not null;index:idx_ledger_direction" json:"direction"`
	EntryType      string     `gorm:"type:varchar(32);not null;index:idx_ledger_type" json:"entryType"`
	Amount         float64    `gorm:"not null" json:"amount"`
	Currency       string     `gorm:"type:varchar(8);not null;default:'NGN'" json:"currency"`
	Counterparty   string     `gorm:"type:varchar(160)" json:"counterparty"`
	Description    string     `gorm:"type:text" json:"description"`
	SourceType     string     `gorm:"type:varchar(32);index:idx_ledger_source" json:"sourceType"`
	SourceID       *uuid.UUID `gorm:"type:uuid;index:idx_ledger_source" json:"sourceId,omitempty"`
	Status         string     `gorm:"type:varchar(16);not null;default:'posted'" json:"status"`
	ReversalOfID   *uuid.UUID `gorm:"type:uuid;index:idx_ledger_reversal_of" json:"reversalOfId,omitempty"`
	Metadata       string     `gorm:"type:text" json:"metadata,omitempty"`
	CreatedBy      uuid.UUID  `gorm:"type:uuid;not null" json:"createdBy"`
	CreatedAt      time.Time  `gorm:"not null;index:idx_ledger_created" json:"createdAt"`
}

func (FinancialLedger) TableName() string {
	return "financial_ledgers"
}
