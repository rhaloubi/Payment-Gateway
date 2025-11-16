package service

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/crypto"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/repository"
	"go.uber.org/zap"
)

type KeyManagementService struct {
	keyRepo           *repository.EncryptionKeyRepository
	encryptionService *crypto.EncryptionService
	keyCache          map[string][]byte
	cacheMutex        sync.RWMutex
	vaultEnabled      bool
}

func NewKeyManagementService() *KeyManagementService {
	vaultEnabled := os.Getenv("VAULT_ENABLED") == "true"

	return &KeyManagementService{
		keyRepo:           repository.NewEncryptionKeyRepository(),
		encryptionService: crypto.NewEncryptionService(),
		keyCache:          make(map[string][]byte),
		vaultEnabled:      vaultEnabled,
	}
}

func (s *KeyManagementService) GetOrCreateMerchantKey(merchantID uuid.UUID) ([]byte, string, error) {
	// Try to get existing active key
	keyMetadata, err := s.keyRepo.FindActiveByMerchant(merchantID)

	if err != nil {
		// No active key found, create one
		logger.Log.Info("No active key found for merchant, creating new one",
			zap.String("merchant_id", merchantID.String()),
		)
		return s.CreateMerchantKey(merchantID)
	}

	// Check if key is still valid
	if !keyMetadata.IsValid() {
		logger.Log.Warn("Active key is expired or invalid, creating new one",
			zap.String("merchant_id", merchantID.String()),
			zap.String("key_id", keyMetadata.KeyID),
		)
		return s.CreateMerchantKey(merchantID)
	}

	// Get the actual key (from cache or Vault)
	key, err := s.GetKeyByID(keyMetadata.KeyID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve key: %w", err)
	}

	return key, keyMetadata.KeyID, nil
}

// GetKeyByID retrieves an encryption key by its ID
func (s *KeyManagementService) GetKeyByID(keyID string) ([]byte, error) {
	// Check cache first
	s.cacheMutex.RLock()
	if cachedKey, exists := s.keyCache[keyID]; exists {
		s.cacheMutex.RUnlock()
		logger.Log.Debug("Key retrieved from cache", zap.String("key_id", keyID))
		return cachedKey, nil
	}
	s.cacheMutex.RUnlock()

	// Get key metadata
	keyMetadata, err := s.keyRepo.FindByKeyID(keyID)
	if err != nil {
		return nil, fmt.Errorf("key metadata not found: %w", err)
	}

	if !keyMetadata.IsValid() {
		return nil, errors.New("key is inactive or expired")
	}

	var key []byte
	if s.vaultEnabled {
		key, err = s.fetchKeyFromVault(keyID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch key from Vault: %w", err)
		}
	} else {
		key, err = s.generateDevelopmentKey(keyID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate development key: %w", err)
		}

		logger.Log.Warn("Using development key generation - NOT PRODUCTION SAFE",
			zap.String("key_id", keyID),
		)
	}

	// Cache the key
	s.cacheMutex.Lock()
	s.keyCache[keyID] = key
	s.cacheMutex.Unlock()

	logger.Log.Debug("Key retrieved and cached", zap.String("key_id", keyID))
	return key, nil
}

// Returns: (key bytes, keyID, error)
func (s *KeyManagementService) CreateMerchantKey(merchantID uuid.UUID) ([]byte, string, error) {
	// Deactivate existing active keys
	existingKeys, _ := s.keyRepo.FindByMerchant(merchantID)
	for _, existingKey := range existingKeys {
		if existingKey.IsActive {
			s.keyRepo.DeactivateKey(existingKey.KeyID)
		}
	}

	keyVersion := len(existingKeys) + 1

	keyID := s.encryptionService.GenerateKeyID(merchantID.String(), keyVersion)

	var key []byte
	var err error

	if s.vaultEnabled {

		key, err = s.createKeyInVault(keyID, merchantID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create key in Vault: %w", err)
		}
	} else {
		key, err = s.encryptionService.GenerateKey()
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate key: %w", err)
		}

		logger.Log.Warn("Generated key locally - NOT PRODUCTION SAFE",
			zap.String("key_id", keyID),
		)
	}

	// Create key metadata
	keyMetadata := &model.EncryptionKeyMetadata{
		MerchantID:       merchantID,
		KeyID:            keyID,
		KeyVersion:       keyVersion,
		Algorithm:        "AES-256-GCM",
		Purpose:          "card_data",
		IsActive:         true,
		EncryptedRecords: 0,
		LastUsedAt:       time.Now(),
	}

	if err := s.keyRepo.Create(keyMetadata); err != nil {
		return nil, "", fmt.Errorf("failed to save key metadata: %w", err)
	}

	s.cacheMutex.Lock()
	s.keyCache[keyID] = key
	s.cacheMutex.Unlock()

	logger.Log.Info("Created new encryption key",
		zap.String("merchant_id", merchantID.String()),
		zap.String("key_id", keyID),
		zap.Int("version", keyVersion),
	)

	return key, keyID, nil
}

func (s *KeyManagementService) RotateMerchantKey(merchantID uuid.UUID, rotatedBy uuid.UUID) (string, error) {
	// Get current active key
	currentKey, err := s.keyRepo.FindActiveByMerchant(merchantID)
	if err != nil {
		return "", fmt.Errorf("no active key found: %w", err)
	}

	// Mark old key as rotated
	currentKey.RotatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	currentKey.IsActive = false
	if err := s.keyRepo.Update(currentKey); err != nil {
		return "", fmt.Errorf("failed to mark key as rotated: %w", err)
	}

	// Create new key
	_, newKeyID, err := s.CreateMerchantKey(merchantID)
	if err != nil {
		return "", fmt.Errorf("failed to create new key: %w", err)
	}

	logger.Log.Info("Key rotated successfully",
		zap.String("merchant_id", merchantID.String()),
		zap.String("old_key_id", currentKey.KeyID),
		zap.String("new_key_id", newKeyID),
		zap.String("rotated_by", rotatedBy.String()),
	)

	return newKeyID, nil
}

// RevokeMerchantKey revokes an encryption key
func (s *KeyManagementService) RevokeMerchantKey(keyID string, revokedBy uuid.UUID) error {
	// Revoke key in metadata
	if err := s.keyRepo.RevokeKey(keyID, revokedBy); err != nil {
		return fmt.Errorf("failed to revoke key: %w", err)
	}

	// Remove from cache
	s.cacheMutex.Lock()
	delete(s.keyCache, keyID)
	s.cacheMutex.Unlock()

	logger.Log.Info("Key revoked",
		zap.String("key_id", keyID),
		zap.String("revoked_by", revokedBy.String()),
	)

	return nil
}

// =========================================================================
// Key Statistics & Monitoring
// =========================================================================

// GetKeyStatistics returns statistics about encryption keys for a merchant
type KeyStatistics struct {
	TotalKeys    int
	ActiveKeys   int
	RotatedKeys  int
	RevokedKeys  int
	OldestKeyAge time.Duration
	LastRotation *time.Time
}

func (s *KeyManagementService) GetKeyStatistics(merchantID uuid.UUID) (*KeyStatistics, error) {
	keys, err := s.keyRepo.FindByMerchant(merchantID)
	if err != nil {
		return nil, err
	}

	stats := &KeyStatistics{
		TotalKeys: len(keys),
	}

	var oldestKey *model.EncryptionKeyMetadata

	for i := range keys {
		key := &keys[i]

		if key.IsActive {
			stats.ActiveKeys++
		}

		if key.RotatedAt.Valid {
			stats.RotatedKeys++
			if stats.LastRotation == nil || key.RotatedAt.Time.After(*stats.LastRotation) {
				stats.LastRotation = &key.RotatedAt.Time
			}
		}

		if key.RevokedAt.Valid {
			stats.RevokedKeys++
		}

		if oldestKey == nil || key.CreatedAt.Before(oldestKey.CreatedAt) {
			oldestKey = key
		}
	}

	if oldestKey != nil {
		stats.OldestKeyAge = time.Since(oldestKey.CreatedAt)
	}

	return stats, nil
}

// CheckKeyRotationNeeded checks if key rotation is needed
func (s *KeyManagementService) CheckKeyRotationNeeded(merchantID uuid.UUID) (bool, string) {
	// Get current active key
	currentKey, err := s.keyRepo.FindActiveByMerchant(merchantID)
	if err != nil {
		return true, "No active key found"
	}

	// Check age (rotate every 90 days)
	keyAge := time.Since(currentKey.CreatedAt)
	if keyAge > 90*24*time.Hour {
		return true, fmt.Sprintf("Key is %d days old, exceeds 90-day limit", int(keyAge.Hours()/24))
	}

	// Check usage count (rotate after 1 million encryptions)
	if currentKey.EncryptedRecords > 1000000 {
		return true, fmt.Sprintf("Key has encrypted %d records, exceeds 1M limit", currentKey.EncryptedRecords)
	}

	return false, ""
}

func (s *KeyManagementService) fetchKeyFromVault(keyID string) ([]byte, error) {
	return nil, errors.New("Vault integration not yet implemented")
}

func (s *KeyManagementService) createKeyInVault(keyID string, merchantID uuid.UUID) ([]byte, error) {
	return nil, errors.New("Vault integration not yet implemented")
}

func (s *KeyManagementService) generateDevelopmentKey(keyID string) ([]byte, error) {
	return s.encryptionService.GenerateKey()
}

func (s *KeyManagementService) ClearKeyCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.keyCache = make(map[string][]byte)

	logger.Log.Info("Key cache cleared")
}

func (s *KeyManagementService) GetCacheSize() int {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	return len(s.keyCache)
}
