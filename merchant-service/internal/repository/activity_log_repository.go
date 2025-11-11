package repository

import (
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
)

type ActivityLogRepository struct{}

// NewActivityLogRepository creates a new activity log repository
func NewActivityLogRepository() *ActivityLogRepository {
	return &ActivityLogRepository{}
}

// Create creates an activity log entry
func (r *ActivityLogRepository) Create(log *model.MerchantActivityLog) error {
	return inits.DB.Create(log).Error
}

// FindByMerchant finds activity logs for a merchant
func (r *ActivityLogRepository) FindByMerchant(merchantID uuid.UUID, limit int) ([]model.MerchantActivityLog, error) {
	var logs []model.MerchantActivityLog
	err := inits.DB.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error

	return logs, err
}

// FindByUser finds activity logs for a user across all merchants
func (r *ActivityLogRepository) FindByUser(userID uuid.UUID, limit int) ([]model.MerchantActivityLog, error) {
	var logs []model.MerchantActivityLog
	err := inits.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error

	return logs, err
}
