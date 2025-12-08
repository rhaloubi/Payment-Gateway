package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"github.com/rhaloubi/payment-gateway/transaction-service/internal/repository"
	"go.uber.org/zap"
)

type ChargebackService struct {
	chargebackRepo *repository.ChargebackRepository
	txnRepo        *repository.TransactionRepository
}

func NewChargebackService() *ChargebackService {
	return &ChargebackService{
		chargebackRepo: repository.NewChargebackRepository(),
		txnRepo:        repository.NewTransactionRepository(),
	}
}

// =========================================================================
// Request/Response DTOs
// =========================================================================

type CreateChargebackRequest struct {
	TransactionID     uuid.UUID
	Reason            model.ChargebackReason
	ReasonCode        string
	Amount            int64
	CustomerStatement string
	IssuerReference   string
	IssuerBank        string
}

type SubmitEvidenceRequest struct {
	ChargebackID      uuid.UUID
	MerchantID        uuid.UUID
	Evidence          map[string]interface{} // Evidence documents
	MerchantStatement string
}

type AcceptChargebackRequest struct {
	ChargebackID uuid.UUID
	MerchantID   uuid.UUID
	Reason       string
}

// =========================================================================
// Create Chargeback (Initiated by customer via issuing bank)
// =========================================================================

func (s *ChargebackService) CreateChargeback(ctx context.Context, req *CreateChargebackRequest) (*model.Chargeback, error) {
	logger.Log.Info("Creating chargeback",
		zap.String("transaction_id", req.TransactionID.String()),
		zap.String("reason", string(req.Reason)),
	)

	// Step 1: Get transaction
	txn, err := s.txnRepo.FindByID(req.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// Step 2: Validate transaction is eligible for chargeback
	if txn.Status != model.TransactionStatusCaptured && txn.Status != model.TransactionStatusSettled {
		return nil, errors.New("transaction is not eligible for chargeback (must be captured or settled)")
	}

	// Step 3: Check if chargeback already exists
	existing, _ := s.chargebackRepo.FindByTransaction(req.TransactionID)
	if len(existing) > 0 {
		for _, cb := range existing {
			if cb.IsOpen() {
				return nil, errors.New("an open chargeback already exists for this transaction")
			}
		}
	}

	// Step 4: Calculate chargeback fee and net loss
	chargebackFee := int64(1500) // $15.00 in cents
	netLoss := req.Amount + chargebackFee

	// Step 5: Create chargeback record
	chargeback := &model.Chargeback{
		TransactionID: req.TransactionID,
		MerchantID:    txn.MerchantID,
		Status:        model.ChargebackStatusNeedsResponse,
		Reason:        req.Reason,
		ReasonCode:    req.ReasonCode,
		Amount:        req.Amount,
		Currency:      txn.Currency,
		ChargebackFee: chargebackFee,
		NetLoss:       netLoss,
		DisputedAt:    time.Now(),
	}

	// Set response deadline (typically 7-10 days)
	responseDue := time.Now().Add(7 * 24 * time.Hour)
	chargeback.ResponseDueDate = sql.NullTime{Time: responseDue, Valid: true}

	// Set issuer info
	if req.IssuerReference != "" {
		chargeback.IssuerReference = sql.NullString{String: req.IssuerReference, Valid: true}
	}
	if req.IssuerBank != "" {
		chargeback.IssuerBank = sql.NullString{String: req.IssuerBank, Valid: true}
	}
	if req.CustomerStatement != "" {
		chargeback.CustomerStatement = sql.NullString{String: req.CustomerStatement, Valid: true}
	}

	// Step 6: Save chargeback
	if err := s.chargebackRepo.Create(chargeback); err != nil {
		logger.Log.Error("Failed to create chargeback", zap.Error(err))
		return nil, fmt.Errorf("failed to create chargeback: %w", err)
	}

	// Step 7: Log event
	go s.chargebackRepo.CreateEvent(&model.ChargebackEvent{
		ChargebackID: chargeback.ID,
		EventType:    "chargeback_created",
		OldStatus:    "",
		NewStatus:    model.ChargebackStatusNeedsResponse,
	})

	logger.Log.Info("Chargeback created",
		zap.String("chargeback_id", chargeback.ID.String()),
		zap.String("transaction_id", req.TransactionID.String()),
		zap.Int64("amount", req.Amount),
	)

	// TODO: Send notification to merchant (email, webhook)

	return chargeback, nil
}

// =========================================================================
// Submit Evidence (Merchant disputes chargeback)
// =========================================================================

func (s *ChargebackService) SubmitEvidence(ctx context.Context, req *SubmitEvidenceRequest) error {
	logger.Log.Info("Submitting chargeback evidence",
		zap.String("chargeback_id", req.ChargebackID.String()),
	)

	// Step 1: Get chargeback
	chargeback, err := s.chargebackRepo.FindByID(req.ChargebackID)
	if err != nil {
		return fmt.Errorf("chargeback not found: %w", err)
	}

	// Step 2: Verify merchant ownership
	if chargeback.MerchantID != req.MerchantID {
		return errors.New("access denied: chargeback belongs to different merchant")
	}

	// Step 3: Validate can submit evidence
	if !chargeback.NeedsResponse() {
		return errors.New("chargeback is not in a state that accepts evidence")
	}

	// Step 4: Store evidence (as JSON)
	evidenceJSON, _ := sql.NullString{String: fmt.Sprintf("%v", req.Evidence), Valid: true}.Value()
	chargeback.MerchantEvidence = sql.NullString{String: evidenceJSON.(string), Valid: true}
	chargeback.ResponseSubmittedAt = sql.NullTime{Time: time.Now(), Valid: true}
	chargeback.Status = model.ChargebackStatusUnderReview

	// Step 5: Update chargeback
	if err := s.chargebackRepo.Update(chargeback); err != nil {
		return fmt.Errorf("failed to update chargeback: %w", err)
	}

	// Step 6: Log event
	go s.chargebackRepo.CreateEvent(&model.ChargebackEvent{
		ChargebackID: req.ChargebackID,
		EventType:    "evidence_submitted",
		OldStatus:    model.ChargebackStatusNeedsResponse,
		NewStatus:    model.ChargebackStatusUnderReview,
		Note:         sql.NullString{String: "Merchant submitted evidence", Valid: true},
	})

	logger.Log.Info("Evidence submitted",
		zap.String("chargeback_id", req.ChargebackID.String()),
	)

	// TODO: Send evidence to issuing bank/card network

	return nil
}

// =========================================================================
// Accept Chargeback (Merchant accepts and won't dispute)
// =========================================================================

func (s *ChargebackService) AcceptChargeback(ctx context.Context, req *AcceptChargebackRequest) error {
	logger.Log.Info("Accepting chargeback",
		zap.String("chargeback_id", req.ChargebackID.String()),
	)

	// Step 1: Get chargeback
	chargeback, err := s.chargebackRepo.FindByID(req.ChargebackID)
	if err != nil {
		return fmt.Errorf("chargeback not found: %w", err)
	}

	// Step 2: Verify merchant ownership
	if chargeback.MerchantID != req.MerchantID {
		return errors.New("access denied")
	}

	// Step 3: Validate can accept
	if !chargeback.IsOpen() {
		return errors.New("chargeback cannot be accepted (already resolved)")
	}

	// Step 4: Update status
	oldStatus := chargeback.Status
	chargeback.Status = model.ChargebackStatusAccepted
	chargeback.ResolutionReason = sql.NullString{String: req.Reason, Valid: true}
	chargeback.ResolvedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err := s.chargebackRepo.Update(chargeback); err != nil {
		return fmt.Errorf("failed to accept chargeback: %w", err)
	}

	// Step 5: Log event
	go s.chargebackRepo.CreateEvent(&model.ChargebackEvent{
		ChargebackID: req.ChargebackID,
		EventType:    "chargeback_accepted",
		OldStatus:    oldStatus,
		NewStatus:    model.ChargebackStatusAccepted,
		Note:         sql.NullString{String: req.Reason, Valid: true},
	})

	logger.Log.Info("Chargeback accepted",
		zap.String("chargeback_id", req.ChargebackID.String()),
	)

	return nil
}

// =========================================================================
// Resolve Chargeback (Bank/network decision)
// =========================================================================

func (s *ChargebackService) ResolveChargeback(ctx context.Context, chargebackID uuid.UUID, merchantWon bool, reason string) error {
	chargeback, err := s.chargebackRepo.FindByID(chargebackID)
	if err != nil {
		return err
	}

	oldStatus := chargeback.Status
	if merchantWon {
		chargeback.Status = model.ChargebackStatusWon
	} else {
		chargeback.Status = model.ChargebackStatusLost
	}

	chargeback.ResolutionReason = sql.NullString{String: reason, Valid: true}
	chargeback.ResolvedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err := s.chargebackRepo.Update(chargeback); err != nil {
		return err
	}

	go s.chargebackRepo.CreateEvent(&model.ChargebackEvent{
		ChargebackID: chargebackID,
		EventType:    "chargeback_resolved",
		OldStatus:    oldStatus,
		NewStatus:    chargeback.Status,
		Note:         sql.NullString{String: reason, Valid: true},
	})

	logger.Log.Info("Chargeback resolved",
		zap.String("chargeback_id", chargebackID.String()),
		zap.Bool("merchant_won", merchantWon),
	)

	return nil
}

// =========================================================================
// Get Merchant Chargebacks
// =========================================================================

func (s *ChargebackService) GetMerchantChargebacks(merchantID uuid.UUID) ([]model.Chargeback, error) {
	return s.chargebackRepo.FindByMerchant(merchantID)
}

// GetChargebackByID retrieves a specific chargeback
func (s *ChargebackService) GetChargebackByID(chargebackID uuid.UUID) (*model.Chargeback, error) {
	return s.chargebackRepo.FindByID(chargebackID)
}
