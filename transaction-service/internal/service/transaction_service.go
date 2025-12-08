package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/transaction-service/internal/client"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"github.com/rhaloubi/payment-gateway/transaction-service/internal/repository"
	"go.uber.org/zap"
)

type TransactionService struct {
	txnRepo             *repository.TransactionRepository
	currencyService     *CurrencyService
	tokenizationClient  *client.TokenizationClient
	cardSimulatorClient *client.CardSimulatorClient
}

func NewTransactionService() (*TransactionService, error) {
	tokenClient, err := client.NewTokenizationClient()
	if err != nil {
		logger.Log.Warn("Failed to connect to tokenization service", zap.Error(err))
	}

	return &TransactionService{
		txnRepo:             repository.NewTransactionRepository(),
		currencyService:     NewCurrencyService(),
		tokenizationClient:  tokenClient,
		cardSimulatorClient: client.NewCardSimulatorClient(),
	}, nil
}

type AuthorizeRequest struct {
	MerchantID    uuid.UUID
	Amount        int64
	Currency      string
	CardToken     string
	CardBrand     string
	CardLast4     string
	FraudScore    int
	CustomerEmail string
	Description   string
	IPAddress     string
	UserAgent     string
}

type AuthorizeResponse struct {
	TransactionID   uuid.UUID
	Status          model.TransactionStatus
	Approved        bool
	AuthCode        string
	ResponseCode    string
	ResponseMessage string
	DeclineReason   string
	Amount          int64
	AmountMAD       int64
	ExchangeRate    float64
	ProcessingFee   int64
	NetAmount       int64
}

type CaptureRequest struct {
	TransactionID uuid.UUID
	Amount        int64
	MerchantID    uuid.UUID
}

type CaptureResponse struct {
	TransactionID   uuid.UUID
	Status          model.TransactionStatus
	CapturedAmount  int64
	ResponseMessage string
}

type VoidRequest struct {
	TransactionID uuid.UUID
	MerchantID    uuid.UUID
	Reason        string
}

type VoidResponse struct {
	TransactionID   uuid.UUID
	Status          model.TransactionStatus
	ResponseMessage string
}

type RefundRequest struct {
	TransactionID uuid.UUID
	Amount        int64
	Reason        string
	MerchantID    uuid.UUID
}

type RefundResponse struct {
	RefundID        uuid.UUID
	TransactionID   uuid.UUID
	RefundedAmount  int64
	RemainingAmount int64
	ResponseMessage string
}

// =========================================================================
// AUTHORIZE - Hold funds on customer's card
// =========================================================================

func (s *TransactionService) Authorize(ctx context.Context, req *AuthorizeRequest) (*AuthorizeResponse, error) {
	startTime := time.Now()
	logger.Log.Info("Processing authorization",
		zap.String("merchant_id", req.MerchantID.String()),
		zap.Int64("amount", req.Amount),
		zap.String("currency", req.Currency),
	)

	// Step 1: Validate request
	if err := s.validateAuthorizationRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Convert amount to MAD
	amountMAD, exchangeRate, err := s.currencyService.ConvertToMAD(req.Amount, req.Currency)
	if err != nil {
		return nil, fmt.Errorf("currency conversion failed: %w", err)
	}

	// Step 3: Calculate processing fee (2.9% + $0.30 in MAD)
	processingFee := s.currencyService.CalculateProcessingFee(amountMAD)
	netAmount := amountMAD - processingFee

	// Step 4: Check fraud score (auto-decline if > 70)
	if req.FraudScore > 70 {
		logger.Log.Warn("Transaction declined by fraud detection",
			zap.Int("fraud_score", req.FraudScore),
		)
		return s.createFailedTransaction(req, "Declined by fraud detection", amountMAD, exchangeRate, processingFee)
	}

	// Step 5: Detokenize card data
	cardData, err := s.tokenizationClient.Detokenize(ctx, req.CardToken, req.MerchantID.String())
	if err != nil {
		logger.Log.Error("Detokenization failed", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve card data: %w", err)
	}

	// Step 6: Call Card Simulator (issuer authorization)
	issuerResp, err := s.cardSimulatorClient.Authorize(ctx, &client.AuthorizeCardRequest{
		CardNumber: cardData.CardNumber,
		ExpMonth:   cardData.ExpMonth,
		ExpYear:    cardData.ExpYear,
		Amount:     req.Amount,
		Currency:   req.Currency,
		MerchantID: req.MerchantID.String(),
	})
	if err != nil {
		logger.Log.Error("Issuer authorization failed", zap.Error(err))
		return nil, fmt.Errorf("issuer authorization failed: %w", err)
	}

	// Step 7: Create transaction record
	txn := &model.Transaction{
		MerchantID:    req.MerchantID,
		Type:          model.TransactionTypeAuthorize,
		Amount:        req.Amount,
		Currency:      req.Currency,
		AmountMAD:     amountMAD,
		ExchangeRate:  exchangeRate,
		CardToken:     req.CardToken,
		CardBrand:     req.CardBrand,
		CardLast4:     req.CardLast4,
		FraudScore:    req.FraudScore,
		ProcessingFee: processingFee,
		NetAmount:     netAmount,
		IPAddress:     req.IPAddress,
	}

	if req.UserAgent != "" {
		txn.UserAgent = sql.NullString{String: req.UserAgent, Valid: true}
	}
	if req.Description != "" {
		txn.Description = sql.NullString{String: req.Description, Valid: true}
	}

	// Step 8: Set status based on issuer response
	if issuerResp.Approved {
		txn.Status = model.TransactionStatusAuthorized
		txn.AuthCode = sql.NullString{String: issuerResp.AuthCode, Valid: true}
		txn.ResponseCode = sql.NullString{String: issuerResp.ResponseCode, Valid: true}
		txn.ResponseMessage = sql.NullString{String: issuerResp.ResponseMessage, Valid: true}
		now := time.Now()
		txn.AuthorizedAt = sql.NullTime{Time: now, Valid: true}
		txn.ExpiresAt = sql.NullTime{Time: now.Add(7 * 24 * time.Hour), Valid: true}

		if issuerResp.AVSResult != "" {
			txn.AVSResult = sql.NullString{String: issuerResp.AVSResult, Valid: true}
		}
		if issuerResp.CVVResult != "" {
			txn.CVVResult = sql.NullString{String: issuerResp.CVVResult, Valid: true}
		}
	} else {
		txn.Status = model.TransactionStatusFailed
		txn.ResponseCode = sql.NullString{String: issuerResp.ResponseCode, Valid: true}
		txn.ResponseMessage = sql.NullString{String: issuerResp.DeclineReason, Valid: true}
	}

	// Step 9: Save transaction
	if err := s.txnRepo.Create(txn); err != nil {
		logger.Log.Error("Failed to save transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	// Step 10: Log transaction event
	go s.txnRepo.CreateEvent(&model.TransactionEvent{
		TransactionID: txn.ID,
		EventType:     "authorized",
		OldStatus:     model.TransactionStatusPending,
		NewStatus:     txn.Status,
		Amount:        txn.Amount,
	})

	// Step 11: Store issuer response for debugging
	s.storeIssuerResponse(txn.ID, issuerResp, time.Since(startTime))

	logger.Log.Info("Authorization completed",
		zap.String("transaction_id", txn.ID.String()),
		zap.String("status", string(txn.Status)),
		zap.Bool("approved", issuerResp.Approved),
		zap.Duration("processing_time", time.Since(startTime)),
	)

	// Step 12: Build response
	response := &AuthorizeResponse{
		TransactionID: txn.ID,
		Status:        txn.Status,
		Approved:      issuerResp.Approved,
		Amount:        txn.Amount,
		AmountMAD:     amountMAD,
		ExchangeRate:  exchangeRate,
		ProcessingFee: processingFee,
		NetAmount:     netAmount,
	}

	if issuerResp.Approved {
		response.AuthCode = issuerResp.AuthCode
		response.ResponseCode = issuerResp.ResponseCode
		response.ResponseMessage = issuerResp.ResponseMessage
	} else {
		response.ResponseCode = issuerResp.ResponseCode
		response.DeclineReason = issuerResp.DeclineReason
	}

	return response, nil
}

// =========================================================================
// CAPTURE - Charge previously authorized funds
// =========================================================================

func (s *TransactionService) Capture(ctx context.Context, req *CaptureRequest) (*CaptureResponse, error) {
	logger.Log.Info("Processing capture",
		zap.String("transaction_id", req.TransactionID.String()),
		zap.Int64("amount", req.Amount),
	)

	// Step 1: Get transaction
	txn, err := s.txnRepo.FindByIDAndMerchant(req.TransactionID, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// Step 2: Validate can capture
	if !txn.CanCapture() {
		return nil, errors.New("transaction cannot be captured (not in authorized state or expired)")
	}

	// Step 3: Validate capture amount
	if req.Amount > txn.Amount {
		return nil, errors.New("capture amount exceeds authorized amount")
	}

	// Step 4: Call card simulator to finalize capture
	captureResp, err := s.cardSimulatorClient.Capture(ctx, &client.CaptureCardRequest{
		TransactionID: req.TransactionID.String(),
		Amount:        req.Amount,
	})
	if err != nil {
		logger.Log.Error("Capture failed at issuer", zap.Error(err))
		return nil, fmt.Errorf("capture failed: %w", err)
	}

	if !captureResp.Success {
		return nil, errors.New("capture declined by issuer")
	}

	// Step 5: Update transaction
	if err := s.txnRepo.MarkCaptured(req.TransactionID, req.Amount); err != nil {
		return nil, err
	}

	// Step 6: Log event
	go s.txnRepo.CreateEvent(&model.TransactionEvent{
		TransactionID: req.TransactionID,
		EventType:     "captured",
		OldStatus:     model.TransactionStatusAuthorized,
		NewStatus:     model.TransactionStatusCaptured,
		Amount:        req.Amount,
	})

	logger.Log.Info("Capture completed",
		zap.String("transaction_id", req.TransactionID.String()),
		zap.Int64("amount", req.Amount),
	)

	return &CaptureResponse{
		TransactionID:   req.TransactionID,
		Status:          model.TransactionStatusCaptured,
		CapturedAmount:  req.Amount,
		ResponseMessage: "Capture successful",
	}, nil
}

// =========================================================================
// VOID - Cancel authorization before capture
// =========================================================================

func (s *TransactionService) Void(ctx context.Context, req *VoidRequest) (*VoidResponse, error) {
	logger.Log.Info("Processing void",
		zap.String("transaction_id", req.TransactionID.String()),
	)

	// Step 1: Get transaction
	txn, err := s.txnRepo.FindByIDAndMerchant(req.TransactionID, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// Step 2: Validate can void
	if !txn.CanVoid() {
		return nil, errors.New("transaction cannot be voided (not in authorized state)")
	}

	// Step 3: Call card simulator to void
	voidResp, err := s.cardSimulatorClient.Void(ctx, &client.VoidCardRequest{
		TransactionID: req.TransactionID.String(),
		Reason:        req.Reason,
	})
	if err != nil {
		logger.Log.Error("Void failed at issuer", zap.Error(err))
		return nil, fmt.Errorf("void failed: %w", err)
	}

	if !voidResp.Success {
		return nil, errors.New("void declined by issuer")
	}

	// Step 4: Update transaction
	if err := s.txnRepo.MarkVoided(req.TransactionID); err != nil {
		return nil, err
	}

	// Step 5: Log event
	go s.txnRepo.CreateEvent(&model.TransactionEvent{
		TransactionID: req.TransactionID,
		EventType:     "voided",
		OldStatus:     model.TransactionStatusAuthorized,
		NewStatus:     model.TransactionStatusVoided,
		Amount:        txn.Amount,
		Metadata:      sql.NullString{String: fmt.Sprintf(`{"reason":"%s"}`, req.Reason), Valid: true},
	})

	logger.Log.Info("Void completed",
		zap.String("transaction_id", req.TransactionID.String()),
	)

	return &VoidResponse{
		TransactionID:   req.TransactionID,
		Status:          model.TransactionStatusVoided,
		ResponseMessage: "Authorization voided successfully",
	}, nil
}

// =========================================================================
// REFUND - Return funds to customer
// =========================================================================

func (s *TransactionService) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	logger.Log.Info("Processing refund",
		zap.String("transaction_id", req.TransactionID.String()),
		zap.Int64("amount", req.Amount),
	)

	// Step 1: Get original transaction
	originalTxn, err := s.txnRepo.FindByIDAndMerchant(req.TransactionID, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// Step 2: Validate can refund
	if !originalTxn.CanRefund() {
		return nil, errors.New("transaction cannot be refunded")
	}

	// Step 3: Validate refund amount
	if req.Amount > originalTxn.RemainingRefundableAmount() {
		return nil, fmt.Errorf("refund amount exceeds remaining refundable amount (%d)",
			originalTxn.RemainingRefundableAmount())
	}

	// Step 4: Call card simulator to process refund
	refundResp, err := s.cardSimulatorClient.Refund(ctx, &client.RefundCardRequest{
		TransactionID: req.TransactionID.String(),
		Amount:        req.Amount,
		Reason:        req.Reason,
	})
	if err != nil {
		logger.Log.Error("Refund failed at issuer", zap.Error(err))
		return nil, fmt.Errorf("refund failed: %w", err)
	}

	if !refundResp.Success {
		return nil, errors.New("refund declined by issuer")
	}

	// Step 5: Create refund transaction record
	refundTxn := &model.Transaction{
		MerchantID:          req.MerchantID,
		ParentTransactionID: sql.NullString{String: req.TransactionID.String(), Valid: true},
		Type:                model.TransactionTypeRefund,
		Status:              model.TransactionStatusRefunded,
		Amount:              -req.Amount, // Negative amount for refund
		Currency:            originalTxn.Currency,
		AmountMAD:           -originalTxn.AmountMAD * req.Amount / originalTxn.CapturedAmount,
		ExchangeRate:        originalTxn.ExchangeRate,
		CardToken:           originalTxn.CardToken,
		CardBrand:           originalTxn.CardBrand,
		CardLast4:           originalTxn.CardLast4,
		Description:         sql.NullString{String: req.Reason, Valid: true},
	}

	now := time.Now()
	refundTxn.RefundedAt = sql.NullTime{Time: now, Valid: true}

	// Step 6: Save refund transaction
	if err := s.txnRepo.Create(refundTxn); err != nil {
		return nil, fmt.Errorf("failed to save refund transaction: %w", err)
	}

	// Step 7: Update original transaction refunded amount
	if err := s.txnRepo.AddRefundAmount(req.TransactionID, req.Amount); err != nil {
		return nil, err
	}

	// Step 8: Log event
	go s.txnRepo.CreateEvent(&model.TransactionEvent{
		TransactionID: req.TransactionID,
		EventType:     "refunded",
		OldStatus:     originalTxn.Status,
		NewStatus:     model.TransactionStatusRefunded,
		Amount:        req.Amount,
	})

	logger.Log.Info("Refund completed",
		zap.String("refund_id", refundTxn.ID.String()),
		zap.String("transaction_id", req.TransactionID.String()),
		zap.Int64("amount", req.Amount),
	)

	// Refresh original transaction to get updated amounts
	originalTxn, _ = s.txnRepo.FindByID(req.TransactionID)

	return &RefundResponse{
		RefundID:        refundTxn.ID,
		TransactionID:   req.TransactionID,
		RefundedAmount:  req.Amount,
		RemainingAmount: originalTxn.RemainingRefundableAmount(),
		ResponseMessage: "Refund processed successfully",
	}, nil
}

// =========================================================================
// Helper Methods
// =========================================================================

func (s *TransactionService) validateAuthorizationRequest(req *AuthorizeRequest) error {
	if req.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	if req.Currency == model.CurrencyMAD || req.Currency == model.CurrencyEUR {
		if 500 > req.Amount || req.Amount > 2500000 {
			return errors.New("transaction amount must be between $5 and $25,000")
		}
	}
	if req.Currency == model.CurrencyMAD {
		if 5000 > req.Amount || req.Amount > 25000000 {
			return errors.New("transaction amount must be between DH50 and DH250,000")
		}
	}

	if req.Currency != model.CurrencyUSD && req.Currency != model.CurrencyEUR && req.Currency != model.CurrencyMAD {
		return errors.New("unsupported currency (only USD, EUR, MAD supported)")
	}

	return nil
}

func (s *TransactionService) createFailedTransaction(req *AuthorizeRequest, reason string, amountMAD int64, exchangeRate float64, processingFee int64) (*AuthorizeResponse, error) {
	txn := &model.Transaction{
		MerchantID:      req.MerchantID,
		Type:            model.TransactionTypeAuthorize,
		Status:          model.TransactionStatusFailed,
		Amount:          req.Amount,
		Currency:        req.Currency,
		AmountMAD:       amountMAD,
		ExchangeRate:    exchangeRate,
		CardToken:       req.CardToken,
		CardBrand:       req.CardBrand,
		CardLast4:       req.CardLast4,
		FraudScore:      req.FraudScore,
		ProcessingFee:   processingFee,
		ResponseMessage: sql.NullString{String: reason, Valid: true},
		IPAddress:       req.IPAddress,
	}

	s.txnRepo.Create(txn)

	return &AuthorizeResponse{
		TransactionID: txn.ID,
		Status:        model.TransactionStatusFailed,
		Approved:      false,
		DeclineReason: reason,
		Amount:        req.Amount,
		AmountMAD:     amountMAD,
	}, nil
}

func (s *TransactionService) storeIssuerResponse(txnID uuid.UUID, resp *client.AuthorizeCardResponse, processingTime time.Duration) {
	// Store for debugging
	s.txnRepo.CreateIssuerResponse(&model.IssuerResponse{
		TransactionID:    txnID,
		Approved:         resp.Approved,
		AuthCode:         sql.NullString{String: resp.AuthCode, Valid: resp.Approved},
		ResponseCode:     sql.NullString{String: resp.ResponseCode, Valid: true},
		ResponseMessage:  sql.NullString{String: resp.ResponseMessage, Valid: true},
		DeclineReason:    sql.NullString{String: resp.DeclineReason, Valid: !resp.Approved},
		AVSResult:        sql.NullString{String: resp.AVSResult, Valid: true},
		CVVResult:        sql.NullString{String: resp.CVVResult, Valid: true},
		ProcessingTimeMs: int(processingTime.Milliseconds()),
	})
}

func (s *TransactionService) GetTransaction(txnID, merchantID uuid.UUID) (*model.Transaction, error) {
	return s.txnRepo.FindByIDAndMerchant(txnID, merchantID)
}

func (s *TransactionService) FindByStatus(merchantID uuid.UUID, status model.TransactionStatus) ([]model.Transaction, error) {
	return s.txnRepo.FindByStatus(merchantID, status)
}

func (s *TransactionService) FindByMerchant(merchantID uuid.UUID, limit, offset int) ([]model.Transaction, error) {
	return s.txnRepo.FindByMerchant(merchantID, limit, offset)
}
