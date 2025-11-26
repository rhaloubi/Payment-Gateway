package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/api"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/grpc"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/util"
	pb "github.com/rhaloubi/payment-gateway/tokenization-service/proto"
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

	// Initialize gRPC server and register service
	grpcServer := util.InitGRPC()
	pb.RegisterTokenizationServiceServer(grpcServer, grpc.NewTokenizationServer())

	httpServer := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: inits.R,
	}

	// Run HTTP server in goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Log.Info("🚀 HTTP (Gin) server running on :" + os.Getenv("PORT"))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Log.Warn("🛑 Shutting down gracefully...")

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Log.Error("HTTP server shutdown error", zap.Error(err))
	} else {
		logger.Log.Info("🧹 HTTP server stopped.")
	}

	// Shutdown gRPC server
	if grpcServer != nil {
		logger.Log.Info("🧹 Stopping gRPC server...")
		grpcServer.GracefulStop()
	}

	// Wait for HTTP goroutine to finish
	wg.Wait()

	// Close Redis connection
	if err := inits.RDB.Close(); err != nil {
		logger.Log.Error("Error closing Redis", zap.Error(err))
	} else {
		logger.Log.Info("🧹 Redis connection closed.")
	}

	logger.Log.Info("✅ Shutdown complete.")
}
