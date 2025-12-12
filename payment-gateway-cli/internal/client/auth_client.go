package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rhaloubi/payment-gateway-cli/internal/config"
)

type AuthClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewAuthClient() *AuthClient {
	return &AuthClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    config.GetAPIURL(),
	}
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

func (c *AuthClient) Register(email, name, password string) (*User, error) {
	payload := map[string]string{
		"email":    email,
		"name":     name,
		"password": password,
	}

	resp, err := c.post("/api/v1/auth/register", payload, "")
	if err != nil {
		return nil, err
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			User User `json:"user"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("registration failed")
	}

	return &result.Data.User, nil
}

func (c *AuthClient) Login(email, password string) (*Tokens, *User, error) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	resp, err := c.post("/api/v1/auth/login", payload, "")
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			User         User   `json:"user"`
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, nil, err
	}

	if !result.Success {
		return nil, nil, fmt.Errorf("login failed")
	}

	tokens := &Tokens{
		AccessToken:  result.Data.AccessToken,
		RefreshToken: result.Data.RefreshToken,
	}

	return tokens, &result.Data.User, nil
}

func (c *AuthClient) post(endpoint string, payload interface{}, token string) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
