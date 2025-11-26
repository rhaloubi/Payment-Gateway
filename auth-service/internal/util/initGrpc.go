package util

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

// InitGRPC initializes and returns the gRPC server (without registering services)
func InitGRPC() *grpc.Server {
	lis, err := net.Listen("tcp", ":"+os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", os.Getenv("GRPC_PORT"), err)
	}

	grpcServer := grpc.NewServer()

	// Start serving in a goroutine
	go func() {
		log.Printf("🚀 gRPC server running on :%s", os.Getenv("GRPC_PORT"))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("❌ Failed to serve gRPC: %v", err)
		}
	}()

	return grpcServer
}
