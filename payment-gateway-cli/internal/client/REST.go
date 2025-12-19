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

type AuthOptions struct {
	BearerToken string
	APIKey      string
}
type RESTClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewRESTClient() *RESTClient {
	return &RESTClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: config.GetAPIURL(),
	}
}

func applyAuthHeaders(req *http.Request, auth *AuthOptions) {
	if auth == nil {
		return
	}

	if auth.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+auth.BearerToken)
	}

	if auth.APIKey != "" {
		req.Header.Set("X-API-Key", auth.APIKey)
	}
}

/*
doRequest is the internal request handler used by all HTTP methods.
*/
func (c *RESTClient) doRequest(
	method string,
	endpoint string,
	payload interface{},
	auth *AuthOptions,
) ([]byte, error) {

	var body io.Reader

	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(
		method,
		c.baseURL+endpoint,
		body,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	applyAuthHeaders(req, auth)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf(
			"HTTP %d: %s",
			resp.StatusCode,
			string(respBody),
		)
	}

	return respBody, nil
}

/*
Post sends a POST request with an optional payload and authentication.
*/
func (c *RESTClient) Post(
	endpoint string,
	payload interface{},
	auth *AuthOptions,
) ([]byte, error) {
	return c.doRequest(http.MethodPost, endpoint, payload, auth)
}

/*
Get sends a GET request with optional authentication.
*/
func (c *RESTClient) Get(
	endpoint string,
	auth *AuthOptions,
) ([]byte, error) {
	return c.doRequest(http.MethodGet, endpoint, nil, auth)
}
