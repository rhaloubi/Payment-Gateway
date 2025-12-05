package repository

import (
	"github.com/rhaloubi/payment-gateway/transaction-service/inits"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"gorm.io/gorm"
)

type ExchangeRateRepository struct {
	db *gorm.DB
}

func NewExchangeRateRepository() *ExchangeRateRepository {
	return &ExchangeRateRepository{
		db: inits.DB,
	}
}

func (r *ExchangeRateRepository) Create(rate *model.ExchangeRate) error {
	return r.db.Create(rate).Error
}

func (r *ExchangeRateRepository) FindLatestRate(fromCurrency, toCurrency string) (*model.ExchangeRate, error) {
	var rate model.ExchangeRate
	if err := r.db.Where("from_currency = ? AND to_currency = ?", fromCurrency, toCurrency).
		Order("effective_at DESC").
		First(&rate).Error; err != nil {
		return nil, err
	}
	return &rate, nil
}
