package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantUserStatus string

const (
	MerchantUserStatusActive    MerchantUserStatus = "active"
	MerchantUserStatusPending   MerchantUserStatus = "pending"
	MerchantUserStatusSuspended MerchantUserStatus = "suspended"
)

// MerchantUser represents a team member (user assigned to merchant)
type MerchantUser struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	RoleID     uuid.UUID `gorm:"type:uuid;not null"`              // References auth.roles
	RoleName   string    `gorm:"type:varchar(50);not null;index"` // NEW: Cache role name (Admin, Manager, Staff)

	// Status
	Status MerchantUserStatus `gorm:"type:varchar(20);not null;default:'active'"`

	// Invitation tracking
	InvitedBy uuid.UUID    `gorm:"type:uuid"` // Who invited this user
	InvitedAt time.Time    `gorm:"not null;default:now()"`
	JoinedAt  sql.NullTime `gorm:"type:timestamp"`

	// Relationships
	Merchant *Merchant `gorm:"foreignKey:MerchantID"`

	// Timestamps
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for MerchantUser
func (MerchantUser) TableName() string {
	return "merchant_users"
}

// BeforeCreate hook
func (mu *MerchantUser) BeforeCreate(tx *gorm.DB) error {
	if mu.ID == uuid.Nil {
		mu.ID = uuid.New()
	}
	return nil
}
