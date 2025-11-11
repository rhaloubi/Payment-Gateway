package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"gorm.io/gorm"
)

type VerificationRepository struct{}

// NewVerificationRepository creates a new verification repository
func NewVerificationRepository() *VerificationRepository {
	return &VerificationRepository{}
}

// Create creates merchant verification record
func (r *VerificationRepository) Create(verification *model.MerchantVerification) error {
	return inits.DB.Create(verification).Error
}

// FindByMerchantID finds verification by merchant ID
func (r *VerificationRepository) FindByMerchantID(merchantID uuid.UUID) (*model.MerchantVerification, error) {
	var verification model.MerchantVerification
	err := inits.DB.Where("merchant_id = ?", merchantID).First(&verification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("verification not found")
		}
		return nil, err
	}
	return &verification, nil
}

// Update updates merchant verification
func (r *VerificationRepository) Update(verification *model.MerchantVerification) error {
	return inits.DB.Save(verification).Error
}

// MarkAsVerified marks merchant as verified
func (r *VerificationRepository) MarkAsVerified(merchantID, verifiedBy uuid.UUID) error {
	now := time.Now()
	return inits.DB.Model(&model.MerchantVerification{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"verification_status": model.VerificationStatusVerified,
			"verified_at":         now,
			"verified_by":         verifiedBy.String(),
			"can_process_live":    true,
		}).Error
}

// MarkAsRejected marks merchant verification as rejected
func (r *VerificationRepository) MarkAsRejected(merchantID uuid.UUID, reason string) error {
	return inits.DB.Model(&model.MerchantVerification{}).
		Where("merchant_id = ?", merchantID).
		Updates(map[string]interface{}{
			"verification_status": model.VerificationStatusRejected,
			"rejection_reason":    reason,
		}).Error
}
