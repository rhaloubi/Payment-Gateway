package util

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

// InitGRPC initializes and returns the gRPC server and listener (without starting it)
func InitGRPC() (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", ":"+os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", os.Getenv("GRPC_PORT"), err)
	}

	grpcServer := grpc.NewServer()

	return grpcServer, lis
}
