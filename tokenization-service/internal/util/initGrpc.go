package util

import (
	"log"
	"net"

	"github.com/rhaloubi/payment-gateway/tokenization-service/config"
	"google.golang.org/grpc"
)

// InitGRPC initializes and returns the gRPC server and listener (without starting it)
func InitGRPC() (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", ":"+config.GetEnv("GRPC_PORT"))
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", config.GetEnv("GRPC_PORT"), err)
	}

	grpcServer := grpc.NewServer()

	return grpcServer, lis
}
