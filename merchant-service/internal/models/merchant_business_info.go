package model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MerchantBusinessInfo struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	MerchantID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`

	// Tax and registration
	TaxID              sql.NullString `gorm:"type:varchar(100)"` // ICE in Morocco
	RegistrationNumber sql.NullString `gorm:"type:varchar(100)"` // RC in Morocco
	VATNumber          sql.NullString `gorm:"type:varchar(100)"` // IF in Morocco

	// Business details
	BusinessDescription sql.NullString `gorm:"type:text"`
	Industry            sql.NullString `gorm:"type:varchar(100)"`
	FoundedYear         sql.NullInt32  `gorm:"type:integer"`
	EmployeeCount       sql.NullInt32  `gorm:"type:integer"`

	// Address (Morocco)
	AddressLine1 sql.NullString `gorm:"type:varchar(255)"`
	AddressLine2 sql.NullString `gorm:"type:varchar(255)"`
	City         sql.NullString `gorm:"type:varchar(100)"`
	Region       sql.NullString `gorm:"type:varchar(100)"` // e.g., "Tanger-Tétouan-Al Hoceïma"
	PostalCode   sql.NullString `gorm:"type:varchar(20)"`
	CountryCode  string         `gorm:"type:char(2);default:'MA'"`

	// Contact person
	ContactName  sql.NullString `gorm:"type:varchar(255)"`
	ContactPhone sql.NullString `gorm:"type:varchar(50)"`
	ContactEmail sql.NullString `gorm:"type:varchar(255)"`

	// Relationships
	Merchant *Merchant `gorm:"foreignKey:MerchantID"`

	// Timestamps
	CreatedAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for MerchantBusinessInfo
func (MerchantBusinessInfo) TableName() string {
	return "merchant_business_info"
}

// BeforeCreate hook
func (mbi *MerchantBusinessInfo) BeforeCreate(tx *gorm.DB) error {
	if mbi.ID == uuid.Nil {
		mbi.ID = uuid.New()
	}
	mbi.CountryCode = "MA" // Always Morocco
	return nil
}
