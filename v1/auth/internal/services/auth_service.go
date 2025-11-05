package services

import (
	"context"
	"fmt"

	pb "github.com/unwelcome/FrameWorkTask1/v1/auth/api"
)

type AuthService struct {
	pb.UnimplementedAuthServiceServer
}

func NewauthService() *AuthService {
	return &AuthService{}
}

// Health
func (s *AuthService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	fmt.Printf("ID: %s Msg: health check\n", req.OperationId)
	return &pb.HealthResponse{Health: "healthy"}, nil
}
