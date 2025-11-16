package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EncryptionKeyMetadata struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index"`

	KeyID      string `gorm:"type:varchar(100);not null;uniqueIndex"` // Reference to key in Vault
	KeyVersion int    `gorm:"type:integer;not null;default:1"`        // For key rotation

	// Key metadata
	Algorithm string `gorm:"type:varchar(50);not null;default:'AES-256-GCM'"` // Encryption algorithm
	Purpose   string `gorm:"type:varchar(50);not null;default:'card_data'"`   // What this key encrypts

	IsActive  bool         `gorm:"type:boolean;not null;default:true;index"`
	RotatedAt sql.NullTime `gorm:"type:timestamp"`
	ExpiresAt sql.NullTime `gorm:"type:timestamp"`

	EncryptedRecords int       `gorm:"type:integer;default:0"`
	LastUsedAt       time.Time `gorm:"type:timestamp"`

	CreatedBy uuid.UUID    `gorm:"type:uuid"`
	RevokedBy uuid.UUID    `gorm:"type:uuid"`
	RevokedAt sql.NullTime `gorm:"type:timestamp"`

	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (EncryptionKeyMetadata) TableName() string {
	return "encryption_key_metadata"
}

func (ekm *EncryptionKeyMetadata) BeforeCreate(tx *gorm.DB) error {
	if ekm.ID == uuid.Nil {
		ekm.ID = uuid.New()
	}
	return nil
}

func (ekm *EncryptionKeyMetadata) IsValid() bool {
	if !ekm.IsActive {
		return false
	}

	if ekm.ExpiresAt.Valid && time.Now().After(ekm.ExpiresAt.Time) {
		return false
	}

	return true
}
