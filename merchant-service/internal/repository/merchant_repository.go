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

// MerchantRepository handles merchant database operations
type MerchantRepository struct{}

// NewMerchantRepository creates a new merchant repository
func NewMerchantRepository() *MerchantRepository {
	return &MerchantRepository{}
}

// Cache keys
const (
	merchantCacheKeyByID   = "merchant:id:%s"
	merchantCacheKeyByCode = "merchant:code:%s"
	userMerchantsCacheKey  = "user:merchants:%s"
	merchantCacheTTL       = 15 * time.Minute
)

// Create creates a new merchant
func (r *MerchantRepository) Create(merchant *model.Merchant) error {
	err := inits.DB.Create(merchant).Error
	if err != nil {
		return err
	}

	// Invalidate user's merchants cache
	r.invalidateUserMerchantsCache(merchant.OwnerID)

	return nil
}

// FindByID finds a merchant by ID (with Redis caching)
func (r *MerchantRepository) FindByID(id uuid.UUID) (*model.Merchant, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(merchantCacheKeyByID, id.String())
	cachedMerchant, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedMerchant != "" {
		var merchant model.Merchant
		if err = json.Unmarshal([]byte(cachedMerchant), &merchant); err == nil {
			return &merchant, nil
		}
	}

	// Get from database
	var merchant model.Merchant
	err = inits.DB.Where("id = ? AND deleted_at IS NULL", id).First(&merchant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("merchant not found")
		}
		return nil, err
	}

	// Cache the merchant
	r.cacheMerchant(&merchant)

	return &merchant, nil
}

// FindByCode finds a merchant by merchant code (with Redis caching)
func (r *MerchantRepository) FindByCode(code string) (*model.Merchant, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(merchantCacheKeyByCode, code)
	cachedMerchant, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedMerchant != "" {
		var merchant model.Merchant
		if err = json.Unmarshal([]byte(cachedMerchant), &merchant); err == nil {
			return &merchant, nil
		}
	}

	// Get from database
	var merchant model.Merchant
	err = inits.DB.Where("merchant_code = ? AND deleted_at IS NULL", code).First(&merchant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("merchant not found")
		}
		return nil, err
	}

	// Cache the merchant
	r.cacheMerchant(&merchant)

	return &merchant, nil
}

// FindByOwnerID finds all merchants owned by a user (with Redis caching)
func (r *MerchantRepository) FindByOwnerID(ownerID uuid.UUID) ([]model.Merchant, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(userMerchantsCacheKey, ownerID.String())
	cachedMerchants, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedMerchants != "" {
		var merchants []model.Merchant
		if err = json.Unmarshal([]byte(cachedMerchants), &merchants); err == nil {
			return merchants, nil
		}
	}

	// Get from database
	var merchants []model.Merchant
	err = inits.DB.Where("owner_id = ? AND deleted_at IS NULL", ownerID).
		Order("created_at DESC").
		Find(&merchants).Error

	if err != nil {
		return nil, err
	}

	// Cache the merchants
	merchantsJSON, _ := json.Marshal(merchants)
	inits.RDB.Set(inits.Ctx, cacheKey, merchantsJSON, merchantCacheTTL)

	return merchants, nil
}

// FindMerchantsWhereUserIsTeamMember finds all merchants where user is a team member
func (r *MerchantRepository) FindMerchantsWhereUserIsTeamMember(userID uuid.UUID) ([]model.Merchant, error) {
	var merchants []model.Merchant
	err := inits.DB.
		Joins("JOIN merchant_users ON merchant_users.merchant_id = merchants.id").
		Where("merchant_users.user_id = ? AND merchant_users.deleted_at IS NULL", userID).
		Where("merchants.deleted_at IS NULL").
		Order("merchants.created_at DESC").
		Find(&merchants).Error

	return merchants, err
}

// Update updates a merchant
func (r *MerchantRepository) Update(merchant *model.Merchant) error {
	err := inits.DB.Save(merchant).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateMerchantCache(merchant.ID, merchant.MerchantCode, merchant.OwnerID)

	return nil
}

// Delete soft deletes a merchant
func (r *MerchantRepository) Delete(id uuid.UUID) error {
	merchant, err := r.FindByID(id)
	if err != nil {
		return err
	}

	err = inits.DB.Where("id = ?", id).Delete(&model.Merchant{}).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateMerchantCache(id, merchant.MerchantCode, merchant.OwnerID)

	return nil
}

// UpdateStatus updates merchant status
func (r *MerchantRepository) UpdateStatus(id uuid.UUID, status model.MerchantStatus) error {
	merchant, err := r.FindByID(id)
	if err != nil {
		return err
	}

	err = inits.DB.Model(&model.Merchant{}).
		Where("id = ?", id).
		Update("status", status).Error

	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateMerchantCache(id, merchant.MerchantCode, merchant.OwnerID)

	return nil
}

// ExistsByCode checks if merchant code already exists
func (r *MerchantRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := inits.DB.Model(&model.Merchant{}).
		Where("merchant_code = ? AND deleted_at IS NULL", code).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMerchantWithRelations gets merchant with all relations loaded
func (r *MerchantRepository) GetMerchantWithRelations(id uuid.UUID) (*model.Merchant, error) {
	var merchant model.Merchant
	err := inits.DB.
		Preload("Settings").
		Preload("BusinessInfo").
		Preload("Branding").
		Preload("Verification").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&merchant).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("merchant not found")
		}
		return nil, err
	}

	return &merchant, nil
}

// Helper: Cache merchant in Redis
func (r *MerchantRepository) cacheMerchant(merchant *model.Merchant) {
	merchantJSON, err := json.Marshal(merchant)
	if err != nil {
		return
	}

	// Cache by ID
	cacheKeyID := fmt.Sprintf(merchantCacheKeyByID, merchant.ID.String())
	inits.RDB.Set(inits.Ctx, cacheKeyID, merchantJSON, merchantCacheTTL)

	// Cache by code
	cacheKeyCode := fmt.Sprintf(merchantCacheKeyByCode, merchant.MerchantCode)
	inits.RDB.Set(inits.Ctx, cacheKeyCode, merchantJSON, merchantCacheTTL)
}

// Helper: Invalidate merchant cache
func (r *MerchantRepository) invalidateMerchantCache(merchantID uuid.UUID, merchantCode string, ownerID uuid.UUID) {
	cacheKeyID := fmt.Sprintf(merchantCacheKeyByID, merchantID.String())
	cacheKeyCode := fmt.Sprintf(merchantCacheKeyByCode, merchantCode)

	inits.RDB.Del(inits.Ctx, cacheKeyID, cacheKeyCode)
	r.invalidateUserMerchantsCache(ownerID)
}

// Helper: Invalidate user merchants cache
func (r *MerchantRepository) invalidateUserMerchantsCache(userID uuid.UUID) {
	cacheKey := fmt.Sprintf(userMerchantsCacheKey, userID.String())
	inits.RDB.Del(inits.Ctx, cacheKey)
}
