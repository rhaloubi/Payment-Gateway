package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MerchantStatus represents the status of a merchant account
type MerchantStatus string

const (
	MerchantStatusPendingReview MerchantStatus = "pending_review"
	MerchantStatusActive        MerchantStatus = "active"
	MerchantStatusSuspended     MerchantStatus = "suspended"
	MerchantStatusClosed        MerchantStatus = "closed"
)

// BusinessType represents the type of business
type BusinessType string

const (
	BusinessTypeIndividual     BusinessType = "individual"
	BusinessTypeSoleProprietor BusinessType = "sole_proprietor"
	BusinessTypePartnership    BusinessType = "partnership"
	BusinessTypeCorporation    BusinessType = "corporation"
	BusinessTypeNonProfit      BusinessType = "non_profit"
)

// Merchant represents a business account (merchant)
type Merchant struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	OwnerID      uuid.UUID `gorm:"type:uuid;not null;index"`              // References auth.users
	MerchantCode string    `gorm:"type:varchar(50);uniqueIndex;not null"` // e.g., "mch_abc123"

	// Business info
	BusinessName string         `gorm:"type:varchar(255);not null"`
	LegalName    sql.NullString `gorm:"type:varchar(255)"`
	Email        string         `gorm:"type:varchar(255);not null"`
	Phone        sql.NullString `gorm:"type:varchar(50)"`
	Website      sql.NullString `gorm:"type:varchar(255)"`

	// Status
	Status       MerchantStatus `gorm:"type:varchar(20);not null;default:'pending_review'"`
	BusinessType BusinessType   `gorm:"type:varchar(50);not null;default:'individual'"`

	// Location (Morocco only)
	CountryCode string `gorm:"type:char(2);not null;default:'MA'"` // Always "MA" for Morocco

	// Settings
	CurrencyCode string `gorm:"type:char(3);not null;default:'MAD'"` // Default currency
	Timezone     string `gorm:"type:varchar(50);default:'Africa/Casablanca'"`

	// Relationships
	Settings     *MerchantSettings     `gorm:"foreignKey:MerchantID"`
	BusinessInfo *MerchantBusinessInfo `gorm:"foreignKey:MerchantID"`
	Branding     *MerchantBranding     `gorm:"foreignKey:MerchantID"`
	Verification *MerchantVerification `gorm:"foreignKey:MerchantID"`
	TeamMembers  []MerchantUser        `gorm:"foreignKey:MerchantID"`
	Invitations  []MerchantInvitation  `gorm:"foreignKey:MerchantID"`
	ActivityLogs []MerchantActivityLog `gorm:"foreignKey:MerchantID"`

	// Timestamps
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for Merchant
func (Merchant) TableName() string {
	return "merchants"
}

// BeforeCreate hook
func (m *Merchant) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	// Auto-generate merchant code if not set
	if m.MerchantCode == "" {
		m.MerchantCode = "mch_" + uuid.New().String()[:16]
	}
	// Force Morocco
	m.CountryCode = "MA"
	return nil
}

// IsActive checks if merchant can process payments
func (m *Merchant) IsActive() bool {
	return m.Status == MerchantStatusActive
}
