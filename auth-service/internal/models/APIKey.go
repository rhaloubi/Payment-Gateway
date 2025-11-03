package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIKey struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Key details
	KeyHash   string `gorm:"type:varchar(255);not null;uniqueIndex"` // SHA-256 hash of actual key
	KeyPrefix string `gorm:"type:varchar(20);not null"`              // e.g., 'pk_live_', 'sk_test_'
	Name      string `gorm:"type:varchar(100)"`                      // User-friendly name

	// Status
	IsActive  bool         `gorm:"default:true;index"`
	ExpiresAt sql.NullTime `gorm:"type:timestamp;index"`

	// Usage tracking
	LastUsedAt sql.NullTime `gorm:"type:timestamp"`

	// Audit
	CreatedBy uuid.UUID `gorm:"type:uuid"`

	// Relationships
	Creator *User `gorm:"foreignKey:CreatedBy"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for APIKey
func (APIKey) TableName() string {
	return "api_keys"
}

// BeforeCreate hook
func (a *APIKey) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
