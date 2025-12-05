package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"gorm.io/gorm"
)

type SettlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository() *SettlementRepository {
	return &SettlementRepository{db: inits.DB}
}

func (r *SettlementRepository) Create(batch *model.SettlementBatch) error {
	return r.db.Create(batch).Error
}

func (r *SettlementRepository) FindByID(id uuid.UUID) (*model.SettlementBatch, error) {
	var batch model.SettlementBatch
	if err := r.db.Where("id = ?", id).First(&batch).Error; err != nil {
		return nil, err
	}
	return &batch, nil
}

func (r *SettlementRepository) FindByMerchantAndDate(merchantID uuid.UUID, date time.Time) (*model.SettlementBatch, error) {
	var batch model.SettlementBatch
	if err := r.db.Where("merchant_id = ? AND batch_date = ?", merchantID, date).First(&batch).Error; err != nil {
		return nil, err
	}
	return &batch, nil
}

func (r *SettlementRepository) FindPendingBatches() ([]model.SettlementBatch, error) {
	var batches []model.SettlementBatch
	if err := r.db.Where("status = ? AND settlement_date <= ?",
		model.SettlementStatusPending,
		time.Now()).
		Find(&batches).Error; err != nil {
		return nil, err
	}
	return batches, nil
}

func (r *SettlementRepository) Update(batch *model.SettlementBatch) error {
	return r.db.Save(batch).Error
}

func (r *SettlementRepository) MarkSettled(id uuid.UUID) error {
	return r.db.Model(&model.SettlementBatch{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     model.SettlementStatusSettled,
			"settled_at": time.Now(),
		}).Error
}
