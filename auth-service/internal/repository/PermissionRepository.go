package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"gorm.io/gorm"
)

type PermissionRepository struct{}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository() *PermissionRepository {
	return &PermissionRepository{}
}

// Create creates a new permission
func (r *PermissionRepository) Create(permission *model.Permission) error {
	return inits.DB.Create(permission).Error
}

// FindByID finds a permission by ID
func (r *PermissionRepository) FindByID(id uuid.UUID) (*model.Permission, error) {
	var permission model.Permission
	err := inits.DB.Where("id = ?", id).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}
	return &permission, nil
}

// FindByResourceAndAction finds a permission by resource and action
func (r *PermissionRepository) FindByResourceAndAction(resource, action string) (*model.Permission, error) {
	var permission model.Permission
	err := inits.DB.Where("resource = ? AND action = ?", resource, action).First(&permission).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permission not found")
		}
		return nil, err
	}
	return &permission, nil
}

// FindAll gets all permissions
func (r *PermissionRepository) FindAll() ([]model.Permission, error) {
	var permissions []model.Permission
	err := inits.DB.Find(&permissions).Error
	return permissions, err
}

// FindByResource gets all permissions for a resource
func (r *PermissionRepository) FindByResource(resource string) ([]model.Permission, error) {
	var permissions []model.Permission
	err := inits.DB.Where("resource = ?", resource).Find(&permissions).Error
	return permissions, err
}

// Update updates a permission
func (r *PermissionRepository) Update(permission *model.Permission) error {
	return inits.DB.Save(permission).Error
}

// Delete deletes a permission
func (r *PermissionRepository) Delete(id uuid.UUID) error {
	return inits.DB.Where("id = ?", id).Delete(&model.Permission{}).Error
}
