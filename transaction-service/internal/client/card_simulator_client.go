package client

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	"go.uber.org/zap"
)

// CardSimulatorClient simulates issuer bank responses
type CardSimulatorClient struct {
	enabled bool
}

func NewCardSimulatorClient() *CardSimulatorClient {
	return &CardSimulatorClient{
		enabled: true,
	}
}

// =========================================================================
// Request/Response Types
// =========================================================================

type AuthorizeCardRequest struct {
	CardNumber string
	ExpMonth   int32
	ExpYear    int32
	Amount     int64
	Currency   string
	MerchantID string
}

type AuthorizeCardResponse struct {
	Approved        bool
	AuthCode        string
	ResponseCode    string
	ResponseMessage string
	DeclineReason   string
	AVSResult       string // Address Verification System
	CVVResult       string // CVV Check Result
}

type CaptureCardRequest struct {
	TransactionID string
	Amount        int64
}

type CaptureCardResponse struct {
	Success         bool
	ResponseMessage string
}

type VoidCardRequest struct {
	TransactionID string
	Reason        string
}

type VoidCardResponse struct {
	Success         bool
	ResponseMessage string
}

type RefundCardRequest struct {
	TransactionID string
	Amount        int64
	Reason        string
}

type RefundCardResponse struct {
	Success         bool
	RefundID        string
	ResponseMessage string
}

// =========================================================================
// Authorization
// =========================================================================

func (c *CardSimulatorClient) Authorize(ctx context.Context, req *AuthorizeCardRequest) (*AuthorizeCardResponse, error) {
	logger.Log.Info("Simulating card authorization",
		zap.String("card_last4", req.CardNumber[len(req.CardNumber)-4:]),
		zap.Int64("amount", req.Amount),
	)

	/* Simulate processing time (100-500ms)
	processingTime := time.Duration(100+rand.Intn(400)) * time.Millisecond
	time.Sleep(processingTime) */

	// Get last 4 digits for test card detection
	last4 := req.CardNumber[len(req.CardNumber)-4:]

	// Simulate authorization based on test cards
	response := c.simulateAuthorization(last4)

	logger.Log.Info("Authorization simulation complete",
		zap.Bool("approved", response.Approved),
		zap.String("response_code", response.ResponseCode),
		//zap.Duration("processing_time", processingTime),
	)

	return response, nil
}

// simulateAuthorization simulates issuer response based on card number
func (c *CardSimulatorClient) simulateAuthorization(last4 string) *AuthorizeCardResponse {
	// Test cards (based on last 4 digits)
	switch last4 {
	case "4242": // Success - Visa
		return &AuthorizeCardResponse{
			Approved:        true,
			AuthCode:        c.generateAuthCode(),
			ResponseCode:    "00",
			ResponseMessage: "Approved",
			AVSResult:       "Y", // Address match
			CVVResult:       "M", // CVV match
		}

	case "4444": // Success - Mastercard
		return &AuthorizeCardResponse{
			Approved:        true,
			AuthCode:        c.generateAuthCode(),
			ResponseCode:    "00",
			ResponseMessage: "Approved",
			AVSResult:       "Y",
			CVVResult:       "M",
		}

	case "0002": // Generic decline
		return &AuthorizeCardResponse{
			Approved:      false,
			ResponseCode:  "05",
			DeclineReason: "Do not honor",
		}

	case "9995": // Insufficient funds
		return &AuthorizeCardResponse{
			Approved:      false,
			ResponseCode:  "51",
			DeclineReason: "Insufficient funds",
		}

	case "0069": // Expired card
		return &AuthorizeCardResponse{
			Approved:      false,
			ResponseCode:  "54",
			DeclineReason: "Expired card",
		}

	case "0127": // CVV mismatch
		return &AuthorizeCardResponse{
			Approved:      false,
			ResponseCode:  "N7",
			DeclineReason: "CVV verification failed",
			CVVResult:     "N", // No match
		}

	case "0119": // Processing error
		return &AuthorizeCardResponse{
			Approved:      false,
			ResponseCode:  "96",
			DeclineReason: "System error - please retry",
		}

	default:
		// Real card simulation - approve
		return &AuthorizeCardResponse{
			Approved:        true,
			AuthCode:        c.generateAuthCode(),
			ResponseCode:    "00",
			ResponseMessage: "Approved",
			AVSResult:       "Y",
			CVVResult:       "M",
		}
	}
}

// =========================================================================
// Capture
// =========================================================================

func (c *CardSimulatorClient) Capture(ctx context.Context, req *CaptureCardRequest) (*CaptureCardResponse, error) {
	logger.Log.Info("Simulating card capture",
		zap.String("transaction_id", req.TransactionID),
		zap.Int64("amount", req.Amount),
	)

	// Simulate processing
	time.Sleep(30 * time.Millisecond)

	// Mock: Always succeed
	return &CaptureCardResponse{
		Success:         true,
		ResponseMessage: "Capture successful",
	}, nil
}

// =========================================================================
// Void
// =========================================================================

func (c *CardSimulatorClient) Void(ctx context.Context, req *VoidCardRequest) (*VoidCardResponse, error) {
	logger.Log.Info("Simulating card void",
		zap.String("transaction_id", req.TransactionID),
	)

	// Simulate processing
	time.Sleep(30 * time.Millisecond)

	// Mock: Always succeed
	return &VoidCardResponse{
		Success:         true,
		ResponseMessage: "Authorization voided successfully",
	}, nil
}

// =========================================================================
// Refund
// =========================================================================

func (c *CardSimulatorClient) Refund(ctx context.Context, req *RefundCardRequest) (*RefundCardResponse, error) {
	logger.Log.Info("Simulating card refund",
		zap.String("transaction_id", req.TransactionID),
		zap.Int64("amount", req.Amount),
	)

	// Simulate processing
	time.Sleep(50 * time.Millisecond)

	// Mock: Always succeed
	return &RefundCardResponse{
		Success:         true,
		RefundID:        c.generateRefundID(),
		ResponseMessage: "Refund processed successfully",
	}, nil
}

// =========================================================================
// Helper Methods
// =========================================================================

func (c *CardSimulatorClient) generateAuthCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (c *CardSimulatorClient) generateRefundID() string {
	return fmt.Sprintf("REF%d", time.Now().UnixNano())
}

func (c *CardSimulatorClient) Close() error {
	return nil
}
