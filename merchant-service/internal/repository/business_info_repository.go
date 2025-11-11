package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"gorm.io/gorm"
)

type BusinessInfoRepository struct{}

// NewBusinessInfoRepository creates a new business info repository
func NewBusinessInfoRepository() *BusinessInfoRepository {
	return &BusinessInfoRepository{}
}

// Create creates merchant business info
func (r *BusinessInfoRepository) Create(info *model.MerchantBusinessInfo) error {
	return inits.DB.Create(info).Error
}

// FindByMerchantID finds business info by merchant ID
func (r *BusinessInfoRepository) FindByMerchantID(merchantID uuid.UUID) (*model.MerchantBusinessInfo, error) {
	var info model.MerchantBusinessInfo
	err := inits.DB.Where("merchant_id = ?", merchantID).First(&info).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("business info not found")
		}
		return nil, err
	}
	return &info, nil
}

// Update updates merchant business info
func (r *BusinessInfoRepository) Update(info *model.MerchantBusinessInfo) error {
	return inits.DB.Save(info).Error
}
