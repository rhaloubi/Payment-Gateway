package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rhaloubi/payment-gateway/payment-api-service/config"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	pb "github.com/rhaloubi/payment-gateway/payment-api-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TokenizationClient communicates with Tokenization Service via gRPC
type TokenizationClient struct {
	httpClient         *http.Client
	grpcConn           *grpc.ClientConn
	grpcTimeout        time.Duration
	tokenizationClient pb.TokenizationServiceClient
}

func NewTokenizationClient() (*TokenizationClient, error) {

	grpcAddress := config.GetEnv("TOKENIZATION_SERVICE_GRPC_URL") // From your response
	if grpcAddress == "" {
		grpcAddress = "localhost:50052"
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

// TokenizeCardResponse represents tokenization response
type TokenizeCardResponse struct {
	Token       string
	CardBrand   string
	CardType    string
	Last4       string
	ExpMonth    int
	ExpYear     int
	Fingerprint string
	IsNewToken  bool
	Error       string
}

// TokenizeCard tokenizes card data
func (c *TokenizationClient) TokenizeCard(ctx context.Context, req *pb.TokenizeCardRequest) (*TokenizeCardResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	logger.Log.Info("Tokenizing card (simulated)",
		zap.String("merchant_id", req.MerchantId),
		zap.String("last4", req.CardNumber[len(req.CardNumber)-4:]),
	)

	resp, err := c.tokenizationClient.TokenizeCard(ctx, req)
	if err != nil {
		logger.Log.Error("Tokenization service gRPC request failed", zap.Error(err))
		return nil, fmt.Errorf("tokenization service unavailable or invalid key: %w", err)
	}

	if resp.Card == nil {
		if resp.Error != "" {
			return nil, fmt.Errorf("tokenization failed: %s", resp.Error)
		}
		return nil, fmt.Errorf("tokenization failed: unknown error")
	}

	// Simulate tokenization response
	response := &TokenizeCardResponse{
		Token:       resp.Token,
		CardBrand:   resp.Card.Brand,
		CardType:    resp.Card.Type,
		Last4:       resp.Card.Last4,
		ExpMonth:    int(resp.Card.ExpMonth),
		ExpYear:     int(resp.Card.ExpYear),
		Fingerprint: resp.Card.Fingerprint,
		IsNewToken:  resp.IsNewToken,
	}

	return response, nil
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
