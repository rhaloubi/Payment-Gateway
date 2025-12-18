package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rhaloubi/api-gateway/internal/config"
)

type AuthClient struct {
	httpClient *http.Client
	config     *config.Config
}

func NewAuthClient(cfg *config.Config) *AuthClient {
	return &AuthClient{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		config: cfg,
	}
}

func (ac *AuthClient) ValidateJWT(ctx context.Context, token string) (map[string]interface{}, error) {
	url := ac.config.Services.Auth.URL + "/internal/validate-jwt"

	reqBody := map[string]string{"token": token}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("token validation failed")
	}

	return result.Data, nil
}

func (ac *AuthClient) ValidateAPIKey(ctx context.Context, apiKey string) (map[string]interface{}, error) {
	url := ac.config.Services.Auth.URL + "/internal/validate-api-key"

	reqBody := map[string]string{"api_key": apiKey}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate API key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API key validation failed: status %d", resp.StatusCode)
	}

	var result struct {
		Success bool                   `json:"success"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API key validation failed")
	}

	return result.Data, nil
}
