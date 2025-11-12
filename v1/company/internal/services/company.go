package services

import (
	"context"

	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/v1/company/api"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/company/internal/database/postgres"
)

type CompanyService struct {
	db *postgresDB.DatabaseRepository
	pb.UnimplementedCompanyServiceServer
}

func NewCompanyService(db *postgresDB.DatabaseRepository) *CompanyService {
	return &CompanyService{
		db: db,
	}
}

// Health Проверка состояния сервиса
func (s *CompanyService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "health").Msg("success")
	return &pb.HealthResponse{Health: "healthy"}, nil
}
