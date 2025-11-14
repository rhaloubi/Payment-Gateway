package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	service "github.com/rhaloubi/payment-gateway/merchant-service/internal/service"
)

type SettingsHandler struct {
	settingsService *service.SettingsService
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{
		settingsService: service.NewSettingsService(),
	}
}

// UpdateSettingsRequest represents settings update request
type UpdateSettingsRequest struct {
	DefaultCurrency   string `json:"default_currency" binding:"omitempty,len=3"`
	AutoSettle        *bool  `json:"auto_settle"`
	SettleSchedule    string `json:"settle_schedule" binding:"omitempty,oneof=daily weekly monthly"`
	WebhookURL        string `json:"webhook_url" binding:"omitempty,url"`
	NotificationEmail string `json:"notification_email" binding:"omitempty,email"`
	SendEmailReceipts *bool  `json:"send_email_receipts"`
}

// GET /api/v1/merchants/:id/settings
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	settings, err := h.settingsService.GetSettings(merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "settings not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"settings": settings,
		},
	})
}

// PATCH /api/v1/merchants/:id/settings
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	// Prepare updates
	updates := make(map[string]interface{})
	if req.DefaultCurrency != "" {
		updates["default_currency"] = req.DefaultCurrency
	}
	if req.AutoSettle != nil {
		updates["auto_settle"] = *req.AutoSettle
	}
	if req.SettleSchedule != "" {
		updates["settle_schedule"] = req.SettleSchedule
	}
	if req.WebhookURL != "" {
		updates["webhook_url"] = req.WebhookURL
	}

	// Update settings
	if err := h.settingsService.UpdateSettings(merchantID, updates, userUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated successfully",
	})
}
