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

type UserRepository struct{}

// NewUserRepository creates a new user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Cache keys
const (
	userCacheKeyByID    = "user:id:%s"
	userCacheKeyByEmail = "user:email:%s"
	userCacheTTL        = 15 * time.Minute
)

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	err := inits.DB.Create(user).Error
	if err != nil {
		return err
	}

	// Invalidate cache after create
	r.invalidateUserCache(user.ID, user.Email)

	return nil
}

// FindByID finds a user by ID (with Redis caching)
func (r *UserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf(userCacheKeyByID, id.String())
	cachedUser, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedUser != "" {
		// Cache hit - unmarshal and return
		var user model.User
		if err = json.Unmarshal([]byte(cachedUser), &user); err == nil {
			return &user, nil
		}
	}

	// Cache miss - get from database
	var user model.User
	err = inits.DB.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Store in cache
	r.cacheUser(&user)

	return &user, nil
}

// FindByEmail finds a user by email (with Redis caching)
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf(userCacheKeyByEmail, email)
	cachedUser, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedUser != "" {
		// Cache hit
		var user model.User
		if err = json.Unmarshal([]byte(cachedUser), &user); err == nil {
			return &user, nil
		}
	}

	// Cache miss - get from database
	var user model.User
	err = inits.DB.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Store in cache
	r.cacheUser(&user)

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *model.User) error {
	err := inits.DB.Save(user).Error
	if err != nil {
		return err
	}

	// Invalidate cache after update
	r.invalidateUserCache(user.ID, user.Email)

	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uuid.UUID) error {
	// Get user first to get email for cache invalidation
	user, err := r.FindByID(id)
	if err != nil {
		return err
	}

	err = inits.DB.Where("id = ?", id).Delete(&model.User{}).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateUserCache(id, user.Email)

	return nil
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
	err := inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at":         time.Now(),
			"last_login_ip":         ipAddress,
			"failed_login_attempts": 0, // Reset failed attempts on successful login
		}).Error

	if err != nil {
		return err
	}

	// Invalidate user cache
	user, _ := r.FindByID(userID)
	if user != nil {
		r.invalidateUserCache(userID, user.Email)
	}

	return nil
}

// IncrementFailedLoginAttempts increments failed login attempts
func (r *UserRepository) IncrementFailedLoginAttempts(userID uuid.UUID) error {
	err := inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("failed_login_attempts", gorm.Expr("failed_login_attempts + ?", 1)).
		Error

	if err != nil {
		return err
	}

	// Invalidate user cache
	user, _ := r.FindByID(userID)
	if user != nil {
		r.invalidateUserCache(userID, user.Email)
	}

	return nil
}

// LockAccount locks a user account until the specified time
func (r *UserRepository) LockAccount(userID uuid.UUID, until time.Time) error {
	err := inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Update("locked_until", until).
		Error

	if err != nil {
		return err
	}

	// Invalidate user cache
	user, _ := r.FindByID(userID)
	if user != nil {
		r.invalidateUserCache(userID, user.Email)
	}

	return nil
}

// UnlockAccount unlocks a user account
func (r *UserRepository) UnlockAccount(userID uuid.UUID) error {
	err := inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"locked_until":          nil,
			"failed_login_attempts": 0,
		}).Error

	if err != nil {
		return err
	}

	// Invalidate user cache
	user, _ := r.FindByID(userID)
	if user != nil {
		r.invalidateUserCache(userID, user.Email)
	}

	return nil
}

// VerifyEmail marks a user's email as verified
func (r *UserRepository) VerifyEmail(userID uuid.UUID) error {
	err := inits.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"email_verified": true,
			"status":         model.UserStatusActive,
		}).Error

	if err != nil {
		return err
	}

	// Invalidate user cache
	user, _ := r.FindByID(userID)
	if user != nil {
		r.invalidateUserCache(userID, user.Email)
	}

	return nil
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

// Helper: Cache user in Redis
func (r *UserRepository) cacheUser(user *model.User) {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return // Silent fail - caching is not critical
	}

	// Cache by ID
	cacheKeyID := fmt.Sprintf(userCacheKeyByID, user.ID.String())
	inits.RDB.Set(inits.Ctx, cacheKeyID, userJSON, userCacheTTL)

	// Cache by email
	cacheKeyEmail := fmt.Sprintf(userCacheKeyByEmail, user.Email)
	inits.RDB.Set(inits.Ctx, cacheKeyEmail, userJSON, userCacheTTL)
}

// Helper: Invalidate user cache
func (r *UserRepository) invalidateUserCache(userID uuid.UUID, email string) {
	cacheKeyID := fmt.Sprintf(userCacheKeyByID, userID.String())
	cacheKeyEmail := fmt.Sprintf(userCacheKeyByEmail, email)

	inits.RDB.Del(inits.Ctx, cacheKeyID, cacheKeyEmail)
}
