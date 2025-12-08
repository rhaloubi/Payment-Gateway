package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"github.com/rhaloubi/payment-gateway/transaction-service/internal/repository"
	"go.uber.org/zap"
)

type SettlementService struct {
	settlementRepo  *repository.SettlementRepository
	txnRepo         *repository.TransactionRepository
	currencyService *CurrencyService
}

func NewSettlementService() *SettlementService {
	return &SettlementService{
		settlementRepo:  repository.NewSettlementRepository(),
		txnRepo:         repository.NewTransactionRepository(),
		currencyService: NewCurrencyService(),
	}
}

// =========================================================================
// Daily Settlement Batch Creation (Runs at midnight)
// =========================================================================

func (s *SettlementService) CreateDailySettlementBatches(ctx context.Context) error {
	batchDate := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour) // Yesterday

	logger.Log.Info("Creating daily settlement batches",
		zap.Time("batch_date", batchDate),
	)

	// Get all captured transactions from yesterday
	transactions, err := s.txnRepo.FindCapturedForSettlement(batchDate)
	if err != nil {
		logger.Log.Error("Failed to find transactions for settlement", zap.Error(err))
		return err
	}

	if len(transactions) == 0 {
		logger.Log.Info("No transactions to settle")
		return nil
	}

	// Group transactions by merchant
	merchantTxns := s.groupTransactionsByMerchant(transactions)

	// Create batch for each merchant
	for merchantID, txns := range merchantTxns {
		if err := s.createMerchantSettlementBatch(merchantID, batchDate, txns); err != nil {
			logger.Log.Error("Failed to create settlement batch",
				zap.Error(err),
				zap.String("merchant_id", merchantID.String()),
			)
		}
	}

	logger.Log.Info("Daily settlement batches created",
		zap.Int("merchant_count", len(merchantTxns)),
		zap.Int("transaction_count", len(transactions)),
	)

	return nil
}

func (s *SettlementService) createMerchantSettlementBatch(
	merchantID uuid.UUID,
	batchDate time.Time,
	transactions []model.Transaction,
) error {
	logger.Log.Info("Creating settlement batch for merchant",
		zap.String("merchant_id", merchantID.String()),
		zap.Int("transaction_count", len(transactions)),
	)

	// Calculate totals
	var grossAmount int64
	var refundAmount int64
	var feeAmount int64
	transactionCount := 0
	refundCount := 0
	currencyBreakdown := make(map[string]int64)

	for _, txn := range transactions {
		if txn.Type == model.TransactionTypeRefund {
			refundAmount += -txn.AmountMAD // Refunds are negative
			refundCount++
		} else {
			grossAmount += txn.AmountMAD
			transactionCount++
			feeAmount += txn.ProcessingFee
		}

		// Track currency breakdown
		currencyBreakdown[txn.Currency] += txn.Amount
	}

	netAmount := grossAmount - refundAmount - feeAmount

	// Serialize currency breakdown
	breakdownJSON, _ := json.Marshal(currencyBreakdown)

	// Create settlement batch
	batch := &model.SettlementBatch{
		MerchantID:        merchantID,
		BatchDate:         batchDate,
		GrossAmount:       grossAmount,
		RefundAmount:      refundAmount,
		FeeAmount:         feeAmount,
		NetAmount:         netAmount,
		TransactionCount:  transactionCount,
		RefundCount:       refundCount,
		CurrencyBreakdown: sql.NullString{String: string(breakdownJSON), Valid: true},
		Status:            model.SettlementStatusPending,
		SettlementDate:    batchDate.AddDate(0, 0, 2), // T+2 settlement
		SettlementMethod:  "bank_transfer",
	}

	// TODO: Get merchant bank details from merchant service
	// batch.BankAccount = merchantBankAccount
	// batch.BankName = merchantBankName

	// Save batch
	if err := s.settlementRepo.Create(batch); err != nil {
		return fmt.Errorf("failed to save settlement batch: %w", err)
	}

	// Link transactions to batch
	txnIDs := make([]uuid.UUID, len(transactions))
	for i, txn := range transactions {
		txnIDs[i] = txn.ID
	}

	if err := s.txnRepo.LinkToSettlementBatch(txnIDs, batch.ID); err != nil {
		return fmt.Errorf("failed to link transactions to batch: %w", err)
	}

	logger.Log.Info("Settlement batch created",
		zap.String("batch_id", batch.ID.String()),
		zap.String("merchant_id", merchantID.String()),
		zap.Int64("net_amount", netAmount),
		zap.Int("transaction_count", transactionCount),
	)

	// TODO: Send notification to merchant
	// TODO: Generate settlement report (CSV)

	return nil
}

// =========================================================================
// Process Pending Settlements (Runs on T+2)
// =========================================================================

// ProcessPendingSettlements processes settlements that are due
func (s *SettlementService) ProcessPendingSettlements(ctx context.Context) error {
	logger.Log.Info("Processing pending settlements")

	// Get pending batches due for settlement
	batches, err := s.settlementRepo.FindPendingBatches()
	if err != nil {
		logger.Log.Error("Failed to find pending settlements", zap.Error(err))
		return err
	}

	if len(batches) == 0 {
		logger.Log.Info("No pending settlements to process")
		return nil
	}

	for _, batch := range batches {
		if err := s.processSettlementBatch(&batch); err != nil {
			logger.Log.Error("Failed to process settlement batch",
				zap.Error(err),
				zap.String("batch_id", batch.ID.String()),
			)
			// Continue with other batches
		}
	}

	logger.Log.Info("Pending settlements processed",
		zap.Int("batch_count", len(batches)),
	)

	return nil
}

// processSettlementBatch processes a single settlement batch
func (s *SettlementService) processSettlementBatch(batch *model.SettlementBatch) error {
	logger.Log.Info("Processing settlement batch",
		zap.String("batch_id", batch.ID.String()),
		zap.String("merchant_id", batch.MerchantID.String()),
		zap.Int64("net_amount", batch.NetAmount),
	)

	// TODO: Integrate with payment provider (bank transfer, ACH, wire)
	// For now, simulate successful settlement

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Mark batch as settled
	if err := s.settlementRepo.MarkSettled(batch.ID); err != nil {
		return fmt.Errorf("failed to mark batch as settled: %w", err)
	}

	logger.Log.Info("Settlement batch processed successfully",
		zap.String("batch_id", batch.ID.String()),
	)

	// TODO: Send settlement confirmation email to merchant
	// TODO: Update accounting records

	return nil
}

// =========================================================================
// Auto-Void Expired Authorizations (Runs hourly)
// =========================================================================

// AutoVoidExpiredAuthorizations voids authorizations older than 7 days
func (s *SettlementService) AutoVoidExpiredAuthorizations(ctx context.Context) error {
	logger.Log.Info("Auto-voiding expired authorizations")

	// Find expired authorizations
	expiredTxns, err := s.txnRepo.FindExpiredAuthorizations()
	if err != nil {
		logger.Log.Error("Failed to find expired authorizations", zap.Error(err))
		return err
	}

	if len(expiredTxns) == 0 {
		logger.Log.Info("No expired authorizations found")
		return nil
	}

	voidedCount := 0
	for _, txn := range expiredTxns {
		// Mark as voided
		if err := s.txnRepo.MarkVoided(txn.ID); err != nil {
			logger.Log.Error("Failed to auto-void transaction",
				zap.Error(err),
				zap.String("transaction_id", txn.ID.String()),
			)
			continue
		}

		// Log event
		s.txnRepo.CreateEvent(&model.TransactionEvent{
			TransactionID: txn.ID,
			EventType:     "auto_voided",
			OldStatus:     model.TransactionStatusAuthorized,
			NewStatus:     model.TransactionStatusVoided,
			Amount:        txn.Amount,
			Metadata:      sql.NullString{String: `{"reason":"Authorization expired after 7 days"}`, Valid: true},
		})

		voidedCount++

		logger.Log.Info("Authorization auto-voided",
			zap.String("transaction_id", txn.ID.String()),
			zap.String("merchant_id", txn.MerchantID.String()),
		)
	}

	logger.Log.Info("Expired authorizations auto-voided",
		zap.Int("count", voidedCount),
	)

	return nil
}

// =========================================================================
// Helper Methods
// =========================================================================

func (s *SettlementService) groupTransactionsByMerchant(txns []model.Transaction) map[uuid.UUID][]model.Transaction {
	grouped := make(map[uuid.UUID][]model.Transaction)

	for _, txn := range txns {
		grouped[txn.MerchantID] = append(grouped[txn.MerchantID], txn)
	}

	return grouped
}

// GetMerchantSettlements retrieves settlement history for a merchant
func (s *SettlementService) GetMerchantSettlements(merchantID uuid.UUID) ([]model.SettlementBatch, error) {
	// This would be implemented in the repository
	return nil, nil
}

// GetSettlementByID retrieves a specific settlement batch
func (s *SettlementService) GetSettlementByID(batchID uuid.UUID) (*model.SettlementBatch, error) {
	return s.settlementRepo.FindByID(batchID)
}
