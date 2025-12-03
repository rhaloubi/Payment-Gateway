package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ChargebackStatus represents the status of a chargeback/dispute
type ChargebackStatus string

const (
	ChargebackStatusOpen              ChargebackStatus = "open"
	ChargebackStatusUnderReview       ChargebackStatus = "under_review"
	ChargebackStatusNeedsResponse     ChargebackStatus = "needs_response"
	ChargebackStatusWon               ChargebackStatus = "won"               // Merchant wins
	ChargebackStatusLost              ChargebackStatus = "lost"              // Customer wins
	ChargebackStatusAccepted          ChargebackStatus = "accepted"          // Merchant accepts
	ChargebackStatusClosed            ChargebackStatus = "closed"
)

// ChargebackReason represents the reason for chargeback
type ChargebackReason string

const (
	ChargebackReasonFraud              ChargebackReason = "fraud"
	ChargebackReasonProductNotReceived ChargebackReason = "product_not_received"
	ChargebackReasonProductDefective   ChargebackReason = "product_defective"
	ChargebackReasonDuplicate          ChargebackReason = "duplicate"
	ChargebackReasonCreditNotProcessed ChargebackReason = "credit_not_processed"
	ChargebackReasonUnauthorized       ChargebackReason = "unauthorized"
	ChargebackReasonOther              ChargebackReason = "other"
)

// Chargeback represents a dispute/chargeback
type Chargeback struct {
	ID                uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TransactionID     uuid.UUID        `gorm:"type:uuid;not null;index" json:"transaction_id"`
	MerchantID        uuid.UUID        `gorm:"type:uuid;not null;index" json:"merchant_id"`
	
	// Chargeback Details
	Status            ChargebackStatus `gorm:"type:varchar(30);not null;index" json:"status"`
	Reason            ChargebackReason `gorm:"type:varchar(50);not null" json:"reason"`
	ReasonCode        string           `gorm:"type:varchar(10)" json:"reason_code"` // Issuer code (4853, etc.)
	Amount            int64            `gorm:"not null" json:"amount"`              // Disputed amount
	Currency          string           `gorm:"type:varchar(3)" json:"currency"`
	
	// Issuer Information
	IssuerReference   sql.NullString   `gorm:"type:varchar(100)" json:"issuer_reference,omitempty"`
	IssuerBank        sql.NullString   `gorm:"type:varchar(255)" json:"issuer_bank,omitempty"`
	
	// Response Details
	ResponseDueDate   sql.NullTime     `json:"response_due_date,omitempty"`
	ResponseSubmittedAt sql.NullTime   `json:"response_submitted_at,omitempty"`
	
	// Evidence & Documentation
	MerchantEvidence  sql.NullString   `gorm:"type:jsonb" json:"merchant_evidence,omitempty"` // JSON with evidence docs
	CustomerStatement sql.NullString   `gorm:"type:text" json:"customer_statement,omitempty"`
	
	// Resolution
	ResolutionReason  sql.NullString   `gorm:"type:text" json:"resolution_reason,omitempty"`
	ResolvedAt        sql.NullTime     `json:"resolved_at,omitempty"`
	ResolvedBy        sql.NullString   `gorm:"type:uuid" json:"resolved_by,omitempty"`
	
	// Financial Impact
	ChargebackFee     int64            `gorm:"default:1500" json:"chargeback_fee"` // $15.00 fee
	NetLoss           int64            `json:"net_loss"`                            // Amount + Fee
	
	// Timestamps
	DisputedAt        time.Time        `gorm:"not null" json:"disputed_at"`
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name
func (Chargeback) TableName() string {
	return "chargebacks"
}

// IsOpen checks if chargeback is still open
func (c *Chargeback) IsOpen() bool {
	return c.Status == ChargebackStatusOpen || 
		   c.Status == ChargebackStatusUnderReview || 
		   c.Status == ChargebackStatusNeedsResponse
}

// NeedsResponse checks if merchant needs to respond
func (c *Chargeback) NeedsResponse() bool {
	return c.Status == ChargebackStatusNeedsResponse && 
		   c.ResponseDueDate.Valid && 
		   time.Now().Before(c.ResponseDueDate.Time)
}

// IsOverdue checks if response deadline has passed
func (c *Chargeback) IsOverdue() bool {
	return c.Status == ChargebackStatusNeedsResponse && 
		   c.ResponseDueDate.Valid && 
		   time.Now().After(c.ResponseDueDate.Time)
}

// ChargebackEvent tracks chargeback state changes
type ChargebackEvent struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ChargebackID  uuid.UUID        `gorm:"type:uuid;not null;index" json:"chargeback_id"`
	EventType     string           `gorm:"type:varchar(50);not null" json:"event_type"`
	OldStatus     ChargebackStatus `gorm:"type:varchar(30)" json:"old_status"`
	NewStatus     ChargebackStatus `gorm:"type:varchar(30)" json:"new_status"`
	Note          sql.NullString   `gorm:"type:text" json:"note,omitempty"`
	CreatedBy     sql.NullString   `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt     time.Time        `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (ChargebackEvent) TableName() string {
	return "chargeback_events"
}
