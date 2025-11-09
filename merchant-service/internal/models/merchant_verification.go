package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VerificationStatus string

const (
	VerificationStatusUnverified VerificationStatus = "unverified"
	VerificationStatusPending    VerificationStatus = "pending"
	VerificationStatusVerified   VerificationStatus = "verified"
	VerificationStatusRejected   VerificationStatus = "rejected"
)

// RiskLevel represents the risk assessment
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

// MerchantVerification represents KYC/business verification status
type MerchantVerification struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`

	// Verification status
	VerificationStatus VerificationStatus `gorm:"type:varchar(20);not null;default:'unverified'"`

	// Verification details
	VerifiedAt      sql.NullTime   `gorm:"type:timestamp"`
	VerifiedBy      sql.NullString `gorm:"type:uuid"` // Admin who verified
	RejectionReason sql.NullString `gorm:"type:text"`

	// Document tracking (simple for now - just count/status)
	DocumentsSubmitted []byte `gorm:"type:jsonb"` // JSON: [{"type": "id_card", "submitted": true}]
	DocumentsRequired  bool   `gorm:"default:true"`

	// Risk assessment
	RiskLevel RiskLevel      `gorm:"type:varchar(20);default:'medium'"`
	RiskNotes sql.NullString `gorm:"type:text"`

	// Limits (based on verification)
	CanProcessLive bool          `gorm:"default:false"` // Can process live transactions
	DailyLimit     sql.NullInt64 `gorm:"type:bigint"`   // In MAD cents
	MonthlyLimit   sql.NullInt64 `gorm:"type:bigint"`   // In MAD cents

	// Relationships
	Merchant *Merchant `gorm:"foreignKey:MerchantID"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for MerchantVerification
func (MerchantVerification) TableName() string {
	return "merchant_verification"
}

// BeforeCreate hook
func (mv *MerchantVerification) BeforeCreate(tx *gorm.DB) error {
	if mv.ID == uuid.Nil {
		mv.ID = uuid.New()
	}
	return nil
}

// IsVerified checks if merchant is verified
func (mv *MerchantVerification) IsVerified() bool {
	return mv.VerificationStatus == VerificationStatusVerified
}
