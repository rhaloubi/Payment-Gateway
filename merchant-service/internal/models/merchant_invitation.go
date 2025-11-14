package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "pending"
	InvitationStatusAccepted  InvitationStatus = "accepted"
	InvitationStatusExpired   InvitationStatus = "expired"
	InvitationStatusCancelled InvitationStatus = "cancelled"
)

// MerchantInvitation represents an invitation to join a merchant team
type MerchantInvitation struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Invitation details
	Email           string    `gorm:"type:varchar(255);not null;index"`
	RoleID          uuid.UUID `gorm:"type:uuid;not null"`         // Role to be assigned
	RoleName        string    `gorm:"type:varchar(255);not null"` // Role name to be assigned
	InvitationToken string    `gorm:"type:varchar(255);uniqueIndex;not null"`

	// Inviter
	InvitedBy uuid.UUID `gorm:"type:uuid;not null"` // Who sent the invitation

	// Status
	Status InvitationStatus `gorm:"type:varchar(20);not null;default:'pending'"`

	// Expiration
	ExpiresAt  time.Time    `gorm:"not null;index"`
	AcceptedAt sql.NullTime `gorm:"type:timestamp"`

	// Relationships
	Merchant *Merchant `gorm:"foreignKey:MerchantID"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for MerchantInvitation
func (MerchantInvitation) TableName() string {
	return "merchant_invitations"
}

// BeforeCreate hook
func (mi *MerchantInvitation) BeforeCreate(tx *gorm.DB) error {
	if mi.ID == uuid.Nil {
		mi.ID = uuid.New()
	}
	// Auto-generate invitation token if not set
	if mi.InvitationToken == "" {
		mi.InvitationToken = "inv_" + uuid.New().String()
	}
	// Set expiration (7 days from now)
	if mi.ExpiresAt.IsZero() {
		mi.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	}
	return nil
}

// IsExpired checks if the invitation has expired
func (mi *MerchantInvitation) IsExpired() bool {
	return time.Now().After(mi.ExpiresAt)
}

// IsValid checks if invitation can be accepted
func (mi *MerchantInvitation) IsValid() bool {
	return mi.Status == InvitationStatusPending && !mi.IsExpired()
}
