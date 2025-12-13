package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rhaloubi/payment-gateway-cli/internal/config"
)

type MerchantClient struct {
	httpClient *http.Client
	baseURL    string
	restClient *RESTClient
}

func NewMerchantClient() *MerchantClient {
	return &MerchantClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "http://localhost:8002",
		restClient: NewHttpClient(),
	}
}

type Merchant struct {
	ID           string `json:"id"`
	BusinessName string `json:"business_name"`
	LegalName    string `json:"legal_name"`
	Email        string `json:"email"`
	BusinessType string `json:"business_type"`
	Status       string `json:"status"`
	CountryCode  string `json:"country_code"`
	CurrencyCode string `json:"currency_code"`
	OwnerID      string `json:"owner_id"`
	MerchantCode string `json:"merchant_code"`
}
type Invitation struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	Status          string    `json:"status"`
	RoleName        string    `json:"role_name"`
	InvitationToken string    `json:"invitation_token"`
	ExpiresAt       time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at"`
}

func (c *MerchantClient) Create(BusinessName, LegalName, email, BusinessType string) (*Merchant, error) {
	payload := map[string]string{
		"business_name": BusinessName,
		"legal_name":    LegalName,
		"email":         email,
		"business_type": BusinessType,
	}

	// TODO: Implement HTTP POST to merchant service
	resp, err := c.restClient.Post(c.baseURL+"/api/v1/merchants", payload, config.GetAccessToken())
	if err != nil {
		return nil, err
	}
	result := struct {
		Success bool `json:"success"`
		Data    struct {
			Merchant Merchant `json:"merchant"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse roles response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to create merchant")
	}
	return &Merchant{
		ID:           result.Data.Merchant.ID,
		BusinessName: result.Data.Merchant.BusinessName,
		Email:        result.Data.Merchant.Email,
		BusinessType: result.Data.Merchant.BusinessType,
		Status:       result.Data.Merchant.Status,
		OwnerID:      result.Data.Merchant.OwnerID,
	}, nil
}

/*
	func (c *MerchantClient) List() ([]Merchant, error) {
		// TODO: Implement HTTP GET to merchant service
		// For now, return mock data
		return []Merchant{
			{ID: "mer_1", Name: "Test Merchant 1", Email: "test1@example.com", Status: "active"},
			{ID: "mer_2", Name: "Test Merchant 2", Email: "test2@example.com", Status: "active"},
		}, nil
	}
*/
func (c *MerchantClient) GetMerchant(id string) (*Merchant, error) {
	accessToken := config.GetAccessToken()

	resp, err := c.restClient.Get(c.baseURL+"/api/v1/merchants/"+id, accessToken)
	if err != nil {
		return nil, err
	}

	result := struct {
		Success bool `json:"success"`
		Data    struct {
			Merchant Merchant `json:"merchant"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse roles response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to get merchant")
	}
	return &Merchant{
		ID:           result.Data.Merchant.ID,
		BusinessName: result.Data.Merchant.BusinessName,
		LegalName:    result.Data.Merchant.LegalName,
		Email:        result.Data.Merchant.Email,
		BusinessType: result.Data.Merchant.BusinessType,
		Status:       result.Data.Merchant.Status,
		CountryCode:  result.Data.Merchant.CountryCode,
		CurrencyCode: result.Data.Merchant.CurrencyCode,
		OwnerID:      result.Data.Merchant.OwnerID,
		MerchantCode: result.Data.Merchant.MerchantCode,
	}, nil
}

func (c *MerchantClient) InviteUser(merchantID, email, rolename, roleID string) (*Invitation, error) {
	if config.GetAccessToken() == "" {
		return nil, fmt.Errorf("access token not set")
	}
	payload := map[string]string{
		"email":     email,
		"role_name": rolename,
		"role_id":   roleID,
	}

	resp, err := c.restClient.Post(c.baseURL+"/api/v1/merchants/"+merchantID+"/team/invite", payload, config.GetAccessToken())
	if err != nil {
		return nil, err
	}

	result := struct {
		Success bool `json:"success"`
		Data    struct {
			Invitation Invitation `json:"invitation"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse roles response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to invite user")
	}
	return &result.Data.Invitation, nil
}
