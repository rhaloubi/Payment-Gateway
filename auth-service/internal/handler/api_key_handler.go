package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/service"
)

type APIKeyHandler struct {
	apiKeyService *service.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler() *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyService: service.NewAPIKeyService(),
	}
}

// CreateAPIKeyRequest represents API key creation request
type CreateAPIKeyRequest struct {
	MerchantID string `json:"merchant_id" binding:"required,uuid"`
	Name       string `json:"name" binding:"required"`
}

// CreateAPIKey creates a new API key
// POST /api/v1/api-keys
func (h *APIKeyHandler) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get creator from context
	createdBy, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	// Call service
	resp, err := h.apiKeyService.CreateAPIKey(&service.CreateAPIKeyRequest{
		MerchantID: uuid.MustParse(req.MerchantID),
		Name:       req.Name,
		CreatedBy:  uuid.MustParse(createdBy.(string)),
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"api_key": gin.H{
				"id":         resp.APIKey.ID,
				"name":       resp.APIKey.Name,
				"key_prefix": resp.APIKey.KeyPrefix,
				"created_at": resp.APIKey.CreatedAt,
			},
			"plain_key": resp.PlainKey, // ⚠️ Only shown once!
		},
		"message": "⚠️ Save this API key! It won't be shown again.",
	})
}

// GetMerchantAPIKeys gets all API keys for a merchant
// GET /api/v1/api-keys/merchant/:merchant_id
func (h *APIKeyHandler) GetMerchantAPIKeys(c *gin.Context) {
	merchantID := c.Param("merchant_id")

	apiKeys, err := h.apiKeyService.GetMerchantAPIKeys(uuid.MustParse(merchantID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch API keys",
		})
		return
	}

	// Don't return key hashes
	var response []gin.H
	for _, key := range apiKeys {
		response = append(response, gin.H{
			"id":           key.ID,
			"name":         key.Name,
			"key_prefix":   key.KeyPrefix,
			"is_active":    key.IsActive,
			"last_used_at": key.LastUsedAt,
			"created_at":   key.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"api_keys": response,
		},
	})
}

// DeactivateAPIKey deactivates an API key
// PATCH /api/v1/api-keys/:id/deactivate
func (h *APIKeyHandler) DeactivateAPIKey(c *gin.Context) {
	keyID := c.Param("id")

	if err := h.apiKeyService.DeactivateAPIKey(uuid.MustParse(keyID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API key deactivated successfully",
	})
}

// DeleteAPIKey deletes an API key
// DELETE /api/v1/api-keys/:id
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
	keyID := c.Param("id")

	if err := h.apiKeyService.DeleteAPIKey(uuid.MustParse(keyID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "API key deleted successfully",
	})
}
