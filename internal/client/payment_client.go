package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rhaloubi/payment-gateway-cli/internal/validation"
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
type Transaction struct {
	ID         string `json:"id"`
	MerchantID string `json:"merchant_id"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Amount     int    `json:"amount"`
	Currency   string `json:"currency"`
	CardBrand  string `json:"card_brand"`
	CardLast4  string `json:"card_last4"`
	FraudScore int    `json:"fraud_score"`
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

func (c *PaymentClient) ListTransactions(
	apiKey string,
	filters *validation.TransactionFilters,
) ([]Transaction, error) {

	endpoint := "/api/v1/transactions"

	query := []string{}

	if filters != nil {
		if filters.Limit != nil {
			query = append(query, fmt.Sprintf("limit=%d", *filters.Limit))
		}
		if filters.Offset != nil {
			query = append(query, fmt.Sprintf("offset=%d", *filters.Offset))
		}
		if filters.Status != nil {
			query = append(query, fmt.Sprintf("status=%s", *filters.Status))
		}
	}

	if len(query) > 0 {
		endpoint += "?" + strings.Join(query, "&")
	}

	resp, err := c.restClient.Get(
		endpoint,
		&AuthOptions{APIKey: apiKey},
	)
	if err != nil {
		return nil, err
	}

	result := struct {
		Success bool `json:"success"`
		Data    struct {
			Transactions []Transaction `json:"transactions"`
			Total        int           `json:"total"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf(
			"failed to parse transaction list response: %w",
			err,
		)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to list transactions")
	}

	return result.Data.Transactions, nil
}
