package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PaymentIntentClient struct {
	httpClient *http.Client
	restClient *RESTClient
}

func NewPaymentIntentClient() *PaymentIntentClient {
	return &PaymentIntentClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		restClient: NewRESTClient(),
	}
}

type PaymentIntentRequest struct {
	Amount        int64             `json:"amount"`
	Currency      string            `json:"currency"`
	SuccessURL    string            `json:"success_url"`
	CancelURL     string            `json:"cancel_url,omitempty"`
	OrderID       string            `json:"order_id,omitempty"`
	Description   string            `json:"description,omitempty"`
	CaptureMethod string            `json:"capture_method,omitempty"`
	CustomerEmail string            `json:"customer_email,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type PaymentIntent struct {
	ID           string            `json:"id"`
	ClientSecret string            `json:"client_secret"`
	CheckoutURL  string            `json:"checkout_url"`
	Amount       int64             `json:"amount"`
	Currency     string            `json:"currency"`
	Status       string            `json:"status"`
	SuccessURL   string            `json:"success_url"`
	CancelURL    string            `json:"cancel_url"`
	Description  string            `json:"description"`
	ExpiresAt    time.Time         `json:"expires_at"`
	CreatedAt    time.Time         `json:"created_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type PaymentIntentStatus struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	PaymentID string `json:"payment_id,omitempty"`
}

// CreatePaymentIntent creates a new payment intent
func (c *PaymentIntentClient) CreatePaymentIntent(req *PaymentIntentRequest, apiKey string) (*PaymentIntent, error) {
	resp, err := c.restClient.Post("/api/v1/payment-intents", req, &AuthOptions{APIKey: apiKey})
	if err != nil {
		return nil, err
	}

	var result struct {
		Success bool          `json:"success"`
		Data    PaymentIntent `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse payment intent response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to create payment intent")
	}

	return &result.Data, nil
}

// GetPaymentIntent retrieves a payment intent status (public endpoint)
func (c *PaymentIntentClient) GetPaymentIntent(intentID string) (*PaymentIntentStatus, error) {
	resp, err := c.restClient.Get("/api/public/payment-intents/"+intentID, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			Amount    int64  `json:"amount"`
			Currency  string `json:"currency"`
			PaymentID string `json:"payment_id,omitempty"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse payment intent status: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to get payment intent")
	}

	return &PaymentIntentStatus{
		ID:        result.Data.ID,
		Status:    result.Data.Status,
		PaymentID: result.Data.PaymentID,
	}, nil
}
