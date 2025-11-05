package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles all database operations for users
type UserRepository struct{}

// NewUserRepository creates a new user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	return inits.DB.Create(user).Error
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := inits.DB.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	err := inits.DB.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *model.User) error {
	return inits.DB.Save(user).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uuid.UUID) error {
	return inits.DB.Where("id = ?", id).Delete(&model.User{}).Error
}

// ExistsByEmail checks if a user exists with the given email
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := inits.DB.Model(&model.User{}).Where("email = ? AND deleted_at IS NULL", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateLastLogin updates the last login timestamp and IP
func (r *UserRepository) UpdateLastLogin(userID uuid.UUID, ipAddress string) error {
	return inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at":         time.Now(),
			"last_login_ip":         ipAddress,
			"failed_login_attempts": 0, // Reset failed attempts on successful login
		}).Error
}

// IncrementFailedLoginAttempts increments failed login attempts
func (r *UserRepository) IncrementFailedLoginAttempts(userID uuid.UUID) error {
	return inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("failed_login_attempts", gorm.Expr("failed_login_attempts + ?", 1)).
		Error
}

// LockAccount locks a user account until the specified time
func (r *UserRepository) LockAccount(userID uuid.UUID, until time.Time) error {
	return inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Update("locked_until", until).
		Error
}

// UnlockAccount unlocks a user account
func (r *UserRepository) UnlockAccount(userID uuid.UUID) error {
	return inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"locked_until":          nil,
			"failed_login_attempts": 0,
		}).Error
}

// VerifyEmail marks a user's email as verified
func (r *UserRepository) VerifyEmail(userID uuid.UUID) error {
	return inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"email_verified": true,
			"status":         model.UserStatusActive,
		}).Error
}

// GetUserWithRoles gets a user with their roles
func (r *UserRepository) GetUserWithRoles(userID uuid.UUID) (*model.User, error) {
	var user model.User
	err := inits.DB.Preload("Roles").Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
