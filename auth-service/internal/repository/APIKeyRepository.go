package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"gorm.io/gorm"
)

type APIKeyRepository struct{}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository() *APIKeyRepository {
	return &APIKeyRepository{}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(apiKey *model.APIKey) error {
	return inits.DB.Create(apiKey).Error
}

// FindByID finds an API key by ID
func (r *APIKeyRepository) FindByID(id uuid.UUID) (*model.APIKey, error) {
	var apiKey model.APIKey
	err := inits.DB.Where("id = ?", id).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("api key not found")
		}
		return nil, err
	}
	return &apiKey, nil
}

// FindByKeyHash finds an API key by its hash
func (r *APIKeyRepository) FindByKeyHash(keyHash string) (*model.APIKey, error) {
	var apiKey model.APIKey
	err := inits.DB.Where("key_hash = ?", keyHash).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("api key not found")
		}
		return nil, err
	}
	return &apiKey, nil
}

// FindByMerchantID finds all API keys for a merchant
func (r *APIKeyRepository) FindByMerchantID(merchantID uuid.UUID) ([]model.APIKey, error) {
	var apiKeys []model.APIKey
	err := inits.DB.Where("merchant_id = ? AND is_active = true", merchantID).
		Order("created_at DESC").
		Find(&apiKeys).Error
	return apiKeys, err
}

// Update updates an API key
func (r *APIKeyRepository) Update(apiKey *model.APIKey) error {
	return inits.DB.Save(apiKey).Error
}

// Deactivate deactivates an API key
func (r *APIKeyRepository) Deactivate(id uuid.UUID) error {
	return inits.DB.Model(&model.APIKey{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// Delete deletes an API key
func (r *APIKeyRepository) Delete(id uuid.UUID) error {
	return inits.DB.Where("id = ?", id).Delete(&model.APIKey{}).Error
}

// UpdateLastUsed updates the last used timestamp
func (r *APIKeyRepository) UpdateLastUsed(id uuid.UUID) error {
	return inits.DB.Model(&model.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", time.Now()).Error
}

// IsKeyValid checks if an API key is valid
func (r *APIKeyRepository) IsKeyValid(keyHash string) (bool, error) {
	var count int64
	err := inits.DB.Model(&model.APIKey{}).
		Where("key_hash = ? AND is_active = true", keyHash).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
