package service

import (
	"context"
	"time"

	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"github.com/rhaloubi/payment-gateway/transaction-service/internal/repository"
	"go.uber.org/zap"
)

type CurrencyService struct {
	exchangeRateRepo *repository.ExchangeRateRepository
}

func NewCurrencyService() *CurrencyService {
	return &CurrencyService{
		exchangeRateRepo: repository.NewExchangeRateRepository(),
	}
}

// ConvertToMAD converts amount from any currency to MAD
func (s *CurrencyService) ConvertToMAD(amount int64, fromCurrency string) (int64, float64, error) {
	// If already MAD, no conversion needed
	if fromCurrency == model.CurrencyMAD {
		return amount, 1.0, nil
	}

	// Get exchange rate
	rate, err := s.GetExchangeRate(fromCurrency, model.CurrencyMAD)
	if err != nil {
		logger.Log.Error("Failed to get exchange rate",
			zap.Error(err),
			zap.String("from", fromCurrency),
		)
		return 0, 0, err
	}

	// Convert (amount is in cents, rate is per unit)
	amountMAD := int64(float64(amount) * rate)

	logger.Log.Debug("Currency conversion",
		zap.Int64("original_amount", amount),
		zap.String("from_currency", fromCurrency),
		zap.Float64("rate", rate),
		zap.Int64("converted_amount", amountMAD),
	)

	return amountMAD, rate, nil
}

// GetExchangeRate retrieves the current exchange rate
func (s *CurrencyService) GetExchangeRate(fromCurrency, toCurrency string) (float64, error) {
	// Try to get from database (cached rates)
	rate, err := s.exchangeRateRepo.FindLatestRate(fromCurrency, toCurrency)
	if err == nil && rate != nil {
		// Check if rate is still fresh (< 1 hour old)
		if time.Since(rate.EffectiveAt) < 1*time.Hour {
			return rate.Rate, nil
		}
	}

	// Rate not found or stale, use default rates
	// In production, this would call an external API (e.g., OpenExchangeRates)
	rateValue := s.getDefaultRate(fromCurrency, toCurrency)

	// Store in database for future use
	newRate := &model.ExchangeRate{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Rate:         rateValue,
		EffectiveAt:  time.Now(),
		Source:       "default",
	}
	s.exchangeRateRepo.Create(newRate)

	return rateValue, nil
}

func (s *CurrencyService) getDefaultRate(fromCurrency, toCurrency string) float64 {
	key := fromCurrency + "_" + toCurrency
	if rate, exists := model.DefaultExchangeRates[key]; exists {
		return rate
	}

	// If not found, return 1.0 (no conversion)
	logger.Log.Warn("Exchange rate not found, using 1.0",
		zap.String("from", fromCurrency),
		zap.String("to", toCurrency),
	)
	return 1.0
}

// UpdateExchangeRates fetches latest rates from external API
// This should be called periodically (every hour) via cron job
func (s *CurrencyService) UpdateExchangeRates(ctx context.Context) error {
	logger.Log.Info("Updating exchange rates from external API")

	// TODO: Call external API (e.g., OpenExchangeRates, CurrencyLayer)
	// For now, using default rates

	rates := []struct {
		From string
		To   string
		Rate float64
	}{
		{model.CurrencyUSD, model.CurrencyMAD, 10.00},
		{model.CurrencyEUR, model.CurrencyMAD, 11.00},
		{model.CurrencyMAD, model.CurrencyMAD, 1.00},
	}

	for _, r := range rates {
		exchangeRate := &model.ExchangeRate{
			FromCurrency: r.From,
			ToCurrency:   r.To,
			Rate:         r.Rate,
			EffectiveAt:  time.Now(),
			Source:       "manual_update",
		}

		if err := s.exchangeRateRepo.Create(exchangeRate); err != nil {
			logger.Log.Error("Failed to save exchange rate",
				zap.Error(err),
				zap.String("from", r.From),
				zap.String("to", r.To),
			)
		}
	}

	logger.Log.Info("Exchange rates updated successfully")
	return nil
}

// CalculateProcessingFee calculates fee: 2.9% + $0.30 (converted to MAD)
func (s *CurrencyService) CalculateProcessingFee(amountMAD int64) int64 {
	// Base fee: $0.30 = 300 MAD cents (assuming 1 USD = 10 MAD)
	baseFeeMAD := int64(300) // 3 MAD in cents

	// Percentage fee: 2.9%
	percentageFee := int64(float64(amountMAD) * 0.029)

	totalFee := baseFeeMAD + percentageFee

	logger.Log.Debug("Processing fee calculated",
		zap.Int64("amount_mad", amountMAD),
		zap.Int64("base_fee", baseFeeMAD),
		zap.Int64("percentage_fee", percentageFee),
		zap.Int64("total_fee", totalFee),
	)

	return totalFee
}

// ConvertBack converts MAD back to original currency (for refunds)
func (s *CurrencyService) ConvertBack(amountMAD int64, toCurrency string, originalRate float64) int64 {
	if toCurrency == model.CurrencyMAD {
		return amountMAD
	}

	// Use original rate to convert back
	originalAmount := int64(float64(amountMAD) / originalRate)

	logger.Log.Debug("Converting back from MAD",
		zap.Int64("amount_mad", amountMAD),
		zap.String("to_currency", toCurrency),
		zap.Float64("rate", originalRate),
		zap.Int64("original_amount", originalAmount),
	)

	return originalAmount
}

// GetCurrencyBreakdown returns breakdown of amounts by currency
func (s *CurrencyService) GetCurrencyBreakdown(transactions []model.Transaction) (map[string]int64, error) {
	breakdown := make(map[string]int64)

	for _, txn := range transactions {
		breakdown[txn.Currency] += txn.Amount
	}

	return breakdown, nil
}
