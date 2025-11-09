package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantSettings struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`

	// Payment settings
	PaymentMethods  []byte `gorm:"type:jsonb"` // JSON array: ["card", "bank_transfer"]
	Currencies      []byte `gorm:"type:jsonb"` // JSON array: ["MAD", "USD", "EUR"]
	DefaultCurrency string `gorm:"type:char(3);default:'MAD'"`

	// Display settings
	StatementDescriptor sql.NullString `gorm:"type:varchar(22)"` // Shows on customer card statements (max 22 chars)

	// Webhook settings
	WebhookURL    sql.NullString `gorm:"type:varchar(500)"`
	WebhookSecret sql.NullString `gorm:"type:varchar(255)"` // HMAC secret

	// Notification settings
	NotificationEmail sql.NullString `gorm:"type:varchar(255)"`
	SendEmailReceipts bool           `gorm:"default:true"`

	// Settlement settings
	AutoSettle     bool   `gorm:"default:true"`
	SettleSchedule string `gorm:"type:varchar(20);default:'daily'"` // daily, weekly, monthly

	// Relationships
	Merchant *Merchant `gorm:"foreignKey:MerchantID"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for MerchantSettings
func (MerchantSettings) TableName() string {
	return "merchant_settings"
}

// BeforeCreate hook
func (ms *MerchantSettings) BeforeCreate(tx *gorm.DB) error {
	if ms.ID == uuid.Nil {
		ms.ID = uuid.New()
	}
	return nil
}
