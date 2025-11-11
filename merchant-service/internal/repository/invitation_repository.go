package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"gorm.io/gorm"
)

type InvitationRepository struct{}

// NewInvitationRepository creates a new invitation repository
func NewInvitationRepository() *InvitationRepository {
	return &InvitationRepository{}
}

// Create creates a new invitation
func (r *InvitationRepository) Create(invitation *model.MerchantInvitation) error {
	return inits.DB.Create(invitation).Error
}

// FindByID finds an invitation by ID
func (r *InvitationRepository) FindByID(id uuid.UUID) (*model.MerchantInvitation, error) {
	var invitation model.MerchantInvitation
	err := inits.DB.Where("id = ?", id).First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return &invitation, nil
}

// FindByToken finds an invitation by token
func (r *InvitationRepository) FindByToken(token string) (*model.MerchantInvitation, error) {
	var invitation model.MerchantInvitation
	err := inits.DB.Where("invitation_token = ?", token).First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invitation not found")
		}
		return nil, err
	}
	return &invitation, nil
}

// FindPendingByEmail finds pending invitations for an email
func (r *InvitationRepository) FindPendingByEmail(email string) ([]model.MerchantInvitation, error) {
	var invitations []model.MerchantInvitation
	err := inits.DB.Where("email = ? AND status = ?", email, model.InvitationStatusPending).
		Order("created_at DESC").
		Find(&invitations).Error

	return invitations, err
}

// FindByMerchant finds all invitations for a merchant
func (r *InvitationRepository) FindByMerchant(merchantID uuid.UUID) ([]model.MerchantInvitation, error) {
	var invitations []model.MerchantInvitation
	err := inits.DB.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Find(&invitations).Error

	return invitations, err
}

// Update updates an invitation
func (r *InvitationRepository) Update(invitation *model.MerchantInvitation) error {
	return inits.DB.Save(invitation).Error
}

// MarkAsAccepted marks an invitation as accepted
func (r *InvitationRepository) MarkAsAccepted(id uuid.UUID) error {
	now := time.Now()
	return inits.DB.Model(&model.MerchantInvitation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.InvitationStatusAccepted,
			"accepted_at": now,
		}).Error
}

// MarkAsExpired marks expired invitations as expired
func (r *InvitationRepository) MarkAsExpired(merchantID uuid.UUID) error {
	return inits.DB.Model(&model.MerchantInvitation{}).
		Where("merchant_id = ? AND status = ? AND expires_at < ?",
			merchantID, model.InvitationStatusPending, time.Now()).
		Update("status", model.InvitationStatusExpired).Error
}

// Cancel cancels an invitation
func (r *InvitationRepository) Cancel(id uuid.UUID) error {
	return inits.DB.Model(&model.MerchantInvitation{}).
		Where("id = ?", id).
		Update("status", model.InvitationStatusCancelled).Error
}

// ExistsPendingForEmail checks if there's already a pending invitation
func (r *InvitationRepository) ExistsPendingForEmail(merchantID uuid.UUID, email string) (bool, error) {
	var count int64
	err := inits.DB.Model(&model.MerchantInvitation{}).
		Where("merchant_id = ? AND email = ? AND status = ?",
			merchantID, email, model.InvitationStatusPending).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
