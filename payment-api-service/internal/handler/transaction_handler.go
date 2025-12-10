package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/service"
	pb "github.com/rhaloubi/payment-gateway/payment-api-service/proto"
)

type TransactionHandler struct {
	transactionService *service.TransactionService
}

func NewTransactionHandler() (*TransactionHandler, error) {
	transactionService, err := service.NewTransactionService()
	if err != nil {
		return nil, err
	}

	return &TransactionHandler{
		transactionService: transactionService,
	}, nil
}

func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	// Get transaction ID from request
	transactionID := c.Param("id")
	if transactionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "transaction ID is required",
		})
		return
	}

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "invalid merchant context",
		})
		return
	}
	serviceReq := &pb.GetTransactionRequest{
		TransactionId: transactionID,
		MerchantId:    merchantID.String(),
	}
	resp, err := h.transactionService.GetTransaction(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

func (h *TransactionHandler) ListTransactions(c *gin.Context) {

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "invalid merchant context",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	serviceReq := &pb.ListTransactionsRequest{
		MerchantId: merchantID.String(),
		Status:     c.Query("status"),
		Limit:      int32(limit),
		Offset:     int32(offset),
	}
	resp, err := h.transactionService.ListTransactions(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}
