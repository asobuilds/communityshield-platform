package models

// LedgerSequence tracks the last sequence number used per year for
// financial ledger references (format NG-YYYY-NNNNNNNN).
// Updated atomically inside the same transaction as the ledger insert
// to guarantee uniqueness under concurrency.
type LedgerSequence struct {
	Year    int64 `gorm:"primaryKey" json:"year"`
	LastSeq int64 `gorm:"not null" json:"lastSeq"`
}

func (LedgerSequence) TableName() string {
	return "ledger_sequences"
}
