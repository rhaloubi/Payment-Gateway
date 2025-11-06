package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"gorm.io/gorm"
)

type RoleRepository struct{}

// NewRoleRepository creates a new role repository
func NewRoleRepository() *RoleRepository {
	return &RoleRepository{}
}

// Cache keys for roles
const (
	roleCacheKeyByID   = "role:id:%s"
	roleCacheKeyByName = "role:name:%s"
	rolesCacheKey      = "roles:all"
	roleCacheTTL       = 30 * time.Minute // Roles don't change often
)

// Create creates a new role
func (r *RoleRepository) Create(role *model.Role) error {
	err := inits.DB.Create(role).Error
	if err != nil {
		return err
	}

	// Invalidate all roles cache
	inits.RDB.Del(inits.Ctx, rolesCacheKey)

	return nil
}

// FindByID finds a role by ID (with Redis caching)
func (r *RoleRepository) FindByID(id uuid.UUID) (*model.Role, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(roleCacheKeyByID, id.String())
	cachedRole, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedRole != "" {
		var role model.Role
		if err = json.Unmarshal([]byte(cachedRole), &role); err == nil {
			return &role, nil
		}
	}

	// Get from database
	var role model.Role
	err = inits.DB.Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	// Cache the role
	r.cacheRole(&role)

	return &role, nil
}

// FindByName finds a role by name (with Redis caching)
func (r *RoleRepository) FindByName(name string) (*model.Role, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(roleCacheKeyByName, name)
	cachedRole, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedRole != "" {
		var role model.Role
		if err = json.Unmarshal([]byte(cachedRole), &role); err == nil {
			return &role, nil
		}
	}

	// Get from database
	var role model.Role
	err = inits.DB.Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}

	// Cache the role
	r.cacheRole(&role)

	return &role, nil
}

// FindAll gets all roles (with Redis caching)
func (r *RoleRepository) FindAll() ([]model.Role, error) {
	// Try cache first
	cachedRoles, err := inits.RDB.Get(inits.Ctx, rolesCacheKey).Result()

	if err == nil && cachedRoles != "" {
		var roles []model.Role
		if err = json.Unmarshal([]byte(cachedRoles), &roles); err == nil {
			return roles, nil
		}
	}

	// Get from database
	var roles []model.Role
	err = inits.DB.Find(&roles).Error
	if err != nil {
		return nil, err
	}

	// Cache all roles
	rolesJSON, _ := json.Marshal(roles)
	inits.RDB.Set(inits.Ctx, rolesCacheKey, rolesJSON, roleCacheTTL)

	return roles, nil
}

// Update updates a role
func (r *RoleRepository) Update(role *model.Role) error {
	err := inits.DB.Save(role).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateRoleCache(role.ID, role.Name)

	return nil
}

// Delete deletes a role
func (r *RoleRepository) Delete(id uuid.UUID) error {
	role, err := r.FindByID(id)
	if err != nil {
		return err
	}

	err = inits.DB.Where("id = ?", id).Delete(&model.Role{}).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateRoleCache(id, role.Name)

	return nil
}

// GetRoleWithPermissions gets a role with all its permissions
func (r *RoleRepository) GetRoleWithPermissions(roleID uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := inits.DB.Preload("Permissions").Where("id = ?", roleID).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// AssignPermissionToRole assigns a permission to a role
func (r *RoleRepository) AssignPermissionToRole(roleID, permissionID uuid.UUID) error {
	role, err := r.FindByID(roleID)
	if err != nil {
		return err
	}

	permission := &model.Permission{ID: permissionID}
	err = inits.DB.Model(role).Association("Permissions").Append(permission)

	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateRoleCache(roleID, role.Name)

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *RoleRepository) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	role, err := r.FindByID(roleID)
	if err != nil {
		return err
	}

	permission := &model.Permission{ID: permissionID}
	err = inits.DB.Model(role).Association("Permissions").Delete(permission)

	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateRoleCache(roleID, role.Name)

	return nil
}

// Helper: Cache role in Redis
func (r *RoleRepository) cacheRole(role *model.Role) {
	roleJSON, err := json.Marshal(role)
	if err != nil {
		return
	}

	// Cache by ID
	cacheKeyID := fmt.Sprintf(roleCacheKeyByID, role.ID.String())
	inits.RDB.Set(inits.Ctx, cacheKeyID, roleJSON, roleCacheTTL)

	// Cache by name
	cacheKeyName := fmt.Sprintf(roleCacheKeyByName, role.Name)
	inits.RDB.Set(inits.Ctx, cacheKeyName, roleJSON, roleCacheTTL)
}

// Helper: Invalidate role cache
func (r *RoleRepository) invalidateRoleCache(roleID uuid.UUID, roleName string) {
	cacheKeyID := fmt.Sprintf(roleCacheKeyByID, roleID.String())
	cacheKeyName := fmt.Sprintf(roleCacheKeyByName, roleName)

	inits.RDB.Del(inits.Ctx, cacheKeyID, cacheKeyName, rolesCacheKey)
}
