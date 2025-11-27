package client

import (
	"context"
	"os"
	"time"

	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TokenizationClient communicates with Tokenization Service via gRPC
type TokenizationClient struct {
	conn    *grpc.ClientConn
	address string
}

func NewTokenizationClient() (*TokenizationClient, error) {
	address := os.Getenv("TOKENIZATION_SERVICE_GRPC")
	if address == "" {
		address = "localhost:50051" // Default
	}

	// Connect to tokenization service
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		logger.Log.Error("Failed to connect to tokenization service",
			zap.Error(err),
			zap.String("address", address),
		)
		return nil, err
	}

	logger.Log.Info("Connected to tokenization service", zap.String("address", address))

	return &TokenizationClient{
		conn:    conn,
		address: address,
	}, nil
}

// Close closes the gRPC connection
func (c *TokenizationClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// TokenizeCardRequest represents tokenization request
type TokenizeCardRequest struct {
	MerchantID     string
	CardNumber     string
	CardholderName string
	ExpMonth       int
	ExpYear        int
	CVV            string
	IsSingleUse    bool
	RequestID      string
	IPAddress      string
	UserAgent      string
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
func (c *TokenizationClient) TokenizeCard(ctx context.Context, req *TokenizeCardRequest) (*TokenizeCardResponse, error) {
	// TODO: Implement actual gRPC call when proto is generated
	// For now, simulate tokenization

	logger.Log.Info("Tokenizing card (simulated)",
		zap.String("merchant_id", req.MerchantID),
		zap.String("last4", req.CardNumber[len(req.CardNumber)-4:]),
	)

	// Simulate tokenization response
	response := &TokenizeCardResponse{
		Token:       "tok_live_" + generateRandomString(40),
		CardBrand:   detectCardBrand(req.CardNumber),
		CardType:    "credit",
		Last4:       req.CardNumber[len(req.CardNumber)-4:],
		ExpMonth:    req.ExpMonth,
		ExpYear:     req.ExpYear,
		Fingerprint: generateRandomString(32),
		IsNewToken:  true,
	}

	return response, nil
}

// ValidateToken validates a token
func (c *TokenizationClient) ValidateToken(ctx context.Context, token string, merchantID string) (bool, error) {
	// TODO: Implement actual gRPC call
	logger.Log.Debug("Validating token (simulated)", zap.String("token", token))
	return true, nil
}

// Helper functions
func detectCardBrand(cardNumber string) string {
	if len(cardNumber) < 2 {
		return "unknown"
	}

	switch cardNumber[0:1] {
	case "4":
		return "visa"
	case "5":
		return "mastercard"
	default:
		return "unknown"
	}
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
