package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/v1/application/api"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/v1/application/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/application/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"
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

// CreateApplication Создание новой заявки
func (s *ApplicationService) CreateApplication(ctx context.Context, req *pb.CreateApplicationRequest) (*pb.CreateApplicationResponse, error) {
	// Генерируем uuid для новой заявки
	applicationUUID := uuid.New().String()

	// Создаем заявку
	createErr := s.db.ApplicationRepository.CreateApplication(ctx, entities.CreateApplicationDTO{
		ApplicationUUID: applicationUUID,
		CompanyUUID:     req.GetCompanyUuid(),
		Title:           req.GetTitle(),
		Description:     req.GetDescription(),
		CreatedBy:       req.GetUserUuid(),
	})
	err := Error.HandleError(createErr, req.GetOperationId(), "create application")
	if err != nil {
		log.Error().Str("id", req.GetOperationId()).Str("method", "create application").Err(fmt.Errorf("create application error: %w", err)).Msg("error")
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "create application").Msg("success")
	return &pb.CreateApplicationResponse{ApplicationUuid: applicationUUID}, nil
}

// AddApplicationFixLog Добавление новой записи в fixlog заявки
func (s *ApplicationService) AddApplicationFixLog(ctx context.Context, req *pb.AddApplicationFixLogRequest) (*emptypb.Empty, error) {
	// Создаем новую запись в fix log
	addErr := s.db.ApplicationRepository.AddApplicationFixLog(ctx, entities.CreateFixLogDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		Text:            req.GetLogText(),
		CreatedBy:       req.GetUserUuid(),
	})
	err := Error.HandleError(addErr, req.GetOperationId(), "add fixlog")
	if err != nil {
		log.Error().Str("id", req.GetOperationId()).Str("method", "add fixlog").Err(fmt.Errorf("add fixlog error: %w", err)).Msg("error")
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "add fixlog").Msg("success")
	return &emptypb.Empty{}, nil
}

// AssignApplicationToEmployee Назначение заявки инженеру
func (s *ApplicationService) AssignApplicationToEmployee(ctx context.Context, req *pb.AssignApplicationToEmployeeRequest) (*emptypb.Empty, error) {
	// Проверяем роль инициатора
	// initiator, err :=

	// TODO
	// [ ] Добавить в company service бэкдор в метод GetEmployee чтобы другой сервис мог получить информацию по сотруднику

	// Проверяем роль цели

	// Сохраняем копию старой заявки

	// Назначаем заявку

	// Сохраняем копию старой заявки в архив

	log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Msg("success")
	return &emptypb.Empty{}, nil
}

func (s *ApplicationService) GetApplication(ctx context.Context, req *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "get application").Msg("success")
	return &pb.GetApplicationResponse{}, nil
}

func (s *ApplicationService) GetApplications(ctx context.Context, req *pb.GetApplicationsRequest) (*pb.GetApplicationsResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Msg("success")
	return &pb.GetApplicationsResponse{}, nil
}

func (s *ApplicationService) GetCompanyApplicationStatistic(ctx context.Context, req *pb.GetCompanyApplicationStatisticRequest) (*pb.GetCompanyApplicationStatisticResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "get company statistic").Msg("success")
	return &pb.GetCompanyApplicationStatisticResponse{}, nil
}

func (s *ApplicationService) GetEmployeeApplicationStatistic(ctx context.Context, req *pb.GetEmployeeApplicationStatisticRequest) (*pb.GetEmployeeApplicationStatisticResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "get employee statistic").Msg("success")
	return &pb.GetEmployeeApplicationStatisticResponse{}, nil
}

func (s *ApplicationService) UpdateApplicationData(ctx context.Context, req *pb.UpdateApplicationDataRequest) (*emptypb.Empty, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "update application data").Msg("success")
	return &emptypb.Empty{}, nil
}

func (s *ApplicationService) UpdateApplicationStatus(ctx context.Context, req *pb.UpdateApplicationStatusRequest) (*emptypb.Empty, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").Msg("success")
	return &emptypb.Empty{}, nil
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, req *pb.DeleteApplicationRequest) (*emptypb.Empty, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "delete application").Msg("success")
	return &emptypb.Empty{}, nil
}
