package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/service"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
	webhookService *service.WebhookService
}

func NewPaymentHandler() (*PaymentHandler, error) {
	paymentService, err := service.NewPaymentService()
	if err != nil {
		return nil, err
	}

	return &PaymentHandler{
		paymentService: paymentService,
		webhookService: service.NewWebhookService(),
	}, nil
}

type CardRequest struct {
	Number         string `json:"number" binding:"required,min=13,max=19"`
	CardholderName string `json:"cardholder_name" binding:"required"`
	ExpMonth       int    `json:"exp_month" binding:"required,min=1,max=12"`
	ExpYear        int    `json:"exp_year" binding:"required,min=2024"`
	CVV            string `json:"cvv" binding:"required,min=3,max=4"`
}

type CustomerRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
	Name  string `json:"name"`
}

type AuthorizeRequest struct {
	Amount      int64                  `json:"amount" binding:"required,min=1"`
	Currency    string                 `json:"currency" binding:"required,len=3"`
	Card        CardRequest            `json:"card" binding:"required"`
	Customer    CustomerRequest        `json:"customer"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type CaptureRequest struct {
	Amount int64 `json:"amount" binding:"required,min=1"`
}

type VoidRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type RefundRequest struct {
	Amount int64  `json:"amount" binding:"required,min=1"`
	Reason string `json:"reason" binding:"required"`
}

// =========================================================================
// POST /v1/payments/authorize
// =========================================================================

func (h *PaymentHandler) AuthorizePayment(c *gin.Context) {
	var req AuthorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	// Validate currency
	if req.Currency != "USD" && req.Currency != "EUR" && req.Currency != "MAD" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "unsupported currency (only USD, EUR, and MAD supported)",
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

	// Get idempotency key
	idempotencyKey := c.GetHeader("Idempotency-Key")

	// Build service request
	serviceReq := &service.AuthorizePaymentRequest{
		MerchantID:     merchantID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		CardNumber:     req.Card.Number,
		CardholderName: req.Card.CardholderName,
		ExpMonth:       req.Card.ExpMonth,
		ExpYear:        req.Card.ExpYear,
		CVV:            req.Card.CVV,
		CustomerEmail:  req.Customer.Email,
		CustomerName:   req.Customer.Name,
		Description:    req.Description,
		Metadata:       req.Metadata,
		IdempotencyKey: idempotencyKey,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
	}

	// Process authorization
	response, err := h.paymentService.AuthorizePayment(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Log.Error("Authorization failed",
			zap.Error(err),
			zap.String("merchant_id", merchantID.String()),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// TODO: Send webhook (if configured)
	// webhookURL := getMerchantWebhookURL(merchantID)
	// if webhookURL != "" {
	//     h.webhookService.SendPaymentWebhook(...)
	// }

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// =========================================================================
// POST /v1/payments/sale
// =========================================================================

func (h *PaymentHandler) SalePayment(c *gin.Context) {
	var req AuthorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	idempotencyKey := c.GetHeader("Idempotency-Key")

	serviceReq := &service.AuthorizePaymentRequest{
		MerchantID:     merchantID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		CardNumber:     req.Card.Number,
		CardholderName: req.Card.CardholderName,
		ExpMonth:       req.Card.ExpMonth,
		ExpYear:        req.Card.ExpYear,
		CVV:            req.Card.CVV,
		CustomerEmail:  req.Customer.Email,
		CustomerName:   req.Customer.Name,
		Description:    req.Description,
		Metadata:       req.Metadata,
		IdempotencyKey: idempotencyKey,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
	}

	// Process sale (authorize + capture)
	response, err := h.paymentService.SalePayment(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Log.Error("Sale failed", zap.Error(err))
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
// POST /v1/payments/:id/capture
// =========================================================================

func (h *PaymentHandler) CapturePayment(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid payment ID",
		})
		return
	}

	var req CaptureRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	response, err := h.paymentService.CapturePayment(c.Request.Context(), paymentID, merchantID, req.Amount)
	if err != nil {
		logger.Log.Error("Capture failed", zap.Error(err))
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
// POST /v1/payments/:id/void
// =========================================================================

func (h *PaymentHandler) VoidPayment(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid payment ID",
		})
		return
	}

	var req VoidRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	response, err := h.paymentService.VoidPayment(c.Request.Context(), paymentID, merchantID, req.Reason)
	if err != nil {
		logger.Log.Error("Void failed", zap.Error(err))
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
// POST /v1/payments/:id/refund
// =========================================================================

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid payment ID",
		})
		return
	}

	var req RefundRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid request: " + err.Error(),
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	response, err := h.paymentService.RefundPayment(c.Request.Context(), paymentID, merchantID, req.Amount, req.Reason)
	if err != nil {
		logger.Log.Error("Refund failed", zap.Error(err))
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
// GET /v1/payments/:id
// =========================================================================

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid payment ID",
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, _ := uuid.Parse(merchantIDStr.(string))

	payment, err := h.paymentService.GetPayment(paymentID, merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "payment not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    payment,
	})
}
