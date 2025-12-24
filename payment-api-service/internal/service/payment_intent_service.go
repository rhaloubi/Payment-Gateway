package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/payment-api-service/internal/models"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/repository"
	"go.uber.org/zap"
)

type PaymentIntentService struct {
	intentRepo     *repository.PaymentIntentRepository
	paymentService *PaymentService
}

func NewPaymentIntentService(paymentService *PaymentService) *PaymentIntentService {
	return &PaymentIntentService{
		intentRepo:     repository.NewPaymentIntentRepository(),
		paymentService: paymentService,
	}
}

// =========================================================================
// Request/Response DTOs
// =========================================================================

type CreatePaymentIntentRequest struct {
	MerchantID    uuid.UUID
	Amount        int64
	Currency      string
	OrderID       string // Optional
	Description   string
	CaptureMethod model.CaptureMethod
	SuccessURL    string
	CancelURL     string
	CustomerEmail string
	Metadata      map[string]interface{}
}

type PaymentIntentResponse struct {
	ID           uuid.UUID                 `json:"id"`
	ClientSecret string                    `json:"client_secret"`
	Status       model.PaymentIntentStatus `json:"status"`
	Amount       int64                     `json:"amount"`
	Currency     string                    `json:"currency"`
	SuccessURL   string                    `json:"success_url"`
	CancelURL    string                    `json:"cancel_url"`
	CheckoutURL  string                    `json:"checkout_url"`
	ExpiresAt    time.Time                 `json:"expires_at"`
	CreatedAt    time.Time                 `json:"created_at"`
}

type ConfirmPaymentIntentRequest struct {
	PaymentIntentID string
	ClientSecret    string
	CardNumber      string
	CardholderName  string
	ExpMonth        int
	ExpYear         int
	CVV             string
	CustomerEmail   string // Can override
	IdempotencyKey  string // Optional
	IPAddress       string
	UserAgent       string
}
type PaymentIntentError struct {
	Code           string
	Message        string
	RemainingTries int
}

func (e *PaymentIntentError) Error() string {
	return e.Message
}

// =========================================================================
// Create Payment Intent
// =========================================================================

func (s *PaymentIntentService) CreatePaymentIntent(ctx context.Context, req *CreatePaymentIntentRequest) (*PaymentIntentResponse, error) {
	logger.Log.Info("Creating payment intent",
		zap.String("merchant_id", req.MerchantID.String()),
		zap.Int64("amount", req.Amount),
		zap.String("currency", req.Currency),
	)

	// Validate
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if req.Currency != "USD" && req.Currency != "EUR" && req.Currency != "MAD" {
		return nil, errors.New("unsupported currency")
	}
	if req.SuccessURL == "" {
		return nil, errors.New("success_url is required")
	}

	// Set defaults
	if req.CaptureMethod == "" {
		req.CaptureMethod = model.CaptureMethodAutomatic
	}

	// Generate client secret (browser authentication)
	clientSecret, err := generateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client secret: %w", err)
	}

	// Create payment intent with 1-hour expiration
	intent := &model.PaymentIntent{
		MerchantID:    req.MerchantID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        model.PaymentIntentStatusAwaitingPayment,
		CaptureMethod: req.CaptureMethod,
		SuccessURL:    req.SuccessURL,
		CancelURL:     req.CancelURL,
		ClientSecret:  clientSecret,
		MaxAttempts:   7,
		AttemptCount:  0,
		ExpiresAt:     time.Now().Add(1 * time.Hour), // 1 HOUR EXPIRATION
	}

	if req.OrderID != "" {
		intent.OrderID = sql.NullString{String: req.OrderID, Valid: true}
	}
	if req.Description != "" {
		intent.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.CustomerEmail != "" {
		intent.CustomerEmail = sql.NullString{String: req.CustomerEmail, Valid: true}
	}

	if err := s.intentRepo.Create(intent); err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	logger.Log.Info("Payment intent created",
		zap.String("intent_id", intent.ID.String()),
		zap.Time("expires_at", intent.ExpiresAt),
	)

	return &PaymentIntentResponse{
		ID:           intent.ID,
		ClientSecret: intent.ClientSecret,
		Status:       intent.Status,
		Amount:       intent.Amount,
		Currency:     intent.Currency,
		CheckoutURL:  intent.GetCheckoutURL(os.Getenv("CHECKOUT_URL")),
		ExpiresAt:    intent.ExpiresAt,
		CreatedAt:    intent.CreatedAt,
	}, nil
}

// =========================================================================
// Get Payment Intent (Browser-Safe)
// =========================================================================

func (s *PaymentIntentService) GetPaymentIntent(ctx context.Context, intentID uuid.UUID) (*PaymentIntentResponse, error) {
	intent, err := s.intentRepo.FindByID(intentID)
	if err != nil {
		return nil, fmt.Errorf("payment intent not found: %w", err)
	}

	// Check expiration
	if intent.IsExpired() && intent.Status == model.PaymentIntentStatusAwaitingPayment {
		s.intentRepo.MarkExpired(intentID)
		intent.Status = model.PaymentIntentStatusExpired
	}

	// Return safe data (no client_secret)
	return &PaymentIntentResponse{
		ID:         intent.ID,
		Status:     intent.Status,
		Amount:     intent.Amount,
		Currency:   intent.Currency,
		SuccessURL: intent.SuccessURL,
		CancelURL:  intent.CancelURL,
		ExpiresAt:  intent.ExpiresAt,
		CreatedAt:  intent.CreatedAt,
	}, nil
}

// =========================================================================
// Confirm Payment Intent (Process Payment)
// =========================================================================

func (s *PaymentIntentService) ConfirmPaymentIntent(ctx context.Context, req *ConfirmPaymentIntentRequest) (*PaymentResponse, error) {
	logger.Log.Info("Confirming payment intent",
		zap.String("intent_id", req.PaymentIntentID),
	)

	// Parse intent ID
	intentID, err := uuid.Parse(req.PaymentIntentID)
	if err != nil {
		return nil, &PaymentIntentError{
			Code:    "INVALID_INTENT_ID",
			Message: "Invalid payment intent ID",
		}
	}

	// Verify client secret
	intent, err := s.intentRepo.FindByClientSecret(req.ClientSecret)
	if err != nil || intent.ID != intentID {
		return nil, &PaymentIntentError{
			Code:    "INVALID_CLIENT_SECRET",
			Message: "Invalid client secret",
		}
	}

	// ===================================================================
	// VALIDATION CHECKS
	// ===================================================================

	// Check if expired
	if intent.IsExpired() {
		s.intentRepo.UpdateStatus(intentID, model.PaymentIntentStatusExpired)
		return nil, &PaymentIntentError{
			Code:    "INTENT_EXPIRED",
			Message: fmt.Sprintf("Payment intent expired at %s. Please create a new payment.", intent.ExpiresAt.Format("15:04:05")),
		}
	}

	// Check if max attempts reached
	if intent.AttemptCount >= intent.MaxAttempts {
		s.intentRepo.UpdateStatus(intentID, model.PaymentIntentStatusFailed)
		return nil, &PaymentIntentError{
			Code:    "MAX_ATTEMPTS_REACHED",
			Message: fmt.Sprintf("Maximum payment attempts (%d) reached. Please create a new payment intent.", intent.MaxAttempts),
		}
	}

	// Check if can confirm
	if !intent.CanConfirm() {
		return nil, &PaymentIntentError{
			Code:           "CANNOT_CONFIRM",
			Message:        fmt.Sprintf("Payment intent cannot be confirmed (status: %s)", intent.Status),
			RemainingTries: intent.GetRemainingAttempts(),
		}
	}

	// ===================================================================
	// INCREMENT ATTEMPT COUNTER
	// ===================================================================
	if err = s.intentRepo.IncrementAttemptCount(intentID); err != nil {
		logger.Log.Error("Failed to increment attempt count", zap.Error(err))
	}

	// Refresh intent to get updated attempt count
	intent, _ = s.intentRepo.FindByID(intentID)

	logger.Log.Info("Processing payment attempt",
		zap.String("intent_id", intentID.String()),
		zap.Int("attempt", intent.AttemptCount),
		zap.Int("remaining", intent.GetRemainingAttempts()),
	)

	// ===================================================================
	// BUILD PAYMENT REQUEST
	// ===================================================================
	authReq := &AuthorizePaymentRequest{
		MerchantID:     intent.MerchantID,
		Amount:         intent.Amount,
		Currency:       intent.Currency,
		CardNumber:     req.CardNumber,
		CardholderName: req.CardholderName,
		ExpMonth:       req.ExpMonth,
		ExpYear:        req.ExpYear,
		CVV:            req.CVV,
		CustomerEmail:  req.CustomerEmail,
		IdempotencyKey: req.IdempotencyKey,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
	}

	// Use customer email from request or intent
	if req.CustomerEmail != "" {
		authReq.CustomerEmail = req.CustomerEmail
	} else if intent.CustomerEmail.Valid {
		authReq.CustomerEmail = intent.CustomerEmail.String
	}

	// Set description from intent
	if intent.Description.Valid {
		authReq.Description = intent.Description.String
	}

	// ===================================================================
	// PROCESS PAYMENT
	// ===================================================================
	var paymentResp *PaymentResponse
	if intent.CaptureMethod == model.CaptureMethodAutomatic {
		paymentResp, err = s.paymentService.SalePayment(ctx, authReq)
	} else {
		paymentResp, err = s.paymentService.AuthorizePayment(ctx, authReq)
	}

	// ===================================================================
	// HANDLE PAYMENT RESULT
	// ===================================================================
	if err != nil {
		logger.Log.Warn("Payment authorization failed",
			zap.Error(err),
			zap.String("intent_id", intentID.String()),
			zap.Int("attempt", intent.AttemptCount),
			zap.Int("remaining", intent.GetRemainingAttempts()),
		)

		// Check if this was the last attempt
		if intent.GetRemainingAttempts() == 0 {
			s.intentRepo.UpdateStatus(intentID, model.PaymentIntentStatusFailed)
			return nil, &PaymentIntentError{
				Code:           "MAX_ATTEMPTS_REACHED",
				Message:        "Payment failed. Maximum attempts reached. Please create a new payment intent.",
				RemainingTries: 0,
			}
		}

		// Return error with remaining attempts
		return nil, &PaymentIntentError{
			Code:           "PAYMENT_FAILED",
			Message:        fmt.Sprintf("Payment failed: %s", err.Error()),
			RemainingTries: intent.GetRemainingAttempts(),
		}
	}

	// ===================================================================
	// PAYMENT SUCCESSFUL
	// ===================================================================
	logger.Log.Info("Payment authorization successful",
		zap.String("intent_id", intentID.String()),
		zap.String("payment_id", paymentResp.ID.String()),
		zap.String("status", string(paymentResp.Status)),
	)

	// Update intent status based on payment result
	if paymentResp.Status == model.PaymentStatusAuthorized ||
		paymentResp.Status == model.PaymentStatusCaptured {

		// Mark as confirmed and reset attempts
		s.intentRepo.MarkConfirmed(intentID, paymentResp.ID)
		s.intentRepo.ResetAttempts(intentID)

		logger.Log.Info("Payment intent confirmed",
			zap.String("intent_id", intentID.String()),
			zap.String("payment_id", paymentResp.ID.String()),
		)
	} else {
		// Payment was processed but not successful (declined by bank)
		if intent.GetRemainingAttempts() == 0 {
			s.intentRepo.UpdateStatus(intentID, model.PaymentIntentStatusFailed)
		}

		return nil, &PaymentIntentError{
			Code:           "PAYMENT_DECLINED",
			Message:        paymentResp.ResponseMsg,
			RemainingTries: intent.GetRemainingAttempts(),
		}
	}

	return paymentResp, nil
}

// =========================================================================
// Cancel Payment Intent
// =========================================================================

func (s *PaymentIntentService) CancelPaymentIntent(ctx context.Context, intentID, merchantID uuid.UUID) error {
	intent, err := s.intentRepo.FindByIDAndMerchant(intentID, merchantID)
	if err != nil {
		return fmt.Errorf("payment intent not found: %w", err)
	}

	if !intent.CanCancel() {
		return fmt.Errorf("payment intent cannot be canceled (status: %s)", intent.Status)
	}

	// If already authorized, void the payment
	if intent.Status == model.PaymentIntentStatusAuthorized && intent.PaymentID.Valid {
		paymentID, _ := uuid.Parse(intent.PaymentID.String)
		_, err := s.paymentService.VoidPayment(ctx, paymentID, merchantID, "Payment intent canceled")
		if err != nil {
			logger.Log.Error("Failed to void payment",
				zap.Error(err),
				zap.String("payment_id", paymentID.String()),
			)
		}
	}

	// Mark intent as canceled
	if err := s.intentRepo.MarkCanceled(intentID); err != nil {
		return err
	}

	logger.Log.Info("Payment intent canceled",
		zap.String("intent_id", intentID.String()),
	)

	return nil
}

// =========================================================================
// Helpers
// =========================================================================

func generateClientSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "pi_secret_" + base64.URLEncoding.EncodeToString(bytes), nil
}
