package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	service "github.com/rhaloubi/payment-gateway/merchant-service/internal/service"
)

type TeamHandler struct {
	teamService *service.TeamService
}

// NewTeamHandler creates a new team handler
func NewTeamHandler() *TeamHandler {
	return &TeamHandler{
		teamService: service.NewTeamService(),
	}
}

// InviteTeamMemberRequest represents team invitation request
type InviteTeamMemberRequest struct {
	Email    string `json:"email" binding:"required,email"`
	RoleID   string `json:"role_id" binding:"required,uuid"`
	RoleName string `json:"role_name" binding:"required"`
}

// InviteTeamMember invites a user to join the team
// POST /api/v1/merchants/:id/team/invite
func (h *TeamHandler) InviteTeamMember(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	var req InviteTeamMemberRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid role ID",
		})
		return
	}
	roleName := req.RoleName

	// Get inviter ID
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

	// Invite team member
	invitation, err := h.teamService.InviteTeamMember(&service.InviteTeamMemberRequest{
		MerchantID: merchantID,
		Email:      req.Email,
		RoleID:     roleID,
		RoleName:   roleName,
		InvitedBy:  userUUID,
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
			"invitation": gin.H{
				"id":               invitation.ID,
				"email":            invitation.Email,
				"status":           invitation.Status,
				"role_name":        invitation.RoleName,
				"invitation_token": invitation.InvitationToken,
				"expires_at":       invitation.ExpiresAt,
				"created_at":       invitation.CreatedAt,
			},
		},
		"message": "Invitation sent successfully",
	})
}

// AcceptInvitation accepts a team invitation
// POST /api/v1/invitations/:token/accept
func (h *TeamHandler) AcceptInvitation(c *gin.Context) {
	token := c.Param("token")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	userUUID, _ := uuid.Parse(userID.(string))

	// Accept invitation
	if err := h.teamService.AcceptInvitation(token, userUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invitation accepted successfully",
	})
}

// GetTeamMembers gets all team members for a merchant
// GET /api/v1/merchants/:id/team
func (h *TeamHandler) GetTeamMembers(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	teamMembers, err := h.teamService.GetTeamMembers(merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch team members",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"team_members": teamMembers,
			"count":        len(teamMembers),
		},
	})
}

// RemoveTeamMember removes a user from the team
// DELETE /api/v1/merchants/:id/team/:user_id
func (h *TeamHandler) RemoveTeamMember(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	removeUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	// Get who is removing
	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	// Remove team member
	if err := h.teamService.RemoveTeamMember(merchantID, removeUserID, userUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Team member removed successfully",
	})
}

// UpdateTeamMemberRole updates a team member's role
// PATCH /api/v1/merchants/:id/team/:user_id
func (h *TeamHandler) UpdateTeamMemberRole(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID",
		})
		return
	}

	var req struct {
		RoleID   string `json:"role_id" binding:"required,uuid"`
		RoleName string `json:"role_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	newRoleID, _ := uuid.Parse(req.RoleID)
	newRoleName := req.RoleName

	// Get who is updating
	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	// Update role
	if err := h.teamService.UpdateTeamMemberRole(merchantID, targetUserID, newRoleID, userUUID, newRoleName); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Team member role updated successfully",
	})
}

// GetPendingInvitations gets pending invitations
// GET /api/v1/merchants/:id/invitations
func (h *TeamHandler) GetPendingInvitations(c *gin.Context) {
	merchantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid merchant ID",
		})
		return
	}

	invitations, err := h.teamService.GetPendingInvitations(merchantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to fetch invitations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"invitations": invitations,
			"count":       len(invitations),
		},
	})
}

// CancelInvitation cancels a pending invitation
// DELETE /api/v1/invitations/:id
func (h *TeamHandler) CancelInvitation(c *gin.Context) {
	invitationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid invitation ID",
		})
		return
	}

	userID, _ := c.Get("user_id")
	userUUID, _ := uuid.Parse(userID.(string))

	if err := h.teamService.CancelInvitation(invitationID, userUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Invitation cancelled successfully",
	})
}
