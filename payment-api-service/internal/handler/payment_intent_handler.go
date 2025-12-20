package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/payment-api-service/internal/models"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/service"
	"go.uber.org/zap"
)

type PaymentIntentHandler struct {
	intentService *service.PaymentIntentService
}

func NewPaymentIntentHandler(paymentService *service.PaymentService) *PaymentIntentHandler {
	return &PaymentIntentHandler{
		intentService: service.NewPaymentIntentService(paymentService),
	}
}

// =========================================================================
// Request/Response DTOs
// =========================================================================

type CreateIntentRequest struct {
	Amount        int64                  `json:"amount" binding:"required,min=1"`
	Currency      string                 `json:"currency" binding:"required,len=3"`
	OrderID       string                 `json:"order_id"`
	Description   string                 `json:"description"`
	CaptureMethod model.CaptureMethod    `json:"capture_method"` // "automatic" or "manual"
	SuccessURL    string                 `json:"success_url" binding:"required,url"`
	CancelURL     string                 `json:"cancel_url" binding:"omitempty,url"`
	CustomerEmail string                 `json:"customer_email" binding:"omitempty,email"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type ConfirmIntentRequest struct {
	Card struct {
		Number         string `json:"number" binding:"required,min=13,max=19"`
		CardholderName string `json:"cardholder_name" binding:"required"`
		ExpMonth       int    `json:"exp_month" binding:"required,min=1,max=12"`
		ExpYear        int    `json:"exp_year" binding:"required,min=2024"`
		CVV            string `json:"cvv" binding:"required,min=3,max=4"`
	} `json:"card" binding:"required"`
	CustomerEmail string `json:"customer_email" binding:"omitempty,email"`
}

// =========================================================================
// POST /payment-intents (Server-to-Server - Requires API Key)
// =========================================================================

func (h *PaymentIntentHandler) CreatePaymentIntent(c *gin.Context) {
	var req CreateIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	// Get merchant ID from auth middleware
	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "invalid merchant context",
		})
		return
	}

	// Create payment intent
	serviceReq := &service.CreatePaymentIntentRequest{
		MerchantID:    merchantID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		OrderID:       req.OrderID,
		Description:   req.Description,
		CaptureMethod: req.CaptureMethod,
		SuccessURL:    req.SuccessURL,
		CancelURL:     req.CancelURL,
		CustomerEmail: req.CustomerEmail,
		Metadata:      req.Metadata,
	}

	response, err := h.intentService.CreatePaymentIntent(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Log.Error("Failed to create payment intent",
			zap.Error(err),
			zap.String("merchant_id", merchantID.String()),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// =========================================================================
// GET /payment-intents/:id (Browser-Safe - No Auth Required)
// =========================================================================

func (h *PaymentIntentHandler) GetPaymentIntent(c *gin.Context) {
	intentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid payment_intent_id",
		})
		return
	}

	response, err := h.intentService.GetPaymentIntent(c.Request.Context(), intentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "payment intent not found",
		})
		return
	}

	// Return ONLY safe data (no client_secret)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":         response.ID,
			"status":     response.Status,
			"amount":     response.Amount,
			"currency":   response.Currency,
			"expires_at": response.ExpiresAt,
		},
	})
}

// =========================================================================
// POST /payment-intents/:id/confirm (Browser - Requires client_secret)
// =========================================================================

func (h *PaymentIntentHandler) ConfirmPaymentIntent(c *gin.Context) {
	intentID := c.Param("id")

	var req ConfirmIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	// Get client_secret from header or body
	clientSecret := c.GetHeader("X-Client-Secret")
	if clientSecret == "" {
		clientSecret = c.Query("client_secret")
	}

	if clientSecret == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "client_secret is required",
		})
		return
	}

	// Confirm payment
	serviceReq := &service.ConfirmPaymentIntentRequest{
		PaymentIntentID: intentID,
		ClientSecret:    clientSecret,
		CardNumber:      req.Card.Number,
		CardholderName:  req.Card.CardholderName,
		ExpMonth:        req.Card.ExpMonth,
		ExpYear:         req.Card.ExpYear,
		CVV:             req.Card.CVV,
		CustomerEmail:   req.CustomerEmail,
		IPAddress:       c.ClientIP(),
		UserAgent:       c.Request.UserAgent(),
	}

	response, err := h.intentService.ConfirmPaymentIntent(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Log.Error("Failed to confirm payment intent",
			zap.Error(err),
			zap.String("intent_id", intentID),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// =========================================================================
// POST /payment-intents/:id/cancel (Requires API Key)
// =========================================================================

func (h *PaymentIntentHandler) CancelPaymentIntent(c *gin.Context) {
	intentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid payment_intent_id",
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	err = h.intentService.CancelPaymentIntent(c.Request.Context(), intentID, merchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "payment intent canceled",
	})
}
