package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"gorm.io/gorm"
)

type ChargebackRepository struct {
	db *gorm.DB
}

func NewChargebackRepository() *ChargebackRepository {
	return &ChargebackRepository{db: inits.DB}
}

func (r *ChargebackRepository) Create(chargeback *model.Chargeback) error {
	return r.db.Create(chargeback).Error
}

func (r *ChargebackRepository) CreateEvent(event *model.ChargebackEvent) error {
	return r.db.Create(event).Error
}

func (r *ChargebackRepository) FindByID(id uuid.UUID) (*model.Chargeback, error) {
	var chargeback model.Chargeback
	if err := r.db.Where("id = ?", id).First(&chargeback).Error; err != nil {
		return nil, err
	}
	return &chargeback, nil
}

func (r *ChargebackRepository) FindByTransaction(txnID uuid.UUID) ([]model.Chargeback, error) {
	var chargebacks []model.Chargeback
	if err := r.db.Where("transaction_id = ?", txnID).Find(&chargebacks).Error; err != nil {
		return nil, err
	}
	return chargebacks, nil
}

func (r *ChargebackRepository) FindByMerchant(merchantID uuid.UUID) ([]model.Chargeback, error) {
	var chargebacks []model.Chargeback
	if err := r.db.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&chargebacks).Error; err != nil {
		return nil, err
	}
	return chargebacks, nil
}

func (r *ChargebackRepository) FindNeedingResponse() ([]model.Chargeback, error) {
	var chargebacks []model.Chargeback
	if err := r.db.Where("status = ? AND response_due_date > ?",
		model.ChargebackStatusNeedsResponse,
		time.Now()).
		Find(&chargebacks).Error; err != nil {
		return nil, err
	}
	return chargebacks, nil
}

func (r *ChargebackRepository) Update(chargeback *model.Chargeback) error {
	return r.db.Save(chargeback).Error
}

func (r *ChargebackRepository) UpdateStatus(id uuid.UUID, status model.ChargebackStatus) error {
	return r.db.Model(&model.Chargeback{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}
