package main

import (
	pb "auth/api"
	"auth/internal/services"
	"fmt"
	"net"
	"os"

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
