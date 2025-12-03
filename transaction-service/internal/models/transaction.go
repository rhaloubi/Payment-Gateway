package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeAuthorize TransactionType = "authorize"
	TransactionTypeCapture   TransactionType = "capture"
	TransactionTypeVoid      TransactionType = "void"
	TransactionTypeRefund    TransactionType = "refund"
	TransactionTypeSale      TransactionType = "sale" // Authorize + Capture
)

// TransactionStatus represents the current state of a transaction
type TransactionStatus string

const (
	TransactionStatusPending           TransactionStatus = "pending"
	TransactionStatusAuthorized        TransactionStatus = "authorized"
	TransactionStatusCaptured          TransactionStatus = "captured"
	TransactionStatusVoided            TransactionStatus = "voided"
	TransactionStatusSettled           TransactionStatus = "settled"
	TransactionStatusRefunded          TransactionStatus = "refunded"
	TransactionStatusPartiallyRefunded TransactionStatus = "partially_refunded"
	TransactionStatusFailed            TransactionStatus = "failed"
)

// Transaction represents a payment transaction
type Transaction struct {
	ID                  uuid.UUID         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	MerchantID          uuid.UUID         `gorm:"type:uuid;not null;index" json:"merchant_id"`
	ParentTransactionID sql.NullString    `gorm:"type:uuid;index" json:"parent_transaction_id,omitempty"` // For refunds
	
	// Transaction Details
	Type                TransactionType   `gorm:"type:varchar(20);not null" json:"type"`
	Status              TransactionStatus `gorm:"type:varchar(30);not null;index" json:"status"`
	Amount              int64             `gorm:"not null" json:"amount"`          // Amount in cents
	Currency            string            `gorm:"type:varchar(3);not null" json:"currency"` // USD, EUR, MAD
	AmountMAD           int64             `gorm:"not null" json:"amount_mad"`      // Converted to MAD
	ExchangeRate        float64           `gorm:"type:decimal(10,6)" json:"exchange_rate"` // Rate used
	
	// Card Information (from tokenization)
	CardToken           string            `gorm:"type:varchar(255);index" json:"card_token"`
	CardBrand           string            `gorm:"type:varchar(50)" json:"card_brand"`
	CardLast4           string            `gorm:"type:varchar(4)" json:"card_last4"`
	
	// Authorization Details
	AuthCode            sql.NullString    `gorm:"type:varchar(50)" json:"auth_code,omitempty"`
	ResponseCode        sql.NullString    `gorm:"type:varchar(10)" json:"response_code,omitempty"`
	ResponseMessage     sql.NullString    `gorm:"type:text" json:"response_message,omitempty"`
	AVSResult           sql.NullString    `gorm:"type:varchar(1)" json:"avs_result,omitempty"` // Address Verification
	CVVResult           sql.NullString    `gorm:"type:varchar(1)" json:"cvv_result,omitempty"` // CVV Check
	
	// Fraud Information
	FraudScore          int               `gorm:"default:0" json:"fraud_score"`
	FraudDecision       string            `gorm:"type:varchar(20)" json:"fraud_decision"` // approve, review, decline
	
	// Amounts Tracking
	CapturedAmount      int64             `gorm:"default:0" json:"captured_amount"`
	RefundedAmount      int64             `gorm:"default:0" json:"refunded_amount"`
	
	// Processing Fees (2.9% + $0.30)
	ProcessingFee       int64             `gorm:"default:0" json:"processing_fee"` // In cents
	NetAmount           int64             `gorm:"default:0" json:"net_amount"`     // Amount - Fee
	
	// Settlement Information
	SettlementBatchID   sql.NullString    `gorm:"type:uuid" json:"settlement_batch_id,omitempty"`
	
	// Metadata
	Description         sql.NullString    `gorm:"type:text" json:"description,omitempty"`
	Metadata            sql.NullString    `gorm:"type:jsonb" json:"metadata,omitempty"`
	
	// IP & Device Info
	IPAddress           string            `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent           sql.NullString    `gorm:"type:text" json:"user_agent,omitempty"`
	
	// Timestamps
	AuthorizedAt        sql.NullTime      `json:"authorized_at,omitempty"`
	CapturedAt          sql.NullTime      `json:"captured_at,omitempty"`
	VoidedAt            sql.NullTime      `json:"voided_at,omitempty"`
	RefundedAt          sql.NullTime      `json:"refunded_at,omitempty"`
	SettledAt           sql.NullTime      `json:"settled_at,omitempty"`
	ExpiresAt           sql.NullTime      `json:"expires_at,omitempty"` // Auto-void after 7 days
	CreatedAt           time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name
func (Transaction) TableName() string {
	return "transactions"
}

// IsAuthorized checks if transaction is authorized
func (t *Transaction) IsAuthorized() bool {
	return t.Status == TransactionStatusAuthorized
}

// IsCaptured checks if transaction is captured
func (t *Transaction) IsCaptured() bool {
	return t.Status == TransactionStatusCaptured
}

// CanCapture checks if transaction can be captured
func (t *Transaction) CanCapture() bool {
	return t.Status == TransactionStatusAuthorized && !t.IsExpired()
}

// CanVoid checks if transaction can be voided
func (t *Transaction) CanVoid() bool {
	return t.Status == TransactionStatusAuthorized && !t.IsExpired()
}

// CanRefund checks if transaction can be refunded
func (t *Transaction) CanRefund() bool {
	return (t.Status == TransactionStatusCaptured || t.Status == TransactionStatusSettled || 
			t.Status == TransactionStatusPartiallyRefunded) && 
			t.RefundedAmount < t.CapturedAmount
}

// IsExpired checks if authorization has expired
func (t *Transaction) IsExpired() bool {
	if !t.ExpiresAt.Valid {
		return false
	}
	return time.Now().After(t.ExpiresAt.Time)
}

// RemainingRefundableAmount returns amount that can still be refunded
func (t *Transaction) RemainingRefundableAmount() int64 {
	return t.CapturedAmount - t.RefundedAmount
}

// TransactionEvent represents a state change in transaction lifecycle
type TransactionEvent struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TransactionID  uuid.UUID         `gorm:"type:uuid;not null;index" json:"transaction_id"`
	EventType      string            `gorm:"type:varchar(50);not null" json:"event_type"`
	OldStatus      TransactionStatus `gorm:"type:varchar(30)" json:"old_status"`
	NewStatus      TransactionStatus `gorm:"type:varchar(30)" json:"new_status"`
	Amount         int64             `json:"amount"`
	Metadata       sql.NullString    `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedBy      uuid.UUID         `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt      time.Time         `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (TransactionEvent) TableName() string {
	return "transaction_events"
}

// IssuerResponse stores raw issuer/simulator responses for debugging
type IssuerResponse struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TransactionID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"transaction_id"`
	RequestPayload   sql.NullString `gorm:"type:jsonb" json:"request_payload,omitempty"`
	ResponsePayload  sql.NullString `gorm:"type:jsonb" json:"response_payload,omitempty"`
	ProcessingTimeMs int            `json:"processing_time_ms"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (IssuerResponse) TableName() string {
	return "issuer_responses"
}
