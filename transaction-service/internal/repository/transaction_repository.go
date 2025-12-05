package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db  *gorm.DB
	ctx context.Context
}

func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{
		db:  inits.DB,
		ctx: context.Background(),
	}
}

// =========================================================================
// Create Operations
// =========================================================================

func (r *TransactionRepository) Create(txn *model.Transaction) error {
	if err := r.db.Create(txn).Error; err != nil {
		logger.Log.Error("Failed to create transaction", zap.Error(err))
		return err
	}

	// Cache transaction
	r.cacheTransaction(txn)

	return nil
}

func (r *TransactionRepository) CreateEvent(event *model.TransactionEvent) error {
	if err := r.db.Create(event).Error; err != nil {
		logger.Log.Error("Failed to create transaction event", zap.Error(err))
		return err
	}
	return nil
}

func (r *TransactionRepository) CreateIssuerResponse(response *model.IssuerResponse) error {
	return r.db.Create(response).Error
}

// =========================================================================
// Read Operations
// =========================================================================

func (r *TransactionRepository) FindByID(id uuid.UUID) (*model.Transaction, error) {
	// Try cache first
	if cached := r.getCachedTransaction(id); cached != nil {
		return cached, nil
	}

	var txn model.Transaction
	if err := r.db.Where("id = ?", id).First(&txn).Error; err != nil {
		return nil, err
	}

	r.cacheTransaction(&txn)
	return &txn, nil
}

func (r *TransactionRepository) FindByIDAndMerchant(id, merchantID uuid.UUID) (*model.Transaction, error) {
	var txn model.Transaction
	if err := r.db.Where("id = ? AND merchant_id = ?", id, merchantID).First(&txn).Error; err != nil {
		return nil, err
	}
	return &txn, nil
}

func (r *TransactionRepository) FindByMerchant(merchantID uuid.UUID, limit, offset int) ([]model.Transaction, error) {
	var txns []model.Transaction
	if err := r.db.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&txns).Error; err != nil {
		return nil, err
	}
	return txns, nil
}

func (r *TransactionRepository) FindByStatus(merchantID uuid.UUID, status model.TransactionStatus) ([]model.Transaction, error) {
	var txns []model.Transaction
	if err := r.db.Where("merchant_id = ? AND status = ?", merchantID, status).
		Order("created_at DESC").
		Find(&txns).Error; err != nil {
		return nil, err
	}
	return txns, nil
}

// FindExpiredAuthorizations finds authorizations that have expired (> 7 days)
func (r *TransactionRepository) FindExpiredAuthorizations() ([]model.Transaction, error) {
	var txns []model.Transaction
	if err := r.db.Where("status = ? AND expires_at < ?",
		model.TransactionStatusAuthorized,
		time.Now()).
		Find(&txns).Error; err != nil {
		return nil, err
	}
	return txns, nil
}

// FindCapturedForSettlement finds captured transactions for settlement batch
func (r *TransactionRepository) FindCapturedForSettlement(batchDate time.Time) ([]model.Transaction, error) {
	startDate := batchDate.Truncate(24 * time.Hour)
	endDate := startDate.Add(24 * time.Hour)

	var txns []model.Transaction
	if err := r.db.Where("status = ? AND captured_at >= ? AND captured_at < ? AND settlement_batch_id IS NULL",
		model.TransactionStatusCaptured,
		startDate,
		endDate).
		Find(&txns).Error; err != nil {
		return nil, err
	}
	return txns, nil
}

func (r *TransactionRepository) GetTransactionEvents(txnID uuid.UUID) ([]model.TransactionEvent, error) {
	var events []model.TransactionEvent
	if err := r.db.Where("transaction_id = ?", txnID).
		Order("created_at ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// =========================================================================
// Update Operations
// =========================================================================

func (r *TransactionRepository) Update(txn *model.Transaction) error {
	if err := r.db.Save(txn).Error; err != nil {
		logger.Log.Error("Failed to update transaction", zap.Error(err))
		return err
	}

	r.invalidateCache(txn.ID)
	return nil
}

func (r *TransactionRepository) UpdateStatus(id uuid.UUID, status model.TransactionStatus) error {
	if err := r.db.Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}

	r.invalidateCache(id)
	return nil
}

func (r *TransactionRepository) MarkAuthorized(id uuid.UUID, authCode string) error {
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // Expires in 7 days

	if err := r.db.Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        model.TransactionStatusAuthorized,
			"authorized_at": now,
			"expires_at":    expiresAt,
			"auth_code":     authCode,
			"updated_at":    now,
		}).Error; err != nil {
		return err
	}

	r.invalidateCache(id)
	return nil
}

func (r *TransactionRepository) MarkCaptured(id uuid.UUID, amount int64) error {
	now := time.Now()
	if err := r.db.Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":          model.TransactionStatusCaptured,
			"captured_at":     now,
			"captured_amount": amount,
			"updated_at":      now,
		}).Error; err != nil {
		return err
	}

	r.invalidateCache(id)
	return nil
}

func (r *TransactionRepository) MarkVoided(id uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     model.TransactionStatusVoided,
			"voided_at":  now,
			"updated_at": now,
		}).Error; err != nil {
		return err
	}

	r.invalidateCache(id)
	return nil
}

func (r *TransactionRepository) AddRefundAmount(id uuid.UUID, refundAmount int64) error {
	// Get current transaction
	txn, err := r.FindByID(id)
	if err != nil {
		return err
	}

	newRefundedAmount := txn.RefundedAmount + refundAmount

	// Determine new status
	var newStatus model.TransactionStatus
	if newRefundedAmount >= txn.CapturedAmount {
		newStatus = model.TransactionStatusRefunded
	} else {
		newStatus = model.TransactionStatusPartiallyRefunded
	}

	now := time.Now()
	if err := r.db.Model(&model.Transaction{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"refunded_amount": newRefundedAmount,
			"status":          newStatus,
			"refunded_at":     now,
			"updated_at":      now,
		}).Error; err != nil {
		return err
	}

	r.invalidateCache(id)
	return nil
}

func (r *TransactionRepository) LinkToSettlementBatch(txnIDs []uuid.UUID, batchID uuid.UUID) error {
	if err := r.db.Model(&model.Transaction{}).
		Where("id IN ?", txnIDs).
		Updates(map[string]interface{}{
			"settlement_batch_id": batchID,
			"status":              model.TransactionStatusSettled,
			"settled_at":          time.Now(),
			"updated_at":          time.Now(),
		}).Error; err != nil {
		return err
	}

	// Invalidate cache for all transactions
	for _, id := range txnIDs {
		r.invalidateCache(id)
	}

	return nil
}

// =========================================================================
// Statistics
// =========================================================================

type TransactionStatistics struct {
	TotalTransactions int64
	TotalAmount       int64
	TotalAmountMAD    int64
	AuthorizedAmount  int64
	CapturedAmount    int64
	RefundedAmount    int64
	SettledAmount     int64
	AverageFraudScore float64
	SuccessRate       float64
}

func (r *TransactionRepository) GetStatistics(merchantID uuid.UUID, startDate, endDate time.Time) (*TransactionStatistics, error) {
	stats := &TransactionStatistics{}

	query := r.db.Model(&model.Transaction{}).
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?", merchantID, startDate, endDate)

	// Total transactions
	query.Count(&stats.TotalTransactions)

	// Total amount in original currency
	query.Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalAmount)

	// Total amount in MAD
	query.Select("COALESCE(SUM(amount_mad), 0)").Scan(&stats.TotalAmountMAD)

	// Authorized amount
	r.db.Model(&model.Transaction{}).
		Where("merchant_id = ? AND status = ? AND created_at BETWEEN ? AND ?",
			merchantID, model.TransactionStatusAuthorized, startDate, endDate).
		Select("COALESCE(SUM(amount_mad), 0)").
		Scan(&stats.AuthorizedAmount)

	// Captured amount
	r.db.Model(&model.Transaction{}).
		Where("merchant_id = ? AND status IN ? AND created_at BETWEEN ? AND ?",
			merchantID, []model.TransactionStatus{model.TransactionStatusCaptured, model.TransactionStatusSettled},
			startDate, endDate).
		Select("COALESCE(SUM(captured_amount), 0)").
		Scan(&stats.CapturedAmount)

	// Refunded amount
	r.db.Model(&model.Transaction{}).
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?",
			merchantID, startDate, endDate).
		Select("COALESCE(SUM(refunded_amount), 0)").
		Scan(&stats.RefundedAmount)

	// Average fraud score
	query.Select("COALESCE(AVG(fraud_score), 0)").Scan(&stats.AverageFraudScore)

	// Success rate
	var successCount int64
	r.db.Model(&model.Transaction{}).
		Where("merchant_id = ? AND status IN ? AND created_at BETWEEN ? AND ?",
			merchantID,
			[]model.TransactionStatus{model.TransactionStatusAuthorized, model.TransactionStatusCaptured, model.TransactionStatusSettled},
			startDate, endDate).
		Count(&successCount)

	if stats.TotalTransactions > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalTransactions) * 100
	}

	return stats, nil
}

// =========================================================================
// Cache Operations (Redis)
// =========================================================================

func (r *TransactionRepository) cacheTransaction(txn *model.Transaction) {
	key := fmt.Sprintf("transaction:%s", txn.ID.String())
	data, _ := json.Marshal(txn)
	inits.RDB.Set(r.ctx, key, data, 5*time.Minute)
}

func (r *TransactionRepository) getCachedTransaction(id uuid.UUID) *model.Transaction {
	key := fmt.Sprintf("transaction:%s", id.String())
	data, err := inits.RDB.Get(r.ctx, key).Result()
	if err != nil {
		return nil
	}

	var txn model.Transaction
	if err := json.Unmarshal([]byte(data), &txn); err != nil {
		return nil
	}

	return &txn
}

func (r *TransactionRepository) invalidateCache(id uuid.UUID) {
	key := fmt.Sprintf("transaction:%s", id.String())
	inits.RDB.Del(r.ctx, key)
}
