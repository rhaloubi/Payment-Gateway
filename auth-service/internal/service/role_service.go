package service

import (
	"errors"

	"github.com/google/uuid"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/repository"
)

type RoleService struct {
	roleRepo     *repository.RoleRepository
	userRoleRepo *repository.UserRoleRepository
}

// NewRoleService creates a new role service
func NewRoleService() *RoleService {
	return &RoleService{
		roleRepo:     repository.NewRoleRepository(),
		userRoleRepo: repository.NewUserRoleRepository(),
	}
}

// GetAllRoles gets all available roles
func (s *RoleService) GetAllRoles() ([]model.Role, error) {
	return s.roleRepo.FindAll()
}

// GetRoleByID gets a role by ID
func (s *RoleService) GetRoleByID(roleID uuid.UUID) (*model.Role, error) {
	return s.roleRepo.FindByID(roleID)
}

// GetRoleByName gets a role by name
func (s *RoleService) GetRoleByName(name string) (*model.Role, error) {
	return s.roleRepo.FindByName(name)
}

// GetRoleWithPermissions gets a role with all its permissions
func (s *RoleService) GetRoleWithPermissions(roleID uuid.UUID) (*model.Role, error) {
	return s.roleRepo.GetRoleWithPermissions(roleID)
}

// AssignRoleToUser assigns a role to a user in a merchant
func (s *RoleService) AssignRoleToUser(userID, roleID, merchantID, assignedBy uuid.UUID) error {
	// Verify role exists
	_, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		return errors.New("role not found")
	}

	return s.userRoleRepo.AssignRoleToUser(userID, roleID, merchantID, assignedBy)
}

// RemoveRoleFromUser removes a role from a user
func (s *RoleService) RemoveRoleFromUser(userID, roleID, merchantID uuid.UUID) error {
	return s.userRoleRepo.RemoveRoleFromUser(userID, roleID, merchantID)
}

// GetUserRoles gets all roles for a user in a merchant
func (s *RoleService) GetUserRoles(userID, merchantID uuid.UUID) ([]model.Role, error) {
	return s.userRoleRepo.GetUserRoles(userID, merchantID)
}

// GetUserPermissions gets all permissions for a user in a merchant
func (s *RoleService) GetUserPermissions(userID, merchantID uuid.UUID) ([]model.Permission, error) {
	return s.userRoleRepo.GetUserPermissions(userID, merchantID)
}

// HasPermission checks if a user has a specific permission
func (s *RoleService) HasPermission(userID, merchantID uuid.UUID, resource, action string) (bool, error) {
	return s.userRoleRepo.HasPermission(userID, merchantID, resource, action)
}

// UpdateUserRole changes a user's role in a merchant
func (s *RoleService) UpdateUserRole(userID, oldRoleID, newRoleID, merchantID uuid.UUID) error {
	// Verify new role exists
	_, err := s.roleRepo.FindByID(newRoleID)
	if err != nil {
		return errors.New("new role not found")
	}

	return s.userRoleRepo.UpdateUserRole(userID, oldRoleID, newRoleID, merchantID)
}
