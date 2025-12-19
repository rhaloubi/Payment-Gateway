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
	restClient *RESTClient
}

func NewMerchantClient() *MerchantClient {
	return &MerchantClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
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

type APIKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	KeyPrefix string    `json:"key_prefix"`
	CreatedAt time.Time `json:"created_at"`
}
type Data struct {
	APIKey   APIKey `json:"api_key"`
	PlainKey string `json:"plain_key"`
}

type TeamMember struct {
	ID        string `json:"ID"`
	UserID    string `json:"UserID"`
	RoleName  string `json:"RoleName"`
	Status    string `json:"Status"`
	InvitedAt string `json:"InvitedAt"`
	JoinedAt  struct {
		Time  string `json:"Time"`
		Valid bool   `json:"Valid"`
	} `json:"JoinedAt"`
}

type Settings struct {
	ID                  string `json:"ID"`
	DefaultCurrency     string `json:"DefaultCurrency"`
	StatementDescriptor struct {
		String string `json:"String"`
		Valid  bool   `json:"Valid"`
	} `json:"StatementDescriptor"`
	WebhookURL struct {
		String string `json:"String"`
		Valid  bool   `json:"Valid"`
	} `json:"WebhookURL"`
	WebhookSecret struct {
		String string `json:"String"`
		Valid  bool   `json:"Valid"`
	} `json:"WebhookSecret"`
	NotificationEmail struct {
		String string `json:"String"`
		Valid  bool   `json:"Valid"`
	} `json:"NotificationEmail"`
	SendEmailReceipts bool   `json:"SendEmailReceipts"`
	AutoSettle        bool   `json:"AutoSettle"`
	SettleSchedule    string `json:"SettleSchedule"`
}

type InvitationResponse struct {
	ID              string `json:"ID"`
	Email           string `json:"Email"`
	RoleName        string `json:"RoleName"`
	InvitationToken string `json:"InvitationToken"`
	Status          string `json:"Status"`
	ExpiresAt       string `json:"ExpiresAt"`
	AcceptedAt      struct {
		Time  string `json:"Time"`
		Valid bool   `json:"Valid"`
	} `json:"AcceptedAt"`
}

func (c *MerchantClient) Create(BusinessName, LegalName, email, BusinessType string) (*Merchant, error) {
	payload := map[string]string{
		"business_name": BusinessName,
		"legal_name":    LegalName,
		"email":         email,
		"business_type": BusinessType,
	}

	// TODO: Implement HTTP POST to merchant service
	resp, err := c.restClient.Post("/api/v1/merchants", payload, config.GetAccessToken())
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

func (c *MerchantClient) GetMerchant(id string) (*Merchant, error) {
	accessToken := config.GetAccessToken()

	resp, err := c.restClient.Get("/api/v1/merchants/"+id, accessToken)
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

func (c *MerchantClient) List() ([]Merchant, error) {
	if config.GetAccessToken() == "" {
		return nil, fmt.Errorf("access token not set")
	}

	resp, err := c.restClient.Get("/api/v1/merchants", config.GetAccessToken())
	if err != nil {
		return nil, err
	}

	result := struct {
		Success bool `json:"success"`
		Data    struct {
			Merchants []Merchant `json:"merchants"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse roles response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to list merchants")
	}
	return result.Data.Merchants, nil
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

	resp, err := c.restClient.Post("/api/v1/merchants/"+merchantID+"/team/invite", payload, config.GetAccessToken())
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

func (c *MerchantClient) CreateAPIKey(merchantID, name string) (*Data, error) {
	if config.GetAccessToken() == "" {
		return nil, fmt.Errorf("access token not set")
	}
	payload := map[string]string{
		"merchant_id": merchantID,
		"name":        name,
	}

	resp, err := c.restClient.Post("/api/v1/merchants/api-keys", payload, config.GetAccessToken())
	if err != nil {
		return nil, err
	}

	result := struct {
		Success bool `json:"success"`
		Data    Data `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse roles response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to create api key")
	}
	return &result.Data, nil
}

func (c *MerchantClient) ListTeamMembers(merchantID string) ([]TeamMember, error) {
	if config.GetAccessToken() == "" {
		return nil, fmt.Errorf("access token not set")
	}

	resp, err := c.restClient.Get("/api/v1/merchants/"+merchantID+"/team", config.GetAccessToken())
	if err != nil {
		return nil, err
	}

	result := struct {
		Success bool `json:"success"`
		Data    struct {
			TeamMembers []TeamMember `json:"team_members"`
			Count       int          `json:"count"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse team response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to list team members")
	}
	return result.Data.TeamMembers, nil
}

func (c *MerchantClient) ListInvitations(merchantID string) ([]InvitationResponse, error) {
	if config.GetAccessToken() == "" {
		return nil, fmt.Errorf("access token not set")
	}

	resp, err := c.restClient.Get("/api/v1/merchants/"+merchantID+"/invitations", config.GetAccessToken())
	if err != nil {
		return nil, err
	}

	result := struct {
		Success bool `json:"success"`
		Data    struct {
			Invitations []InvitationResponse `json:"invitations"`
			Count       int                  `json:"count"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse invitations response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to list invitations")
	}
	return result.Data.Invitations, nil
}

func (c *MerchantClient) GetSettings(merchantID string) (*Settings, error) {
	if config.GetAccessToken() == "" {
		return nil, fmt.Errorf("access token not set")
	}

	resp, err := c.restClient.Get("/api/v1/merchants/"+merchantID+"/settings", config.GetAccessToken())
	if err != nil {
		return nil, err
	}

	result := struct {
		Success bool `json:"success"`
		Data    struct {
			Settings Settings `json:"settings"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse settings response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to get settings")
	}
	return &result.Data.Settings, nil
}
