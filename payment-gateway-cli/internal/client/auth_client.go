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
	ID    string `json:"ID"`
	Name  string `json:"Name"`
	Email string `json:"Email"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

// Register registers a new user with the provided email, name, and password.
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

// Login authenticates a user with the provided email and password.
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

// GetUserProfile retrieves the profile of a user with the provided email.
func (c *AuthClient) GetUserProfile(email string) (*User, error) {
	accessToken := config.GetAccessToken()

	resp, err := c.get("/api/v1/auth/profile", accessToken)
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
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to get user profile")
	}

	if result.Data.User.Name == "" {
	}

	return &result.Data.User, nil
}

// Logout logs out the user by invalidating the access token.
func (c *AuthClient) Logout() error {
	accessToken := config.GetAccessToken()

	resp, err := c.post("/api/v1/auth/logout", map[string]string{}, accessToken)
	if err != nil {
		return err
	}

	var result struct {
		Success bool `json:"success"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}
	config.ClearCredentials()

	if !result.Success {
		return fmt.Errorf("logout failed")
	}

	return nil
}

// ChangePassword changes the password of a user with the provided email.
func (c *AuthClient) ChangePassword(email, oldPassword, newPassword string) error {
	accessToken := config.GetAccessToken()

	payload := map[string]string{
		"email":        email,
		"old_password": oldPassword,
		"new_password": newPassword,
	}

	resp, err := c.post("/api/v1/auth/change-password", payload, accessToken)
	if err != nil {
		return err
	}

	var result struct {
		Success bool `json:"success"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("failed to change password")
	}

	return nil
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

func (c *AuthClient) get(endpoint string, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

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
