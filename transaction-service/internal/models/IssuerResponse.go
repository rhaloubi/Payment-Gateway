package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

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
