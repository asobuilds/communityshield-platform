package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type IdentityVerification struct {
    ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

    UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`

    DocumentType string `gorm:"not null" json:"documentType"`

    // We deliberately do not store the raw government ID number.
    // The submitted number is hashed before storage.
    DocumentNumberHash string `gorm:"not null" json:"-"`

    // Reference to the securely stored identity document.
    // This should eventually point to private object storage.
    DocumentURL string `json:"-"`

    Status string `gorm:"not null;default:pending;index" json:"status"`

    RejectionReason string `json:"rejectionReason,omitempty"`

    VerifiedBy *uuid.UUID `gorm:"type:uuid" json:"verifiedBy,omitempty"`

    SubmittedAt time.Time `gorm:"not null;default:now()" json:"submittedAt"`
    VerifiedAt  *time.Time `json:"verifiedAt,omitempty"`

    CreatedAt time.Time      `json:"createdAt"`
    UpdatedAt time.Time      `json:"updatedAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

    User       User  `gorm:"foreignKey:UserID" json:"-"`
    Verifier   *User `gorm:"foreignKey:VerifiedBy" json:"-"`
}

func (IdentityVerification) TableName() string {
    return "identity_verifications"
}
