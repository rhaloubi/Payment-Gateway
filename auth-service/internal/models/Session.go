package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Token info
	JWTToken string `gorm:"type:text;not null;index"`

	// Session metadata
	IPAddress sql.NullString `gorm:"type:varchar(45)"`
	UserAgent sql.NullString `gorm:"type:text"`

	// Session control
	ExpiresAt time.Time `gorm:"not null;index"`
	IsRevoked bool      `gorm:"default:false;index"` // Can manually revoke/logout

	// Relationships
	User *User `gorm:"foreignKey:UserID"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for Session
func (Session) TableName() string {
	return "sessions"
}

// BeforeCreate hook
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsActive checks if the session is still active
func (s *Session) IsActive() bool {
	return !s.IsRevoked && time.Now().Before(s.ExpiresAt)
}
