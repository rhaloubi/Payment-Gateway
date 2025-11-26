package util

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

// InitGRPC initializes and returns the gRPC server (without registering services)
func InitGRPC() *grpc.Server {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ Failed to listen on port 50051: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Start serving in a goroutine
	go func() {
		log.Println("🚀 gRPC server running on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("❌ Failed to serve gRPC: %v", err)
		}
	}()

	return grpcServer
}
