package validation

import (
	"os"
	"testing"

	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
)

func TestCardValidator_DetectCardBrand(t *testing.T) {
	validator := NewCardValidator()

	tests := []struct {
		name       string
		cardNumber string
		want       model.CardBrand
	}{
		// Visa Tests
		{"Visa Standard 16", "4000123456789010", model.CardBrandVisa},
		{"Visa 13 Digits", "4000123456789", model.CardBrandVisa},
		{"Visa 19 Digits", "4000123456789012345", model.CardBrandVisa},
		
		// Mastercard Tests
		{"Mastercard Standard 51", "5100123456789010", model.CardBrandMastercard},
		{"Mastercard Standard 55", "5500123456789010", model.CardBrandMastercard},
		{"Mastercard 2-series Low (2221)", "2221000000000000", model.CardBrandMastercard},
		{"Mastercard 2-series Mid (2500)", "2500000000000000", model.CardBrandMastercard},
		{"Mastercard 2-series High (2720)", "2720000000000000", model.CardBrandMastercard},

		// Negative Tests
		{"Unknown Brand (3...)", "3000000000000000", model.CardBrandUnknown},
		{"Mastercard 2-series Too Low (2220)", "2220000000000000", model.CardBrandUnknown}, // Should fail
		{"Mastercard 2-series Too High (2721)", "2721000000000000", model.CardBrandUnknown}, // Should fail
		{"Invalid Length Visa", "4000", model.CardBrandUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validator.DetectCardBrand(tt.cardNumber); got != tt.want {
				t.Errorf("CardValidator.DetectCardBrand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCardValidator_ValidateCardNumber(t *testing.T) {
	validator := NewCardValidator()
	
	tests := []struct {
		name       string
		cardNumber string
		skipLuhn   string
		wantErr    bool
	}{
		{"Valid Visa", "4242424242424242", "", false},
		{"Invalid Luhn", "4242424242424243", "", true}, 
		{"Too Short", "4242", "", true},
		{"Invalid Luhn Skipped", "4242424242424243", "true", false}, // Should pass with skip
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipLuhn != "" {
				os.Setenv("SKIP_LUHN_VALIDATION", tt.skipLuhn)
				defer os.Unsetenv("SKIP_LUHN_VALIDATION")
			} else {
				os.Unsetenv("SKIP_LUHN_VALIDATION")
			}

			err := validator.ValidateCardNumber(tt.cardNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("CardValidator.ValidateCardNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
