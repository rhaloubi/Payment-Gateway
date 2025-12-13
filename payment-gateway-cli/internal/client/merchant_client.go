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
		Success  bool     `json:"success"`
		Merchant Merchant `json:"merchant"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse roles response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to create merchant")
	}
	return &Merchant{
		ID:           result.Merchant.ID,
		BusinessName: result.Merchant.BusinessName,
		LegalName:    result.Merchant.LegalName,
		Email:        result.Merchant.Email,
		BusinessType: result.Merchant.BusinessType,
		Status:       result.Merchant.Status,
	}, nil
}

/*func (c *MerchantClient) List() ([]Merchant, error) {
	// TODO: Implement HTTP GET to merchant service
	// For now, return mock data
	return []Merchant{
		{ID: "mer_1", Name: "Test Merchant 1", Email: "test1@example.com", Status: "active"},
		{ID: "mer_2", Name: "Test Merchant 2", Email: "test2@example.com", Status: "active"},
	}, nil
}

func (c *MerchantClient) Get(id string) (*Merchant, error) {
	// TODO: Implement HTTP GET to merchant service
	// For now, return mock data
	return &Merchant{
		ID:     id,
		Name:   "Test Merchant",
		Email:  "test@example.com",
		Status: "active",
	}, nil
}
*/
