package util

import (
	"log"
	"net"

	"github.com/rhaloubi/payment-gateway/auth-service/config"
	"google.golang.org/grpc"
)

// InitGRPC initializes and returns the gRPC server (without registering services)
func InitGRPC() *grpc.Server {
	lis, err := net.Listen("tcp", ":"+config.GetEnv("GRPC_PORT"))
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", config.GetEnv("GRPC_PORT"), err)
	}

	grpcServer := grpc.NewServer()

	// Start serving in a goroutine
	go func() {
		log.Printf("🚀 gRPC server running on :%s", config.GetEnv("GRPC_PORT"))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("❌ Failed to serve gRPC: %v", err)
		}
	}()

	return grpcServer
}
