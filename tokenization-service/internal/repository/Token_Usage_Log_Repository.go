package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"gorm.io/gorm"
)

type TokenUsageLogRepository struct{}

func NewTokenUsageLogRepository() *TokenUsageLogRepository {
	return &TokenUsageLogRepository{}
}

func (r *TokenUsageLogRepository) Create(log *model.TokenUsageLog) error {
	return inits.DB.Create(log).Error
}

func (r *TokenUsageLogRepository) FindByToken(tokenID uuid.UUID) ([]model.TokenUsageLog, error) {
	var logs []model.TokenUsageLog
	err := inits.DB.Where("token_id = ?", tokenID).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, err
}

func (r *TokenUsageLogRepository) FindByTransaction(transactionID uuid.UUID) (*model.TokenUsageLog, error) {
	var log model.TokenUsageLog
	err := inits.DB.Where("transaction_id = ?", transactionID).First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token usage log not found")
		}
		return nil, err
	}
	return &log, nil
}

func (r *TokenUsageLogRepository) FindByMerchant(merchantID uuid.UUID, limit int, offset int) ([]model.TokenUsageLog, error) {
	var logs []model.TokenUsageLog
	err := inits.DB.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	return logs, err
}

func (r *TokenUsageLogRepository) CountByToken(tokenID uuid.UUID) (int64, error) {
	var count int64
	err := inits.DB.Model(&model.TokenUsageLog{}).
		Where("token_id = ?", tokenID).
		Count(&count).Error

	return count, err
}

func (r *TokenUsageLogRepository) FindFailedUsages(tokenID uuid.UUID) ([]model.TokenUsageLog, error) {
	var logs []model.TokenUsageLog
	err := inits.DB.Where("token_id = ? AND success = ?", tokenID, false).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, err
}
