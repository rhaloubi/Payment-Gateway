package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenUsageLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	TokenID    uuid.UUID `gorm:"type:uuid;not null;index"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index"`

	TransactionID   uuid.UUID `gorm:"type:uuid;index"`  // Reference to transaction
	TransactionType string    `gorm:"type:varchar(50)"` // authorize, sale, etc.
	Amount          int64     `gorm:"type:bigint"`      // Amount in cents
	Currency        string    `gorm:"type:char(3)"`     // ISO 4217 currency code

	UsageType string         `gorm:"type:varchar(50);not null"` // payment, verification, recurring
	IPAddress string         `gorm:"type:varchar(45)"`
	UserAgent sql.NullString `gorm:"type:text"`

	Success   bool           `gorm:"type:boolean;not null"`
	ErrorCode sql.NullString `gorm:"type:text"`

	Token *CardVault `gorm:"foreignKey:TokenID"`

	CreatedAt time.Time `gorm:"not null;default:now();index"`
}

func (TokenUsageLog) TableName() string {
	return "token_usage_logs"
}

func (tul *TokenUsageLog) BeforeCreate(tx *gorm.DB) error {
	if tul.ID == uuid.Nil {
		tul.ID = uuid.New()
	}
	return nil
}
