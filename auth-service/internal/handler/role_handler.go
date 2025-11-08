package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/service"
)

// RoleHandler handles role-related requests
type RoleHandler struct {
	roleService *service.RoleService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler() *RoleHandler {
	return &RoleHandler{
		roleService: service.NewRoleService(),
	}
}

// AssignRoleRequest represents role assignment request
type AssignRoleRequest struct {
	UserID     string `json:"user_id" binding:"required,uuid"`
	RoleID     string `json:"role_id" binding:"required,uuid"`
	MerchantID string `json:"merchant_id" binding:"required,uuid"`
}

// GetAllRoles gets all available roles
// GET /api/v1/roles
func (h *RoleHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.roleService.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch roles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"roles": roles,
		},
	})
}

// GetRoleByID gets a role by ID with its permissions
// GET /api/v1/roles/:id
func (h *RoleHandler) GetRoleByID(c *gin.Context) {
	roleID := c.Param("id")
	parsedRoleID, err := uuid.Parse(roleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid role ID format",
		})
		return
	}

	role, err := h.roleService.GetRoleWithPermissions(parsedRoleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "role not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"role": role,
		},
	})
}

// AssignRoleToUser assigns a role to a user
// POST /api/v1/roles/assign
func (h *RoleHandler) AssignRoleToUser(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get who is assigning (from auth context)
	assignedBy, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	// Call service
	err := h.roleService.AssignRoleToUser(
		uuid.MustParse(req.UserID),
		uuid.MustParse(req.RoleID),
		uuid.MustParse(req.MerchantID),
		uuid.MustParse(assignedBy.(string)),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Role assigned successfully",
	})
}

// RemoveRoleFromUser removes a role from a user
// DELETE /api/v1/roles/assign
func (h *RoleHandler) RemoveRoleFromUser(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Call service
	err := h.roleService.RemoveRoleFromUser(
		uuid.MustParse(req.UserID),
		uuid.MustParse(req.RoleID),
		uuid.MustParse(req.MerchantID),
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Role removed successfully",
	})
}

// GetUserRoles gets all roles for a user in a merchant
// GET /api/v1/roles/user/:user_id/merchant/:merchant_id
func (h *RoleHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("user_id")
	merchantID := c.Param("merchant_id")

	roles, err := h.roleService.GetUserRoles(uuid.MustParse(userID), uuid.MustParse(merchantID))
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
			"roles": roles,
		},
	})
}

// GetUserPermissions gets all permissions for a user in a merchant
// GET /api/v1/roles/user/:user_id/merchant/:merchant_id/permissions
func (h *RoleHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("user_id")
	merchantID := c.Param("merchant_id")

	permissions, err := h.roleService.GetUserPermissions(uuid.MustParse(userID), uuid.MustParse(merchantID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch user permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"permissions": permissions,
		},
	})
}
