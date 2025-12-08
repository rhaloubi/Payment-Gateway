package main

import (
	"context"
	"net"
	"time"

	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	grpcServer "github.com/rhaloubi/payment-gateway/transaction-service/internal/grpc"
	"github.com/rhaloubi/payment-gateway/transaction-service/internal/service"
	pb "github.com/rhaloubi/payment-gateway/transaction-service/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// =========================================================================
// gRPC Server
// =========================================================================

func startGRPCServer(port string) {
	// Create TCP listener
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Log.Fatal("Failed to listen on gRPC port", zap.Error(err))
	}

	// Create gRPC server
	grpcSrv := grpc.NewServer()

	// Register transaction service
	transactionServer, err := grpcServer.NewTransactionServer()
	if err != nil {
		logger.Log.Fatal("Failed to create transaction server", zap.Error(err))
	}
	pb.RegisterTransactionServiceServer(grpcSrv, transactionServer)

	logger.Log.Info("gRPC server starting", zap.String("port", port))

	// Start serving
	if err := grpcSrv.Serve(lis); err != nil {
		logger.Log.Fatal("Failed to serve gRPC", zap.Error(err))
	}
}

// =========================================================================
// Background Workers
// =========================================================================

// Settlement Worker - Runs daily at midnight
func startSettlementWorker(ctx context.Context, settlementService *service.SettlementService) {
	logger.Log.Info("Settlement worker started")

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Calculate time until next midnight
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	initialDelay := time.Until(nextMidnight)

	logger.Log.Info("Next settlement batch creation scheduled",
		zap.Duration("in", initialDelay),
		zap.Time("at", nextMidnight),
	)

	// Wait until midnight
	select {
	case <-time.After(initialDelay):
		// Run settlement
		if err := settlementService.CreateDailySettlementBatches(ctx); err != nil {
			logger.Log.Error("Settlement batch creation failed", zap.Error(err))
		}
	case <-ctx.Done():
		return
	}

	// Then run daily
	for {
		select {
		case <-ticker.C:
			logger.Log.Info("Running daily settlement batch creation")
			if err := settlementService.CreateDailySettlementBatches(ctx); err != nil {
				logger.Log.Error("Settlement batch creation failed", zap.Error(err))
			}

			// Also process pending settlements (T+2)
			if err := settlementService.ProcessPendingSettlements(ctx); err != nil {
				logger.Log.Error("Pending settlement processing failed", zap.Error(err))
			}

		case <-ctx.Done():
			logger.Log.Info("Settlement worker stopped")
			return
		}
	}
}

// Auto-Void Worker - Runs every hour
func startAutoVoidWorker(ctx context.Context, settlementService *service.SettlementService) {
	logger.Log.Info("Auto-void worker started")

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run immediately on startup
	if err := settlementService.AutoVoidExpiredAuthorizations(ctx); err != nil {
		logger.Log.Error("Auto-void failed", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
			logger.Log.Info("Running auto-void expired authorizations")
			if err := settlementService.AutoVoidExpiredAuthorizations(ctx); err != nil {
				logger.Log.Error("Auto-void failed", zap.Error(err))
			}

		case <-ctx.Done():
			logger.Log.Info("Auto-void worker stopped")
			return
		}
	}
}

// Currency Update Worker - Updates exchange rates every 24 hour
func startCurrencyUpdateWorker(ctx context.Context, currencyService *service.CurrencyService) {
	logger.Log.Info("Currency update worker started")

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run immediately on startup
	if err := currencyService.UpdateExchangeRates(ctx); err != nil {
		logger.Log.Error("Currency update failed", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
			logger.Log.Info("Updating exchange rates")
			if err := currencyService.UpdateExchangeRates(ctx); err != nil {
				logger.Log.Error("Currency update failed", zap.Error(err))
			}

		case <-ctx.Done():
			logger.Log.Info("Currency update worker stopped")
			return
		}
	}
}
