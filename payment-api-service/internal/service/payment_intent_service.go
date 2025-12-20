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
	IPAddress       string
	UserAgent       string
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

	// Create payment intent
	intent := &model.PaymentIntent{
		MerchantID:    req.MerchantID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        model.PaymentIntentStatusAwaitingPayment,
		CaptureMethod: req.CaptureMethod,
		SuccessURL:    req.SuccessURL,
		CancelURL:     req.CancelURL,
		ClientSecret:  clientSecret,
		ExpiresAt:     time.Now().Add(24 * time.Hour), // 24 hour expiration
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
		ID:        intent.ID,
		Status:    intent.Status,
		Amount:    intent.Amount,
		Currency:  intent.Currency,
		ExpiresAt: intent.ExpiresAt,
		CreatedAt: intent.CreatedAt,
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
		return nil, errors.New("invalid payment_intent_id")
	}

	// Verify client secret
	intent, err := s.intentRepo.FindByClientSecret(req.ClientSecret)
	if err != nil || intent.ID != intentID {
		return nil, errors.New("invalid client_secret")
	}

	// Check if can confirm
	if !intent.CanConfirm() {
		return nil, fmt.Errorf("payment intent cannot be confirmed (status: %s)", intent.Status)
	}

	// Build authorize request using intent's amount/currency
	authReq := &AuthorizePaymentRequest{
		MerchantID:     intent.MerchantID,
		Amount:         intent.Amount,      // FROM INTENT (server-set)
		Currency:       intent.Currency,    // FROM INTENT (server-set)
		CardNumber:     req.CardNumber,     // FROM BROWSER
		CardholderName: req.CardholderName, // FROM BROWSER
		ExpMonth:       req.ExpMonth,       // FROM BROWSER
		ExpYear:        req.ExpYear,        // FROM BROWSER
		CVV:            req.CVV,            // FROM BROWSER
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

	// Process payment authorization
	var paymentResp *PaymentResponse
	if intent.CaptureMethod == model.CaptureMethodAutomatic {
		// Sale (authorize + capture)
		paymentResp, err = s.paymentService.SalePayment(ctx, authReq)
	} else {
		// Authorize only
		paymentResp, err = s.paymentService.AuthorizePayment(ctx, authReq)
	}

	if err != nil {
		// Mark intent as failed
		intent.Status = model.PaymentIntentStatusFailed
		s.intentRepo.Update(intent)

		logger.Log.Error("Payment authorization failed",
			zap.Error(err),
			zap.String("intent_id", intentID.String()),
		)
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	// Update intent with payment reference
	if paymentResp.Status == model.PaymentStatusAuthorized ||
		paymentResp.Status == model.PaymentStatusCaptured {
		s.intentRepo.MarkConfirmed(intentID, paymentResp.ID)

		logger.Log.Info("Payment intent confirmed",
			zap.String("intent_id", intentID.String()),
			zap.String("payment_id", paymentResp.ID.String()),
		)
	} else {
		// Mark as failed
		intent.Status = model.PaymentIntentStatusFailed
		s.intentRepo.Update(intent)
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
