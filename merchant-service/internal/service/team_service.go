package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/merchant-service/internal/models"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/repository"
	"go.uber.org/zap"
)

type TeamService struct {
	merchantUserRepo *repository.MerchantUserRepository
	invitationRepo   *repository.InvitationRepository
	merchantRepo     *repository.MerchantRepository
	activityLogRepo  *repository.ActivityLogRepository
	emailService     *EmailService
}

// NewTeamService creates a new team service
func NewTeamService() *TeamService {
	return &TeamService{
		merchantUserRepo: repository.NewMerchantUserRepository(),
		invitationRepo:   repository.NewInvitationRepository(),
		merchantRepo:     repository.NewMerchantRepository(),
		activityLogRepo:  repository.NewActivityLogRepository(),
		emailService:     NewEmailService(),
	}
}

// InviteTeamMemberRequest represents team invitation data
type InviteTeamMemberRequest struct {
	MerchantID uuid.UUID
	Email      string
	RoleID     uuid.UUID
	RoleName   string
	InvitedBy  uuid.UUID
}

// InviteTeamMember  ##### ------- i removed send email for testing purposes ---------- #######
func (s *TeamService) InviteTeamMember(req *InviteTeamMemberRequest) (*model.MerchantInvitation, error) {
	// Validate merchant exists
	merchant, err := s.merchantRepo.FindByID(req.MerchantID)
	if err != nil {
		return nil, err
	}

	// Check if already has pending invitation
	hasPending, err := s.invitationRepo.ExistsPendingForEmail(req.MerchantID, req.Email)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, errors.New("pending invitation already exists for this email")
	}

	// Create invitation
	invitation := &model.MerchantInvitation{
		MerchantID: req.MerchantID,
		Email:      req.Email,
		RoleID:     req.RoleID,
		RoleName:   req.RoleName,
		InvitedBy:  req.InvitedBy,
		Status:     model.InvitationStatusPending,
	}

	if err := s.invitationRepo.Create(invitation); err != nil {
		return nil, err
	}

	// Send invitation email via Mailtrap
	go func(invitation *model.MerchantInvitation, merchant *model.Merchant) {
		if err := s.emailService.SendInvitationEmail(invitation, merchant); err != nil {
			// Log error but don't fail the invitation
			logger.Log.Error("Failed to send invitation email", zap.Error(err))
		}
	}(invitation, merchant)
	//
	// Log activity
	changes := map[string]interface{}{
		"email":     req.Email,
		"role_id":   req.RoleID.String(),
		"role_name": req.RoleName,
	}
	s.logActivity(req.MerchantID, req.InvitedBy, "team_member_invited", "invitation", invitation.ID, changes)

	return invitation, nil
}

func (s *TeamService) AcceptInvitation(token string, userID uuid.UUID) error {
	// Find invitation
	invitation, err := s.invitationRepo.FindByToken(token)
	if err != nil {
		return err
	}

	// Validate invitation
	if !invitation.IsValid() {
		if invitation.IsExpired() {
			return errors.New("invitation has expired")
		}
		return errors.New("invitation is not valid")
	}

	// Check if user is already a team member
	isTeamMember, err := s.merchantUserRepo.IsUserInMerchant(invitation.MerchantID, userID)
	if err != nil {
		return err
	}
	if isTeamMember {
		return errors.New("user is already a team member")
	}

	// Add user to team
	merchantUser := &model.MerchantUser{
		MerchantID: invitation.MerchantID,
		UserID:     userID,
		RoleID:     invitation.RoleID,
		RoleName:   invitation.RoleName,
		InvitedBy:  invitation.InvitedBy,
		Status:     model.MerchantUserStatusActive,
	}
	merchantUser.JoinedAt = toNullTime(time.Now())

	if err := s.merchantUserRepo.Create(merchantUser); err != nil {
		return err
	}

	// Mark invitation as accepted
	if err := s.invitationRepo.MarkAsAccepted(invitation.ID); err != nil {
		return err
	}

	// Log activity
	changes := map[string]interface{}{
		"user_id": userID.String(),
		"role_id": invitation.RoleID.String(),
	}
	s.logActivity(invitation.MerchantID, userID, "team_member_joined", "merchant_user", merchantUser.ID, changes)

	return nil
}

// GetTeamMembers gets all team members for a merchant
func (s *TeamService) GetTeamMembers(merchantID uuid.UUID) ([]model.MerchantUser, error) {
	return s.merchantUserRepo.GetTeamMembers(merchantID)
}

// RemoveTeamMember removes a user from the merchant team
func (s *TeamService) RemoveTeamMember(merchantID, userID, removedBy uuid.UUID) error {
	// Find merchant user
	merchantUser, err := s.merchantUserRepo.FindByMerchantAndUser(merchantID, userID)
	if err != nil {
		return err
	}

	// Check if trying to remove owner
	merchant, err := s.merchantRepo.FindByID(merchantID)
	if err != nil {
		return err
	}
	if merchant.OwnerID == userID {
		return errors.New("cannot remove the owner from the team")
	}

	// Remove from team
	if err := s.merchantUserRepo.Delete(merchantUser.ID); err != nil {
		return err
	}

	// Log activity
	changes := map[string]interface{}{
		"removed_user_id": userID.String(),
	}
	s.logActivity(merchantID, removedBy, "team_member_removed", "merchant_user", merchantUser.ID, changes)

	return nil
}

// UpdateTeamMemberRole updates a team member's role
func (s *TeamService) UpdateTeamMemberRole(merchantID, userID uuid.UUID, newRoleID uuid.UUID, updatedBy uuid.UUID, newRoleName string) error {
	merchantUser, err := s.merchantUserRepo.FindByMerchantAndUser(merchantID, userID)
	if err != nil {
		return err
	}

	oldRoleID := merchantUser.RoleID
	merchantUser.RoleID = newRoleID
	merchantUser.RoleName = newRoleName

	if err := s.merchantUserRepo.Update(merchantUser); err != nil {
		return err
	}

	// Log activity
	changes := map[string]interface{}{
		"user_id": userID.String(),
		"role": map[string]interface{}{
			"old": oldRoleID.String(),
			"new": newRoleID.String(),
		},
		"role_name": map[string]interface{}{
			"old": merchantUser.RoleName,
			"new": newRoleName,
		},
	}
	go s.logActivity(merchantID, updatedBy, "team_member_role_updated", "merchant_user", merchantUser.ID, changes)

	return nil
}

// GetPendingInvitations gets pending invitations for a merchant
func (s *TeamService) GetPendingInvitations(merchantID uuid.UUID) ([]model.MerchantInvitation, error) {
	// Mark expired invitations
	s.invitationRepo.MarkAsExpired(merchantID)

	// Get all invitations
	return s.invitationRepo.FindByMerchant(merchantID)
}

// CancelInvitation cancels a pending invitation
func (s *TeamService) CancelInvitation(invitationID, cancelledBy uuid.UUID) error {
	invitation, err := s.invitationRepo.FindByID(invitationID)
	if err != nil {
		return err
	}

	if err := s.invitationRepo.Cancel(invitationID); err != nil {
		return err
	}

	// Log activity
	changes := map[string]interface{}{
		"email": invitation.Email,
	}
	go s.logActivity(invitation.MerchantID, cancelledBy, "invitation_cancelled", "invitation", invitationID, changes)

	return nil
}

// IsUserInMerchant checks if user has access to merchant
func (s *TeamService) IsUserInMerchant(merchantID, userID uuid.UUID) (bool, error) {
	// Check if owner
	merchant, err := s.merchantRepo.FindByID(merchantID)
	if err != nil {
		return false, err
	}
	if merchant.OwnerID == userID {
		return true, nil
	}

	// Check if team member
	return s.merchantUserRepo.IsUserInMerchant(merchantID, userID)
}

// CheckUserPermission checks if user has specific permission for the merchant
func (s *TeamService) CheckUserPermission(merchantID, userID uuid.UUID, action string) (bool, error) {
	// Get merchant
	merchant, err := s.merchantRepo.FindByID(merchantID)
	if err != nil {
		return false, err
	}

	// Check if user is owner
	if merchant.OwnerID == userID {
		return true, nil
	}

	// Get user's role in the merchant
	merchantUser, err := s.merchantUserRepo.FindByMerchantAndUser(merchantID, userID)
	if err != nil {
		return false, err
	}

	switch merchantUser.RoleName {
	case "Admin":
		// Admin can do everything except delete
		switch action {
		case "delete":
			return false, nil
		default:
			return true, nil
		}
	case "Manager":
		// Manager can create and read
		switch action {
		case "create", "read":
			return true, nil
		default:
			return false, nil
		}
	case "Staff":
		// Staff can only read
		switch action {
		case "read":
			return true, nil
		default:
			return false, nil
		}
	default:
		return false, nil
	}
}

// logActivity logs team activity
func (s *TeamService) logActivity(merchantID, userID uuid.UUID, action, resourceType string, resourceID uuid.UUID, changes map[string]interface{}) {
	log := &model.MerchantActivityLog{
		MerchantID:   merchantID,
		UserID:       userID,
		Action:       action,
		ResourceType: toNullString(resourceType),
		ResourceID:   toNullString(resourceID.String()),
	}

	if changes != nil {
		changesJSON, _ := json.Marshal(changes)
		log.Changes = changesJSON
	}

	s.activityLogRepo.Create(log)
}
