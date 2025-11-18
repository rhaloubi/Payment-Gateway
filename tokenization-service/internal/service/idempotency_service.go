package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/repository"
	"go.uber.org/zap"
)

type IdempotencyService struct {
	idempotencyRepo *repository.IdempotencyRepository
}

func NewIdempotencyService() *IdempotencyService {
	return &IdempotencyService{
		idempotencyRepo: repository.NewIdempotencyRepository(),
	}
}

// CheckRequest checks if request has been processed before
// Returns:
// - cached response if found
// - nil if this is a new request
// - error if key exists but request body differs (idempotency key reuse)
func (s *IdempotencyService) CheckRequest(
	merchantID uuid.UUID,
	idempotencyKey string,
	requestBody []byte,
	endpoint string,
	method string,
) (*model.IdempotencyKey, error) {

	// Find existing idempotency key
	existing, err := s.idempotencyRepo.Find(merchantID, idempotencyKey)
	if err != nil {
		logger.Log.Error("Failed to check idempotency key", zap.Error(err))
		return nil, err
	}

	// No existing key found - this is a new request
	if existing == nil {
		logger.Log.Debug("New idempotency key (no cache hit)",
			zap.String("merchant_id", merchantID.String()),
			zap.String("key", idempotencyKey),
		)
		return nil, nil
	}

	// Key exists - check if request body matches
	requestHash := s.hashRequest(requestBody)

	if existing.RequestHash != requestHash {
		// Same key, different request = ERROR (key reuse)
		logger.Log.Warn("Idempotency key reused with different request",
			zap.String("merchant_id", merchantID.String()),
			zap.String("key", idempotencyKey),
			zap.String("original_hash", existing.RequestHash),
			zap.String("new_hash", requestHash),
		)

		return nil, errors.New("idempotency key already used for different request")
	}

	// Same key, same request = Return cached response
	logger.Log.Info("Idempotency key cache hit (returning cached response)",
		zap.String("merchant_id", merchantID.String()),
		zap.String("key", idempotencyKey),
		zap.Int("status", existing.ResponseStatus),
	)

	return existing, nil
}

func (s *IdempotencyService) StoreResponse(
	merchantID uuid.UUID,
	idempotencyKey string,
	requestBody []byte,
	responseBody []byte,
	responseStatus int,
	endpoint string,
	method string,
	ipAddress string,
	userAgent string,
) error {

	record := &model.IdempotencyKey{
		ID:             uuid.New(),
		Key:            idempotencyKey,
		MerchantID:     merchantID,
		RequestHash:    s.hashRequest(requestBody),
		ResponseBody:   responseBody,
		ResponseStatus: responseStatus,
		Endpoint:       endpoint,
		Method:         method,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	// Store in Redis
	if err := s.idempotencyRepo.Save(record); err != nil {
		logger.Log.Error("Failed to store idempotency response",
			zap.Error(err),
			zap.String("merchant_id", merchantID.String()),
			zap.String("key", idempotencyKey),
		)
		return err
	}

	logger.Log.Debug("Idempotency response stored",
		zap.String("merchant_id", merchantID.String()),
		zap.String("key", idempotencyKey),
		zap.Int("status", responseStatus),
	)

	return nil
}

// hashRequest generates SHA-256 hash of request body
func (s *IdempotencyService) hashRequest(requestBody []byte) string {
	// Normalize JSON (ensure consistent formatting)
	var normalized interface{}
	if err := json.Unmarshal(requestBody, &normalized); err != nil {
		// If not JSON, hash raw bytes
		hash := sha256.Sum256(requestBody)
		return hex.EncodeToString(hash[:])
	}

	// Re-marshal for consistent formatting
	normalizedBytes, err := json.Marshal(normalized)
	if err != nil {
		// Fallback to raw bytes
		hash := sha256.Sum256(requestBody)
		return hex.EncodeToString(hash[:])
	}

	// Hash normalized JSON
	hash := sha256.Sum256(normalizedBytes)
	return hex.EncodeToString(hash[:])
}

// ValidateIdempotencyKey checks if key format is valid
func (s *IdempotencyService) ValidateIdempotencyKey(key string) error {
	if key == "" {
		return errors.New("idempotency key cannot be empty")
	}

	// Minimum length check
	if len(key) < 16 {
		return errors.New("idempotency key must be at least 16 characters")
	}

	// Maximum length check (prevent abuse)
	if len(key) > 255 {
		return errors.New("idempotency key must be less than 255 characters")
	}

	return nil
}

// GetKeyStatistics returns statistics about idempotency keys (for monitoring)
type IdempotencyStatistics struct {
	TotalKeys     int64
	CacheHitRate  float64
	AverageKeyAge time.Duration
}

func (s *IdempotencyService) GetStatistics() (*IdempotencyStatistics, error) {
	totalKeys, err := s.idempotencyRepo.CountKeys()
	if err != nil {
		return nil, err
	}

	return &IdempotencyStatistics{
		TotalKeys: totalKeys,
		// TODO: Implement cache hit rate tracking
		CacheHitRate: 0.0,
	}, nil
}
