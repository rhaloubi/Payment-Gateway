package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"gorm.io/gorm"
)

type BrandingRepository struct{}

// NewBrandingRepository creates a new branding repository
func NewBrandingRepository() *BrandingRepository {
	return &BrandingRepository{}
}

// Create creates merchant branding
func (r *BrandingRepository) Create(branding *model.MerchantBranding) error {
	return inits.DB.Create(branding).Error
}

// FindByMerchantID finds branding by merchant ID
func (r *BrandingRepository) FindByMerchantID(merchantID uuid.UUID) (*model.MerchantBranding, error) {
	var branding model.MerchantBranding
	err := inits.DB.Where("merchant_id = ?", merchantID).First(&branding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("branding not found")
		}
		return nil, err
	}
	return &branding, nil
}

// Update updates merchant branding
func (r *BrandingRepository) Update(branding *model.MerchantBranding) error {
	return inits.DB.Save(branding).Error
}
