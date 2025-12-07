package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ChargebackEvent tracks chargeback state changes
type ChargebackEvent struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ChargebackID uuid.UUID        `gorm:"type:uuid;not null;index" json:"chargeback_id"`
	EventType    string           `gorm:"type:varchar(50);not null" json:"event_type"`
	OldStatus    ChargebackStatus `gorm:"type:varchar(30)" json:"old_status"`
	NewStatus    ChargebackStatus `gorm:"type:varchar(30)" json:"new_status"`
	Note         sql.NullString   `gorm:"type:text" json:"note,omitempty"`
	CreatedBy    sql.NullString   `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt    time.Time        `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (ChargebackEvent) TableName() string {
	return "chargeback_events"
}
