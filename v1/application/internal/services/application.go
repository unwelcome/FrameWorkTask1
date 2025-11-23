package services

import (
	"context"

	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/v1/application/api"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/application/internal/database/postgres"
)

type ApplicationService struct {
	db *postgresDB.DatabaseRepository
	pb.UnimplementedApplicationServiceServer
}

func NewApplicationService(db *postgresDB.DatabaseRepository) *ApplicationService {
	return &ApplicationService{
		db: db,
	}
}

// Health Проверка состояния сервиса
func (s *ApplicationService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "health").Msg("success")
	return &pb.HealthResponse{Health: "healthy"}, nil
}
