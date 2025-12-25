package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/api"
	"go.uber.org/zap"
)

func init() {
	if os.Getenv("APP_MODE") == "" {
		inits.InitDotEnv()
	}
	inits.InitDB()
	inits.InitRedis()
	logger.Init()
	api.SetupMerchantRoutes()
}

func main() {
	defer logger.Sync()

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
