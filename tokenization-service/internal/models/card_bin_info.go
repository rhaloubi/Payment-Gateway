package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CardBINInfo struct {
	ID  uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	BIN string    `gorm:"type:char(6);uniqueIndex;not null"` // First 6 digits

	CardBrand    CardBrand `gorm:"type:varchar(20);not null;index"`
	CardType     CardType  `gorm:"type:varchar(20);not null;index"`
	CardCategory string    `gorm:"type:varchar(50)"`

	// Issuer information
	BankName    string         `gorm:"type:varchar(255)"`
	BankCountry string         `gorm:"type:char(2)"` // ISO 3166-1 alpha-2
	BankWebsite sql.NullString `gorm:"type:varchar(255)"`
	BankPhone   sql.NullString `gorm:"type:varchar(50)"`

	IsContactless bool `gorm:"type:boolean;default:false"`
	IsCommercial  bool `gorm:"type:boolean;default:false"`
	IsPrepaid     bool `gorm:"type:boolean;default:false"`

	// Restrictions
	AllowedCountries []byte `gorm:"type:jsonb"`
	BlockedCountries []byte `gorm:"type:jsonb"`

	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

func (CardBINInfo) TableName() string {
	return "card_bin_info"
}

func (cbi *CardBINInfo) BeforeCreate(tx *gorm.DB) error {
	if cbi.ID == uuid.Nil {
		cbi.ID = uuid.New()
	}
	return nil
}
