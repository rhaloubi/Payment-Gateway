package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/payment-api-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentRepository struct {
	db  *gorm.DB
	ctx context.Context
}

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{
		db:  inits.DB,
		ctx: context.Background(),
	}
}

func (r *PaymentRepository) Create(payment *model.Payment) error {
	if err := r.db.Create(payment).Error; err != nil {
		logger.Log.Error("Failed to create payment", zap.Error(err))
		return err
	}

	// Cache payment in Redis
	r.cachePayment(payment)

	return nil
}

func (r *PaymentRepository) CreateEvent(event *model.PaymentEvent) error {
	if err := r.db.Create(event).Error; err != nil {
		logger.Log.Error("Failed to create payment event", zap.Error(err))
		return err
	}
	return nil
}

// =========================================================================
// Read Operations
// =========================================================================

func (r *PaymentRepository) FindByID(id uuid.UUID) (*model.Payment, error) {
	// Try cache first
	if cached := r.getCachedPayment(id); cached != nil {
		return cached, nil
	}

	var payment model.Payment
	if err := r.db.Where("id = ?", id).First(&payment).Error; err != nil {
		return nil, err
	}

	// Cache for future requests
	r.cachePayment(&payment)

	return &payment, nil
}

func (r *PaymentRepository) FindByIDAndMerchant(id, merchantID uuid.UUID) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("id = ? AND merchant_id = ?", id, merchantID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) FindByIdempotencyKey(merchantID uuid.UUID, key string) (*model.Payment, error) {
	var payment model.Payment
	if err := r.db.Where("merchant_id = ? AND idempotency_key = ?", merchantID, key).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *PaymentRepository) FindByMerchant(merchantID uuid.UUID, limit, offset int) ([]model.Payment, error) {
	var payments []model.Payment
	if err := r.db.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *PaymentRepository) FindByStatus(merchantID uuid.UUID, status model.PaymentStatus, limit int) ([]model.Payment, error) {
	var payments []model.Payment
	if err := r.db.Where("merchant_id = ? AND status = ?", merchantID, status).
		Order("created_at DESC").
		Limit(limit).
		Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *PaymentRepository) GetPaymentEvents(paymentID uuid.UUID) ([]model.PaymentEvent, error) {
	var events []model.PaymentEvent
	if err := r.db.Where("payment_id = ?", paymentID).
		Order("created_at ASC").
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// =========================================================================
// Update Operations
// =========================================================================

func (r *PaymentRepository) Update(payment *model.Payment) error {
	if err := r.db.Save(payment).Error; err != nil {
		logger.Log.Error("Failed to update payment", zap.Error(err))
		return err
	}

	// Invalidate cache
	r.invalidateCache(payment.ID)

	return nil
}

func (r *PaymentRepository) UpdateStatus(id uuid.UUID, status model.PaymentStatus) error {
	if err := r.db.Model(&model.Payment{}).
		Where("id = ?", id).
		Update("status", status).
		Update("updated_at", time.Now()).
		Error; err != nil {
		return err
	}

	r.invalidateCache(id)
	return nil
}

func (r *PaymentRepository) MarkCaptured(id uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&model.Payment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.PaymentStatusCaptured,
			"captured_at": now,
			"updated_at":  now,
		}).Error; err != nil {
		return err
	}

	r.invalidateCache(id)
	return nil
}

func (r *PaymentRepository) MarkVoided(id uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&model.Payment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     model.PaymentStatusVoided,
			"voided_at":  now,
			"updated_at": now,
		}).Error; err != nil {
		return err
	}

	r.invalidateCache(id)
	return nil
}

func (r *PaymentRepository) MarkRefunded(id uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&model.Payment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.PaymentStatusRefunded,
			"refunded_at": now,
			"updated_at":  now,
		}).Error; err != nil {
		return err
	}

	r.invalidateCache(id)
	return nil
}

// =========================================================================
// Statistics & Analytics
// =========================================================================

type PaymentStatistics struct {
	TotalPayments     int64
	TotalAmount       int64
	AuthorizedAmount  int64
	CapturedAmount    int64
	RefundedAmount    int64
	SuccessRate       float64
	AverageFraudScore float64
}

func (r *PaymentRepository) GetStatistics(merchantID uuid.UUID, startDate, endDate time.Time) (*PaymentStatistics, error) {
	stats := &PaymentStatistics{}

	// Total payments
	r.db.Model(&model.Payment{}).
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?", merchantID, startDate, endDate).
		Count(&stats.TotalPayments)

	// Total amount
	r.db.Model(&model.Payment{}).
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?", merchantID, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.TotalAmount)

	// Authorized amount
	r.db.Model(&model.Payment{}).
		Where("merchant_id = ? AND status = ? AND created_at BETWEEN ? AND ?",
			merchantID, model.PaymentStatusAuthorized, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.AuthorizedAmount)

	// Captured amount
	r.db.Model(&model.Payment{}).
		Where("merchant_id = ? AND status = ? AND created_at BETWEEN ? AND ?",
			merchantID, model.PaymentStatusCaptured, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.CapturedAmount)

	// Refunded amount
	r.db.Model(&model.Payment{}).
		Where("merchant_id = ? AND status = ? AND created_at BETWEEN ? AND ?",
			merchantID, model.PaymentStatusRefunded, startDate, endDate).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&stats.RefundedAmount)

	// Success rate
	var successCount int64
	r.db.Model(&model.Payment{}).
		Where("merchant_id = ? AND status IN ? AND created_at BETWEEN ? AND ?",
			merchantID, []model.PaymentStatus{model.PaymentStatusAuthorized, model.PaymentStatusCaptured},
			startDate, endDate).
		Count(&successCount)

	if stats.TotalPayments > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalPayments) * 100
	}

	// Average fraud score
	r.db.Model(&model.Payment{}).
		Where("merchant_id = ? AND created_at BETWEEN ? AND ?", merchantID, startDate, endDate).
		Select("COALESCE(AVG(fraud_score), 0)").
		Scan(&stats.AverageFraudScore)

	return stats, nil
}

// =========================================================================
// Cache Operations (Redis)
// =========================================================================

func (r *PaymentRepository) cachePayment(payment *model.Payment) {
	key := fmt.Sprintf("payment:%s", payment.ID.String())
	data, _ := json.Marshal(payment)
	inits.RDB.Set(r.ctx, key, data, 15*time.Minute)
}

func (r *PaymentRepository) getCachedPayment(id uuid.UUID) *model.Payment {
	key := fmt.Sprintf("payment:%s", id.String())
	data, err := inits.RDB.Get(r.ctx, key).Result()
	if err != nil {
		return nil
	}

	var payment model.Payment
	if err := json.Unmarshal([]byte(data), &payment); err != nil {
		return nil
	}

	return &payment
}

func (r *PaymentRepository) invalidateCache(id uuid.UUID) {
	key := fmt.Sprintf("payment:%s", id.String())
	inits.RDB.Del(r.ctx, key)
}
