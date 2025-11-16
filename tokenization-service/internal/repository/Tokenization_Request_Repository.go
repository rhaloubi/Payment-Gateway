package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"gorm.io/gorm"
)

type TokenizationRequestRepository struct{}

func NewTokenizationRequestRepository() *TokenizationRequestRepository {
	return &TokenizationRequestRepository{}
}

func (r *TokenizationRequestRepository) Create(request *model.TokenizationRequest) error {
	return inits.DB.Create(request).Error
}

func (r *TokenizationRequestRepository) FindByID(id uuid.UUID) (*model.TokenizationRequest, error) {
	var request model.TokenizationRequest
	err := inits.DB.Where("id = ?", id).First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tokenization request not found")
		}
		return nil, err
	}
	return &request, nil
}

func (r *TokenizationRequestRepository) FindByRequestID(requestID string) (*model.TokenizationRequest, error) {
	var request model.TokenizationRequest
	err := inits.DB.Where("request_id = ?", requestID).First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tokenization request not found")
		}
		return nil, err
	}
	return &request, nil
}

func (r *TokenizationRequestRepository) FindByToken(tokenID uuid.UUID) ([]model.TokenizationRequest, error) {
	var requests []model.TokenizationRequest
	err := inits.DB.Where("token_id = ?", tokenID).
		Order("created_at DESC").
		Find(&requests).Error

	return requests, err
}

func (r *TokenizationRequestRepository) FindByMerchant(merchantID uuid.UUID, limit int, offset int) ([]model.TokenizationRequest, error) {
	var requests []model.TokenizationRequest
	err := inits.DB.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&requests).Error

	return requests, err
}

func (r *TokenizationRequestRepository) FindFailedRequests(merchantID uuid.UUID, since time.Time) ([]model.TokenizationRequest, error) {
	var requests []model.TokenizationRequest
	err := inits.DB.Where("merchant_id = ? AND success = ? AND created_at > ?",
		merchantID, false, since).
		Order("created_at DESC").
		Find(&requests).Error

	return requests, err
}

func (r *TokenizationRequestRepository) CountByMerchant(merchantID uuid.UUID) (int64, error) {
	var count int64
	err := inits.DB.Model(&model.TokenizationRequest{}).
		Where("merchant_id = ?", merchantID).
		Count(&count).Error

	return count, err
}
