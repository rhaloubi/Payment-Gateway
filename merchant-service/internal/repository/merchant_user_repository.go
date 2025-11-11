package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"gorm.io/gorm"
)

type MerchantUserRepository struct{}

// NewMerchantUserRepository creates a new merchant user repository
func NewMerchantUserRepository() *MerchantUserRepository {
	return &MerchantUserRepository{}
}

// Cache keys for team members
const (
	merchantTeamCacheKey = "merchant:team:%s"
	merchantTeamCacheTTL = 10 * time.Minute
)

// Create adds a user to a merchant team
func (r *MerchantUserRepository) Create(merchantUser *model.MerchantUser) error {
	err := inits.DB.Create(merchantUser).Error
	if err != nil {
		return err
	}

	// Invalidate team cache
	r.invalidateTeamCache(merchantUser.MerchantID)

	return nil
}

// FindByID finds a merchant user by ID
func (r *MerchantUserRepository) FindByID(id uuid.UUID) (*model.MerchantUser, error) {
	var merchantUser model.MerchantUser
	err := inits.DB.Where("id = ? AND deleted_at IS NULL", id).First(&merchantUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("merchant user not found")
		}
		return nil, err
	}
	return &merchantUser, nil
}

// FindByMerchantAndUser finds a specific merchant user relation
func (r *MerchantUserRepository) FindByMerchantAndUser(merchantID, userID uuid.UUID) (*model.MerchantUser, error) {
	var merchantUser model.MerchantUser
	err := inits.DB.Where("merchant_id = ? AND user_id = ? AND deleted_at IS NULL", merchantID, userID).
		First(&merchantUser).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found in merchant team")
		}
		return nil, err
	}
	return &merchantUser, nil
}

// GetTeamMembers gets all team members for a merchant (with Redis caching)
func (r *MerchantUserRepository) GetTeamMembers(merchantID uuid.UUID) ([]model.MerchantUser, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(merchantTeamCacheKey, merchantID.String())
	cachedTeam, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedTeam != "" {
		var team []model.MerchantUser
		if err = json.Unmarshal([]byte(cachedTeam), &team); err == nil {
			return team, nil
		}
	}

	// Get from database
	var team []model.MerchantUser
	err = inits.DB.Where("merchant_id = ? AND deleted_at IS NULL", merchantID).
		Order("created_at ASC").
		Find(&team).Error

	if err != nil {
		return nil, err
	}

	// Cache the team
	teamJSON, _ := json.Marshal(team)
	inits.RDB.Set(inits.Ctx, cacheKey, teamJSON, merchantTeamCacheTTL)

	return team, nil
}

// IsUserInMerchant checks if a user is part of a merchant team
func (r *MerchantUserRepository) IsUserInMerchant(merchantID, userID uuid.UUID) (bool, error) {
	var count int64
	err := inits.DB.Model(&model.MerchantUser{}).
		Where("merchant_id = ? AND user_id = ? AND deleted_at IS NULL", merchantID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update updates a merchant user
func (r *MerchantUserRepository) Update(merchantUser *model.MerchantUser) error {
	err := inits.DB.Save(merchantUser).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTeamCache(merchantUser.MerchantID)

	return nil
}

// Delete removes a user from merchant team (soft delete)
func (r *MerchantUserRepository) Delete(id uuid.UUID) error {
	merchantUser, err := r.FindByID(id)
	if err != nil {
		return err
	}

	err = inits.DB.Where("id = ?", id).Delete(&model.MerchantUser{}).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTeamCache(merchantUser.MerchantID)

	return nil
}

// Helper: Invalidate team cache
func (r *MerchantUserRepository) invalidateTeamCache(merchantID uuid.UUID) {
	cacheKey := fmt.Sprintf(merchantTeamCacheKey, merchantID.String())
	inits.RDB.Del(inits.Ctx, cacheKey)
}
