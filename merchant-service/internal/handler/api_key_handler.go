package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/client"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/service"
	"google.golang.org/grpc/status"
)

type APIKeyHandler struct {
	authClient  *client.AuthServiceClient
	teamService *service.TeamService
}

func NewAPIKeyHandler(authClient *client.AuthServiceClient, teamService *service.TeamService) *APIKeyHandler {
	return &APIKeyHandler{authClient: authClient, teamService: teamService}
}

type CreateAPIKeyRequest struct {
	MerchantID string `json:"merchant_id" binding:"required,uuid"`
	Name       string `json:"name" binding:"required"`
}

func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Get user ID from auth middleware
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid user ID"})
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid merchant ID"})
		return
	}

	hasPermission, err := h.teamService.CheckUserPermission(merchantID, userID, "delete")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "forbidden"})
		return
	}

	resp, err := h.authClient.CreateAPIKey(merchantID, userID, req.Name)
	if err != nil {
		st := status.Convert(err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": st.Message()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"api_key": gin.H{
				"id":         resp.Id,
				"name":       resp.Name,
				"key_prefix": resp.KeyPrefix,
				"created_at": resp.CreatedAt,
			},
			"plain_key": resp.PlainKey,
		},
		"message": resp.Message,
	})
}

func (h *APIKeyHandler) GetMerchantAPIKeys(c *gin.Context) {
	merchantIDStr := c.Param("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid merchant ID"})
		return
	}

	// Get user ID from auth middleware
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid user ID"})
		return
	}

	hasPermission, err := h.teamService.CheckUserPermission(merchantID, userID, "delete")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "forbidden"})
		return
	}

	resp, err := h.authClient.GetMerchantAPIKeys(merchantID)
	if err != nil {
		st := status.Convert(err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": st.Message()})
		return
	}

	// Map to response format
	var apiKeys []gin.H
	for _, key := range resp.ApiKeys {
		apiKeys = append(apiKeys, gin.H{
			"id":           key.Id,
			"name":         key.Name,
			"key_prefix":   key.KeyPrefix,
			"is_active":    key.IsActive,
			"last_used_at": key.LastUsedAt,
			"created_at":   key.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"api_keys": apiKeys},
	})
}

func (h *APIKeyHandler) DeactivateAPIKey(c *gin.Context) {
	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid key ID"})
		return
	}
	merchantIDStr := c.Param("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid merchant ID"})
		return
	}

	// Get user ID from auth middleware
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid user ID"})
		return
	}

	hasPermission, err := h.teamService.CheckUserPermission(merchantID, userID, "delete")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "forbidden"})
		return
	}

	err = h.authClient.DeactivateAPIKey(keyID, merchantID)
	if err != nil {
		st := status.Convert(err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": st.Message()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "API key deactivated successfully"})
}

func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	keyIDStr := c.Param("id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid key ID"})
		return
	}
	merchantIDStr := c.Param("merchant_id")
	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid merchant ID"})
		return
	}

	// Get user ID from auth middleware
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "unauthorized"})
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid user ID"})
		return
	}

	hasPermission, err := h.teamService.CheckUserPermission(merchantID, userID, "delete")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "forbidden"})
		return
	}
	err = h.authClient.DeleteAPIKey(keyID, merchantID)
	if err != nil {
		st := status.Convert(err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": st.Message()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "API key deleted successfully"})
}
