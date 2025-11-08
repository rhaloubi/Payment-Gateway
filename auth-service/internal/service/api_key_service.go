package service

import (
	"errors"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/inits/jwt"
	model "github.com/rhaloubi/payment-gateway/auth-service/internal/models"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/repository"
)

type APIKeyService struct {
	apiKeyRepo *repository.APIKeyRepository
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService() *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: repository.NewAPIKeyRepository(),
	}
}

// CreateAPIKeyRequest represents API key creation data
type CreateAPIKeyRequest struct {
	MerchantID uuid.UUID
	Name       string
	CreatedBy  uuid.UUID
}

// CreateAPIKeyResponse represents created API key data
type CreateAPIKeyResponse struct {
	APIKey   *model.APIKey
	PlainKey string // Only returned once!
}

// CreateAPIKey creates a new API key
func (s *APIKeyService) CreateAPIKey(req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	// Generate random API key
	plainKey := s.generateAPIKey()

	// Hash the key for storage
	keyHash := jwt.HashSHA256(plainKey)

	// Determine key prefix
	keyPrefix := "pk_"

	// Create API key
	apiKey := &model.APIKey{
		MerchantID: req.MerchantID,
		KeyHash:    keyHash,
		KeyPrefix:  keyPrefix,
		Name:       req.Name,
		IsActive:   true,
		CreatedBy:  req.CreatedBy,
	}

	if err := s.apiKeyRepo.Create(apiKey); err != nil {
		return nil, err
	}

	return &CreateAPIKeyResponse{
		APIKey:   apiKey,
		PlainKey: plainKey, // Return plain key only once
	}, nil
}

// ValidateAPIKey validates an API key
func (s *APIKeyService) ValidateAPIKey(plainKey string) (*model.APIKey, error) {
	// Hash the provided key
	keyHash := jwt.HashSHA256(plainKey)

	// Check if key is valid
	isValid, err := s.apiKeyRepo.IsKeyValid(keyHash)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, errors.New("invalid or inactive API key")
	}

	// Get full API key details
	apiKey, err := s.apiKeyRepo.FindByKeyHash(keyHash)
	if err != nil {
		return nil, err
	}

	// Update last used timestamp
	s.apiKeyRepo.UpdateLastUsed(apiKey.ID)

	return apiKey, nil
}

// GetMerchantAPIKeys gets all API keys for a merchant
func (s *APIKeyService) GetMerchantAPIKeys(merchantID uuid.UUID) ([]model.APIKey, error) {
	return s.apiKeyRepo.FindByMerchantID(merchantID)
}

// DeactivateAPIKey deactivates an API key
func (s *APIKeyService) DeactivateAPIKey(keyID uuid.UUID) error {
	return s.apiKeyRepo.Deactivate(keyID)
}

// DeleteAPIKey deletes an API key
func (s *APIKeyService) DeleteAPIKey(keyID uuid.UUID) error {
	return s.apiKeyRepo.Delete(keyID)
}

// generateAPIKey generates a random API key
func (s *APIKeyService) generateAPIKey() string {
	// Generate random 32 character string
	randomBytes := uuid.New().String() + uuid.New().String()
	return "pk_" + randomBytes[:32]
}
