package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/service"
	pb "github.com/rhaloubi/payment-gateway/tokenization-service/proto"
	"go.uber.org/zap"
)

type TokenizationServer struct {
	pb.UnimplementedTokenizationServiceServer
	tokenizationService *service.TokenizationService
}

func NewTokenizationServer() *TokenizationServer {
	return &TokenizationServer{
		tokenizationService: service.NewTokenizationService(),
	}
}

// =========================================================================
// TokenizeCard
// =========================================================================

func (s *TokenizationServer) TokenizeCard(ctx context.Context, req *pb.TokenizeCardRequest) (*pb.TokenizeCardResponse, error) {
	logger.Log.Info("gRPC TokenizeCard called",
		zap.String("merchant_id", req.MerchantId),
		zap.String("request_id", req.RequestId),
	)

	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.TokenizeCardResponse{
			Error: "invalid merchant_id",
		}, nil
	}

	// Parse created_by (optional)
	var createdBy uuid.UUID
	if req.CreatedBy != "" {
		createdBy, _ = uuid.Parse(req.CreatedBy)
	}

	// Build service request
	serviceReq := &service.TokenizeCardRequest{
		MerchantID:     merchantID,
		CardNumber:     req.CardNumber,
		CardholderName: req.CardholderName,
		ExpiryMonth:    int(req.ExpMonth),
		ExpiryYear:     int(req.ExpYear),
		CVV:            req.Cvv,
		IsSingleUse:    req.IsSingleUse,
		RequestID:      req.RequestId,
		IPAddress:      req.IpAddress,
		UserAgent:      req.UserAgent,
		CreatedBy:      createdBy,
	}

	// Tokenize card
	response, err := s.tokenizationService.TokenizeCard(serviceReq)
	if err != nil {
		logger.Log.Error("gRPC tokenization failed", zap.Error(err))
		return &pb.TokenizeCardResponse{
			Error: err.Error(),
		}, nil
	}

	// Build gRPC response
	return &pb.TokenizeCardResponse{
		Token: response.Token,
		Card: &pb.CardMetadata{
			Brand:       string(response.CardBrand),
			Type:        string(response.CardType),
			Last4:       response.Last4Digits,
			ExpMonth:    int32(response.ExpiryMonth),
			ExpYear:     int32(response.ExpiryYear),
			Fingerprint: response.Fingerprint,
		},
		IsNewToken: response.IsNewToken,
	}, nil
}

// =========================================================================
// Detokenize (Internal Only)
// =========================================================================

func (s *TokenizationServer) Detokenize(ctx context.Context, req *pb.DetokenizeRequest) (*pb.DetokenizeResponse, error) {
	logger.Log.Info("gRPC Detokenize called",
		zap.String("token", req.Token),
		zap.String("merchant_id", req.MerchantId),
		zap.String("usage_type", req.UsageType),
	)

	// Parse UUIDs
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.DetokenizeResponse{
			Error: "invalid merchant_id",
		}, nil
	}

	transactionID, _ := uuid.Parse(req.TransactionId)

	// Build service request
	serviceReq := &service.DetokenizeRequest{
		Token:         req.Token,
		MerchantID:    merchantID,
		TransactionID: transactionID,
		UsageType:     req.UsageType,
		Amount:        req.Amount,
		Currency:      req.Currency,
		IPAddress:     req.IpAddress,
		UserAgent:     req.UserAgent,
	}

	// Detokenize
	response, err := s.tokenizationService.Detokenize(serviceReq)
	if err != nil {
		logger.Log.Error("gRPC detokenization failed", zap.Error(err))
		return &pb.DetokenizeResponse{
			Error: err.Error(),
		}, nil
	}

	// Build gRPC response
	return &pb.DetokenizeResponse{
		CardNumber:     response.CardNumber,
		CardholderName: response.CardholderName,
		ExpMonth:       int32(response.ExpiryMonth),
		ExpYear:        int32(response.ExpiryYear),
		CardBrand:      string(response.CardBrand),
		Last4:          response.Last4Digits,
	}, nil
}

// =========================================================================
// ValidateToken
// =========================================================================

func (s *TokenizationServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Error: "invalid merchant_id",
		}, nil
	}

	// Validate token
	isValid, err := s.tokenizationService.ValidateToken(req.Token, merchantID)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Error: err.Error(),
		}, nil
	}

	// Get token info
	tokenInfo, err := s.tokenizationService.GetTokenInfo(req.Token, merchantID)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Error: err.Error(),
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid: isValid,
		Card: &pb.CardMetadata{
			Brand:       string(tokenInfo.CardBrand),
			Type:        string(tokenInfo.CardType),
			Last4:       tokenInfo.Last4Digits,
			ExpMonth:    int32(tokenInfo.ExpiryMonth),
			ExpYear:     int32(tokenInfo.ExpiryYear),
			Fingerprint: tokenInfo.Fingerprint,
		},
		Status:      string(tokenInfo.Status),
		UsageCount:  int32(tokenInfo.UsageCount),
		IsSingleUse: tokenInfo.IsSingleUse,
	}, nil
}

// =========================================================================
// RevokeToken
// =========================================================================

func (s *TokenizationServer) RevokeToken(ctx context.Context, req *pb.RevokeTokenRequest) (*pb.RevokeTokenResponse, error) {
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.RevokeTokenResponse{
			Success: false,
			Error:   "invalid merchant_id",
		}, nil
	}

	revokedBy, _ := uuid.Parse(req.RevokedBy)

	// Revoke token
	err = s.tokenizationService.RevokeToken(req.Token, merchantID, revokedBy, req.Reason)
	if err != nil {
		return &pb.RevokeTokenResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &pb.RevokeTokenResponse{
		Success: true,
		Message: "token revoked successfully",
	}, nil
}
