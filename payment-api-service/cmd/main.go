package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rhaloubi/payment-gateway/payment-api-service/inits"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/api"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/service"
	"go.uber.org/zap"
)

func init() {
	inits.InitDotEnv()
	inits.InitDB()
	inits.InitRedis()
	logger.Init()
	api.SetupRoutes(inits.R)
}

func main() {
	defer logger.Sync()

	// Start webhook retry worker
	webhookService := service.NewWebhookService()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := webhookService.RetryFailedWebhooks(ctx); err != nil {
			logger.Log.Error("Webhook retry worker failed", zap.Error(err))
		}
	}()
	logger.Log.Info("Webhook retry worker started")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := inits.R.Run(); err != nil {
			logger.Log.Error("Server error", zap.Error(err))
		}
	}()

	logger.Log.Info("✅ Server running... Press Ctrl+C to stop.")

	<-stop
	logger.Log.Warn("🛑 Shutting down gracefully...")

	// ✅ Close Redis connection
	if err := inits.RDB.Close(); err != nil {
		logger.Log.Error("Error closing Redis", zap.Error(err))
	} else {
		logger.Log.Info("🧹 Redis connection closed.")
	}

	logger.Log.Info("✅ Shutdown complete.")
}
