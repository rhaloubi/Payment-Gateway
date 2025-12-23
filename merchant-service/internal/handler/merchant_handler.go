package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	service "github.com/rhaloubi/payment-gateway/merchant-service/internal/service"
)

// MerchantHandler handles merchant HTTP requests
type MerchantHandler struct {
	merchantService *service.MerchantService
	teamService     *service.TeamService
}

// NewMerchantHandler creates a new merchant handler
func NewMerchantHandler() *MerchantHandler {
	return &MerchantHandler{
		merchantService: service.NewMerchantService(),
		teamService:     service.NewTeamService(),
	}
}

// CreateMerchantRequest represents merchant creation request
type CreateMerchantRequest struct {
	BusinessName string `json:"business_name" binding:"required"`
	LegalName    string `json:"legal_name"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	BusinessType string `json:"business_type" binding:"required,oneof=individual sole_proprietor partnership corporation non_profit"`
}

// UpdateMerchantRequest represents merchant update request
type UpdateMerchantRequest struct {
	BusinessName string `json:"business_name"`
	Email        string `json:"email" binding:"omitempty,email"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
}

// CreateMerchant creates a new merchant
// POST /api/v1/merchants
func (h *MerchantHandler) CreateMerchant(c *gin.Context) {
	var req CreateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorizedd",
		})
		return
	}
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	// Check if user has a merchant
	hasMerchant, err := h.merchantService.CheckUserHasMerchant(userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if hasMerchant {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "user already has a merchant",
		})
		return
	}
	// Create merchant
	merchant, err := h.merchantService.CreateMerchant(&service.CreateMerchantRequest{
		OwnerID:      userUUID,
		BusinessName: req.BusinessName,
		LegalName:    req.LegalName,
		Email:        req.Email,
		Phone:        req.Phone,
		Website:      req.Website,
		BusinessType: model.BusinessType(req.BusinessType),
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
			"merchant": formatMerchant(merchant),
		},
		"message": "Merchant created successfully",
	})
}

// GetMerchant gets a merchant by ID
// GET /api/v1/merchants/:id
func (h *MerchantHandler) GetMerchant(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	// Check if user has access to merchant
	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	hasAccess, err := h.teamService.IsUserInMerchant(merchantID, userUUID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "access denied",
		})
		return
	}

	// Get merchant
	merchant, err := h.merchantService.GetMerchantByID(merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "merchant not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"merchant": formatMerchant(merchant),
		},
	})
}

// GetMerchantDetails gets merchant with all details
// GET /api/v1/merchants/:id/details
func (h *MerchantHandler) GetMerchantDetails(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	// Check access
	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	hasAccess, err := h.teamService.IsUserInMerchant(merchantID, userUUID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "access denied",
		})
		return
	}

	// Get merchant with details
	merchant, err := h.merchantService.GetMerchantWithDetails(merchantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "merchant not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"merchant":      formatMerchant(merchant),
			"settings":      merchant.Settings,
			"business_info": merchant.BusinessInfo,
			"branding":      merchant.Branding,
			"verification":  merchant.Verification,
		},
	})
}

// ListUserMerchants lists all merchants for the authenticated user
// GET /api/v1/merchants
func (h *MerchantHandler) ListUserMerchants(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	merchants, err := h.merchantService.GetUserMerchants(userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch merchants",
		})
		return
	}

	// Format response
	var formattedMerchants []gin.H
	for _, merchant := range merchants {
		formattedMerchants = append(formattedMerchants, formatMerchant(&merchant))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"merchants": formattedMerchants,
			"count":     len(formattedMerchants),
		},
	})
}

// UpdateMerchant updates merchant information
// PATCH /api/v1/merchants/:id
func (h *MerchantHandler) UpdateMerchant(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	var req UpdateMerchantRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get user ID
	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	// Check access
	hasAccess, err := h.teamService.IsUserInMerchant(merchantID, userUUID)
	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "access denied",
		})
		return
	}

	// Prepare updates
	updates := make(map[string]interface{})
	if req.BusinessName != "" {
		updates["business_name"] = req.BusinessName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Website != "" {
		updates["website"] = req.Website
	}
	updates["_user_id"] = userUUID // For audit log

	// Update merchant
	if err := h.merchantService.UpdateMerchant(merchantID, updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Merchant updated successfully",
	})
}

// DeleteMerchant soft deletes a merchant
// DELETE /api/v1/merchants/:id
func (h *MerchantHandler) DeleteMerchant(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	// Delete merchant (only owner can delete)
	if err := h.merchantService.DeleteMerchant(merchantID, userUUID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Merchant deleted successfully",
	})
}

// Helper function to format merchant response
func formatMerchant(merchant *model.Merchant) gin.H {
	return gin.H{
		"id":            merchant.ID,
		"merchant_code": merchant.MerchantCode,
		"owner_id":      merchant.OwnerID,
		"business_name": merchant.BusinessName,
		"legal_name":    merchant.LegalName.String,
		"email":         merchant.Email,
		"phone":         merchant.Phone.String,
		"website":       merchant.Website.String,
		"status":        merchant.Status,
		"business_type": merchant.BusinessType,
		"country_code":  merchant.CountryCode,
		"currency_code": merchant.CurrencyCode,
		"timezone":      merchant.Timezone,
		"created_at":    merchant.CreatedAt,
		"updated_at":    merchant.UpdatedAt,
	}
}
