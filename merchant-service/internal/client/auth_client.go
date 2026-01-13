package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/config"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits/logger"
	pb "github.com/rhaloubi/payment-gateway/merchant-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServiceClient struct {
	baseURL      string
	httpClient   *http.Client
	grpcConn     *grpc.ClientConn
	grpcClient   pb.RoleServiceClient
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
		grpcClient:   pb.NewRoleServiceClient(conn),
		apiKeyClient: pb.NewAPIKeyServiceClient(conn),
		grpcTimeout:  400 * time.Millisecond, // Adjustable timeout for gRPC calls
	}
}

// AssignMerchantOwnerRole assigns the merchant owner role via gRPC
func (c *AuthServiceClient) AssignMerchantOwnerRole(userID, merchantID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	req := &pb.AssignMerchantOwnerRoleRequest{
		UserId:     userID.String(),
		MerchantId: merchantID.String(),
	}

	resp, err := c.grpcClient.AssignMerchantOwnerRole(ctx, req)
	if err != nil {
		logger.Log.Error("gRPC AssignMerchantOwnerRole failed", zap.Error(err))
		return fmt.Errorf("gRPC call failed: %w", err)
	}

	// Log success (optional)
	logger.Log.Info("gRPC role assigned",
		zap.String("user_id", resp.UserId),
		zap.String("role_id", resp.RoleId),
		zap.String("merchant_id", resp.MerchantId))

	return nil
}

// ... (Keep your existing GetUserRoles or other HTTP methods here unchanged)

func (c *AuthServiceClient) AssignRoleToUser(userID, merchantID, roleId, assignedBy uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	req := &pb.AssignRoleToUserRequest{
		UserId:     userID.String(),
		MerchantId: merchantID.String(),
		RoleId:     roleId.String(),
		AssignedBy: assignedBy.String(),
	}

	resp, err := c.grpcClient.AssignRoleToUser(ctx, req)
	if err != nil {
		logger.Log.Error("gRPC AssignMerchantOwnerRole failed", zap.Error(err))
		return fmt.Errorf("gRPC call failed: %w", err)
	}
	logger.Log.Info("gRPC role assigned",
		zap.String("user_id", resp.UserId),
		zap.String("role_id", resp.RoleId),
		zap.String("merchant_id", resp.MerchantId))

	return nil
}

// CreateAPIKey calls gRPC to create an API key
func (c *AuthServiceClient) CreateAPIKey(merchantID, createdBy uuid.UUID, name string) (*pb.CreateAPIKeyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	req := &pb.CreateAPIKeyRequest{
		MerchantId: merchantID.String(),
		Name:       name,
		CreatedBy:  createdBy.String(),
	}

	resp, err := c.apiKeyClient.CreateAPIKey(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC CreateAPIKey failed: %w", err)
	}
	return resp, nil
}

// GetMerchantAPIKeys calls gRPC to get API keys for a merchant
func (c *AuthServiceClient) GetMerchantAPIKeys(merchantID uuid.UUID) (*pb.GetMerchantAPIKeysResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	req := &pb.GetMerchantAPIKeysRequest{
		MerchantId: merchantID.String(),
	}

	resp, err := c.apiKeyClient.GetMerchantAPIKeys(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC GetMerchantAPIKeys failed: %w", err)
	}
	return resp, nil
}

// DeactivateAPIKey calls gRPC to deactivate an API key
func (c *AuthServiceClient) DeactivateAPIKey(keyID, merchantID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	req := &pb.DeactivateAPIKeyRequest{
		Id:         keyID.String(),
		MerchantId: merchantID.String(),
	}

	_, err := c.apiKeyClient.DeactivateAPIKey(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC DeactivateAPIKey failed: %w", err)
	}
	return nil
}

// DeleteAPIKey calls gRPC to delete an API key
func (c *AuthServiceClient) DeleteAPIKey(keyID, merchantID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	req := &pb.DeleteAPIKeyRequest{
		Id:         keyID.String(),
		MerchantId: merchantID.String(),
	}

	_, err := c.apiKeyClient.DeleteAPIKey(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC DeleteAPIKey failed: %w", err)
	}
	return nil
}

// Close closes the gRPC connection
func (c *AuthServiceClient) Close() error {
	return c.grpcConn.Close()
}
