package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits/logger"
	"go.uber.org/zap"
)

// AuthServiceClient handles communication with Auth Service
type AuthServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthServiceClient() *AuthServiceClient {
	baseURL := os.Getenv("AUTH_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8001"
	}

	return &AuthServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type AssignMerchantOwnerRoleRequest struct {
	UserID     string `json:"user_id"`
	MerchantID string `json:"merchant_id"`
}

type AssignMerchantOwnerRoleResponse struct {
	Success bool `json:"success"`
	Data    struct {
		UserID     uuid.UUID `json:"user_id"`
		RoleID     uuid.UUID `json:"role_id"`
		RoleName   string    `json:"role_name"`
		MerchantID uuid.UUID `json:"merchant_id"`
	} `json:"data"`
	Message string `json:"message"`
}

func (c *AuthServiceClient) AssignMerchantOwnerRole(userID, merchantID uuid.UUID) error {
	url := fmt.Sprintf("%s/internal/v1/roles/assign-merchant-owner", c.baseURL)

	reqBody := AssignMerchantOwnerRoleRequest{
		UserID:     userID.String(),
		MerchantID: merchantID.String(),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Log.Error("failed to marshal request:", zap.Error(err))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Log.Error("failed to create request:", zap.Error(err))
	}

	req.Header.Set("Content-Type", "application/json")
	// Add internal service authentication header
	// In production, use a shared secret or service mesh mTLS
	req.Header.Set("X-Internal-Service", "merchant-service")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("request failed:", zap.Error(err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("failed to read response:", zap.Error(err))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth service returned error (status %d): %s", resp.StatusCode, string(body))
	}

	var response AssignMerchantOwnerRoleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		logger.Log.Error("failed to unmarshal response:", zap.Error(err))
	}

	if !response.Success {
		logger.Log.Error("role assignment failed:", zap.Error(fmt.Errorf("%s", response.Message)))
	}

	return nil
}

// UserRole represents a user's role
type UserRole struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

// GetUserRoles gets all roles for a user in a merchant
func (c *AuthServiceClient) GetUserRoles(userID, merchantID uuid.UUID) ([]UserRole, error) {
	url := fmt.Sprintf("%s/internal/v1/users/%s/roles?merchant_id=%s",
		c.baseURL, userID.String(), merchantID.String())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Log.Error("failed to create request:", zap.Error(err))
	}

	req.Header.Set("X-Internal-Service", "merchant-service")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("request failed:", zap.Error(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth service returned error (status %d): %s", resp.StatusCode, string(body))
	}

	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Roles []UserRole `json:"roles"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Data.Roles, nil
}
