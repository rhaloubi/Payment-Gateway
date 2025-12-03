package model

import (
	"time"

	"github.com/google/uuid"
)

// ExchangeRate stores currency conversion rates
type ExchangeRate struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	FromCurrency string    `gorm:"type:varchar(3);not null;index" json:"from_currency"` // USD, EUR
	ToCurrency   string    `gorm:"type:varchar(3);not null;index" json:"to_currency"`   // MAD
	Rate         float64   `gorm:"type:decimal(10,6);not null" json:"rate"`
	EffectiveAt  time.Time `gorm:"not null;index" json:"effective_at"`
	Source       string    `gorm:"type:varchar(100)" json:"source"` // API source
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name
func (ExchangeRate) TableName() string {
	return "exchange_rates"
}

// Supported currencies
const (
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
	CurrencyMAD = "MAD"
)

// Default exchange rates (will be updated from external API)
var DefaultExchangeRates = map[string]float64{
	"USD_MAD": 10.00, // 1 USD = 10 MAD
	"EUR_MAD": 11.00, // 1 EUR = 11 MAD
	"MAD_MAD": 1.00,  // 1 MAD = 1 MAD
}
