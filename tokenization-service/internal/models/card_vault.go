package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenStatus string

const (
	TokenStatusActive  TokenStatus = "active"
	TokenStatusExpired TokenStatus = "expired"
	TokenStatusRevoked TokenStatus = "revoked"
	TokenStatusUsed    TokenStatus = "used"
)

type CardBrand string

const (
	CardBrandVisa       CardBrand = "visa"
	CardBrandMastercard CardBrand = "mastercard"
	CardBrandUnknown    CardBrand = "unknown"
)

type CardType string

const (
	CardTypeCredit  CardType = "credit"
	CardTypeDebit   CardType = "debit"
	CardTypePrepaid CardType = "prepaid"
	CardTypeUnknown CardType = "unknown"
)

type CardVault struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index"`

	Token       string `gorm:"type:varchar(100);uniqueIndex;not null"` // e.g., tok_live_4xJ3kL9mN2pQ
	TokenPrefix string `gorm:"type:varchar(20);not null;index"`        // e.g., tok_live_, tok_test_

	// Encrypted card data (AES-256-GCM)
	// These fields store base64-encoded encrypted data
	EncryptedCardNumber     string `gorm:"type:text;not null"`
	EncryptedCardholderName string `gorm:"type:text"`
	EncryptedExpiryMonth    string `gorm:"type:text;not null"`
	EncryptedExpiryYear     string `gorm:"type:text;not null"`

	// Encryption metadata
	KeyID                string `gorm:"type:varchar(100);not null"`
	EncryptionKeyVersion int    `gorm:"type:integer;not null;default:1"`

	// Card metadata (unencrypted for display/filtering)
	Last4Digits  string    `gorm:"type:char(4);not null;index"`
	First6Digits string    `gorm:"type:char(6);not null;index"`
	CardBrand    CardBrand `gorm:"type:varchar(20);not null;index"`
	CardType     CardType  `gorm:"type:varchar(20);not null"`
	ExpiryMonth  int       `gorm:"type:integer;not null"`
	ExpiryYear   int       `gorm:"type:integer;not null"`

	// Hash of: card_number + exp_month + exp_year
	Fingerprint string `gorm:"type:varchar(64);not null;index"`

	Status      TokenStatus  `gorm:"type:varchar(20);not null;default:'active';index"`
	IsSingleUse bool         `gorm:"type:boolean;default:false"`
	ExpiresAt   sql.NullTime `gorm:"type:timestamp;index"`

	// Usage tracking
	UsageCount  int          `gorm:"type:integer;default:0"` // How many times token was used
	LastUsedAt  sql.NullTime `gorm:"type:timestamp"`         // Last time token was used for a transaction
	FirstUsedAt sql.NullTime `gorm:"type:timestamp"`         // First time token was used

	// Audit fields
	CreatedBy        uuid.UUID      `gorm:"type:uuid"`
	RevokedBy        uuid.UUID      `gorm:"type:uuid"`
	RevokedAt        sql.NullTime   `gorm:"type:timestamp"`
	RevocationReason sql.NullString `gorm:"type:text"`

	TokenizationRequests []TokenizationRequest `gorm:"foreignKey:TokenID"`
	TokenUsageLogs       []TokenUsageLog       `gorm:"foreignKey:TokenID"`

	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CardVault) TableName() string {
	return "card_vault"
}

func (cv *CardVault) BeforeCreate(tx *gorm.DB) error {
	if cv.ID == uuid.Nil {
		cv.ID = uuid.New()
	}
	return nil
}

func (cv *CardVault) IsValid() bool {
	if cv.Status != TokenStatusActive {
		return false
	}

	if cv.ExpiresAt.Valid && time.Now().After(cv.ExpiresAt.Time) {
		return false
	}

	now := time.Now()
	if cv.ExpiryYear < now.Year() {
		return false
	}
	if cv.ExpiryYear == now.Year() && cv.ExpiryMonth < int(now.Month()) {
		return false
	}

	return true
}

func (cv *CardVault) IsExpired() bool {
	now := time.Now()
	if cv.ExpiryYear < now.Year() {
		return true
	}
	if cv.ExpiryYear == now.Year() && cv.ExpiryMonth < int(now.Month()) {
		return true
	}
	return false
}
