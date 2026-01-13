package client

import (
	"context"
	"fmt"
	"net/http"

	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/config"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	pb "github.com/rhaloubi/payment-gateway/payment-api-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServiceClient struct {
	baseURL      string
	httpClient   *http.Client
	grpcConn     *grpc.ClientConn
	grpcTimeout  time.Duration
	apiKeyClient pb.APIKeyServiceClient
}

func NewAuthServiceClient() *AuthServiceClient {
	baseURL := config.GetEnv("AUTH_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8001"
	}

	grpcAddress := config.GetEnv("AUTH_SERVICE_GRPC_URL") // From your response
	if grpcAddress == "" {
		grpcAddress = "localhost:50051"
	}

	// Dial gRPC connection (insecure for dev)
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal("failed to dial gRPC", zap.Error(err))
	}

	return &AuthServiceClient{
		baseURL:      baseURL,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		grpcConn:     conn,
		apiKeyClient: pb.NewAPIKeyServiceClient(conn),
		grpcTimeout:  400 * time.Millisecond,
	}
}

// =========================================================================
// API Key Validation
// =========================================================================

type ValidateAPIKeyResponse struct {
	Valid       bool      `json:"valid"`
	MerchantID  uuid.UUID `json:"merchant_id"`
	KeyID       uuid.UUID `json:"key_id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
}

func (c *AuthServiceClient) ValidateAPIKey(apiKey string) (*ValidateAPIKeyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	resp, err := c.apiKeyClient.GetInfoByAPIKey(ctx, &pb.GetInfoByAPIKeyRequest{
		ApiKey: apiKey,
	})
	if err != nil {
		logger.Log.Error("Auth service gRPC request failed", zap.Error(err))
		return nil, fmt.Errorf("auth service unavailable or invalid key: %w", err)
	}

	merchantID, err := uuid.Parse(resp.MerchantId)
	if err != nil {
		return nil, fmt.Errorf("invalid merchant ID from auth service: %w", err)
	}

	keyID, err := uuid.Parse(resp.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid key ID from auth service: %w", err)
	}

	return &ValidateAPIKeyResponse{
		Valid:       true,
		MerchantID:  merchantID,
		KeyID:       keyID,
		Name:        resp.Name,
		Permissions: []string{}, // Permissions are not returned by GetInfoByAPIKey
	}, nil
}
