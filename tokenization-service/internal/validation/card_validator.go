package validation

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
)

// CardValidator provides card validation services
type CardValidator struct {
	cardPatterns map[model.CardBrand]*regexp.Regexp
}

// CardValidationRequest represents a card validation request
type CardValidationRequest struct {
	CardNumber     string
	CardholderName string
	ExpiryMonth    int
	ExpiryYear     int
	CVV            string
}

// CardValidationResult represents the result of card validation
type CardValidationResult struct {
	IsValid   bool
	CardBrand model.CardBrand
	Errors    []string
}

// NewCardValidator creates a new card validator instance
func NewCardValidator() *CardValidator {
	cv := &CardValidator{
		cardPatterns: make(map[model.CardBrand]*regexp.Regexp),
	}

	// Initialize card brand patterns (Visa and Mastercard only)
	cv.cardPatterns[model.CardBrandVisa] = regexp.MustCompile(`^4[0-9]{12}(?:[0-9]{3})?$`)
	cv.cardPatterns[model.CardBrandMastercard] = regexp.MustCompile(`^5[1-5][0-9]{14}$`)

	return cv
}

// ValidateCard performs comprehensive card validation
func (cv *CardValidator) ValidateCard(req CardValidationRequest) error {
	var validationErrors []string

	// Validate card number
	if err := cv.ValidateCardNumber(req.CardNumber); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate cardholder name
	if err := cv.ValidateCardholderName(req.CardholderName); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate expiry date
	if err := cv.ValidateExpiryDate(req.ExpiryMonth, req.ExpiryYear); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate CVV
	if err := cv.ValidateCVV(req.CVV, req.CardNumber); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateCardNumber validates the card number format and performs Luhn check
func (cv *CardValidator) ValidateCardNumber(cardNumber string) error {
	// Remove all non-digit characters
	sanitized := cv.SanitizeCardNumber(cardNumber)

	// Check if empty
	if sanitized == "" {
		return errors.New("card number is required")
	}

	// Check length (typically 13-19 digits)
	if len(sanitized) < 13 || len(sanitized) > 19 {
		return errors.New("card number must be between 13 and 19 digits")
	}

	// Perform Luhn algorithm check
	if !cv.isLuhnValid(sanitized) {
		return errors.New("card number is invalid")
	}

	// Check if it's a supported card brand
	cardBrand := cv.DetectCardBrand(sanitized)
	if cardBrand == model.CardBrandUnknown {
		return errors.New("unsupported card brand (only Visa and Mastercard are accepted)")
	}

	return nil
}

// ValidateCardholderName validates the cardholder name
func (cv *CardValidator) ValidateCardholderName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("cardholder name is required")
	}

	// Basic validation: at least 2 characters, only letters, spaces, hyphens, and dots
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-\.]{2,100}$`)
	if !nameRegex.MatchString(name) {
		return errors.New("cardholder name contains invalid characters")
	}

	return nil
}

// ValidateExpiryDate validates the expiry date
func (cv *CardValidator) ValidateExpiryDate(month, year int) error {
	if month < 1 || month > 12 {
		return errors.New("expiry month must be between 1 and 12")
	}

	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())

	// Check if year is in the past
	if year < currentYear {
		return errors.New("card has expired")
	}

	// Check if month/year is in the past
	if year == currentYear && month < currentMonth {
		return errors.New("card has expired")
	}

	// Check if year is too far in the future (max 20 years)
	if year > currentYear+20 {
		return errors.New("expiry year is too far in the future")
	}

	return nil
}

// ValidateCVV validates the CVV/Security code
func (cv *CardValidator) ValidateCVV(cvv string, cardNumber string) error {
	if strings.TrimSpace(cvv) == "" {
		return errors.New("CVV is required")
	}

	// Remove any non-digit characters
	sanitized := strings.ReplaceAll(cvv, " ", "")
	sanitized = strings.ReplaceAll(sanitized, "-", "")

	// Check if all digits
	if !regexp.MustCompile(`^\d+$`).MatchString(sanitized) {
		return errors.New("CVV must contain only digits")
	}

	// For Visa and Mastercard, CVV is always 3 digits
	if len(sanitized) != 3 {
		return errors.New("CVV must be 3 digits")
	}

	return nil
}

// DetectCardBrand detects the card brand from card number
func (cv *CardValidator) DetectCardBrand(cardNumber string) model.CardBrand {
	sanitized := cv.SanitizeCardNumber(cardNumber)

	for brand, pattern := range cv.cardPatterns {
		if pattern.MatchString(sanitized) {
			return brand
		}
	}

	return model.CardBrandUnknown
}

// DetectCardNumberBrand is an alias for DetectCardBrand for backward compatibility
func (cv *CardValidator) DetectCardNumberBrand(cardNumber string) model.CardBrand {
	return cv.DetectCardBrand(cardNumber)
}

// SanitizeCardNumber removes spaces and dashes from card number
func (cv *CardValidator) SanitizeCardNumber(cardNumber string) string {
	// Remove spaces and dashes
	sanitized := strings.ReplaceAll(cardNumber, " ", "")
	sanitized = strings.ReplaceAll(sanitized, "-", "")

	// Keep only digits
	sanitized = regexp.MustCompile(`\D`).ReplaceAllString(sanitized, "")

	return sanitized
}

// GetLast4Digits extracts the last 4 digits from a card number
func (cv *CardValidator) GetLast4Digits(cardNumber string) string {
	sanitized := cv.SanitizeCardNumber(cardNumber)
	if len(sanitized) >= 4 {
		return sanitized[len(sanitized)-4:]
	}
	return sanitized
}

// GetFirst6Digits extracts the first 6 digits (BIN) from a card number
func (cv *CardValidator) GetFirst6Digits(cardNumber string) string {
	sanitized := cv.SanitizeCardNumber(cardNumber)
	if len(sanitized) >= 6 {
		return sanitized[:6]
	}
	return sanitized
}

// isLuhnValid performs the Luhn algorithm check
func (cv *CardValidator) isLuhnValid(cardNumber string) bool {
	// Convert string to slice of integers
	digits := make([]int, len(cardNumber))
	for i, char := range cardNumber {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			return false
		}
		digits[i] = digit
	}

	// Luhn algorithm
	sum := 0
	isEven := false

	// Process from right to left
	for i := len(digits) - 1; i >= 0; i-- {
		digit := digits[i]

		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}
