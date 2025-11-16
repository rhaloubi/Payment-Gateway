package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CardVaultRepository struct{}

func NewCardVaultRepository() *CardVaultRepository {
	return &CardVaultRepository{}
}

const (
	tokenCacheKey = "token:%s"
	tokenCacheTTL = 15 * time.Minute
)

func (r *CardVaultRepository) Create(cardVault *model.CardVault) error {
	err := inits.DB.Create(cardVault).Error
	if err != nil {
		return err
	}

	r.cacheToken(cardVault)

	return nil
}

func (r *CardVaultRepository) FindByToken(token string) (*model.CardVault, error) {
	cacheKey := fmt.Sprintf(tokenCacheKey, token)
	cachedData, err := inits.RDB.Get(inits.Ctx, cacheKey).Result()

	if err == nil && cachedData != "" {
		var cardVault model.CardVault
		if err = json.Unmarshal([]byte(cachedData), &cardVault); err == nil {
			return &cardVault, nil
		}
	}

	var cardVault model.CardVault
	err = inits.DB.Where("token = ? AND deleted_at IS NULL", token).First(&cardVault).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}

	r.cacheToken(&cardVault)

	return &cardVault, nil
}

func (r *CardVaultRepository) FindByID(id uuid.UUID) (*model.CardVault, error) {
	var cardVault model.CardVault
	err := inits.DB.Where("id = ? AND deleted_at IS NULL", id).First(&cardVault).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("card vault entry not found")
		}
		return nil, err
	}
	return &cardVault, nil
}

func (r *CardVaultRepository) FindByFingerprint(merchantID uuid.UUID, fingerprint string) (*model.CardVault, error) {
	var cardVault model.CardVault
	err := inits.DB.Where("merchant_id = ? AND fingerprint = ? AND status = ? AND deleted_at IS NULL",
		merchantID, fingerprint, model.TokenStatusActive).
		First(&cardVault).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cardVault, nil
}

func (r *CardVaultRepository) FindByMerchantAndLast4(merchantID uuid.UUID, last4 string) ([]model.CardVault, error) {
	var cards []model.CardVault
	err := inits.DB.Where("merchant_id = ? AND last4_digits = ? AND status = ? AND deleted_at IS NULL",
		merchantID, last4, model.TokenStatusActive).
		Order("created_at DESC").
		Find(&cards).Error

	return cards, err
}

// Update updates a card vault entry
func (r *CardVaultRepository) Update(cardVault *model.CardVault) error {
	err := inits.DB.Save(cardVault).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTokenCache(cardVault.Token)

	return nil
}

// UpdateStatus updates token status
func (r *CardVaultRepository) UpdateStatus(token string, status model.TokenStatus) error {
	err := inits.DB.Model(&model.CardVault{}).
		Where("token = ?", token).
		Update("status", status).Error

	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTokenCache(token)

	return nil
}

// IncrementUsageCount increments the usage count and updates last used timestamp
func (r *CardVaultRepository) IncrementUsageCount(token string) error {
	err := inits.DB.Model(&model.CardVault{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"usage_count":  gorm.Expr("usage_count + 1"),
			"last_used_at": time.Now(),
		}).Error

	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTokenCache(token)

	return nil
}

// SetFirstUsed sets the first_used_at timestamp if not already set
func (r *CardVaultRepository) SetFirstUsed(token string) error {
	var cardVault model.CardVault
	err := inits.DB.Where("token = ? AND first_used_at IS NULL", token).First(&cardVault).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	err = inits.DB.Model(&cardVault).Update("first_used_at", time.Now()).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTokenCache(token)

	return nil
}

// RevokeToken revokes a token
func (r *CardVaultRepository) RevokeToken(token string, revokedBy uuid.UUID, reason string) error {
	err := inits.DB.Model(&model.CardVault{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"status":            model.TokenStatusRevoked,
			"revoked_by":        revokedBy,
			"revoked_at":        time.Now(),
			"revocation_reason": reason,
		}).Error

	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTokenCache(token)

	return nil
}

// Delete soft deletes a card vault entry
func (r *CardVaultRepository) Delete(id uuid.UUID) error {
	var cardVault model.CardVault
	err := inits.DB.Where("id = ?", id).First(&cardVault).Error
	if err != nil {
		return err
	}

	err = inits.DB.Delete(&cardVault).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	r.invalidateTokenCache(cardVault.Token)

	return nil
}

// CountByMerchant counts active tokens for a merchant
func (r *CardVaultRepository) CountByMerchant(merchantID uuid.UUID) (int64, error) {
	var count int64
	err := inits.DB.Model(&model.CardVault{}).
		Where("merchant_id = ? AND status = ? AND deleted_at IS NULL", merchantID, model.TokenStatusActive).
		Count(&count).Error

	return count, err
}

// FindExpiredTokens finds tokens that have expired
func (r *CardVaultRepository) FindExpiredTokens(limit int) ([]model.CardVault, error) {
	var tokens []model.CardVault
	now := time.Now()

	err := inits.DB.Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?",
		model.TokenStatusActive, now).
		Limit(limit).
		Find(&tokens).Error

	return tokens, err
}

// MarkExpiredTokens marks tokens as expired
func (r *CardVaultRepository) MarkExpiredTokens(tokenIDs []uuid.UUID) error {
	return inits.DB.Model(&model.CardVault{}).
		Where("id IN ?", tokenIDs).
		Update("status", model.TokenStatusExpired).Error
}

func (r *CardVaultRepository) cacheToken(cardVault *model.CardVault) {
	data, err := json.Marshal(cardVault)
	if err != nil {
		logger.Log.Error("Failed to marshal card vault for caching", zap.Error(err))
		return
	}

	cacheKey := fmt.Sprintf(tokenCacheKey, cardVault.Token)
	inits.RDB.Set(inits.Ctx, cacheKey, data, tokenCacheTTL)
}

func (r *CardVaultRepository) invalidateTokenCache(token string) {
	cacheKey := fmt.Sprintf(tokenCacheKey, token)
	inits.RDB.Del(inits.Ctx, cacheKey)
}
