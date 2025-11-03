package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	UserID     uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	RoleID     uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	AssignedAt time.Time `gorm:"not null;default:now()"`

	// Relationships
	User *User `gorm:"foreignKey:UserID"`
	Role *Role `gorm:"foreignKey:RoleID"`
}

// TableName specifies the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}
