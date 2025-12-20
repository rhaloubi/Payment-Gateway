package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/payment-api-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentIntentRepository struct {
	db  *gorm.DB
	ctx context.Context
}

func NewPaymentIntentRepository() *PaymentIntentRepository {
	return &PaymentIntentRepository{
		db:  inits.DB,
		ctx: context.Background(),
	}
}

// =========================================================================
// Create Operations
// =========================================================================

func (r *PaymentIntentRepository) Create(intent *model.PaymentIntent) error {
	if err := r.db.Create(intent).Error; err != nil {
		logger.Log.Error("Failed to create payment intent", zap.Error(err))
		return err
	}
	return nil
}

// =========================================================================
// Read Operations
// =========================================================================

func (r *PaymentIntentRepository) FindByID(id uuid.UUID) (*model.PaymentIntent, error) {
	var intent model.PaymentIntent
	if err := r.db.Where("id = ?", id).First(&intent).Error; err != nil {
		return nil, err
	}
	return &intent, nil
}

func (r *PaymentIntentRepository) FindByClientSecret(clientSecret string) (*model.PaymentIntent, error) {
	var intent model.PaymentIntent
	if err := r.db.Where("client_secret = ?", clientSecret).First(&intent).Error; err != nil {
		return nil, err
	}
	return &intent, nil
}

func (r *PaymentIntentRepository) FindByIDAndMerchant(id, merchantID uuid.UUID) (*model.PaymentIntent, error) {
	var intent model.PaymentIntent
	if err := r.db.Where("id = ? AND merchant_id = ?", id, merchantID).First(&intent).Error; err != nil {
		return nil, err
	}
	return &intent, nil
}

func (r *PaymentIntentRepository) FindByOrderID(merchantID uuid.UUID, orderID string) (*model.PaymentIntent, error) {
	var intent model.PaymentIntent
	if err := r.db.Where("merchant_id = ? AND order_id = ?", merchantID, orderID).
		Order("created_at DESC").
		First(&intent).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &intent, nil
}

// =========================================================================
// Update Operations
// =========================================================================

func (r *PaymentIntentRepository) Update(intent *model.PaymentIntent) error {
	if err := r.db.Save(intent).Error; err != nil {
		logger.Log.Error("Failed to update payment intent", zap.Error(err))
		return err
	}
	return nil
}

func (r *PaymentIntentRepository) UpdateStatus(id uuid.UUID, status model.PaymentIntentStatus) error {
	if err := r.db.Model(&model.PaymentIntent{}).
		Where("id = ?", id).
		Update("status", status).
		Update("updated_at", time.Now()).
		Error; err != nil {
		return err
	}
	return nil
}

func (r *PaymentIntentRepository) MarkConfirmed(id uuid.UUID, paymentID uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&model.PaymentIntent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.PaymentIntentStatusAuthorized,
			"payment_id":   paymentID.String(),
			"confirmed_at": now,
			"updated_at":   now,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (r *PaymentIntentRepository) MarkCanceled(id uuid.UUID) error {
	now := time.Now()
	if err := r.db.Model(&model.PaymentIntent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.PaymentIntentStatusCanceled,
			"canceled_at": now,
			"updated_at":  now,
		}).Error; err != nil {
		return err
	}
	return nil
}

// =========================================================================
// Expiration Management
// =========================================================================

func (r *PaymentIntentRepository) MarkExpired(id uuid.UUID) error {
	return r.db.Model(&model.PaymentIntent{}).
		Where("id = ?", id).
		Update("status", model.PaymentIntentStatusExpired).
		Update("updated_at", time.Now()).
		Error
}

func (r *PaymentIntentRepository) FindExpired() ([]model.PaymentIntent, error) {
	var intents []model.PaymentIntent
	if err := r.db.Where("status = ? AND expires_at < ?",
		model.PaymentIntentStatusAwaitingPayment,
		time.Now()).
		Find(&intents).Error; err != nil {
		return nil, err
	}
	return intents, nil
}

// =========================================================================
// List Operations
// =========================================================================

func (r *PaymentIntentRepository) FindByMerchant(merchantID uuid.UUID, limit, offset int) ([]model.PaymentIntent, error) {
	var intents []model.PaymentIntent
	if err := r.db.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&intents).Error; err != nil {
		return nil, err
	}
	return intents, nil
}

func (r *PaymentIntentRepository) CountByMerchant(merchantID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.Model(&model.PaymentIntent{}).
		Where("merchant_id = ?", merchantID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
