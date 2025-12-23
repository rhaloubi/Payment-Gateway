package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PaymentIntentStatus string

const (
	PaymentIntentStatusCreated         PaymentIntentStatus = "created"
	PaymentIntentStatusAwaitingPayment PaymentIntentStatus = "awaiting_payment_method"
	PaymentIntentStatusAuthorized      PaymentIntentStatus = "authorized"
	PaymentIntentStatusCaptured        PaymentIntentStatus = "captured"
	PaymentIntentStatusFailed          PaymentIntentStatus = "failed"
	PaymentIntentStatusCanceled        PaymentIntentStatus = "canceled"
	PaymentIntentStatusExpired         PaymentIntentStatus = "expired"
)

type CaptureMethod string

const (
	CaptureMethodAutomatic CaptureMethod = "automatic" // Capture immediately after auth
	CaptureMethodManual    CaptureMethod = "manual"    // Merchant captures manually
)

type PaymentIntent struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;index" json:"merchant_id"`

	// Order/Reference Info
	OrderID     sql.NullString `gorm:"type:varchar(255);index" json:"order_id,omitempty"`
	Description sql.NullString `gorm:"type:text" json:"description,omitempty"`

	// Payment Details (set by merchant, never by browser)
	Amount   int64  `gorm:"not null" json:"amount"` // Amount in cents
	Currency string `gorm:"type:varchar(3);not null" json:"currency"`

	// Status & Flow
	Status        PaymentIntentStatus `gorm:"type:varchar(30);not null;index" json:"status"`
	CaptureMethod CaptureMethod       `gorm:"type:varchar(20);not null" json:"capture_method"`

	// Payment Reference (once confirmed)
	PaymentID sql.NullString `gorm:"type:uuid;index" json:"payment_id,omitempty"`

	// Redirect URLs
	SuccessURL string `gorm:"type:text" json:"success_url"`
	CancelURL  string `gorm:"type:text" json:"cancel_url"`

	// Customer Info (optional)
	CustomerEmail sql.NullString `gorm:"type:varchar(255)" json:"customer_email,omitempty"`
	CustomerName  sql.NullString `gorm:"type:varchar(255)" json:"customer_name,omitempty"`

	// Metadata
	Metadata sql.NullString `gorm:"type:jsonb" json:"metadata,omitempty"`

	// Security
	ClientSecret string `gorm:"type:varchar(255);uniqueIndex" json:"client_secret"` // For checkout UI auth

	AttemptCount  int          `gorm:"default:0" json:"attempt_count"`
	MaxAttempts   int          `gorm:"default:7" json:"max_attempts"`
	LastAttemptAt sql.NullTime `json:"last_attempt_at,omitempty"`
	BlockedUntil  sql.NullTime `json:"blocked_until,omitempty"`

	// Expiration
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`

	// Timestamps
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	ConfirmedAt sql.NullTime `json:"confirmed_at,omitempty"`
	CanceledAt  sql.NullTime `json:"canceled_at,omitempty"`
}

func (PaymentIntent) TableName() string {
	return "payment_intents"
}

// IsExpired checks if the payment intent has expired
func (pi *PaymentIntent) IsExpired() bool {
	return time.Now().After(pi.ExpiresAt)
}
func (pi *PaymentIntent) CanCancel() bool {
	return pi.Status == PaymentIntentStatusAwaitingPayment ||
		pi.Status == PaymentIntentStatusAuthorized
}
func (pi *PaymentIntent) CanConfirm() bool {
	// Check if expired
	if pi.IsExpired() {
		return false
	}

	// Check if max attempts reached
	if pi.AttemptCount >= pi.MaxAttempts {
		return false
	}

	// Check status
	return pi.Status == PaymentIntentStatusAwaitingPayment
}

// GetRemainingAttempts returns how many attempts are left
func (pi *PaymentIntent) GetRemainingAttempts() int {
	remaining := pi.MaxAttempts - pi.AttemptCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetCheckoutURL returns the hosted checkout URL
func (pi *PaymentIntent) GetCheckoutURL(baseURL string) string {
	return baseURL + "/checkout/" + pi.ID.String()
}
