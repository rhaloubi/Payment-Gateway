package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	model "github.com/rhaloubi/payment-gateway/transaction-service/internal/models"
	"github.com/rhaloubi/payment-gateway/transaction-service/internal/service"
	pb "github.com/rhaloubi/payment-gateway/transaction-service/proto"
	"go.uber.org/zap"
)

type TransactionServer struct {
	pb.UnimplementedTransactionServiceServer
	transactionService *service.TransactionService
}

func NewTransactionServer() (*TransactionServer, error) {
	txnService, err := service.NewTransactionService()
	if err != nil {
		return nil, err
	}

	return &TransactionServer{
		transactionService: txnService,
	}, nil
}

// =========================================================================
// Authorize
// =========================================================================

func (s *TransactionServer) Authorize(ctx context.Context, req *pb.AuthorizeRequest) (*pb.AuthorizeResponse, error) {
	logger.Log.Info("gRPC Authorize called",
		zap.String("merchant_id", req.MerchantId),
		zap.Int64("amount", req.Amount),
		zap.String("currency", req.Currency),
	)

	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.AuthorizeResponse{
			Error: "invalid merchant_id",
		}, nil
	}

	// Build service request
	serviceReq := &service.AuthorizeRequest{
		MerchantID:    merchantID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		CardToken:     req.CardToken,
		CardBrand:     req.CardBrand,
		CardLast4:     req.CardLast4,
		FraudScore:    int(req.FraudScore),
		CustomerEmail: req.CustomerEmail,
		Description:   req.Description,
		IPAddress:     req.IpAddress,
		UserAgent:     req.UserAgent,
	}

	// Process authorization
	response, err := s.transactionService.Authorize(ctx, serviceReq)
	if err != nil {
		logger.Log.Error("gRPC authorization failed", zap.Error(err))
		return &pb.AuthorizeResponse{
			Error: err.Error(),
		}, nil
	}

	// Build gRPC response
	return &pb.AuthorizeResponse{
		TransactionId:   response.TransactionID.String(),
		Status:          string(response.Status),
		Approved:        response.Approved,
		AuthCode:        response.AuthCode,
		ResponseCode:    response.ResponseCode,
		ResponseMessage: response.ResponseMessage,
		DeclineReason:   response.DeclineReason,
		Amount:          response.Amount,
		AmountMad:       response.AmountMAD,
		ExchangeRate:    response.ExchangeRate,
		ProcessingFee:   response.ProcessingFee,
		NetAmount:       response.NetAmount,
	}, nil
}

// =========================================================================
// Capture
// =========================================================================

func (s *TransactionServer) Capture(ctx context.Context, req *pb.CaptureRequest) (*pb.CaptureResponse, error) {
	logger.Log.Info("gRPC Capture called",
		zap.String("transaction_id", req.TransactionId),
		zap.Int64("amount", req.Amount),
	)

	// Parse IDs
	txnID, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return &pb.CaptureResponse{
			Error: "invalid transaction_id",
		}, nil
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.CaptureResponse{
			Error: "invalid merchant_id",
		}, nil
	}

	// Build service request
	serviceReq := &service.CaptureRequest{
		TransactionID: txnID,
		Amount:        req.Amount,
		MerchantID:    merchantID,
	}

	// Process capture
	response, err := s.transactionService.Capture(ctx, serviceReq)
	if err != nil {
		logger.Log.Error("gRPC capture failed", zap.Error(err))
		return &pb.CaptureResponse{
			Error: err.Error(),
		}, nil
	}

	return &pb.CaptureResponse{
		TransactionId:   response.TransactionID.String(),
		Status:          string(response.Status),
		CapturedAmount:  response.CapturedAmount,
		ResponseMessage: response.ResponseMessage,
	}, nil
}

// =========================================================================
// Void
// =========================================================================

func (s *TransactionServer) Void(ctx context.Context, req *pb.VoidRequest) (*pb.VoidResponse, error) {
	logger.Log.Info("gRPC Void called",
		zap.String("transaction_id", req.TransactionId),
	)

	// Parse IDs
	txnID, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return &pb.VoidResponse{
			Error: "invalid transaction_id",
		}, nil
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.VoidResponse{
			Error: "invalid merchant_id",
		}, nil
	}

	// Build service request
	serviceReq := &service.VoidRequest{
		TransactionID: txnID,
		MerchantID:    merchantID,
		Reason:        req.Reason,
	}

	// Process void
	response, err := s.transactionService.Void(ctx, serviceReq)
	if err != nil {
		logger.Log.Error("gRPC void failed", zap.Error(err))
		return &pb.VoidResponse{
			Error: err.Error(),
		}, nil
	}

	return &pb.VoidResponse{
		TransactionId:   response.TransactionID.String(),
		Status:          string(response.Status),
		ResponseMessage: response.ResponseMessage,
	}, nil
}

// =========================================================================
// Refund
// =========================================================================

func (s *TransactionServer) Refund(ctx context.Context, req *pb.RefundRequest) (*pb.RefundResponse, error) {
	logger.Log.Info("gRPC Refund called",
		zap.String("transaction_id", req.TransactionId),
		zap.Int64("amount", req.Amount),
	)

	// Parse IDs
	txnID, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return &pb.RefundResponse{
			Error: "invalid transaction_id",
		}, nil
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.RefundResponse{
			Error: "invalid merchant_id",
		}, nil
	}

	// Build service request
	serviceReq := &service.RefundRequest{
		TransactionID: txnID,
		Amount:        req.Amount,
		Reason:        req.Reason,
		MerchantID:    merchantID,
	}

	// Process refund
	response, err := s.transactionService.Refund(ctx, serviceReq)
	if err != nil {
		logger.Log.Error("gRPC refund failed", zap.Error(err))
		return &pb.RefundResponse{
			Error: err.Error(),
		}, nil
	}

	return &pb.RefundResponse{
		RefundId:        response.RefundID.String(),
		TransactionId:   response.TransactionID.String(),
		RefundedAmount:  response.RefundedAmount,
		RemainingAmount: response.RemainingAmount,
		ResponseMessage: response.ResponseMessage,
	}, nil
}

// =========================================================================
// GetTransaction
// =========================================================================

func (s *TransactionServer) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.TransactionResponse, error) {
	// Parse IDs
	txnID, err := uuid.Parse(req.TransactionId)
	if err != nil {
		return &pb.TransactionResponse{
			Error: "invalid transaction_id",
		}, nil
	}

	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.TransactionResponse{
			Error: "invalid merchant_id",
		}, nil
	}

	// Get transaction
	txn, err := s.transactionService.GetTransaction(txnID, merchantID)
	if err != nil {
		return &pb.TransactionResponse{
			Error: "transaction not found",
		}, nil
	}

	// Build response
	response := &pb.TransactionResponse{
		Id:             txn.ID.String(),
		MerchantId:     txn.MerchantID.String(),
		Type:           string(txn.Type),
		Status:         string(txn.Status),
		Amount:         txn.Amount,
		Currency:       txn.Currency,
		AmountMad:      txn.AmountMAD,
		ExchangeRate:   txn.ExchangeRate,
		CardBrand:      txn.CardBrand,
		CardLast4:      txn.CardLast4,
		FraudScore:     int32(txn.FraudScore),
		CapturedAmount: txn.CapturedAmount,
		RefundedAmount: txn.RefundedAmount,
		ProcessingFee:  txn.ProcessingFee,
		NetAmount:      txn.NetAmount,
		CreatedAt:      txn.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if txn.AuthCode.Valid {
		response.AuthCode = txn.AuthCode.String
	}
	if txn.AuthorizedAt.Valid {
		response.AuthorizedAt = txn.AuthorizedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	if txn.CapturedAt.Valid {
		response.CapturedAt = txn.CapturedAt.Time.Format("2006-01-02T15:04:05Z")
	}

	return response, nil
}

// =========================================================================
// ListTransactions
// =========================================================================

func (s *TransactionServer) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	// Parse merchant ID
	merchantID, err := uuid.Parse(req.MerchantId)
	if err != nil {
		return &pb.ListTransactionsResponse{
			Error: "invalid merchant_id",
		}, nil
	}

	// Get transactions
	var txns []model.Transaction
	if req.Status != "" {
		status := model.TransactionStatus(req.Status)
		txns, err = s.transactionService.FindByStatus(merchantID, status)
	} else {
		limit := int(req.Limit)
		if limit == 0 {
			limit = 50
		}
		offset := int(req.Offset)
		txns, err = s.transactionService.FindByMerchant(merchantID, limit, offset)
	}

	if err != nil {
		return &pb.ListTransactionsResponse{
			Error: err.Error(),
		}, nil
	}

	// Build response
	transactions := make([]*pb.TransactionResponse, len(txns))
	for i, txn := range txns {
		transactions[i] = &pb.TransactionResponse{
			Id:             txn.ID.String(),
			MerchantId:     txn.MerchantID.String(),
			Type:           string(txn.Type),
			Status:         string(txn.Status),
			Amount:         txn.Amount,
			Currency:       txn.Currency,
			AmountMad:      txn.AmountMAD,
			CardBrand:      txn.CardBrand,
			CardLast4:      txn.CardLast4,
			FraudScore:     int32(txn.FraudScore),
			CapturedAmount: txn.CapturedAmount,
			RefundedAmount: txn.RefundedAmount,
			CreatedAt:      txn.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return &pb.ListTransactionsResponse{
		Transactions: transactions,
		Total:        int32(len(txns)),
	}, nil
}
