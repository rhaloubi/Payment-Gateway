package service

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/client"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/repository"
)

// MerchantService handles merchant business logic
type MerchantService struct {
	merchantRepo     *repository.MerchantRepository
	settingsRepo     *repository.SettingsRepository
	businessInfoRepo *repository.BusinessInfoRepository
	brandingRepo     *repository.BrandingRepository
	verificationRepo *repository.VerificationRepository
	activityLogRepo  *repository.ActivityLogRepository
	authClient       *client.AuthServiceClient // NEW: Add auth client

}

// NewMerchantService creates a new merchant service
func NewMerchantService() *MerchantService {
	return &MerchantService{
		merchantRepo:     repository.NewMerchantRepository(),
		settingsRepo:     repository.NewSettingsRepository(),
		businessInfoRepo: repository.NewBusinessInfoRepository(),
		brandingRepo:     repository.NewBrandingRepository(),
		verificationRepo: repository.NewVerificationRepository(),
		activityLogRepo:  repository.NewActivityLogRepository(),
		authClient:       client.NewAuthServiceClient(), // NEW: Initialize auth client
	}
}

// CreateMerchantRequest represents merchant creation data
type CreateMerchantRequest struct {
	OwnerID      uuid.UUID
	BusinessName string
	LegalName    string
	Email        string
	Phone        string
	Website      string
	BusinessType model.BusinessType
}

// CreateMerchant creates a new merchant account
func (s *MerchantService) CreateMerchant(req *CreateMerchantRequest) (*model.Merchant, error) {
	// Validate input
	if err := s.validateMerchantCreation(req); err != nil {
		return nil, err
	}

	// Create merchant
	merchant := &model.Merchant{
		OwnerID:      req.OwnerID,
		BusinessName: req.BusinessName,
		Email:        req.Email,
		BusinessType: req.BusinessType,
		Status:       model.MerchantStatusPendingReview,
		CountryCode:  "MA", // Always Morocco
		CurrencyCode: "MAD",
		Timezone:     "Africa/Casablanca",
	}

	if req.LegalName != "" {
		merchant.LegalName = toNullString(req.LegalName)
	}
	if req.Phone != "" {
		merchant.Phone = toNullString(req.Phone)
	}
	if req.Website != "" {
		merchant.Website = toNullString(req.Website)
	}

	if err := s.merchantRepo.Create(merchant); err != nil {
		return nil, err
	}

	// Create default settings
	if err := s.createDefaultSettings(merchant.ID); err != nil {
		return nil, err
	}

	// Create default verification record
	if err := s.createDefaultVerification(merchant.ID); err != nil {
		return nil, err
	}
	if err := s.authClient.AssignMerchantOwnerRole(req.OwnerID, merchant.ID); err != nil {
		fmt.Printf("WARNING: Failed to assign admin role to merchant owner: %v\n", err)
	}

	// Log activity
	s.logActivity(merchant.ID, req.OwnerID, "merchant_created", "", merchant.ID, nil)

	return merchant, nil
}

// GetMerchantByID gets a merchant by ID
func (s *MerchantService) GetMerchantByID(id uuid.UUID) (*model.Merchant, error) {
	return s.merchantRepo.FindByID(id)
}

// GetMerchantByCode gets a merchant by code
func (s *MerchantService) GetMerchantByCode(code string) (*model.Merchant, error) {
	return s.merchantRepo.FindByCode(code)
}

// GetUserMerchants gets all merchants for a user (owned + team member)
func (s *MerchantService) GetUserMerchants(userID uuid.UUID) ([]model.Merchant, error) {
	// Get owned merchants
	ownedMerchants, err := s.merchantRepo.FindByOwnerID(userID)
	if err != nil {
		return nil, err
	}

	// Get merchants where user is team member
	teamMerchants, err := s.merchantRepo.FindMerchantsWhereUserIsTeamMember(userID)
	if err != nil {
		return nil, err
	}

	// Combine and deduplicate
	merchantMap := make(map[uuid.UUID]model.Merchant)
	for _, m := range ownedMerchants {
		merchantMap[m.ID] = m
	}
	for _, m := range teamMerchants {
		merchantMap[m.ID] = m
	}

	merchants := make([]model.Merchant, 0, len(merchantMap))
	for _, m := range merchantMap {
		merchants = append(merchants, m)
	}

	return merchants, nil
}

// UpdateMerchant updates merchant information
func (s *MerchantService) UpdateMerchant(id uuid.UUID, updates map[string]interface{}) error {
	merchant, err := s.merchantRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Track changes for audit log
	changes := make(map[string]interface{})

	// Update allowed fields
	if businessName, ok := updates["business_name"].(string); ok && businessName != "" {
		changes["business_name"] = map[string]interface{}{
			"old": merchant.BusinessName,
			"new": businessName,
		}
		merchant.BusinessName = businessName
	}

	if email, ok := updates["email"].(string); ok && email != "" {
		changes["email"] = map[string]interface{}{
			"old": merchant.Email,
			"new": email,
		}
		merchant.Email = email
	}

	if phone, ok := updates["phone"].(string); ok {
		changes["phone"] = map[string]interface{}{
			"old": merchant.Phone.String,
			"new": phone,
		}
		merchant.Phone = toNullString(phone)
	}

	if website, ok := updates["website"].(string); ok {
		changes["website"] = map[string]interface{}{
			"old": merchant.Website.String,
			"new": website,
		}
		merchant.Website = toNullString(website)
	}

	if err := s.merchantRepo.Update(merchant); err != nil {
		return err
	}

	// Log activity
	if userID, ok := updates["_user_id"].(uuid.UUID); ok {
		s.logActivity(merchant.ID, userID, "merchant_updated", "merchant", id, changes)
	}

	return nil
}

// UpdateMerchantStatus updates merchant status
func (s *MerchantService) UpdateMerchantStatus(id uuid.UUID, status model.MerchantStatus, userID uuid.UUID) error {
	merchant, err := s.merchantRepo.FindByID(id)
	if err != nil {
		return err
	}

	oldStatus := merchant.Status

	if err := s.merchantRepo.UpdateStatus(id, status); err != nil {
		return err
	}

	// Log activity
	changes := map[string]interface{}{
		"status": map[string]interface{}{
			"old": oldStatus,
			"new": status,
		},
	}
	s.logActivity(merchant.ID, userID, "merchant_status_changed", "merchant", id, changes)

	return nil
}

// DeleteMerchant soft deletes a merchant
func (s *MerchantService) DeleteMerchant(id uuid.UUID, userID uuid.UUID) error {
	merchant, err := s.merchantRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Check if user is owner
	if merchant.OwnerID != userID {
		return errors.New("only the owner can delete a merchant")
	}

	if err := s.merchantRepo.Delete(id); err != nil {
		return err
	}

	// Log activity
	s.logActivity(merchant.ID, userID, "merchant_deleted", "merchant", id, nil)

	return nil
}

// GetMerchantWithDetails gets merchant with all related data
func (s *MerchantService) GetMerchantWithDetails(id uuid.UUID) (*model.Merchant, error) {
	return s.merchantRepo.GetMerchantWithRelations(id)
}

// validateMerchantCreation validates merchant creation input
func (s *MerchantService) validateMerchantCreation(req *CreateMerchantRequest) error {
	if req.BusinessName == "" {
		return errors.New("business name is required")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.BusinessType == "" {
		req.BusinessType = model.BusinessTypeIndividual
	}
	return nil
}

// createDefaultSettings creates default settings for a new merchant
func (s *MerchantService) createDefaultSettings(merchantID uuid.UUID) error {
	settings := &model.MerchantSettings{
		MerchantID:        merchantID,
		DefaultCurrency:   "MAD",
		AutoSettle:        true,
		SettleSchedule:    "daily",
		SendEmailReceipts: true,
	}

	// Default payment methods and currencies (as JSON)
	settings.PaymentMethods = []byte(`["card"]`)
	settings.Currencies = []byte(`["MAD","USD","EUR"]`)

	return s.settingsRepo.Create(settings)
}

// createDefaultVerification creates default verification record
func (s *MerchantService) createDefaultVerification(merchantID uuid.UUID) error {
	verification := &model.MerchantVerification{
		MerchantID:         merchantID,
		VerificationStatus: model.VerificationStatusUnverified,
		RiskLevel:          model.RiskLevelMedium,
		CanProcessLive:     false,
		DocumentsRequired:  true,
	}

	return s.verificationRepo.Create(verification)
}

// logActivity logs merchant activity
func (s *MerchantService) logActivity(merchantID, userID uuid.UUID, action, resourceType string, resourceID uuid.UUID, changes map[string]interface{}) {
	log := &model.MerchantActivityLog{
		MerchantID:   merchantID,
		UserID:       userID,
		Action:       action,
		ResourceType: toNullString(resourceType),
		ResourceID:   toNullString(resourceID.String()),
	}

	if changes != nil {
		changesJSON, _ := json.Marshal(changes)
		log.Changes = changesJSON
	}

	s.activityLogRepo.Create(log)
}
