package services

import (
	"context"
	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/v1/auth/api"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/auth/internal/database/postgres"
)

type AuthService struct {
	db *postgresDB.DatabaseRepository
	pb.UnimplementedAuthServiceServer
}

func NewAuthService(db *postgresDB.DatabaseRepository) *AuthService {
	return &AuthService{db: db}
}

// Health check
func (s *AuthService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Info().Str("id", req.OperationId).Str("method", "health").Msg("request")
	return &pb.HealthResponse{Health: "healthy"}, nil
}
