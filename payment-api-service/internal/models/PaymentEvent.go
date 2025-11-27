package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)



type PaymentEvent struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PaymentID   uuid.UUID     `gorm:"type:uuid;not null;index" json:"payment_id"`
	EventType   string        `gorm:"type:varchar(50);not null" json:"event_type"` // authorized, captured, voided, etc.
	OldStatus   PaymentStatus `gorm:"type:varchar(20)" json:"old_status"`
	NewStatus   PaymentStatus `gorm:"type:varchar(20)" json:"new_status"`
	Amount      int64         `json:"amount"`
	Description sql.NullString `gorm:"type:text" json:"description,omitempty"`
	CreatedBy   uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (PaymentEvent) TableName() string {
	return "payment_events"
}
