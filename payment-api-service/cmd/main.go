package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rhaloubi/payment-gateway/payment-api-service/config"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/api"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/service"
	"go.uber.org/zap"
)

func init() {
	if config.GetEnv("APP_MODE") == "" {
		inits.InitDotEnv()
	}
	logger.Init()
	inits.InitDB()
	inits.InitRedis()
	api.SetupRoutes(inits.R)
}

func main() {
	defer logger.Sync()

	logger.Log.Info("Starting Payment API Service...")

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

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server
	go func() {
		if err := inits.R.Run(); err != nil {
			logger.Log.Error("Server error", zap.Error(err))
		}
	}()

	port := config.GetEnv("PORT")
	if port == "" {
		port = "8004"
	}

	logger.Log.Info("✅ Payment API Service running on port " + port)
	logger.Log.Info("Press Ctrl+C to stop...")

	<-stop
	logger.Log.Warn("🛑 Shutting down gracefully...")

	// Stop webhook worker
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
