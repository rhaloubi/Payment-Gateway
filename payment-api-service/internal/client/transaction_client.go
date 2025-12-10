package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	pb "github.com/rhaloubi/payment-gateway/payment-api-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TransactionClient communicates with Transaction Service
// TODO: Replace with actual gRPC client when transaction service is built
type TransactionClient struct {
	httpClient        *http.Client
	grpcConn          *grpc.ClientConn
	grpcTimeout       time.Duration
	transactionClient pb.TransactionServiceClient
}

func NewTransactionClient() *TransactionClient {
	grpcAddress := os.Getenv("TRANSACTION_SERVICE_GRPC_URL")
	if grpcAddress == "" {
		grpcAddress = "localhost:50053"
	}

	// Dial gRPC connection (insecure for dev)
	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Fatal("failed to dial gRPC", zap.Error(err))
	}

	return &TransactionClient{
		httpClient:        &http.Client{Timeout: 10 * time.Second},
		grpcConn:          conn,
		grpcTimeout:       400 * time.Millisecond,
		transactionClient: pb.NewTransactionServiceClient(conn),
	}
}

// =========================================================================
// Authorization
// =========================================================================

func (c *TransactionClient) Authorize(ctx context.Context, req *pb.AuthorizeRequest) (*pb.AuthorizeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	logger.Log.Info("Processing authorization ",
		zap.String("merchant_id", req.MerchantId),
		zap.Int64("amount", req.Amount),
		zap.String("card_last4", req.CardLast4),
	)
	resp, err := c.transactionClient.Authorize(ctx, &pb.AuthorizeRequest{
		MerchantId:    req.MerchantId,
		Amount:        req.Amount,
		Currency:      req.Currency,
		CardToken:     req.CardToken,
		CardBrand:     req.CardBrand,
		CardLast4:     req.CardLast4,
		FraudScore:    req.FraudScore,
		CustomerEmail: req.CustomerEmail,
		Description:   req.Description,
		IpAddress:     req.IpAddress,
		UserAgent:     req.UserAgent,
	})
	if err != nil {
		logger.Log.Error("Transaction service gRPC request failed", zap.Error(err))
		return nil, fmt.Errorf("transaction service unavailable or invalid key: %w", err)
	}

	return &pb.AuthorizeResponse{
		TransactionId:   resp.TransactionId,
		Status:          resp.Status,
		Approved:        resp.Approved,
		AuthCode:        resp.AuthCode,
		ResponseCode:    resp.ResponseCode,
		ResponseMessage: resp.ResponseMessage,
		DeclineReason:   resp.DeclineReason,
		Amount:          resp.Amount,
		AmountMad:       resp.AmountMad,
		ExchangeRate:    resp.ExchangeRate,
		ProcessingFee:   resp.ProcessingFee,
		NetAmount:       resp.NetAmount,
	}, nil
}

// =========================================================================
// Capture
// =========================================================================

func (c *TransactionClient) Capture(ctx context.Context, req *pb.CaptureRequest) (*pb.CaptureResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	logger.Log.Info("Processing capture (mock)",
		zap.String("transaction_id", req.TransactionId),
		zap.Int64("amount", req.Amount),
	)

	resp, err := c.transactionClient.Capture(ctx, &pb.CaptureRequest{
		TransactionId: req.TransactionId,
		Amount:        req.Amount,
		MerchantId:    req.MerchantId,
	})
	if err != nil {
		logger.Log.Error("Transaction service gRPC request failed", zap.Error(err))
		return nil, fmt.Errorf("transaction service unavailable or invalid key: %w", err)
	}

	return &pb.CaptureResponse{
		TransactionId:   resp.TransactionId,
		Status:          resp.Status,
		CapturedAmount:  resp.CapturedAmount,
		ResponseMessage: resp.ResponseMessage,
	}, nil
}

// =========================================================================
// Void
// =========================================================================

// Void cancels an authorized transaction
func (c *TransactionClient) Void(ctx context.Context, req *pb.VoidRequest) (*pb.VoidResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	logger.Log.Info("Processing void (mock)",
		zap.String("transaction_id", req.TransactionId),
		zap.String("reason", req.Reason),
	)

	resp, err := c.transactionClient.Void(ctx, &pb.VoidRequest{
		TransactionId: req.TransactionId,
		Reason:        req.Reason,
		MerchantId:    req.MerchantId,
	})
	if err != nil {
		logger.Log.Error("Transaction service gRPC request failed", zap.Error(err))
		return nil, fmt.Errorf("transaction service unavailable or invalid key: %w", err)
	}

	return &pb.VoidResponse{
		TransactionId:   resp.TransactionId,
		Status:          resp.Status,
		ResponseMessage: resp.ResponseMessage,
	}, nil
}

// =========================================================================
// Refund
// =========================================================================

// Refund processes a refund
func (c *TransactionClient) Refund(ctx context.Context, req *pb.RefundRequest) (*pb.RefundResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	logger.Log.Info("Processing refund (mock)",
		zap.String("transaction_id", req.TransactionId),
		zap.Int64("amount", req.Amount),
	)

	resp, err := c.transactionClient.Refund(ctx, &pb.RefundRequest{
		TransactionId: req.TransactionId,
		Amount:        req.Amount,
		MerchantId:    req.MerchantId,
	})
	if err != nil {
		logger.Log.Error("Transaction service gRPC request failed", zap.Error(err))
		return nil, fmt.Errorf("transaction service unavailable or invalid key: %w", err)
	}

	return &pb.RefundResponse{
		RefundId:        resp.RefundId,
		RefundedAmount:  resp.RefundedAmount,
		ResponseMessage: resp.ResponseMessage,
	}, nil
}

func (c *TransactionClient) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.TransactionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	logger.Log.Info("Processing get transaction (mock)",
		zap.String("transaction_id", req.TransactionId),
		zap.String("merchant_id", req.MerchantId),
	)

	resp, err := c.transactionClient.GetTransaction(ctx, &pb.GetTransactionRequest{
		TransactionId: req.TransactionId,
		MerchantId:    req.MerchantId,
	})
	if err != nil {
		logger.Log.Error("Transaction service gRPC request failed", zap.Error(err))
		return nil, fmt.Errorf("transaction service unavailable or invalid key: %w", err)
	}

	return &pb.TransactionResponse{
		Id:             resp.Id,
		MerchantId:     resp.MerchantId,
		Type:           resp.Type,
		Status:         resp.Status,
		Amount:         resp.Amount,
		Currency:       resp.Currency,
		CardBrand:      resp.CardBrand,
		CardLast4:      resp.CardLast4,
		AuthCode:       resp.AuthCode,
		AmountMad:      resp.AmountMad,
		ExchangeRate:   resp.ExchangeRate,
		FraudScore:     resp.FraudScore,
		CapturedAmount: resp.CapturedAmount,
		RefundedAmount: resp.RefundedAmount,
		ProcessingFee:  resp.ProcessingFee,
		NetAmount:      resp.NetAmount,
		CreatedAt:      resp.CreatedAt,
		AuthorizedAt:   resp.AuthorizedAt,
		CapturedAt:     resp.CapturedAt,
	}, nil
}

func (c *TransactionClient) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.grpcTimeout)
	defer cancel()

	logger.Log.Info("Processing list transactions (mock)",
		zap.String("merchant_id", req.MerchantId),
	)

	resp, err := c.transactionClient.ListTransactions(ctx, &pb.ListTransactionsRequest{
		MerchantId: req.MerchantId,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	if err != nil {
		logger.Log.Error("Transaction service gRPC request failed", zap.Error(err))
		return nil, fmt.Errorf("transaction service unavailable or invalid key: %w", err)
	}

	return &pb.ListTransactionsResponse{
		Transactions: resp.Transactions,
		Total:        resp.Total,
	}, nil
}

// Close closes the client connection (no-op for mock)
func (c *TransactionClient) Close() error {
	return nil
}
