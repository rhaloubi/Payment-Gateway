package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/client"
	model "github.com/rhaloubi/payment-gateway/payment-api-service/internal/models"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/repository"
	pb "github.com/rhaloubi/payment-gateway/payment-api-service/proto"
	"go.uber.org/zap"
)

type PaymentService struct {
	paymentRepo        *repository.PaymentRepository
	tokenizationClient *client.TokenizationClient
	fraudClient        *client.FraudClient
	transactionClient  *client.TransactionClient
}

func NewPaymentService() (*PaymentService, error) {
	tokenClient, err := client.NewTokenizationClient()
	if err != nil {
		logger.Log.Warn("Failed to connect to tokenization service, using mock", zap.Error(err))
		// Continue without tokenization client (will use mock)
	}

	return &PaymentService{
		paymentRepo:        repository.NewPaymentRepository(),
		tokenizationClient: tokenClient,
		fraudClient:        client.NewFraudClient(),
		transactionClient:  client.NewTransactionClient(),
	}, nil
}

// Request/Response DTOs
type AuthorizePaymentRequest struct {
	MerchantID     uuid.UUID
	Amount         int64
	Currency       string
	CardNumber     string
	CardholderName string
	ExpMonth       int
	ExpYear        int
	CVV            string
	CustomerEmail  string
	CustomerName   string
	Description    string
	Metadata       map[string]interface{}
	IdempotencyKey string
	IPAddress      string
	UserAgent      string
	CreatedBy      uuid.UUID
}

type PaymentResponse struct {
	ID            uuid.UUID           `json:"id"`
	Status        model.PaymentStatus `json:"status"`
	Amount        int64               `json:"amount"`
	Currency      string              `json:"currency"`
	Token         string              `json:"token,omitempty"`
	CardBrand     string              `json:"card_brand"`
	CardLast4     string              `json:"card_last4"`
	AuthCode      string              `json:"auth_code,omitempty"`
	FraudScore    int                 `json:"fraud_score"`
	FraudDecision string              `json:"fraud_decision"`
	ResponseCode  string              `json:"response_code"`
	ResponseMsg   string              `json:"response_message"`
	TransactionID uuid.UUID           `json:"transaction_id,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
}

func (s *PaymentService) AuthorizePayment(ctx context.Context, req *AuthorizePaymentRequest) (*PaymentResponse, error) {
	startTime := time.Now()
	logger.Log.Info("Processing payment authorization",
		zap.String("merchant_id", req.MerchantID.String()),
		zap.Int64("amount", req.Amount),
		zap.String("currency", req.Currency),
	)

	// Step 1: Check idempotency
	if req.IdempotencyKey != "" {
		existing, err := s.paymentRepo.FindByIdempotencyKey(req.MerchantID, req.IdempotencyKey)
		if err == nil && existing != nil {
			logger.Log.Info("Returning cached payment (idempotency)",
				zap.String("payment_id", existing.ID.String()),
			)
			return s.buildPaymentResponse(existing), nil
		}
	}

	// Step 2: Tokenize card
	tokenResp, err := s.tokenizationClient.TokenizeCard(ctx, &pb.TokenizeCardRequest{
		MerchantId:     req.MerchantID.String(),
		CardNumber:     req.CardNumber,
		CardholderName: req.CardholderName,
		ExpMonth:       int32(req.ExpMonth),
		ExpYear:        int32(req.ExpYear),
		Cvv:            req.CVV,
		IsSingleUse:    false,
		IpAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
	})
	if err != nil {
		logger.Log.Error("Tokenization failed", zap.Error(err))
		return nil, fmt.Errorf("failed to tokenize card: %w", err)
	}

	// Step 3: Fraud check
	fraudResp, err := s.fraudClient.CheckFraud(ctx, &client.FraudCheckRequest{
		MerchantID:    req.MerchantID.String(),
		Amount:        req.Amount,
		Currency:      req.Currency,
		CardToken:     tokenResp.Token,
		CardBrand:     tokenResp.CardBrand,
		CardLast4:     tokenResp.Last4,
		CustomerEmail: req.CustomerEmail,
		CustomerIP:    req.IPAddress,
	})
	if err != nil {
		logger.Log.Error("Fraud check failed", zap.Error(err))
		// Continue without fraud check (default to low risk)
		fraudResp = &client.FraudCheckResponse{
			RiskScore: 10,
			Decision:  "approve",
		}
	}

	// Step 4: Check fraud decision
	if fraudResp.Decision == "decline" {
		logger.Log.Warn("Payment declined by fraud system",
			zap.Int("risk_score", fraudResp.RiskScore),
		)
		return s.createFailedPayment(req, tokenResp, fraudResp, "Declined by fraud detection")
	}

	// Step 5: Authorize transaction
	authResp, err := s.transactionClient.Authorize(ctx, &client.AuthorizeRequest{
		MerchantID:    req.MerchantID.String(),
		Amount:        req.Amount,
		Currency:      req.Currency,
		CardToken:     tokenResp.Token,
		CardBrand:     tokenResp.CardBrand,
		CardLast4:     tokenResp.Last4,
		FraudScore:    fraudResp.RiskScore,
		CustomerEmail: req.CustomerEmail,
		Description:   req.Description,
	})
	if err != nil {
		logger.Log.Error("Transaction authorization failed", zap.Error(err))
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	// Step 6: Create payment record
	payment := &model.Payment{
		MerchantID:    req.MerchantID,
		TransactionID: authResp.TransactionID,
		Type:          model.PaymentTypeAuthorize,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Token:         tokenResp.Token,
		CardBrand:     tokenResp.CardBrand,
		CardLast4:     tokenResp.Last4,
		FraudScore:    fraudResp.RiskScore,
		FraudDecision: fraudResp.Decision,
		IPAddress:     req.IPAddress,
		CreatedBy:     req.CreatedBy,
	}

	// Set customer info
	if req.CustomerEmail != "" {
		payment.CustomerEmail = sql.NullString{String: req.CustomerEmail, Valid: true}
	}
	if req.CustomerName != "" {
		payment.CustomerName = sql.NullString{String: req.CustomerName, Valid: true}
	}
	if req.Description != "" {
		payment.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.UserAgent != "" {
		payment.UserAgent = sql.NullString{String: req.UserAgent, Valid: true}
	}
	if req.IdempotencyKey != "" {
		payment.IdempotencyKey = sql.NullString{String: req.IdempotencyKey, Valid: true}
	}

	if authResp.Approved {
		payment.Status = model.PaymentStatusAuthorized
		payment.AuthCode = sql.NullString{String: authResp.AuthCode, Valid: true}
		payment.ResponseCode = sql.NullString{String: authResp.ResponseCode, Valid: true}
		payment.ResponseMsg = sql.NullString{String: authResp.ResponseMsg, Valid: true}
	} else {
		payment.Status = model.PaymentStatusFailed
		payment.ResponseCode = sql.NullString{String: authResp.ResponseCode, Valid: true}
		payment.ResponseMsg = sql.NullString{String: authResp.DeclineReason, Valid: true}
	}

	// Save payment
	if err := s.paymentRepo.Create(payment); err != nil {
		logger.Log.Error("Failed to save payment", zap.Error(err))
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Log event
	go s.paymentRepo.CreateEvent(&model.PaymentEvent{
		PaymentID: payment.ID,
		EventType: string(payment.Type),
		OldStatus: model.PaymentStatusPending,
		NewStatus: payment.Status,
		Amount:    payment.Amount,
		CreatedBy: req.CreatedBy,
	})

	logger.Log.Info("Payment authorization completed",
		zap.String("payment_id", payment.ID.String()),
		zap.String("status", string(payment.Status)),
		zap.Duration("processing_time", time.Since(startTime)),
	)

	return s.buildPaymentResponse(payment), nil
}

// Sale (Authorize + Capture)
func (s *PaymentService) SalePayment(ctx context.Context, req *AuthorizePaymentRequest) (*PaymentResponse, error) {
	// First authorize
	authResp, err := s.AuthorizePayment(ctx, req)
	if err != nil {
		return nil, err
	}

	// If authorized, immediately capture
	if authResp.Status == model.PaymentStatusAuthorized {
		captureResp, err := s.CapturePayment(ctx, authResp.ID, req.MerchantID, authResp.Amount)
		if err != nil {
			logger.Log.Error("Auto-capture failed", zap.Error(err))
			return authResp, nil
		}
		return captureResp, nil
	}

	return authResp, nil
}

// Capture Payment
func (s *PaymentService) CapturePayment(ctx context.Context, paymentID, merchantID uuid.UUID, amount int64) (*PaymentResponse, error) {
	// Get payment
	payment, err := s.paymentRepo.FindByIDAndMerchant(paymentID, merchantID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// Validate can capture
	if !payment.CanCapture() {
		return nil, errors.New("payment cannot be captured (not in authorized state)")
	}

	// Capture via transaction service
	_, err = s.transactionClient.Capture(ctx, &client.CaptureRequest{
		TransactionID: payment.TransactionID,
		Amount:        amount,
	})
	if err != nil {
		return nil, fmt.Errorf("capture failed: %w", err)
	}

	// Update payment status
	if err := s.paymentRepo.MarkCaptured(paymentID); err != nil {
		return nil, err
	}

	// Log event
	go s.paymentRepo.CreateEvent(&model.PaymentEvent{
		PaymentID: paymentID,
		EventType: "captured",
		OldStatus: model.PaymentStatusAuthorized,
		NewStatus: model.PaymentStatusCaptured,
		Amount:    amount,
	})

	// Refresh payment
	payment, _ = s.paymentRepo.FindByID(paymentID)

	logger.Log.Info("Payment captured",
		zap.String("payment_id", paymentID.String()),
		zap.Int64("amount", amount),
	)

	return s.buildPaymentResponse(payment), nil
}

// Void Payment
func (s *PaymentService) VoidPayment(ctx context.Context, paymentID, merchantID uuid.UUID, reason string) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.FindByIDAndMerchant(paymentID, merchantID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if !payment.CanVoid() {
		return nil, errors.New("payment cannot be voided")
	}

	// Void via transaction service
	_, err = s.transactionClient.Void(ctx, &client.VoidRequest{
		TransactionID: payment.TransactionID,
		Reason:        reason,
	})
	if err != nil {
		return nil, fmt.Errorf("void failed: %w", err)
	}

	// Update status
	if err := s.paymentRepo.MarkVoided(paymentID); err != nil {
		return nil, err
	}

	// Log event
	go s.paymentRepo.CreateEvent(&model.PaymentEvent{
		PaymentID:   paymentID,
		EventType:   "voided",
		OldStatus:   payment.Status,
		NewStatus:   model.PaymentStatusVoided,
		Description: sql.NullString{String: reason, Valid: true},
	})

	payment, _ = s.paymentRepo.FindByID(paymentID)

	logger.Log.Info("Payment voided",
		zap.String("payment_id", paymentID.String()),
		zap.String("reason", reason),
	)

	return s.buildPaymentResponse(payment), nil
}

// Refund Payment
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID, merchantID uuid.UUID, amount int64, reason string) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.FindByIDAndMerchant(paymentID, merchantID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if !payment.CanRefund() {
		return nil, errors.New("payment cannot be refunded (not captured)")
	}

	// Refund via transaction service
	_, err = s.transactionClient.Refund(ctx, &client.RefundRequest{
		TransactionID: payment.TransactionID,
		Amount:        amount,
		Reason:        reason,
	})
	if err != nil {
		return nil, fmt.Errorf("refund failed: %w", err)
	}

	// Update status
	if err := s.paymentRepo.MarkRefunded(paymentID); err != nil {
		return nil, err
	}

	// Log event
	go s.paymentRepo.CreateEvent(&model.PaymentEvent{
		PaymentID:   paymentID,
		EventType:   "refunded",
		OldStatus:   payment.Status,
		NewStatus:   model.PaymentStatusRefunded,
		Amount:      amount,
		Description: sql.NullString{String: reason, Valid: true},
	})

	payment, _ = s.paymentRepo.FindByID(paymentID)

	logger.Log.Info("Payment refunded",
		zap.String("payment_id", paymentID.String()),
		zap.Int64("amount", amount),
	)

	return s.buildPaymentResponse(payment), nil
}

// =========================================================================
// Helper Methods
// =========================================================================

func (s *PaymentService) createFailedPayment(
	req *AuthorizePaymentRequest,
	tokenResp *client.TokenizeCardResponse,
	fraudResp *client.FraudCheckResponse,
	reason string,
) (*PaymentResponse, error) {
	payment := &model.Payment{
		MerchantID:    req.MerchantID,
		Type:          model.PaymentTypeAuthorize,
		Status:        model.PaymentStatusFailed,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Token:         tokenResp.Token,
		CardBrand:     tokenResp.CardBrand,
		CardLast4:     tokenResp.Last4,
		FraudScore:    fraudResp.RiskScore,
		FraudDecision: fraudResp.Decision,
		ResponseMsg:   sql.NullString{String: reason, Valid: true},
		IPAddress:     req.IPAddress,
		CreatedBy:     req.CreatedBy,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, err
	}

	return s.buildPaymentResponse(payment), nil
}

func (s *PaymentService) buildPaymentResponse(payment *model.Payment) *PaymentResponse {
	resp := &PaymentResponse{
		ID:            payment.ID,
		Status:        payment.Status,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Token:         payment.Token,
		CardBrand:     payment.CardBrand,
		CardLast4:     payment.CardLast4,
		FraudScore:    payment.FraudScore,
		FraudDecision: payment.FraudDecision,
		TransactionID: payment.TransactionID,
		CreatedAt:     payment.CreatedAt,
	}

	if payment.AuthCode.Valid {
		resp.AuthCode = payment.AuthCode.String
	}
	if payment.ResponseCode.Valid {
		resp.ResponseCode = payment.ResponseCode.String
	}
	if payment.ResponseMsg.Valid {
		resp.ResponseMsg = payment.ResponseMsg.String
	}

	return resp
}

func (s *PaymentService) GetPayment(paymentID, merchantID uuid.UUID) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.FindByIDAndMerchant(paymentID, merchantID)
	if err != nil {
		return nil, err
	}
	return s.buildPaymentResponse(payment), nil
}
