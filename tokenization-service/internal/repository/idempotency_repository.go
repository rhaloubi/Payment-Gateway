package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"go.uber.org/zap"
)

type IdempotencyRepository struct {
	ctx context.Context
}

func NewIdempotencyRepository() *IdempotencyRepository {
	return &IdempotencyRepository{
		ctx: context.Background(),
	}
}

const (
	idempotencyKeyPrefix = "idempotency:"
	idempotencyTTL       = 24 * time.Hour
)

func (r *IdempotencyRepository) generateCacheKey(merchantID uuid.UUID, key string) string {
	return fmt.Sprintf("%s%s:%s", idempotencyKeyPrefix, merchantID.String(), key)
}

func (r *IdempotencyRepository) Save(idempotencyKey *model.IdempotencyKey) error {
	cacheKey := r.generateCacheKey(idempotencyKey.MerchantID, idempotencyKey.Key)

	data, err := json.Marshal(idempotencyKey)
	if err != nil {
		logger.Log.Error("Failed to marshal idempotency key", zap.Error(err))
		return fmt.Errorf("failed to serialize idempotency key: %w", err)
	}

	err = inits.RDB.Set(r.ctx, cacheKey, data, idempotencyTTL).Err()
	if err != nil {
		logger.Log.Error("Failed to save idempotency key to Redis", zap.Error(err))
		return fmt.Errorf("failed to save idempotency key: %w", err)
	}

	logger.Log.Debug("Idempotency key saved",
		zap.String("merchant_id", idempotencyKey.MerchantID.String()),
		zap.String("key", idempotencyKey.Key),
		zap.Duration("ttl", idempotencyTTL),
	)

	return nil
}

// Find retrieves idempotency key from Redis
func (r *IdempotencyRepository) Find(merchantID uuid.UUID, key string) (*model.IdempotencyKey, error) {
	cacheKey := r.generateCacheKey(merchantID, key)

	// Get from Redis
	data, err := inits.RDB.Get(r.ctx, cacheKey).Result()
	if err != nil {
		// Key not found (first request with this key)
		return nil, nil
	}

	// Deserialize
	var idempotencyKey model.IdempotencyKey
	if err := json.Unmarshal([]byte(data), &idempotencyKey); err != nil {
		logger.Log.Error("Failed to unmarshal idempotency key", zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize idempotency key: %w", err)
	}

	// Check if expired
	if idempotencyKey.IsExpired() {
		// Delete expired key
		r.Delete(merchantID, key)
		return nil, nil
	}

	logger.Log.Debug("Idempotency key found",
		zap.String("merchant_id", merchantID.String()),
		zap.String("key", key),
	)

	return &idempotencyKey, nil
}

// Delete removes idempotency key from Redis
func (r *IdempotencyRepository) Delete(merchantID uuid.UUID, key string) error {
	cacheKey := r.generateCacheKey(merchantID, key)

	err := inits.RDB.Del(r.ctx, cacheKey).Err()
	if err != nil {
		logger.Log.Error("Failed to delete idempotency key", zap.Error(err))
		return fmt.Errorf("failed to delete idempotency key: %w", err)
	}

	return nil
}

// Exists checks if idempotency key exists without retrieving it
func (r *IdempotencyRepository) Exists(merchantID uuid.UUID, key string) (bool, error) {
	cacheKey := r.generateCacheKey(merchantID, key)

	result, err := inits.RDB.Exists(r.ctx, cacheKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check idempotency key existence: %w", err)
	}

	return result > 0, nil
}

// GetTTL returns remaining TTL for idempotency key
func (r *IdempotencyRepository) GetTTL(merchantID uuid.UUID, key string) (time.Duration, error) {
	cacheKey := r.generateCacheKey(merchantID, key)

	ttl, err := inits.RDB.TTL(r.ctx, cacheKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

func (r *IdempotencyRepository) CountKeys() (int64, error) {
	pattern := idempotencyKeyPrefix + "*"

	// Use SCAN to count keys (more efficient than KEYS for large datasets)
	var cursor uint64
	var count int64

	for {
		keys, nextCursor, err := inits.RDB.Scan(r.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, fmt.Errorf("failed to scan keys: %w", err)
		}

		count += int64(len(keys))
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	return count, nil
}
