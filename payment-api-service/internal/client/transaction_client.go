package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"go.uber.org/zap"
)

// TransactionClient communicates with Transaction Service
// TODO: Replace with actual gRPC client when transaction service is built
type TransactionClient struct {
	enabled bool
}

func NewTransactionClient() *TransactionClient {
	return &TransactionClient{
		enabled: true,
	}
}

// AuthorizeRequest represents authorization request
type AuthorizeRequest struct {
	MerchantID    string
	Amount        int64
	Currency      string
	CardToken     string
	CardBrand     string
	CardLast4     string
	FraudScore    int
	CustomerEmail string
	Description   string
}

// AuthorizeResponse represents authorization result
type AuthorizeResponse struct {
	TransactionID uuid.UUID
	Approved      bool
	AuthCode      string
	ResponseCode  string
	ResponseMsg   string
	DeclineReason string
}

// CaptureRequest represents capture request
type CaptureRequest struct {
	TransactionID uuid.UUID
	Amount        int64 // Optional: partial capture
}

// CaptureResponse represents capture result
type CaptureResponse struct {
	Success        bool
	CapturedAmount int64
	ResponseMsg    string
}

// VoidRequest represents void request
type VoidRequest struct {
	TransactionID uuid.UUID
	Reason        string
}

// VoidResponse represents void result
type VoidResponse struct {
	Success     bool
	ResponseMsg string
}

// RefundRequest represents refund request
type RefundRequest struct {
	TransactionID uuid.UUID
	Amount        int64 // Optional: partial refund
	Reason        string
}

// RefundResponse represents refund result
type RefundResponse struct {
	RefundID       uuid.UUID
	Success        bool
	RefundedAmount int64
	ResponseMsg    string
}

// =========================================================================
// Authorization
// =========================================================================

// Authorize processes an authorization request
func (c *TransactionClient) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	logger.Log.Info("Processing authorization (mock)",
		zap.String("merchant_id", req.MerchantID),
		zap.Int64("amount", req.Amount),
		zap.String("card_last4", req.CardLast4),
	)

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Use card simulator logic
	approved := simulateCardResponse(req.CardLast4, req.Amount)

	response := &AuthorizeResponse{
		TransactionID: uuid.New(),
		Approved:      approved,
	}

	if approved {
		response.AuthCode = generateAuthCode()
		response.ResponseCode = "00"
		response.ResponseMsg = "Approved"
	} else {
		response.ResponseCode = getDeclineCode(req.CardLast4)
		response.ResponseMsg = getDeclineMessage(response.ResponseCode)
		response.DeclineReason = response.ResponseMsg
	}

	logger.Log.Info("Authorization completed",
		zap.Bool("approved", approved),
		zap.String("response_code", response.ResponseCode),
	)

	return response, nil
}

// =========================================================================
// Capture
// =========================================================================

// Capture captures a previously authorized transaction
func (c *TransactionClient) Capture(ctx context.Context, req *CaptureRequest) (*CaptureResponse, error) {
	logger.Log.Info("Processing capture (mock)",
		zap.String("transaction_id", req.TransactionID.String()),
		zap.Int64("amount", req.Amount),
	)

	time.Sleep(50 * time.Millisecond)

	// Mock: Always succeed
	return &CaptureResponse{
		Success:        true,
		CapturedAmount: req.Amount,
		ResponseMsg:    "Capture successful",
	}, nil
}

// =========================================================================
// Void
// =========================================================================

// Void cancels an authorized transaction
func (c *TransactionClient) Void(ctx context.Context, req *VoidRequest) (*VoidResponse, error) {
	logger.Log.Info("Processing void (mock)",
		zap.String("transaction_id", req.TransactionID.String()),
		zap.String("reason", req.Reason),
	)

	time.Sleep(50 * time.Millisecond)

	// Mock: Always succeed
	return &VoidResponse{
		Success:     true,
		ResponseMsg: "Transaction voided successfully",
	}, nil
}

// =========================================================================
// Refund
// =========================================================================

// Refund processes a refund
func (c *TransactionClient) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	logger.Log.Info("Processing refund (mock)",
		zap.String("transaction_id", req.TransactionID.String()),
		zap.Int64("amount", req.Amount),
	)

	time.Sleep(50 * time.Millisecond)

	// Mock: Always succeed
	return &RefundResponse{
		RefundID:       uuid.New(),
		Success:        true,
		RefundedAmount: req.Amount,
		ResponseMsg:    "Refund processed successfully",
	}, nil
}

// =========================================================================
// Card Simulator Logic
// =========================================================================

// simulateCardResponse simulates issuer bank response based on card number
func simulateCardResponse(last4 string, amount int64) bool {
	// Test cards (based on last 4 digits)
	switch last4 {
	case "4242": // Success
		return true
	case "0002": // Generic decline
		return false
	case "9995": // Insufficient funds
		return false
	case "0069": // Expired card
		return false
	case "0127": // Incorrect CVV
		return false
	case "0119": // Processing error
		return false
	default:
		// Real card - approve
		return true
	}
}

// getDeclineCode returns appropriate decline code
func getDeclineCode(last4 string) string {
	switch last4 {
	case "0002":
		return "05" // Do not honor
	case "9995":
		return "51" // Insufficient funds
	case "0069":
		return "54" // Expired card
	case "0127":
		return "N7" // CVV mismatch
	case "0119":
		return "96" // System error
	default:
		return "05"
	}
}

// getDeclineMessage returns human-readable decline message
func getDeclineMessage(code string) string {
	messages := map[string]string{
		"05": "Transaction declined",
		"51": "Insufficient funds",
		"54": "Expired card",
		"N7": "CVV verification failed",
		"96": "System error - please try again",
	}

	if msg, ok := messages[code]; ok {
		return msg
	}
	return "Transaction declined"
}

// generateAuthCode generates a random authorization code
func generateAuthCode() string {
	return fmt.Sprintf("%06d", time.Now().Unix()%1000000)
}

// Close closes the client connection (no-op for mock)
func (c *TransactionClient) Close() error {
	return nil
}
