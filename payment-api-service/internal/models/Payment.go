package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusAuthorized PaymentStatus = "authorized"
	PaymentStatusCaptured   PaymentStatus = "captured"
	PaymentStatusVoided     PaymentStatus = "voided"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusFailed     PaymentStatus = "failed"
)

// PaymentType represents the type of payment operation
type PaymentType string

const (
	PaymentTypeAuthorize PaymentType = "authorize" // Hold funds
	PaymentTypeSale      PaymentType = "sale"      // Authorize + Capture
	PaymentTypeCapture   PaymentType = "capture"   // Capture held funds
	PaymentTypeVoid      PaymentType = "void"      // Cancel authorization
	PaymentTypeRefund    PaymentType = "refund"    // Return funds
)

// Payment represents a payment record
type Payment struct {
	ID            uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	MerchantID    uuid.UUID     `gorm:"type:uuid;not null;index" json:"merchant_id"`
	TransactionID uuid.UUID     `gorm:"type:uuid;index" json:"transaction_id"` // Link to transaction service
	
	// Payment Details
	Type          PaymentType   `gorm:"type:varchar(20);not null" json:"type"`
	Status        PaymentStatus `gorm:"type:varchar(20);not null;index" json:"status"`
	Amount        int64         `gorm:"not null" json:"amount"`         // Amount in cents
	Currency      string        `gorm:"type:varchar(3);not null" json:"currency"` // USD, EUR, etc.
	
	// Card/Token Info
	Token         string        `gorm:"type:varchar(255);index" json:"token"`
	CardBrand     string        `gorm:"type:varchar(50)" json:"card_brand"`
	CardLast4     string        `gorm:"type:varchar(4)" json:"card_last4"`
	
	// Customer Info
	CustomerEmail sql.NullString `gorm:"type:varchar(255)" json:"customer_email,omitempty"`
	CustomerName  sql.NullString `gorm:"type:varchar(255)" json:"customer_name,omitempty"`
	
	// Payment Response
	AuthCode      sql.NullString `gorm:"type:varchar(50)" json:"auth_code,omitempty"`
	ResponseCode  sql.NullString `gorm:"type:varchar(10)" json:"response_code,omitempty"`
	ResponseMsg   sql.NullString `gorm:"type:text" json:"response_message,omitempty"`
	
	// Fraud
	FraudScore    int           `gorm:"default:0" json:"fraud_score"`
	FraudDecision string        `gorm:"type:varchar(20)" json:"fraud_decision"` // approve, review, decline
	
	// Related Payments
	ParentPaymentID sql.NullString `gorm:"type:uuid" json:"parent_payment_id,omitempty"` // For capture/void/refund
	
	// Metadata
	Description   sql.NullString `gorm:"type:text" json:"description,omitempty"`
	Metadata      sql.NullString `gorm:"type:jsonb" json:"metadata,omitempty"` // Custom merchant data
	
	// Idempotency
	IdempotencyKey sql.NullString `gorm:"type:varchar(255);uniqueIndex" json:"idempotency_key,omitempty"`
	
	// Audit
	IPAddress     string        `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent     sql.NullString `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedBy     uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	
	// Timestamps
	CreatedAt     time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	CapturedAt    sql.NullTime  `json:"captured_at,omitempty"`
	VoidedAt      sql.NullTime  `json:"voided_at,omitempty"`
	RefundedAt    sql.NullTime  `json:"refunded_at,omitempty"`
}

// TableName specifies the table name
func (Payment) TableName() string {
	return "payments"
}

// IsAuthorized checks if payment is in authorized state
func (p *Payment) IsAuthorized() bool {
	return p.Status == PaymentStatusAuthorized
}

// IsCaptured checks if payment is captured
func (p *Payment) IsCaptured() bool {
	return p.Status == PaymentStatusCaptured
}

// CanCapture checks if payment can be captured
func (p *Payment) CanCapture() bool {
	return p.Status == PaymentStatusAuthorized
}

// CanVoid checks if payment can be voided
func (p *Payment) CanVoid() bool {
	return p.Status == PaymentStatusAuthorized
}

// CanRefund checks if payment can be refunded
func (p *Payment) CanRefund() bool {
	return p.Status == PaymentStatusCaptured
}
