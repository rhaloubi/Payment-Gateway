package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
)

type UserRoleRepository struct{}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository() *UserRoleRepository {
	return &UserRoleRepository{}
}

// AssignRoleToUser assigns a role to a user for a specific merchant
func (r *UserRoleRepository) AssignRoleToUser(userID, roleID, merchantID, assignedBy uuid.UUID) error {
	userRole := &model.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		MerchantID: merchantID,
		AssignedBy: assignedBy,
		AssignedAt: time.Now(),
	}
	return inits.DB.Create(userRole).Error
}

// RemoveRoleFromUser removes a role from a user for a specific merchant
func (r *UserRoleRepository) RemoveRoleFromUser(userID, roleID, merchantID uuid.UUID) error {
	return inits.DB.Where("user_id = ? AND role_id = ? AND merchant_id = ?", userID, roleID, merchantID).
		Delete(&model.UserRole{}).Error
}

// GetUserRoles gets all roles for a user in a specific merchant
func (r *UserRoleRepository) GetUserRoles(userID, merchantID uuid.UUID) ([]model.Role, error) {
	var roles []model.Role
	err := inits.DB.
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND user_roles.merchant_id = ?", userID, merchantID).
		Find(&roles).Error
	return roles, err
}

// GetUserPermissions gets all permissions for a user in a specific merchant
func (r *UserRoleRepository) GetUserPermissions(userID, merchantID uuid.UUID) ([]model.Permission, error) {
	var permissions []model.Permission
	err := inits.DB.
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND user_roles.merchant_id = ?", userID, merchantID).
		Distinct().
		Find(&permissions).Error
	return permissions, err
}

// HasPermission checks if a user has a specific permission in a merchant
func (r *UserRoleRepository) HasPermission(userID, merchantID uuid.UUID, resource, action string) (bool, error) {
	var count int64
	err := inits.DB.
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND user_roles.merchant_id = ?", userID, merchantID).
		Where("permissions.resource = ? AND permissions.action = ?", resource, action).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateUserRole updates a user's role in a merchant
func (r *UserRoleRepository) UpdateUserRole(userID, oldRoleID, newRoleID, merchantID uuid.UUID) error {
	return inits.DB.Model(&model.UserRole{}).
		Where("user_id = ? AND role_id = ? AND merchant_id = ?", userID, oldRoleID, merchantID).
		Update("role_id", newRoleID).Error
}

// GetUsersByRole gets all users with a specific role in a merchant
func (r *UserRoleRepository) GetUsersByRole(roleID, merchantID uuid.UUID) ([]model.User, error) {
	var users []model.User
	err := inits.DB.
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id = ? AND user_roles.merchant_id = ?", roleID, merchantID).
		Where("users.deleted_at IS NULL").
		Find(&users).Error
	return users, err
}
