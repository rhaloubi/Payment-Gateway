package repository

import (
	"errors"

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

// Create creates a new role
func (r *RoleRepository) Create(role *model.Role) error {
	return inits.DB.Create(role).Error
}

// FindByID finds a role by ID
func (r *RoleRepository) FindByID(id uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := inits.DB.Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// FindByName finds a role by name
func (r *RoleRepository) FindByName(name string) (*model.Role, error) {
	var role model.Role
	err := inits.DB.Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return &role, nil
}

// FindAll gets all roles
func (r *RoleRepository) FindAll() ([]model.Role, error) {
	var roles []model.Role
	err := inits.DB.Find(&roles).Error
	return roles, err
}

// Update updates a role
func (r *RoleRepository) Update(role *model.Role) error {
	return inits.DB.Save(role).Error
}

// Delete deletes a role
func (r *RoleRepository) Delete(id uuid.UUID) error {
	return inits.DB.Where("id = ?", id).Delete(&model.Role{}).Error
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
	return inits.DB.Model(role).Association("Permissions").Append(permission)
}

// RemovePermissionFromRole removes a permission from a role
func (r *RoleRepository) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	role, err := r.FindByID(roleID)
	if err != nil {
		return err
	}

	permission := &model.Permission{ID: permissionID}
	return inits.DB.Model(role).Association("Permissions").Delete(permission)
}
