package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rhaloubi/payment-gateway-cli/validation"
)

type PaymentClient struct {
	httpClient *http.Client
	restClient *RESTClient
}

func NewPaymentClient() *PaymentClient {
	return &PaymentClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		restClient: NewRESTClient(),
	}
}

// authorization func
type AuthorizeResponse struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Amount          int    `json:"amount"`
	Currency        string `json:"currency"`
	Token           string `json:"token"`
	CardBrand       string `json:"card_brand"`
	CardLast4       string `json:"card_last4"`
	AuthCode        string `json:"auth_code"`
	FraudScore      int    `json:"fraud_score"`
	FraudDecision   string `json:"fraud_decision"`
	ResponseCode    string `json:"response_code"`
	ResponseMessage string `json:"response_message"`
	TransactionID   string `json:"transaction_id"`
}

func (c *PaymentClient) AuthorizePayment(req *validation.AuthorizeRequest, ApiKey string) (*AuthorizeResponse, error) {

	// TODO: Implement HTTP POST to payment service
	resp, err := c.restClient.Post("/api/v1/payments/authorize", req, &AuthOptions{APIKey: ApiKey})
	if err != nil {
		return nil, err
	}
	result := struct {
		Success bool              `json:"success"`
		Data    AuthorizeResponse `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse authorization response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to authorize payment")
	}
	return &result.Data, nil
}
