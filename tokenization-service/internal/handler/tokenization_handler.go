package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/service"
	"go.uber.org/zap"
)

type TokenizationHandler struct {
	tokenizationService *service.TokenizationService
	keyManagementSvc    *service.KeyManagementService
}

func NewTokenizationHandler() *TokenizationHandler {
	return &TokenizationHandler{
		tokenizationService: service.NewTokenizationService(),
		keyManagementSvc:    service.NewKeyManagementService(),
	}
}

type TokenizeRequest struct {
	CardNumber     string `json:"card_number" binding:"required"`
	CardholderName string `json:"cardholder_name" binding:"required"`
	ExpMonth       int    `json:"exp_month" binding:"required,min=1,max=12"`
	ExpYear        int    `json:"exp_year" binding:"required,min=2024"`
	CVV            string `json:"cvv" binding:"required,min=3,max=4"`

	// Optional
	IsSingleUse bool `json:"is_single_use"`
	ExpiresIn   int  `json:"expires_in"`
}

type TokenizeResponse struct {
	Token      string       `json:"token"`
	Card       CardMetadata `json:"card"`
	IsNewToken bool         `json:"is_new_token"`
}

type CardMetadata struct {
	Brand       string `json:"brand"`
	Type        string `json:"type"`
	Last4       string `json:"last4"`
	ExpMonth    int    `json:"exp_month"`
	ExpYear     int    `json:"exp_year"`
	Fingerprint string `json:"fingerprint"`
}

type ValidateTokenResponse struct {
	Valid       bool         `json:"valid"`
	Card        CardMetadata `json:"card,omitempty"`
	Status      string       `json:"status"`
	UsageCount  int          `json:"usage_count"`
	IsSingleUse bool         `json:"is_single_use"`
}

type RevokeTokenRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// TokenizeCard handles POST /v1/tokenize
func (h *TokenizationHandler) TokenizeCard(c *gin.Context) {
	var req TokenizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	// Get merchant ID from auth middleware
	merchantIDStr, exists := c.Get("merchant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "merchant context required",
		})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant_id",
		})
		return
	}

	// Get user ID (for audit)
	userIDStr, _ := c.Get("user_id")
	var createdBy uuid.UUID
	if userIDStr != nil {
		createdBy, _ = uuid.Parse(userIDStr.(string))
	}

	// Get request ID
	requestID, _ := c.Get("request_id")
	requestIDStr := ""
	if requestID != nil {
		requestIDStr = requestID.(string)
	}

	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		expiry := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiresAt = &expiry
	}

	// Build service request
	serviceReq := &service.TokenizeCardRequest{
		MerchantID:     merchantID,
		CardNumber:     req.CardNumber,
		CardholderName: req.CardholderName,
		ExpiryMonth:    req.ExpMonth,
		ExpiryYear:     req.ExpYear,
		CVV:            req.CVV,
		IsSingleUse:    req.IsSingleUse,
		ExpiresAt:      expiresAt,
		RequestID:      requestIDStr,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		CreatedBy:      createdBy,
	}

	response, err := h.tokenizationService.TokenizeCard(serviceReq)
	if err != nil {
		logger.Log.Error("Tokenization failed",
			zap.Error(err),
			zap.String("merchant_id", merchantID.String()),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Build response
	apiResponse := TokenizeResponse{
		Token: response.Token,
		Card: CardMetadata{
			Brand:       string(response.CardBrand),
			Type:        string(response.CardType),
			Last4:       response.Last4Digits,
			ExpMonth:    response.ExpiryMonth,
			ExpYear:     response.ExpiryYear,
			Fingerprint: response.Fingerprint,
		},
		IsNewToken: response.IsNewToken,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apiResponse,
	})
}

// ValidateToken handles GET /v1/tokens/:token/validate
func (h *TokenizationHandler) ValidateToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "token parameter required",
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	// Validate token
	isValid, err := h.tokenizationService.ValidateToken(token, merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "token not found",
		})
		return
	}

	// Get token info
	tokenInfo, err := h.tokenizationService.GetTokenInfo(token, merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "token not found",
		})
		return
	}

	response := ValidateTokenResponse{
		Valid: isValid,
		Card: CardMetadata{
			Brand:       string(tokenInfo.CardBrand),
			Type:        string(tokenInfo.CardType),
			Last4:       tokenInfo.Last4Digits,
			ExpMonth:    tokenInfo.ExpiryMonth,
			ExpYear:     tokenInfo.ExpiryYear,
			Fingerprint: tokenInfo.Fingerprint,
		},
		Status:      string(tokenInfo.Status),
		UsageCount:  tokenInfo.UsageCount,
		IsSingleUse: tokenInfo.IsSingleUse,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetTokenInfo handles GET /v1/tokens/:token
func (h *TokenizationHandler) GetTokenInfo(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "token parameter required",
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	// Get token info
	tokenInfo, err := h.tokenizationService.GetTokenInfo(token, merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "token not found or access denied",
		})
		return
	}

	response := map[string]interface{}{
		"token": tokenInfo.Token,
		"card": CardMetadata{
			Brand:       string(tokenInfo.CardBrand),
			Type:        string(tokenInfo.CardType),
			Last4:       tokenInfo.Last4Digits,
			ExpMonth:    tokenInfo.ExpiryMonth,
			ExpYear:     tokenInfo.ExpiryYear,
			Fingerprint: tokenInfo.Fingerprint,
		},
		"status":        string(tokenInfo.Status),
		"usage_count":   tokenInfo.UsageCount,
		"is_single_use": tokenInfo.IsSingleUse,
		"created_at":    tokenInfo.CreatedAt,
	}

	if tokenInfo.ExpiresAt.Valid {
		response["expires_at"] = tokenInfo.ExpiresAt.Time
	}

	if tokenInfo.FirstUsedAt.Valid {
		response["first_used_at"] = tokenInfo.FirstUsedAt.Time
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// RevokeToken handles DELETE /v1/tokens/:token
func (h *TokenizationHandler) RevokeToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "token parameter required",
		})
		return
	}

	var req RevokeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: reason required",
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	userIDStr, _ := c.Get("user_id")
	var revokedBy uuid.UUID
	if userIDStr != nil {
		revokedBy, _ = uuid.Parse(userIDStr.(string))
	}

	// Revoke token
	err := h.tokenizationService.RevokeToken(token, merchantID, revokedBy, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "token revoked successfully",
	})
}

// GetKeyStatistics handles GET /v1/keys/statistics
func (h *TokenizationHandler) GetKeyStatistics(c *gin.Context) {
	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	stats, err := h.keyManagementSvc.GetKeyStatistics(merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to get statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func (h *TokenizationHandler) RotateKey(c *gin.Context) {
	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	userIDStr, _ := c.Get("user_id")
	rotatedBy, _ := uuid.Parse(userIDStr.(string))

	newKeyID, err := h.keyManagementSvc.RotateMerchantKey(merchantID, rotatedBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to rotate key: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]string{
			"new_key_id": newKeyID,
			"message":    "encryption key rotated successfully",
		},
	})
}

// =========================================================================
// Internal Endpoints (For Service-to-Service Communication)
// =========================================================================

type DetokenizeRequest struct {
	Token         string    `json:"token" binding:"required"`
	MerchantID    uuid.UUID `json:"merchant_id" binding:"required"`
	TransactionID uuid.UUID `json:"transaction_id"`
	UsageType     string    `json:"usage_type"` // "payment", "verification", "recurring"
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
}

type DetokenizeResponse struct {
	CardNumber     string `json:"card_number"`
	CardholderName string `json:"cardholder_name"`
	ExpMonth       int    `json:"exp_month"`
	ExpYear        int    `json:"exp_year"`
	CardBrand      string `json:"card_brand"`
	Last4          string `json:"last4"`
}

// Detokenize handles POST /internal/v1/detokenize (internal service only)
func (h *TokenizationHandler) Detokenize(c *gin.Context) {
	var req DetokenizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	// Build service request
	serviceReq := &service.DetokenizeRequest{
		Token:         req.Token,
		MerchantID:    req.MerchantID,
		TransactionID: req.TransactionID,
		UsageType:     req.UsageType,
		Amount:        req.Amount,
		Currency:      req.Currency,
		IPAddress:     c.ClientIP(),
		UserAgent:     c.Request.UserAgent(),
	}

	// Detokenize
	response, err := h.tokenizationService.Detokenize(serviceReq)
	if err != nil {
		logger.Log.Error("Detokenization failed",
			zap.Error(err),
			zap.String("token", req.Token),
			zap.String("merchant_id", req.MerchantID.String()),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Build response
	apiResponse := DetokenizeResponse{
		CardNumber:     response.CardNumber,
		CardholderName: response.CardholderName,
		ExpMonth:       response.ExpiryMonth,
		ExpYear:        response.ExpiryYear,
		CardBrand:      string(response.CardBrand),
		Last4:          response.Last4Digits,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apiResponse,
	})
}
