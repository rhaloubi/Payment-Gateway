package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/payment-api-service/internal/models"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/repository"
	"go.uber.org/zap"
)

type WebhookService struct {
	webhookRepo *repository.WebhookRepository
	httpClient  *http.Client
}

func NewWebhookService() *WebhookService {
	return &WebhookService{
		webhookRepo: repository.NewWebhookRepository(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WebhookPayload represents the webhook data sent to merchant
type WebhookPayload struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	ID        uuid.UUID              `json:"id"`
}

// SendPaymentWebhook sends a payment event webhook to merchant
func (s *WebhookService) SendPaymentWebhook(ctx context.Context, payment *model.Payment, eventType string, webhookURL string, webhookSecret string) error {

	// Build webhook payload
	payload := WebhookPayload{
		Event:     eventType,
		Timestamp: time.Now(),
		ID:        uuid.New(),
		Data: map[string]interface{}{
			"payment_id":     payment.ID,
			"merchant_id":    payment.MerchantID,
			"status":         payment.Status,
			"amount":         payment.Amount,
			"currency":       payment.Currency,
			"card_brand":     payment.CardBrand,
			"card_last4":     payment.CardLast4,
			"fraud_score":    payment.FraudScore,
			"fraud_decision": payment.FraudDecision,
			"created_at":     payment.CreatedAt,
		},
	}

	// Add optional fields
	if payment.AuthCode.Valid {
		payload.Data["auth_code"] = payment.AuthCode.String
	}
	if payment.ResponseCode.Valid {
		payload.Data["response_code"] = payment.ResponseCode.String
	}
	if payment.ResponseMsg.Valid {
		payload.Data["response_message"] = payment.ResponseMsg.String
	}
	if payment.TransactionID != uuid.Nil {
		payload.Data["transaction_id"] = payment.TransactionID
	}

	// Serialize payload
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		logger.Log.Error("Failed to serialize webhook payload", zap.Error(err))
		return err
	}

	// Create webhook delivery record
	webhookDelivery := &model.WebhookDelivery{
		PaymentID:  payment.ID,
		MerchantID: payment.MerchantID,
		EventType:  eventType,
		WebhookURL: webhookURL,
		Payload:    string(payloadJSON),
	}

	if err := s.webhookRepo.Create(webhookDelivery); err != nil {
		logger.Log.Error("Failed to create webhook delivery record", zap.Error(err))
		return err
	}

	// Send webhook asynchronously
	go s.deliverWebhook(webhookDelivery.ID, webhookURL, payloadJSON, webhookSecret)

	return nil
}

// deliverWebhook sends the actual HTTP request to merchant's webhook endpoint
func (s *WebhookService) deliverWebhook(
	webhookID uuid.UUID,
	url string,
	payload []byte,
	secret string,
) {
	logger.Log.Info("Delivering webhook",
		zap.String("webhook_id", webhookID.String()),
		zap.String("url", url),
	)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		logger.Log.Error("Failed to create webhook request", zap.Error(err))
		s.webhookRepo.MarkFailed(webhookID, 0, err.Error())
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PaymentGateway-Webhook/1.0")
	req.Header.Set("X-Webhook-Timestamp", time.Now().Format(time.RFC3339))

	// Generate HMAC signature
	if secret != "" {
		signature := s.generateSignature(payload, secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Log.Error("Webhook delivery failed",
			zap.Error(err),
			zap.String("url", url),
		)
		s.webhookRepo.MarkFailed(webhookID, 0, err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response
	responseBody := make([]byte, 1024)
	resp.Body.Read(responseBody)

	// Check if successful (2xx status code)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Log.Info("Webhook delivered successfully",
			zap.String("webhook_id", webhookID.String()),
			zap.Int("status_code", resp.StatusCode),
		)
		s.webhookRepo.MarkDelivered(webhookID, resp.StatusCode, string(responseBody))
	} else {
		logger.Log.Warn("Webhook delivery failed",
			zap.String("webhook_id", webhookID.String()),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(responseBody)),
		)
		s.webhookRepo.MarkFailed(webhookID, resp.StatusCode, string(responseBody))
	}
}

// RetryFailedWebhooks retries webhooks that failed previously
func (s *WebhookService) RetryFailedWebhooks(ctx context.Context) error {
	logger.Log.Info("Starting webhook retry worker")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Webhook retry worker stopped")
			return nil
		case <-time.After(5 * time.Minute):
			s.processRetries()
		}
	}
}

func (s *WebhookService) processRetries() {
	// Get pending retries
	webhooks, err := s.webhookRepo.FindPendingRetries()
	if err != nil {
		logger.Log.Error("Failed to fetch pending webhook retries", zap.Error(err))
		return
	}

	if len(webhooks) == 0 {
		return
	}

	logger.Log.Info("Processing webhook retries", zap.Int("count", len(webhooks)))

	for _, webhook := range webhooks {
		// Get webhook secret (should be fetched from merchant settings)
		webhookSecret := "merchant_webhook_secret" // TODO: Fetch from merchant service

		s.deliverWebhook(
			webhook.ID,
			webhook.WebhookURL,
			[]byte(webhook.Payload),
			webhookSecret,
		)

		// Rate limit retries (1 per second)
		time.Sleep(1 * time.Second)
	}
}

// generateSignature creates HMAC-SHA256 signature for webhook verification
func (s *WebhookService) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyWebhookSignature verifies webhook signature (for testing)
func (s *WebhookService) VerifyWebhookSignature(payload []byte, signature, secret string) bool {
	expectedSignature := s.generateSignature(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

const (
	WebhookEventPaymentAuthorized = "payment.authorized"
	WebhookEventPaymentCaptured   = "payment.captured"
	WebhookEventPaymentVoided     = "payment.voided"
	WebhookEventPaymentRefunded   = "payment.refunded"
	WebhookEventPaymentFailed     = "payment.failed"
)

// GetWebhookEventType returns the appropriate webhook event type for payment status
func GetWebhookEventType(status model.PaymentStatus) string {
	switch status {
	case model.PaymentStatusAuthorized:
		return WebhookEventPaymentAuthorized
	case model.PaymentStatusCaptured:
		return WebhookEventPaymentCaptured
	case model.PaymentStatusVoided:
		return WebhookEventPaymentVoided
	case model.PaymentStatusRefunded:
		return WebhookEventPaymentRefunded
	case model.PaymentStatusFailed:
		return WebhookEventPaymentFailed
	default:
		return "payment.unknown"
	}
}
