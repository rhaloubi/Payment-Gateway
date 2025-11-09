package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantBranding struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`

	// Logo and images
	LogoURL    sql.NullString `gorm:"type:varchar(500)"`
	FaviconURL sql.NullString `gorm:"type:varchar(500)"`

	// Colors (hex codes)
	PrimaryColor   sql.NullString `gorm:"type:varchar(7)"` // e.g., "#3B82F6"
	SecondaryColor sql.NullString `gorm:"type:varchar(7)"`
	AccentColor    sql.NullString `gorm:"type:varchar(7)"`

	// Custom styling
	CustomCSS sql.NullString `gorm:"type:text"`

	// Relationships
	Merchant *Merchant `gorm:"foreignKey:MerchantID"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for MerchantBranding
func (MerchantBranding) TableName() string {
	return "merchant_branding"
}

// BeforeCreate hook
func (mb *MerchantBranding) BeforeCreate(tx *gorm.DB) error {
	if mb.ID == uuid.Nil {
		mb.ID = uuid.New()
	}
	return nil
}
