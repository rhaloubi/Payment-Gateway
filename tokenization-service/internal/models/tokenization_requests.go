package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenizationRequest struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenID    uuid.UUID `gorm:"type:uuid;not null;index"`

	RequestID string         `gorm:"type:varchar(100);not null;index"`
	IPAddress string         `gorm:"type:varchar(45)"`
	UserAgent sql.NullString `gorm:"type:text"`

	CardBrand   CardBrand `gorm:"type:varchar(20)"`
	Last4Digits string    `gorm:"type:char(4)"`
	ExpiryMonth int       `gorm:"type:integer"`
	ExpiryYear  int       `gorm:"type:integer"`

	Success      bool           `gorm:"type:boolean;not null"`
	ErrorCode    sql.NullString `gorm:"type:varchar(50)"`
	ErrorMessage sql.NullString `gorm:"type:text"`

	ProcessingTime int64 `gorm:"type:bigint"`

	Token *CardVault `gorm:"foreignKey:TokenID"`

	CreatedAt time.Time `gorm:"not null;default:now();index"`
}

func (TokenizationRequest) TableName() string {
	return "tokenization_requests"
}

func (tr *TokenizationRequest) BeforeCreate(tx *gorm.DB) error {
	if tr.ID == uuid.Nil {
		tr.ID = uuid.New()
	}
	return nil
}
