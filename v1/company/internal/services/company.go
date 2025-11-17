package services

import (
	"context"

	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/v1/company/api"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/company/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/v1/company/internal/database/redis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CompanyService struct {
	db    *postgresDB.DatabaseRepository
	cache *redisDB.CacheRepository
	pb.UnimplementedCompanyServiceServer
}

func NewCompanyService(db *postgresDB.DatabaseRepository, cache *redisDB.CacheRepository) *CompanyService {
	return &CompanyService{
		db:    db,
		cache: cache,
	}
}

// Health Проверка состояния сервиса
func (s *CompanyService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "health").Msg("success")
	return &pb.HealthResponse{Health: "healthy"}, nil
}

// CreateCompany Создает компанию
func (s *CompanyService) CreateCompany(ctx context.Context, req *pb.CreateCompanyRequest) (*pb.CreateCompanyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// GetCompany Возвращает всю информацию о компании
func (s *CompanyService) GetCompany(ctx context.Context, req *pb.GetCompanyRequest) (*pb.GetCompanyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// GetCompanies Возвращает список всех компаний (count, offset)
func (s *CompanyService) GetCompanies(ctx context.Context, req *pb.GetCompaniesRequest) (*pb.GetCompaniesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// UpdateCompanyTitle Обновляет название компании
func (s *CompanyService) UpdateCompanyTitle(ctx context.Context, req *pb.UpdateCompanyTitleRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// UpdateCompanyStatus Обновляет статус компании (open | close)
func (s *CompanyService) UpdateCompanyStatus(ctx context.Context, req *pb.UpdateCompanyStatusRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// DeleteCompany Удаляет компанию
func (s *CompanyService) DeleteCompany(ctx context.Context, req *pb.DeleteCompanyRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// CreateCompanyJoinCode Создает код для добавления в компанию
func (s *CompanyService) CreateCompanyJoinCode(ctx context.Context, req *pb.CreateCompanyJoinCodeRequest) (*pb.CreateCompanyJoinCodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// GetCompanyJoinCodes Возвращает все активные коды для добавления к компании
func (s *CompanyService) GetCompanyJoinCodes(ctx context.Context, req *pb.GetCompanyJoinCodesRequest) (*pb.GetCompanyJoinCodesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// DeleteCompanyJoinCode Удаляет код добавления в компанию
func (s *CompanyService) DeleteCompanyJoinCode(ctx context.Context, req *pb.DeleteCompanyJoinCodeRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// JoinCompany Добавляет пользователя в компанию
func (s *CompanyService) JoinCompany(ctx context.Context, req *pb.JoinCompanyRequest) (*pb.JoinCompanyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// GetCompanyEmployee Возвращает роль сотрудника в компании, иначе возвращает ошибку
func (s *CompanyService) GetCompanyEmployee(ctx context.Context, req *pb.GetCompanyEmployeeRequest) (*pb.GetCompanyEmployeeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// GetCompanyEmployeesSummary Возвращает кол-во сотрудников компании по ролям
func (s *CompanyService) GetCompanyEmployeesSummary(ctx context.Context, req *pb.GetCompanyEmployeesSummaryRequest) (*pb.GetCompanyEmployeesSummaryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// RemoveCompanyEmployee Удаляет сотрудника из компании
func (s *CompanyService) RemoveCompanyEmployee(ctx context.Context, req *pb.RemoveCompanyEmployeeRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}
