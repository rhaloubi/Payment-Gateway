package validation

import (
	"errors"
	"regexp"
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
	// Visa: Starts with 4, length 13-19
	cv.cardPatterns[model.CardBrandVisa] = regexp.MustCompile(`^4[0-9]{12,18}$`)
	// Mastercard: Starts with 51-55 or 2221-2720, length 16
	cv.cardPatterns[model.CardBrandMastercard] = regexp.MustCompile(`^(?:5[1-5][0-9]{14}|2(?:22[1-9]|2[3-9][0-9]|[3-6][0-9]{2}|7[0-1][0-9]|720)[0-9]{12})$`)

	return cv
}

func (cv *CardValidator) ValidateCard(req CardValidationRequest) error {
	var validationErrors []string

	if err := cv.ValidateCardNumber(req.CardNumber); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if err := cv.ValidateCardholderName(req.CardholderName); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if err := cv.ValidateExpiryDate(req.ExpiryMonth, req.ExpiryYear); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if err := cv.ValidateCVV(req.CVV, req.CardNumber); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

func (cv *CardValidator) ValidateCardNumber(cardNumber string) error {
	sanitized := cv.SanitizeCardNumber(cardNumber)

	if sanitized == "" {
		return errors.New("card number is required")
	}

	if len(sanitized) < 13 || len(sanitized) > 19 {
		return errors.New("card number must be between 13 and 19 digits")
	}

	cardBrand := cv.DetectCardBrand(sanitized)
	if cardBrand == model.CardBrandUnknown {
		return errors.New("unsupported card brand (only Visa and Mastercard are accepted)")
	}

	return nil
}

func (cv *CardValidator) ValidateCardholderName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("cardholder name is required")
	}

	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-\.]{2,100}$`)
	if !nameRegex.MatchString(name) {
		return errors.New("cardholder name contains invalid characters")
	}

	return nil
}

func (cv *CardValidator) ValidateExpiryDate(month, year int) error {
	if month < 1 || month > 12 {
		return errors.New("expiry month must be between 1 and 12")
	}

	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())

	if year < currentYear {
		return errors.New("card has expired")
	}

	if year == currentYear && month < currentMonth {
		return errors.New("card has expired")
	}

	if year > currentYear+20 {
		return errors.New("expiry year is too far in the future")
	}

	return nil
}

func (cv *CardValidator) ValidateCVV(cvv string, cardNumber string) error {
	if strings.TrimSpace(cvv) == "" {
		return errors.New("CVV is required")
	}

	sanitized := strings.ReplaceAll(cvv, " ", "")
	sanitized = strings.ReplaceAll(sanitized, "-", "")

	if !regexp.MustCompile(`^\d+$`).MatchString(sanitized) {
		return errors.New("CVV must contain only digits")
	}

	if len(sanitized) != 3 {
		return errors.New("CVV must be 3 digits")
	}

	return nil
}

func (cv *CardValidator) DetectCardBrand(cardNumber string) model.CardBrand {
	sanitized := cv.SanitizeCardNumber(cardNumber)

	for brand, pattern := range cv.cardPatterns {
		if pattern.MatchString(sanitized) {
			return brand
		}
	}

	return model.CardBrandUnknown
}

func (cv *CardValidator) DetectCardNumberBrand(cardNumber string) model.CardBrand {
	return cv.DetectCardBrand(cardNumber)
}

func (cv *CardValidator) SanitizeCardNumber(cardNumber string) string {
	sanitized := strings.ReplaceAll(cardNumber, " ", "")
	sanitized = strings.ReplaceAll(sanitized, "-", "")

	sanitized = regexp.MustCompile(`\D`).ReplaceAllString(sanitized, "")

	return sanitized
}

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
