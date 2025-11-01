package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	"github.com/rhaloubi/payment-gateway/auth-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/api"
	"go.uber.org/zap"
)

func init() {
	inits.InitDotEnv()
	inits.InitDB()
	inits.InitRedis()
	logger.Init()
	api.Routes()
}

func main() {
	defer logger.Sync() // flush logs before exit

	// Create a channel to listen for interrupt (Ctrl+C) or system termination
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run the API server in a goroutine so we can listen for shutdown
	go func() {
		if err := api.R.Run(); err != nil {
			logger.Log.Error("Server error", zap.Error(err))
		}
	}()

	logger.Log.Info("✅ Server running... Press Ctrl+C to stop.")

	// Wait until a stop signal is received
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
