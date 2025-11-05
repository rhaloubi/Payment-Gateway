package model

import (
	"time"

	"github.com/google/uuid"
)

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`

	// Relationships
	Role       *Role       `gorm:"foreignKey:RoleID"`
	Permission *Permission `gorm:"foreignKey:PermissionID"`
}

// TableName specifies the table name for RolePermission
func (RolePermission) TableName() string {
	return "role_permissions"
}
