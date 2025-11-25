package handler

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/service"
	pb "github.com/rhaloubi/payment-gateway/auth-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCAPIKeyService struct {
	pb.UnimplementedAPIKeyServiceServer
	apiKeyService *service.APIKeyService
}

func NewGRPCAPIKeyService() *GRPCAPIKeyService {
	return &GRPCAPIKeyService{
		apiKeyService: service.NewAPIKeyService(),
	}
}

func (s *GRPCAPIKeyService) CreateAPIKey(ctx context.Context, req *pb.CreateAPIKeyRequest) (*pb.CreateAPIKeyResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid merchant_id")
	}

	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid created_by")
	}

	resp, err := s.apiKeyService.CreateAPIKey(&service.CreateAPIKeyRequest{
		MerchantID: merchantID,
		Name:       req.Name,
		CreatedBy:  createdBy,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateAPIKeyResponse{
		Id:        resp.APIKey.ID.String(),
		Name:      resp.APIKey.Name,
		KeyPrefix: resp.APIKey.KeyPrefix,
		PlainKey:  resp.PlainKey, // ⚠️ Only shown once!
		CreatedAt: resp.APIKey.CreatedAt.Format(time.RFC3339),
		Message:   "⚠️ Save this API key! It won't be shown again.",
	}, nil
}

func (s *GRPCAPIKeyService) GetMerchantAPIKeys(ctx context.Context, req *pb.GetMerchantAPIKeysRequest) (*pb.GetMerchantAPIKeysResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid merchant_id")
	}

	apiKeys, err := s.apiKeyService.GetMerchantAPIKeys(merchantID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoAPIKeys := []*pb.APIKey{}
	for _, key := range apiKeys {
		lastUsedAt := ""
		if key.LastUsedAt.Valid {
			lastUsedAt = key.LastUsedAt.Time.Format(time.RFC3339)
		}

		protoAPIKeys = append(protoAPIKeys, &pb.APIKey{
			Id:         key.ID.String(),
			Name:       key.Name,
			KeyPrefix:  key.KeyPrefix,
			IsActive:   key.IsActive,
			LastUsedAt: lastUsedAt,
			CreatedAt:  key.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetMerchantAPIKeysResponse{
		ApiKeys: protoAPIKeys,
	}, nil
}

func (s *GRPCAPIKeyService) DeactivateAPIKey(ctx context.Context, req *pb.DeactivateAPIKeyRequest) (*pb.DeactivateAPIKeyResponse, error) {
	keyID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := s.apiKeyService.DeactivateAPIKey(keyID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeactivateAPIKeyResponse{
		Message: "API key deactivated successfully",
	}, nil
}

func (s *GRPCAPIKeyService) DeleteAPIKey(ctx context.Context, req *pb.DeleteAPIKeyRequest) (*pb.DeleteAPIKeyResponse, error) {
	keyID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	if err := s.apiKeyService.DeleteAPIKey(keyID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.DeleteAPIKeyResponse{
		Message: "API key deleted successfully",
	}, nil
}
