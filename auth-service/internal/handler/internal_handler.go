package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/service"
)

// InternalHandler handles internal service-to-service requests
type InternalHandler struct {
	roleService *service.RoleService
}

// NewInternalHandler creates a new internal handler
func NewInternalHandler() *InternalHandler {
	return &InternalHandler{
		roleService: service.NewRoleService(),
	}
}

// AssignMerchantOwnerRoleRequest represents role assignment for merchant owner
type AssignMerchantOwnerRoleRequest struct {
	UserID     string `json:"user_id" binding:"required,uuid"`
	MerchantID string `json:"merchant_id" binding:"required,uuid"`
}

func (h *InternalHandler) AssignMerchantOwnerRole(c *gin.Context) {
	var req AssignMerchantOwnerRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Parse UUIDs
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user_id format",
		})
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant_id format",
		})
		return
	}

	// Get Admin role by name
	adminRole, err := h.roleService.GetRoleByName("Admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Admin role not found",
		})
		return
	}

	// Assign Admin role to user for this merchant
	// The user themselves is the "assigned_by" since they created the merchant
	err = h.roleService.AssignRoleToUser(userID, adminRole.ID, merchantID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":     userID,
			"role_id":     adminRole.ID,
			"role_name":   adminRole.Name,
			"merchant_id": merchantID,
		},
		"message": "Admin role assigned successfully",
	})
}

// GET /internal/v1/users/:user_id/roles
func (h *InternalHandler) GetUserRolesByUserID(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user_id format",
		})
		return
	}

	merchantIDStr := c.Query("merchant_id")
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "merchant_id query parameter required",
		})
		return
	}

	merchantID, err := uuid.Parse(merchantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant_id format",
		})
		return
	}

	// Get user roles for this merchant
	roles, err := h.roleService.GetUserRoles(userID, merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch user roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":     userID,
			"merchant_id": merchantID,
			"roles":       roles,
		},
	})
}
