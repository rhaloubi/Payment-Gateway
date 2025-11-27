package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/payment-api-service/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WebhookRepository struct {
	db  *gorm.DB
	ctx context.Context
}

func NewWebhookRepository() *WebhookRepository {
	return &WebhookRepository{
		db:  inits.DB,
		ctx: context.Background(),
	}
}

// Create creates a new webhook delivery record
func (r *WebhookRepository) Create(webhook *model.WebhookDelivery) error {
	if err := r.db.Create(webhook).Error; err != nil {
		logger.Log.Error("Failed to create webhook delivery", zap.Error(err))
		return err
	}
	return nil
}

// MarkDelivered marks webhook as successfully delivered
func (r *WebhookRepository) MarkDelivered(id uuid.UUID, statusCode int, response string) error {
	now := time.Now()
	if err := r.db.Model(&model.WebhookDelivery{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"success":      true,
			"status_code":  statusCode,
			"response":     response,
			"delivered_at": now,
		}).Error; err != nil {
		return err
	}
	return nil
}

// MarkFailed marks webhook delivery as failed and schedules retry
func (r *WebhookRepository) MarkFailed(id uuid.UUID, statusCode int, response string) error {
	var webhook model.WebhookDelivery
	if err := r.db.First(&webhook, id).Error; err != nil {
		return err
	}

	// Increment attempt count
	webhook.AttemptCount++
	webhook.StatusCode = statusCode
	webhook.Response.String = response
	webhook.Response.Valid = true

	// Schedule next retry (exponential backoff)
	// 1st retry: 5 min, 2nd: 15 min, 3rd: 1 hour, 4th: 6 hours
	retryDelays := []time.Duration{
		5 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
		6 * time.Hour,
	}

	if webhook.AttemptCount <= len(retryDelays) {
		nextRetry := time.Now().Add(retryDelays[webhook.AttemptCount-1])
		webhook.NextRetryAt.Time = nextRetry
		webhook.NextRetryAt.Valid = true
	}

	if err := r.db.Save(&webhook).Error; err != nil {
		return err
	}

	return nil
}

// FindPendingRetries finds webhooks that need to be retried
func (r *WebhookRepository) FindPendingRetries() ([]model.WebhookDelivery, error) {
	var webhooks []model.WebhookDelivery
	if err := r.db.Where("success = ? AND next_retry_at <= ? AND attempt_count < ?",
		false, time.Now(), 5).
		Find(&webhooks).Error; err != nil {
		return nil, err
	}
	return webhooks, nil
}

// FindByPayment finds all webhook deliveries for a payment
func (r *WebhookRepository) FindByPayment(paymentID uuid.UUID) ([]model.WebhookDelivery, error) {
	var webhooks []model.WebhookDelivery
	if err := r.db.Where("payment_id = ?", paymentID).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return nil, err
	}
	return webhooks, nil
}
