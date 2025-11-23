package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/crypto"
	model "github.com/rhaloubi/payment-gateway/tokenization-service/internal/models"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/repository"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/validation"
	"go.uber.org/zap"
)

type TokenizationService struct {
	cardVaultRepo     *repository.CardVaultRepository
	tokenReqRepo      *repository.TokenizationRequestRepository
	tokenUsageRepo    *repository.TokenUsageLogRepository
	keyRepo           *repository.EncryptionKeyRepository
	binRepo           *repository.CardBINRepository
	encryptionService *crypto.EncryptionService
	validationService *validation.CardValidator
	keyManagementSvc  *KeyManagementService
}

func NewTokenizationService() *TokenizationService {
	return &TokenizationService{
		cardVaultRepo:     repository.NewCardVaultRepository(),
		tokenReqRepo:      repository.NewTokenizationRequestRepository(),
		tokenUsageRepo:    repository.NewTokenUsageLogRepository(),
		keyRepo:           repository.NewEncryptionKeyRepository(),
		binRepo:           repository.NewCardBINRepository(),
		encryptionService: crypto.NewEncryptionService(),
		validationService: validation.NewCardValidator(),
		keyManagementSvc:  NewKeyManagementService(),
	}
}

type TokenizeCardRequest struct {
	MerchantID     uuid.UUID
	CardNumber     string
	CardholderName string
	ExpiryMonth    int
	ExpiryYear     int
	CVV            string

	IsSingleUse bool
	ExpiresAt   *time.Time

	RequestID string
	IPAddress string
	UserAgent string
	CreatedBy uuid.UUID
}

type TokenizeCardResponse struct {
	Token       string
	CardBrand   model.CardBrand
	CardType    model.CardType
	Last4Digits string
	ExpiryMonth int
	ExpiryYear  int
	Fingerprint string
	IsNewToken  bool // true if new, false if returning existing token
}
type DetokenizeRequest struct {
	Token      string
	MerchantID uuid.UUID

	TransactionID uuid.UUID
	UsageType     string
	Amount        int64
	Currency      string
	IPAddress     string
	UserAgent     string
}

type DetokenizeResponse struct {
	CardNumber     string
	CardholderName string
	ExpiryMonth    int
	ExpiryYear     int
	CardBrand      model.CardBrand
	Last4Digits    string
}

func (s *TokenizationService) TokenizeCard(req *TokenizeCardRequest) (*TokenizeCardResponse, error) {
	startTime := time.Now()

	if err := s.validateCardData(req); err != nil {
		go s.logTokenizationRequest(req, nil, false, err, time.Since(startTime))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	req.CardNumber = s.validationService.SanitizeCardNumber(req.CardNumber)

	cardBrand := s.validationService.DetectCardBrand(req.CardNumber)

	fingerprint := s.encryptionService.GenerateCardFingerprint(
		req.CardNumber,
		strconv.Itoa(req.ExpiryMonth),
		strconv.Itoa(req.ExpiryYear),
	)

	existingCard, err := s.cardVaultRepo.FindByFingerprint(req.MerchantID, fingerprint)
	if err != nil {
		logger.Log.Error("Error checking for duplicate", zap.Error(err))
	}

	if existingCard != nil && existingCard.IsValid() {
		logger.Log.Info("Returning existing token for duplicate card",
			zap.String("token", existingCard.Token),
			zap.String("merchant_id", req.MerchantID.String()),
		)

		response := &TokenizeCardResponse{
			Token:       existingCard.Token,
			CardBrand:   existingCard.CardBrand,
			CardType:    existingCard.CardType,
			Last4Digits: existingCard.Last4Digits,
			ExpiryMonth: existingCard.ExpiryMonth,
			ExpiryYear:  existingCard.ExpiryYear,
			Fingerprint: existingCard.Fingerprint,
			IsNewToken:  false,
		}

		go s.logTokenizationRequest(req, existingCard, true, nil, time.Since(startTime))
		return response, nil
	}

	encryptionKey, keyID, err := s.keyManagementSvc.GetOrCreateMerchantKey(req.MerchantID)
	if err != nil {
		go s.logTokenizationRequest(req, nil, false, err, time.Since(startTime))
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	encryptedData, err := s.encryptionService.EncryptCardData(crypto.CardData{
		CardNumber:     req.CardNumber,
		CardholderName: req.CardholderName,
		ExpiryMonth:    strconv.Itoa(req.ExpiryMonth),
		ExpiryYear:     strconv.Itoa(req.ExpiryYear),
	}, encryptionKey)
	if err != nil {
		go s.logTokenizationRequest(req, nil, false, err, time.Since(startTime))
		return nil, fmt.Errorf("encryption failed: %w", err)
	}

	token := s.generateToken("live")

	last4 := s.validationService.GetLast4Digits(req.CardNumber)
	first6 := s.validationService.GetFirst6Digits(req.CardNumber)

	binInfo, _ := s.binRepo.FindByBIN(first6)
	cardType := model.CardTypeUnknown
	if binInfo != nil {
		cardType = binInfo.CardType
	}

	cardVault := &model.CardVault{
		MerchantID:              req.MerchantID,
		Token:                   token,
		TokenPrefix:             s.getTokenPrefix(token),
		EncryptedCardNumber:     encryptedData.EncryptedCardNumber,
		EncryptedCardholderName: encryptedData.EncryptedCardholderName,
		EncryptedExpiryMonth:    encryptedData.EncryptedExpiryMonth,
		EncryptedExpiryYear:     encryptedData.EncryptedExpiryYear,
		KeyID:                   keyID,
		EncryptionKeyVersion:    1,
		Last4Digits:             last4,
		First6Digits:            first6,
		CardBrand:               cardBrand,
		CardType:                cardType,
		ExpiryMonth:             req.ExpiryMonth,
		ExpiryYear:              req.ExpiryYear,
		Fingerprint:             fingerprint,
		Status:                  model.TokenStatusActive,
		IsSingleUse:             req.IsSingleUse,
		CreatedBy:               req.CreatedBy,
	}

	if req.ExpiresAt != nil {
		cardVault.ExpiresAt.Time = *req.ExpiresAt
		cardVault.ExpiresAt.Valid = true
	}

	if err := s.cardVaultRepo.Create(cardVault); err != nil {
		go s.logTokenizationRequest(req, nil, false, err, time.Since(startTime))
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	s.keyRepo.IncrementEncryptedRecords(keyID)

	go s.logTokenizationRequest(req, cardVault, true, nil, time.Since(startTime))

	response := &TokenizeCardResponse{
		Token:       cardVault.Token,
		CardBrand:   cardVault.CardBrand,
		CardType:    cardVault.CardType,
		Last4Digits: cardVault.Last4Digits,
		ExpiryMonth: cardVault.ExpiryMonth,
		ExpiryYear:  cardVault.ExpiryYear,
		Fingerprint: cardVault.Fingerprint,
		IsNewToken:  true,
	}

	logger.Log.Info("Card tokenized successfully",
		zap.String("token", token),
		zap.String("merchant_id", req.MerchantID.String()),
		zap.String("card_brand", string(cardBrand)),
	)

	return response, nil
}

func (s *TokenizationService) Detokenize(req *DetokenizeRequest) (*DetokenizeResponse, error) {
	cardVault, err := s.cardVaultRepo.FindByToken(req.Token)
	if err != nil {
		return nil, fmt.Errorf("token not found: %w", err)
	}

	if cardVault.MerchantID != req.MerchantID {
		logger.Log.Warn("Attempted access to token from different merchant",
			zap.String("token", req.Token),
			zap.String("requesting_merchant", req.MerchantID.String()),
			zap.String("token_owner", cardVault.MerchantID.String()),
		)
		return nil, errors.New("access denied: token does not belong to merchant")
	}

	// Step 3: Validate token status
	if !cardVault.IsValid() {
		s.logTokenUsage(cardVault, req, false, errors.New("token invalid or expired"))
		return nil, errors.New("token is invalid, expired, or revoked")
	}

	// Step 4: Check if single-use token was already used
	if cardVault.IsSingleUse && cardVault.UsageCount > 0 {
		s.logTokenUsage(cardVault, req, false, errors.New("single-use token already consumed"))
		return nil, errors.New("single-use token has already been used")
	}

	// Step 5: Get decryption key
	decryptionKey, err := s.keyManagementSvc.GetKeyByID(cardVault.KeyID)
	if err != nil {
		s.logTokenUsage(cardVault, req, false, err)
		return nil, fmt.Errorf("failed to get decryption key: %w", err)
	}

	// Step 6: Decrypt card data
	decryptedData, err := s.encryptionService.DecryptCardData(crypto.EncryptedCardData{
		EncryptedCardNumber:     cardVault.EncryptedCardNumber,
		EncryptedCardholderName: cardVault.EncryptedCardholderName,
		EncryptedExpiryMonth:    cardVault.EncryptedExpiryMonth,
		EncryptedExpiryYear:     cardVault.EncryptedExpiryYear,
	}, decryptionKey)
	if err != nil {
		s.logTokenUsage(cardVault, req, false, err)
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	// Step 7: Update usage tracking
	s.cardVaultRepo.IncrementUsageCount(req.Token)
	s.cardVaultRepo.SetFirstUsed(req.Token)

	// Step 8: If single-use, mark as used
	if cardVault.IsSingleUse {
		s.cardVaultRepo.UpdateStatus(req.Token, model.TokenStatusUsed)
	}

	// Step 9: Log token usage
	s.logTokenUsage(cardVault, req, true, nil)

	// Step 10: Convert expiry strings to integers
	expiryMonth, _ := strconv.Atoi(decryptedData.ExpiryMonth)
	expiryYear, _ := strconv.Atoi(decryptedData.ExpiryYear)

	// Step 11: Return decrypted data
	response := &DetokenizeResponse{
		CardNumber:     decryptedData.CardNumber,
		CardholderName: decryptedData.CardholderName,
		ExpiryMonth:    expiryMonth,
		ExpiryYear:     expiryYear,
		CardBrand:      cardVault.CardBrand,
		Last4Digits:    cardVault.Last4Digits,
	}

	logger.Log.Info("Token detokenized successfully",
		zap.String("token", req.Token),
		zap.String("merchant_id", req.MerchantID.String()),
		zap.String("usage_type", req.UsageType),
	)

	return response, nil
}

func (s *TokenizationService) ValidateToken(token string, merchantID uuid.UUID) (bool, error) {
	cardVault, err := s.cardVaultRepo.FindByToken(token)
	if err != nil {
		return false, err
	}

	// Verify merchant ownership
	if cardVault.MerchantID != merchantID {
		return false, errors.New("access denied")
	}

	return cardVault.IsValid(), nil
}

// RevokeToken revokes a token
func (s *TokenizationService) RevokeToken(token string, merchantID uuid.UUID, revokedBy uuid.UUID, reason string) error {
	// Verify token exists and belongs to merchant
	cardVault, err := s.cardVaultRepo.FindByToken(token)
	if err != nil {
		return err
	}

	if cardVault.MerchantID != merchantID {
		return errors.New("access denied: token does not belong to merchant")
	}

	// Revoke the token
	err = s.cardVaultRepo.RevokeToken(token, revokedBy, reason)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	logger.Log.Info("Token revoked",
		zap.String("token", token),
		zap.String("merchant_id", merchantID.String()),
		zap.String("reason", reason),
	)

	return nil
}

// GetTokenInfo retrieves token metadata (without decrypting)
func (s *TokenizationService) GetTokenInfo(token string, merchantID uuid.UUID) (*model.CardVault, error) {
	cardVault, err := s.cardVaultRepo.FindByToken(token)
	if err != nil {
		return nil, err
	}

	if cardVault.MerchantID != merchantID {
		return nil, errors.New("access denied")
	}

	return cardVault, nil
}

func (s *TokenizationService) validateCardData(req *TokenizeCardRequest) error {
	validationReq := validation.CardValidationRequest{
		CardNumber:     req.CardNumber,
		CardholderName: req.CardholderName,
		ExpiryMonth:    req.ExpiryMonth,
		ExpiryYear:     req.ExpiryYear,
		CVV:            req.CVV,
	}

	return s.validationService.ValidateCard(validationReq)
}

func (s *TokenizationService) generateToken(environment string) string {
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)

	// Convert to hex string
	randomString := hex.EncodeToString(randomBytes)

	// Format: tok_{environment}_{random}
	return fmt.Sprintf("tok_%s_%s", environment, randomString)
}

// getTokenPrefix extracts the prefix from token
func (s *TokenizationService) getTokenPrefix(token string) string {
	if len(token) > 20 {
		return token[0:8] // "tok_live" or "tok_test"
	}
	return token
}

// logTokenizationRequest logs a tokenization attempt
func (s *TokenizationService) logTokenizationRequest(
	req *TokenizeCardRequest,
	cardVault *model.CardVault,
	success bool,
	err error,
	processingTime time.Duration,
) {
	log := &model.TokenizationRequest{
		MerchantID:     req.MerchantID,
		RequestID:      req.RequestID,
		IPAddress:      req.IPAddress,
		UserAgent:      toNullString(req.UserAgent),
		Success:        success,
		ProcessingTime: processingTime.Milliseconds(),
	}

	if cardVault != nil {
		log.TokenID = cardVault.ID
		log.CardBrand = cardVault.CardBrand
		log.Last4Digits = cardVault.Last4Digits
		log.ExpiryMonth = cardVault.ExpiryMonth
		log.ExpiryYear = cardVault.ExpiryYear
	}

	if err != nil {
		log.ErrorMessage.String = err.Error()
		log.ErrorMessage.Valid = true
		log.ErrorCode.String = "TOKENIZATION_ERROR"
		log.ErrorCode.Valid = true
	}

	s.tokenReqRepo.Create(log)
}

// logTokenUsage logs token usage in a transaction
func (s *TokenizationService) logTokenUsage(
	cardVault *model.CardVault,
	req *DetokenizeRequest,
	success bool,
	err error,
) {
	log := &model.TokenUsageLog{
		TokenID:         cardVault.ID,
		MerchantID:      req.MerchantID,
		TransactionID:   req.TransactionID,
		TransactionType: req.UsageType,
		Amount:          req.Amount,
		Currency:        req.Currency,
		UsageType:       req.UsageType,
		IPAddress:       req.IPAddress,
		UserAgent:       toNullString(req.UserAgent),
		Success:         success,
	}

	if err != nil {
		log.ErrorCode.String = err.Error()
		log.ErrorCode.Valid = true
	}

	s.tokenUsageRepo.Create(log)
}

// toNullString converts string to sql.NullString
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
