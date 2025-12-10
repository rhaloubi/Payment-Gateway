package service

import (
	"context"

	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/client"
	pb "github.com/rhaloubi/payment-gateway/payment-api-service/proto"
)

type TransactionService struct {
	transactionClient *client.TransactionClient
}

func NewTransactionService() (*TransactionService, error) {
	transactionClient := client.NewTransactionClient()
	return &TransactionService{
		transactionClient: transactionClient,
	}, nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.TransactionResponse, error) {
	res, err := s.transactionClient.GetTransaction(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *TransactionService) ListTransactions(ctx context.Context, req *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	res, err := s.transactionClient.ListTransactions(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
