package service

import (
	"errors"

	"github.com/google/uuid"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService() *UserService {
	return &UserService{
		userRepo: repository.NewUserRepository(),
	}
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(userID uuid.UUID) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

// GetUserByEmail gets a user by email
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	return s.userRepo.FindByEmail(email)
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(userID uuid.UUID, name, email string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// Check if email is being changed and if it's already taken
	if email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(email)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("email already in use")
		}
	}

	user.Name = name
	user.Email = email

	return s.userRepo.Update(user)
}

// SuspendUser suspends a user account
func (s *UserService) SuspendUser(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	user.Status = model.UserStatusSuspended
	return s.userRepo.Update(user)
}

// ActivateUser activates a user account
func (s *UserService) ActivateUser(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	user.Status = model.UserStatusActive
	return s.userRepo.Update(user)
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(userID uuid.UUID) error {
	return s.userRepo.Delete(userID)
}
