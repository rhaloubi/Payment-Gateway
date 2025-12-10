package handler

import (
	"net/http"

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

func (h *TransactionHandler) GetTransaction(c *gin.Context, req *pb.GetTransactionRequest) {
	// Get transaction ID from request
	transactionID := req.TransactionId
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

func (h *TransactionHandler) GetTransactions(c *gin.Context, req *pb.ListTransactionsRequest) {

	merchantIDStr, _ := c.Get("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "invalid merchant context",
		})
		return
	}
	serviceReq := &pb.ListTransactionsRequest{
		MerchantId: merchantID.String(),
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
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
