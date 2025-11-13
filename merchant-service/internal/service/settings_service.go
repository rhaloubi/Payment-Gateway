package service

import (
	"encoding/json"

	"github.com/google/uuid"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/repository"
)

type SettingsService struct {
	settingsRepo    *repository.SettingsRepository
	activityLogRepo *repository.ActivityLogRepository
}

// NewSettingsService creates a new settings service
func NewSettingsService() *SettingsService {
	return &SettingsService{
		settingsRepo:    repository.NewSettingsRepository(),
		activityLogRepo: repository.NewActivityLogRepository(),
	}
}

// GetSettings gets merchant settings
func (s *SettingsService) GetSettings(merchantID uuid.UUID) (*model.MerchantSettings, error) {
	return s.settingsRepo.FindByMerchantID(merchantID)
}

// UpdateSettings updates merchant settings
func (s *SettingsService) UpdateSettings(merchantID uuid.UUID, updates map[string]interface{}, userID uuid.UUID) error {
	settings, err := s.settingsRepo.FindByMerchantID(merchantID)
	if err != nil {
		return err
	}

	changes := make(map[string]interface{})

	// Update allowed fields
	if defaultCurrency, ok := updates["default_currency"].(string); ok {
		changes["default_currency"] = map[string]interface{}{
			"old": settings.DefaultCurrency,
			"new": defaultCurrency,
		}
		settings.DefaultCurrency = defaultCurrency
	}

	if autoSettle, ok := updates["auto_settle"].(bool); ok {
		changes["auto_settle"] = map[string]interface{}{
			"old": settings.AutoSettle,
			"new": autoSettle,
		}
		settings.AutoSettle = autoSettle
	}

	if settleSchedule, ok := updates["settle_schedule"].(string); ok {
		changes["settle_schedule"] = map[string]interface{}{
			"old": settings.SettleSchedule,
			"new": settleSchedule,
		}
		settings.SettleSchedule = settleSchedule
	}

	if webhookURL, ok := updates["webhook_url"].(string); ok {
		changes["webhook_url"] = map[string]interface{}{
			"old": settings.WebhookURL.String,
			"new": webhookURL,
		}
		settings.WebhookURL = toNullString(webhookURL)
	}

	if err := s.settingsRepo.Update(settings); err != nil {
		return err
	}

	// Log activity
	s.logActivity(merchantID, userID, "settings_updated", "merchant_settings", settings.ID, changes)

	return nil
}

// logActivity logs settings activity
func (s *SettingsService) logActivity(merchantID, userID uuid.UUID, action, resourceType string, resourceID uuid.UUID, changes map[string]interface{}) {
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

	activityLogRepo := repository.NewActivityLogRepository()
	activityLogRepo.Create(log)
}
