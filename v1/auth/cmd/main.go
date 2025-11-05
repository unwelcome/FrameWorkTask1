package main

import (
	"fmt"
	"net"
	"os"

	pb "github.com/unwelcome/FrameWorkTask1/v1/auth/api"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/internal/services"

	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Printf("Failed to start tcp server: %v", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, services.NewauthService())

	fmt.Printf("Auth service started on port: %s\n", "50051")
	if err := grpcServer.Serve(listener); err != nil {
		fmt.Printf("Failed to start grpc server: %v", err)
		os.Exit(1)
	}
}
