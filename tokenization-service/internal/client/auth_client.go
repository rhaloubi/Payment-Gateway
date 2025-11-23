package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"go.uber.org/zap"
)

type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient() *AuthClient {
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8001"
	}

	return &AuthClient{
		baseURL: authServiceURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type ValidateJWTRequest struct {
	Token string `json:"token"`
}

type ValidateJWTResponse struct {
	Valid      bool      `json:"valid"`
	UserID     uuid.UUID `json:"user_id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	Email      string    `json:"email"`
	Roles      []string  `json:"roles"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (c *AuthClient) ValidateJWT(token string) (*ValidateJWTResponse, error) {
	url := fmt.Sprintf("%s/internal/v1/auth/validate-jwt", c.baseURL)

	reqBody := ValidateJWTRequest{Token: token}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("Auth service request failed", zap.Error(err))
		return nil, fmt.Errorf("auth service unavailable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Log.Warn("JWT validation failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, errors.New("invalid JWT token")
	}

	var result struct {
		Success bool                `json:"success"`
		Data    ValidateJWTResponse `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, errors.New("JWT validation failed")
	}

	return &result.Data, nil
}

// =========================================================================
// API Key Validation
// =========================================================================

type ValidateAPIKeyRequest struct {
	APIKey string `json:"api_key"`
}

type ValidateAPIKeyResponse struct {
	Valid       bool      `json:"valid"`
	MerchantID  uuid.UUID `json:"merchant_id"`
	KeyID       uuid.UUID `json:"key_id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
}

func (c *AuthClient) ValidateAPIKey(apiKey string) (*ValidateAPIKeyResponse, error) {
	url := fmt.Sprintf("%s/internal/v1/auth/validate-api-key", c.baseURL)

	reqBody := ValidateAPIKeyRequest{APIKey: apiKey}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("Auth service request failed", zap.Error(err))
		return nil, fmt.Errorf("auth service unavailable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Log.Warn("API key validation failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, errors.New("invalid API key")
	}

	var result struct {
		Success bool                   `json:"success"`
		Data    ValidateAPIKeyResponse `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		return nil, errors.New("API key validation failed")
	}

	return &result.Data, nil
}

// =========================================================================
// Permission Check
// =========================================================================

type CheckPermissionRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	Permission string    `json:"permission"` // e.g., "tokenization:create"
}

func (c *AuthClient) CheckPermission(userID, merchantID uuid.UUID, permission string) (bool, error) {
	url := fmt.Sprintf("%s/internal/v1/auth/check-permission", c.baseURL)

	reqBody := CheckPermissionRequest{
		UserID:     userID,
		MerchantID: merchantID,
		Permission: permission,
	}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("Permission check failed", zap.Error(err))
		return false, fmt.Errorf("auth service unavailable: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			HasPermission bool `json:"has_permission"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return false, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Data.HasPermission, nil
}
