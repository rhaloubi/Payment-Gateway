package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantActivityLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"` // Who performed the action

	// Action details
	Action       string         `gorm:"type:varchar(100);not null"` // e.g., "settings_updated"
	ResourceType sql.NullString `gorm:"type:varchar(50)"`           // e.g., "merchant_settings"
	ResourceID   sql.NullString `gorm:"type:uuid"`

	// Changes
	Changes []byte `gorm:"type:jsonb"` // JSON: {"old": {...}, "new": {...}}

	// Request context
	IPAddress sql.NullString `gorm:"type:varchar(45)"`
	UserAgent sql.NullString `gorm:"type:text"`

	// Relationships
	Merchant *Merchant `gorm:"foreignKey:MerchantID"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for MerchantActivityLog
func (MerchantActivityLog) TableName() string {
	return "merchant_activity_logs"
}

// BeforeCreate hook
func (mal *MerchantActivityLog) BeforeCreate(tx *gorm.DB) error {
	if mal.ID == uuid.Nil {
		mal.ID = uuid.New()
	}
	return nil
}
