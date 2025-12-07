package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// TransactionEvent represents a state change in transaction lifecycle
type TransactionEvent struct {
	ID            uuid.UUID         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TransactionID uuid.UUID         `gorm:"type:uuid;not null;index" json:"transaction_id"`
	EventType     string            `gorm:"type:varchar(50);not null" json:"event_type"`
	OldStatus     TransactionStatus `gorm:"type:varchar(30)" json:"old_status"`
	NewStatus     TransactionStatus `gorm:"type:varchar(30)" json:"new_status"`
	Amount        int64             `json:"amount"`
	Metadata      sql.NullString    `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedBy     uuid.UUID         `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt     time.Time         `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (TransactionEvent) TableName() string {
	return "transaction_events"
}
