package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rhaloubi/payment-gateway/transaction-service/config"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits"
	"github.com/rhaloubi/payment-gateway/transaction-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/transaction-service/internal/service"
	"go.uber.org/zap"
)

func init() {
	if config.GetEnv("APP_MODE") == "" {
		inits.InitDotEnv()
	}
	logger.Init()
	inits.InitDB()
	inits.InitRedis()
}

func main() {
	defer logger.Sync()

	logger.Log.Info("Starting Transaction Service...")

	// Create services
	settlementService := service.NewSettlementService()
	currencyService := service.NewCurrencyService()

	// Context for background workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background workers
	go startSettlementWorker(ctx, settlementService)
	go startAutoVoidWorker(ctx, settlementService)
	go startCurrencyUpdateWorker(ctx, currencyService)

	// Get gRPC port
	grpcPort := config.GetEnv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50053"
	}

	// Start gRPC server
	go startGRPCServer(grpcPort)

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	port := config.GetEnv("PORT")
	if port == "" {
		port = "8005"
	}

	logger.Log.Info("✅ Transaction Service running",
		zap.String("grpc_port", grpcPort),
	)
	logger.Log.Info("Press Ctrl+C to stop...")

	<-stop
	logger.Log.Warn("🛑 Shutting down gracefully...")

	// Stop background workers
	cancel()

	// Close Redis connection
	if err := inits.RDB.Close(); err != nil {
		logger.Log.Error("Error closing Redis", zap.Error(err))
	} else {
		logger.Log.Info("🧹 Redis connection closed")
	}

	// Close database connection
	sqlDB, err := inits.DB.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.Log.Error("Error closing database", zap.Error(err))
		} else {
			logger.Log.Info("🧹 Database connection closed")
		}
	}

	logger.Log.Info("✅ Shutdown complete")
}
