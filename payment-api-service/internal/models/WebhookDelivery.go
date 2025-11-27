package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)



// WebhookDelivery tracks webhook delivery attempts
type WebhookDelivery struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PaymentID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"payment_id"`
	MerchantID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"merchant_id"`
	EventType    string         `gorm:"type:varchar(50);not null" json:"event_type"`
	WebhookURL   string         `gorm:"type:text;not null" json:"webhook_url"`
	Payload      string         `gorm:"type:jsonb" json:"payload"`
	Response     sql.NullString `gorm:"type:text" json:"response,omitempty"`
	StatusCode   int            `json:"status_code"`
	Success      bool           `gorm:"default:false" json:"success"`
	AttemptCount int            `gorm:"default:1" json:"attempt_count"`
	NextRetryAt  sql.NullTime   `json:"next_retry_at,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeliveredAt  sql.NullTime   `json:"delivered_at,omitempty"`
}

// TableName specifies the table name
func (WebhookDelivery) TableName() string {
	return "webhook_deliveries"
}
