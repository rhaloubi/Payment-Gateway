package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rhaloubi/payment-gateway-cli/internal/config"
)

type MerchantClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewMerchantClient() *MerchantClient {
	return &MerchantClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    config.GetAPIURL(),
	}
}

type Merchant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func (c *MerchantClient) Create(name, email string) (*Merchant, error) {
	/* payload := map[string]string{
		"name":  name,
		"email": email,
	} */

	// TODO: Implement HTTP POST to merchant service
	// For now, return mock data
	return &Merchant{
		ID:     "mer_" + fmt.Sprintf("%d", time.Now().Unix()),
		Name:   name,
		Email:  email,
		Status: "active",
	}, nil
}

func (c *MerchantClient) List() ([]Merchant, error) {
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
