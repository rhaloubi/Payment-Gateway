package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// SettlementStatus represents the status of a settlement batch
type SettlementStatus string

const (
	SettlementStatusPending   SettlementStatus = "pending"
	SettlementStatusProcessing SettlementStatus = "processing"
	SettlementStatusSettled   SettlementStatus = "settled"
	SettlementStatusFailed    SettlementStatus = "failed"
)

// SettlementBatch represents a daily settlement batch
type SettlementBatch struct {
	ID                uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	MerchantID        uuid.UUID        `gorm:"type:uuid;not null;index" json:"merchant_id"`
	BatchDate         time.Time        `gorm:"type:date;not null;index" json:"batch_date"`
	
	// Amounts (all in MAD after conversion)
	GrossAmount       int64            `gorm:"not null" json:"gross_amount"`       // Total captures
	RefundAmount      int64            `gorm:"default:0" json:"refund_amount"`     // Total refunds
	FeeAmount         int64            `gorm:"not null" json:"fee_amount"`         // Processing fees
	NetAmount         int64            `gorm:"not null" json:"net_amount"`         // Amount to merchant
	
	// Transaction Counts
	TransactionCount  int              `gorm:"not null" json:"transaction_count"`
	RefundCount       int              `gorm:"default:0" json:"refund_count"`
	
	// Currency Breakdown
	CurrencyBreakdown sql.NullString   `gorm:"type:jsonb" json:"currency_breakdown,omitempty"` // {"USD": 1000, "EUR": 500}
	
	// Settlement Details
	Status            SettlementStatus `gorm:"type:varchar(20);not null" json:"status"`
	SettlementDate    time.Time        `gorm:"type:date" json:"settlement_date"` // T+2
	SettlementMethod  string           `gorm:"type:varchar(50)" json:"settlement_method"` // bank_transfer, ach, wire
	
	// Bank Information (from merchant settings)
	BankAccount       sql.NullString   `gorm:"type:varchar(255)" json:"bank_account,omitempty"`
	BankName          sql.NullString   `gorm:"type:varchar(255)" json:"bank_name,omitempty"`
	
	// Report & Reference
	ReportURL         sql.NullString   `gorm:"type:text" json:"report_url,omitempty"`
	ReferenceNumber   sql.NullString   `gorm:"type:varchar(100)" json:"reference_number,omitempty"`
	
	// Timestamps
	CreatedAt         time.Time        `gorm:"autoCreateTime" json:"created_at"`
	SettledAt         sql.NullTime     `json:"settled_at,omitempty"`
	FailedAt          sql.NullTime     `json:"failed_at,omitempty"`
}

// TableName specifies the table name
func (SettlementBatch) TableName() string {
	return "settlement_batches"
}

// IsSettled checks if batch is settled
func (s *SettlementBatch) IsSettled() bool {
	return s.Status == SettlementStatusSettled
}

// IsPending checks if batch is pending
func (s *SettlementBatch) IsPending() bool {
	return s.Status == SettlementStatusPending
}
