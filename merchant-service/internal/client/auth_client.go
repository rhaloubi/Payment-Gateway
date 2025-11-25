package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits/logger"
	pb "github.com/rhaloubi/payment-gateway/merchant-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthServiceClient struct {
	baseURL     string
	httpClient  *http.Client
	grpcConn    *grpc.ClientConn
	grpcClient  pb.RoleServiceClient
	grpcTimeout time.Duration
}

func NewAuthServiceClient() *AuthServiceClient {
	baseURL := os.Getenv("AUTH_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8001"
	}

	grpcAddress := os.Getenv("AUTH_SERVICE_GRPC_URL") // From your response
	if grpcAddress == "" {
		grpcAddress = "localhost:50051"
	}

	// Dial gRPC connection (insecure for dev)
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal("failed to dial gRPC", zap.Error(err))
	}

	return &AuthServiceClient{
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		grpcConn:    conn,
		grpcClient:  pb.NewRoleServiceClient(conn),
		grpcTimeout: 400 * time.Millisecond, // Adjustable timeout for gRPC calls
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

// Close closes the gRPC connection (call this on shutdown if needed)
func (c *AuthServiceClient) Close() error {
	return c.grpcConn.Close()
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
