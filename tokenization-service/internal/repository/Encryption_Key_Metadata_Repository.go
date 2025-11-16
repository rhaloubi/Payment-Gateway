package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"gorm.io/gorm"
)

type EncryptionKeyRepository struct{}

func NewEncryptionKeyRepository() *EncryptionKeyRepository {
	return &EncryptionKeyRepository{}
}

func (r *EncryptionKeyRepository) Create(keyMetadata *model.EncryptionKeyMetadata) error {
	return inits.DB.Create(keyMetadata).Error
}

func (r *EncryptionKeyRepository) FindByKeyID(keyID string) (*model.EncryptionKeyMetadata, error) {
	var keyMetadata model.EncryptionKeyMetadata
	err := inits.DB.Where("key_id = ? AND deleted_at IS NULL", keyID).First(&keyMetadata).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("encryption key not found")
		}
		return nil, err
	}
	return &keyMetadata, nil
}

func (r *EncryptionKeyRepository) FindActiveByMerchant(merchantID uuid.UUID) (*model.EncryptionKeyMetadata, error) {
	var keyMetadata model.EncryptionKeyMetadata
	err := inits.DB.Where("merchant_id = ? AND is_active = ? AND deleted_at IS NULL",
		merchantID, true).
		Order("created_at DESC").
		First(&keyMetadata).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active encryption key found for merchant")
		}
		return nil, err
	}
	return &keyMetadata, nil
}

func (r *EncryptionKeyRepository) FindByMerchant(merchantID uuid.UUID) ([]model.EncryptionKeyMetadata, error) {
	var keys []model.EncryptionKeyMetadata
	err := inits.DB.Where("merchant_id = ? AND deleted_at IS NULL", merchantID).
		Order("key_version DESC").
		Find(&keys).Error

	return keys, err
}

func (r *EncryptionKeyRepository) Update(keyMetadata *model.EncryptionKeyMetadata) error {
	return inits.DB.Save(keyMetadata).Error
}

func (r *EncryptionKeyRepository) DeactivateKey(keyID string) error {
	return inits.DB.Model(&model.EncryptionKeyMetadata{}).
		Where("key_id = ?", keyID).
		Update("is_active", false).Error
}

func (r *EncryptionKeyRepository) IncrementEncryptedRecords(keyID string) error {
	return inits.DB.Model(&model.EncryptionKeyMetadata{}).
		Where("key_id = ?", keyID).
		Updates(map[string]interface{}{
			"encrypted_records": gorm.Expr("encrypted_records + 1"),
			"last_used_at":      time.Now(),
		}).Error
}

func (r *EncryptionKeyRepository) RevokeKey(keyID string, revokedBy uuid.UUID) error {
	return inits.DB.Model(&model.EncryptionKeyMetadata{}).
		Where("key_id = ?", keyID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"revoked_by": revokedBy,
			"revoked_at": time.Now(),
		}).Error
}

// CountByMerchant counts encryption keys for a merchant
func (r *EncryptionKeyRepository) CountByMerchant(merchantID uuid.UUID) (int64, error) {
	var count int64
	err := inits.DB.Model(&model.EncryptionKeyMetadata{}).
		Where("merchant_id = ? AND deleted_at IS NULL", merchantID).
		Count(&count).Error

	return count, err
}
