package handler

import (
	"context"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/service"
	pb "github.com/rhaloubi/payment-gateway/auth-service/proto"
)

type GRPCRoleService struct {
	pb.UnimplementedRoleServiceServer
	roleService *service.RoleService
}

func NewGRPCRoleService() *GRPCRoleService {
	return &GRPCRoleService{
		roleService: service.NewRoleService(),
	}
}

// AssignMerchantOwnerRole implements the gRPC method
func (s *GRPCRoleService) AssignMerchantOwnerRole(ctx context.Context, req *pb.AssignMerchantOwnerRoleRequest) (*pb.AssignMerchantOwnerRoleResponse, error) {

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, err
	}

	adminRole, err := s.roleService.GetRoleByName("Admin")
	if err != nil {
		return nil, err
	}

	err = s.roleService.AssignRoleToUser(userID, adminRole.ID, merchantID, userID)
	if err != nil {
		return nil, err
	}

	return &pb.AssignMerchantOwnerRoleResponse{
		UserId:     userID.String(),
		RoleId:     adminRole.ID.String(),
		RoleName:   adminRole.Name,
		MerchantId: merchantID.String(),
		Message:    "Admin role assigned successfully",
	}, nil
}

func (s *GRPCRoleService) GetUserRoles(ctx context.Context, req *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, err
	}

	roles, err := s.roleService.GetUserRoles(userID, merchantID)
	if err != nil {
		return nil, err
	}

	// convert to proto roles
	protoRoles := []*pb.Role{}
	for _, r := range roles {
		protoRoles = append(protoRoles, &pb.Role{
			Id:   r.ID.String(),
			Name: r.Name,
		})
	}

	return &pb.GetUserRolesResponse{
		UserId:     userID.String(),
		MerchantId: merchantID.String(),
		Roles:      protoRoles,
	}, nil
}

// AssignRoleToUser implements the gRPC method
func (s *GRPCRoleService) AssignRoleToUser(ctx context.Context, req *pb.AssignRoleToUserRequest) (*pb.AssignRoleToUserResponse, error) {

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, err
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, err
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, err
	}

	assignedBy, err := uuid.Parse(req.AssignedBy)
	if err != nil {
		return nil, err
	}

	err = s.roleService.AssignRoleToUser(userID, roleID, merchantID, assignedBy)
	if err != nil {
		return nil, err
	}

	return &pb.AssignRoleToUserResponse{
		UserId:     userID.String(),
		RoleId:     roleID.String(),
		RoleName:   s.roleService.GetRoleName(roleID),
		MerchantId: merchantID.String(),
		AssignedBy: assignedBy.String(),
		Message:    "Role assigned successfully",
	}, nil
}
