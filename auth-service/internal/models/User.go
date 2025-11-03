package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive              UserStatus = "active"
	UserStatusSuspended           UserStatus = "suspended"
	UserStatusPendingVerification UserStatus = "pending_verification"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Basic info
	Name          string `gorm:"type:varchar(255);not null"`
	Email         string `gorm:"type:varchar(255);not null;uniqueIndex"`
	EmailVerified bool   `gorm:"default:false"`
	PasswordHash  string `gorm:"type:varchar(255);not null"`

	// Status
	Status UserStatus `gorm:"type:varchar(50);default:'pending_verification'"`

	// Failed login tracking (for security)
	FailedLoginAttempts int            `gorm:"default:0"`
	LockedUntil         sql.NullTime   `gorm:"type:timestamp"`
	LastLoginAt         sql.NullTime   `gorm:"type:timestamp"`
	LastLoginIP         sql.NullString `gorm:"type:varchar(45)"`

	// Relationships
	Roles          []Role    `gorm:"many2many:user_roles;"`
	Sessions       []Session `gorm:"foreignKey:UserID"`
	CreatedAPIKeys []APIKey  `gorm:"foreignKey:CreatedBy"`

	// Timestamps
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft delete
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// IsLocked checks if the user account is locked
func (u *User) IsLocked() bool {
	if u.LockedUntil.Valid {
		return time.Now().Before(u.LockedUntil.Time)
	}
	return false
}
