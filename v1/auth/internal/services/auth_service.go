package services

import (
	pb "auth/api"
	"context"
	"fmt"
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
	return &pb.HealthResponse{Health: "Auth service healthy"}, nil
}
