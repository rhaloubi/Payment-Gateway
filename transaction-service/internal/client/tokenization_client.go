package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	pb "github.com/rhaloubi/payment-gateway/transaction-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TokenizationClient struct {
	httpClient         *http.Client
	grpcConn           *grpc.ClientConn
	grpcTimeout        time.Duration
	tokenizationClient pb.TokenizationServiceClient
}

func NewTokenizationClient() (*TokenizationClient, error) {

	grpcAddress := os.Getenv("TOKENIZATION_SERVICE_GRPC_URL") // From your response
	if grpcAddress == "" {
		grpcAddress = "localhost:50053"
	}

	// Dial gRPC connection (insecure for dev)
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal("failed to dial gRPC", zap.Error(err))
	}

	return &TokenizationClient{
		httpClient:         &http.Client{Timeout: 10 * time.Second},
		grpcConn:           conn,
		grpcTimeout:        400 * time.Millisecond,
		tokenizationClient: pb.NewTokenizationServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection
func (c *TokenizationClient) Close() error {
	if c.grpcConn != nil {
		return c.grpcConn.Close()
	}
	return nil
}

// ValidateToken validates a token
func (c *TokenizationClient) ValidateToken(ctx context.Context, token string, merchantID string) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()
	resp, err := c.tokenizationClient.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Token:      token,
		MerchantId: merchantID,
	})
	if err != nil {
		logger.Log.Error("Tokenization service gRPC request failed", zap.Error(err))
		return false, fmt.Errorf("tokenization service unavailable or invalid key: %w", err)
	}
	return resp.Valid, nil
}

func (c *TokenizationClient) Detokenize(ctx context.Context, token string, merchantID string) (*pb.DetokenizeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()
	resp, err := c.tokenizationClient.Detokenize(ctx, &pb.DetokenizeRequest{
		Token:      token,
		MerchantId: merchantID,
	})
	if err != nil {
		logger.Log.Error("Tokenization service gRPC request failed", zap.Error(err))
		return nil, fmt.Errorf("tokenization service unavailable or invalid key: %w", err)
	}
	return resp, nil
}
