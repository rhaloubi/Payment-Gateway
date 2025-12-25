package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/grpc"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/util"
	pb "github.com/rhaloubi/payment-gateway/tokenization-service/proto"
	"go.uber.org/zap"
)

func init() {
	if os.Getenv("APP_MODE") == "" {
		inits.InitDotEnv()
	}
	inits.InitDB()
	inits.InitRedis()
	logger.Init()
}

func main() {
	defer logger.Sync()

	// Initialize gRPC server and register service
	grpcServer, lis := util.InitGRPC()
	pb.RegisterTokenizationServiceServer(grpcServer, grpc.NewTokenizationServer())

	// Start gRPC server in a goroutine
	go func() {
		logger.Log.Info("🚀 gRPC server running on :" + os.Getenv("GRPC_PORT"))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Log.Fatal("❌ Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Log.Warn("🛑 Shutting down gracefully...")

	// Shutdown gRPC server
	if grpcServer != nil {
		logger.Log.Info("🧹 Stopping gRPC server...")
		grpcServer.GracefulStop()
	}

	// Close Redis connection
	if err := inits.RDB.Close(); err != nil {
		logger.Log.Error("Error closing Redis", zap.Error(err))
	} else {
		logger.Log.Info("🧹 Redis connection closed.")
	}

	logger.Log.Info("✅ Shutdown complete.")
}
