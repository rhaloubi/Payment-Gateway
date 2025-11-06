package repository

import (
	"encoding/json"
	"fmt"
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

// Cache keys for user roles/permissions
const (
	userRolesCacheKey       = "user:roles:%s:%s"       // user_id:merchant_id
	userPermissionsCacheKey = "user:permissions:%s:%s" // user_id:merchant_id
	userRoleCacheTTL        = 10 * time.Minute
)

// AssignRoleToUser assigns a role to a user for a specific merchant
func (r *UserRoleRepository) AssignRoleToUser(userID, roleID, merchantID, assignedBy uuid.UUID) error {
	userRole := &model.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		MerchantID: merchantID,
		AssignedBy: assignedBy,
		AssignedAt: time.Now(),
	}

	err := inits.DB.Create(userRole).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateUserRoleCache(userID, merchantID)

	return nil
}

// RemoveRoleFromUser removes a role from a user for a specific merchant
func (r *UserRoleRepository) RemoveRoleFromUser(userID, roleID, merchantID uuid.UUID) error {
	err := inits.DB.Where("user_id = ? AND role_id = ? AND merchant_id = ?", userID, roleID, merchantID).
		Delete(&model.UserRole{}).Error

	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateUserRoleCache(userID, merchantID)

	return nil
}

// GetUserRoles gets all roles for a user in a specific merchant (with Redis caching)
func (r *UserRoleRepository) GetUserRoles(userID, merchantID uuid.UUID) ([]model.Role, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(userRolesCacheKey, userID.String(), merchantID.String())
	cachedRoles, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedRoles != "" {
		var roles []model.Role
		if err = json.Unmarshal([]byte(cachedRoles), &roles); err == nil {
			return roles, nil
		}
	}

	// Get from database
	var roles []model.Role
	err = inits.DB.
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND user_roles.merchant_id = ?", userID, merchantID).
		Find(&roles).Error

	if err != nil {
		return nil, err
	}

	// Cache the roles
	rolesJSON, _ := json.Marshal(roles)
	inits.RDB.Set(inits.Ctx, cacheKey, rolesJSON, userRoleCacheTTL)

	return roles, nil
}

// GetUserPermissions gets all permissions for a user in a specific merchant (with Redis caching)
func (r *UserRoleRepository) GetUserPermissions(userID, merchantID uuid.UUID) ([]model.Permission, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(userPermissionsCacheKey, userID.String(), merchantID.String())
	cachedPermissions, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedPermissions != "" {
		var permissions []model.Permission
		if err = json.Unmarshal([]byte(cachedPermissions), &permissions); err == nil {
			return permissions, nil
		}
	}

	// Get from database
	var permissions []model.Permission
	err = inits.DB.
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND user_roles.merchant_id = ?", userID, merchantID).
		Distinct().
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	// Cache the permissions
	permissionsJSON, _ := json.Marshal(permissions)
	inits.RDB.Set(inits.Ctx, cacheKey, permissionsJSON, userRoleCacheTTL)

	return permissions, nil
}

// HasPermission checks if a user has a specific permission in a merchant (with Redis caching)
func (r *UserRoleRepository) HasPermission(userID, merchantID uuid.UUID, resource, action string) (bool, error) {
	// Try to get from cache (check if permission exists in cached permissions)
	cacheKey := fmt.Sprintf(userPermissionsCacheKey, userID.String(), merchantID.String())
	cachedPermissions, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedPermissions != "" {
		var permissions []model.Permission
		if err = json.Unmarshal([]byte(cachedPermissions), &permissions); err == nil {
			// Check if permission exists in cached list
			for _, perm := range permissions {
				if perm.Resource == resource && perm.Action == action {
					return true, nil
				}
			}
			return false, nil
		}
	}

	// Cache miss - query database
	var count int64
	err = inits.DB.
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
	err := inits.DB.Model(&model.UserRole{}).
		Where("user_id = ? AND role_id = ? AND merchant_id = ?", userID, oldRoleID, merchantID).
		Update("role_id", newRoleID).Error

	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateUserRoleCache(userID, merchantID)

	return nil
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

// Helper: Invalidate user role/permission cache
func (r *UserRoleRepository) invalidateUserRoleCache(userID, merchantID uuid.UUID) {
	rolesKey := fmt.Sprintf(userRolesCacheKey, userID.String(), merchantID.String())
	permissionsKey := fmt.Sprintf(userPermissionsCacheKey, userID.String(), merchantID.String())

	inits.RDB.Del(inits.Ctx, rolesKey, permissionsKey)
}
