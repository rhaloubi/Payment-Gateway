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

type SettingsRepository struct{}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository() *SettingsRepository {
	return &SettingsRepository{}
}

// Cache keys for settings
const (
	settingsCacheKey = "merchant:settings:%s"
	settingsCacheTTL = 30 * time.Minute
)

// Create creates merchant settings
func (r *SettingsRepository) Create(settings *model.MerchantSettings) error {
	err := inits.DB.Create(settings).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateSettingsCache(settings.MerchantID)

	return nil
}

// FindByMerchantID finds settings by merchant ID (with Redis caching)
func (r *SettingsRepository) FindByMerchantID(merchantID uuid.UUID) (*model.MerchantSettings, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(settingsCacheKey, merchantID.String())
	cachedSettings, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedSettings != "" {
		var settings model.MerchantSettings
		if err = json.Unmarshal([]byte(cachedSettings), &settings); err == nil {
			return &settings, nil
		}
	}

	// Get from database
	var settings model.MerchantSettings
	err = inits.DB.Where("merchant_id = ?", merchantID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("settings not found")
		}
		return nil, err
	}

	// Cache the settings
	settingsJSON, _ := json.Marshal(settings)
	inits.RDB.Set(inits.Ctx, cacheKey, settingsJSON, settingsCacheTTL)

	return &settings, nil
}

// Update updates merchant settings
func (r *SettingsRepository) Update(settings *model.MerchantSettings) error {
	err := inits.DB.Save(settings).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateSettingsCache(settings.MerchantID)

	return nil
}

// Helper: Invalidate settings cache
func (r *SettingsRepository) invalidateSettingsCache(merchantID uuid.UUID) {
	cacheKey := fmt.Sprintf(settingsCacheKey, merchantID.String())
	inits.RDB.Del(inits.Ctx, cacheKey)
}
