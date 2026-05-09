package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/backend/application/api/generated"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/company/api/generated"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var AllApplicationStatuses = []string{"created", "assigned", "redirected", "rejected", "in_progress", "on_hold", "awaiting_approval", "completed", "cancelled", "failed", "archived"}

type ApplicationService struct {
	db            *postgresDB.DatabaseRepository
	companyClient company_proto.CompanyServiceClient
	pb.UnimplementedApplicationServiceServer
}

func NewApplicationService(db *postgresDB.DatabaseRepository, companyClient company_proto.CompanyServiceClient) *ApplicationService {
	return &ApplicationService{
		db:            db,
		companyClient: companyClient,
	}
}

// Health Проверка состояния сервиса
func (s *ApplicationService) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Info().Str("id", req.GetOperationId()).Str("method", "health").Msg("success")
	return &pb.HealthResponse{
		Service:  "healthy",
		Postgres: pingStatus(s.db.Ping(ctx)),
		Redis:    "not implemented",
		Minio:    "not implemented",
		Mongo:    "not implemented",
	}, nil
}

// CreateApplication Создание новой заявки (только inspector)
func (s *ApplicationService) CreateApplication(ctx context.Context, req *pb.CreateApplicationRequest) (*pb.CreateApplicationResponse, error) {
	// Получаем инициатора
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "create application", req.GetCompanyUuid(), req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	// Создавать заявки может только inspector
	if initiator.Role != "inspector" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "create application").Err(fmt.Errorf("not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only inspectors can create applications")
	}

	// Генерируем uuid для новой заявки
	applicationUUID := uuid.New().String()

	// Создаем заявку
	createErr := s.db.ApplicationRepository.CreateApplication(ctx, entities.CreateApplicationDTO{
		ApplicationUUID: applicationUUID,
		CompanyUUID:     req.GetCompanyUuid(),
		DepartmentUUID:  initiator.DepartmentUUID,
		Title:           req.GetTitle(),
		Description:     req.GetDescription(),
		CreatedBy:       req.GetInitiatorUuid(),
	})
	err = Error.HandleError(createErr, req.GetOperationId(), "create application")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "create application").Msg("success")
	return &pb.CreateApplicationResponse{ApplicationUuid: applicationUUID}, nil
}

// GetApplication Получение полной информации о заявке
func (s *ApplicationService) GetApplication(ctx context.Context, req *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "get application")
	if err != nil {
		return nil, err
	}

	// Проверяем, что пользователь принадлежит компании
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "get application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	// Работники без роли не могут просматривать заявки
	if initiator.Role == "unemployed" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get application").Err(fmt.Errorf("not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "not allowed to get application")
	}

	// Получаем fix log-и заявки
	fixLogs, getLogsErr := s.db.ApplicationRepository.GetApplicationFixLogs(ctx, entities.GetApplicationFixLogsDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err = Error.HandleError(getLogsErr, req.GetOperationId(), "get application")
	if err != nil {
		return nil, err
	}

	// Маппинг fix log-ов
	pbFixLogs := make([]*pb.FixLog, 0, len(fixLogs))
	for _, fl := range fixLogs {
		pbFixLogs = append(pbFixLogs, &pb.FixLog{
			Uuid:      fl.UUID,
			Text:      fl.Text,
			CreatedAt: fl.CreatedAt,
			CreatedBy: fl.CreatedBy,
		})
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get application").Msg("success")
	return &pb.GetApplicationResponse{
		Application: &pb.Application{
			ApplicationUuid:     application.ApplicationUUID,
			CompanyUuid:         application.CompanyUUID,
			DepartmentUuid:      application.DepartmentUUID,
			Title:               application.Title,
			Description:         application.Description,
			Status:              application.Status,
			ResponsibleManager:  application.ResponsibleManager,
			ResponsibleEngineer: application.ResponsibleEngineer,
			CreatedAt:           application.CreatedAt,
			CreatedBy:           application.CreatedBy,
			ClosedAt:            application.ClosedAt,
			FixLogs:             pbFixLogs,
		},
	}, nil
}

// GetApplications Получение списка заявок компании
func (s *ApplicationService) GetApplications(ctx context.Context, req *pb.GetApplicationsRequest) (*pb.GetApplicationsResponse, error) {
	// Получаем инициатора
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "get applications", req.GetCompanyUuid(), req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	// Работники без роли не могут просматривать заявки
	if initiator.Role == "unemployed" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "not allowed to get application")
	}

	// Валидация count
	count := req.GetCount()
	if count <= 0 || count > 100 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid count")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid count (1..100)")
	}

	// Валидация offset
	offset := req.GetOffset()
	if offset < 0 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid offset")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}

	// Валидация status
	applicationStatus := req.GetStatus()
	if applicationStatus != "" && !checkArrayContain(AllApplicationStatuses, applicationStatus) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid status")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid status")
	}

	// Получаем список заявок (с фильтрацией по status и department_uuid)
	applications, getErr := s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
		CompanyUUID:    req.GetCompanyUuid(),
		DepartmentUUID: req.GetDepartmentUuid(),
		Status:         applicationStatus,
		Count:          int(count),
		Offset:         int(offset),
	})
	err = Error.HandleError(getErr, req.GetOperationId(), "get applications")
	if err != nil {
		return nil, err
	}

	// Маппинг ответа
	pbApplications := make([]*pb.Application, 0, len(applications))
	for _, app := range applications {
		pbApplications = append(pbApplications, &pb.Application{
			ApplicationUuid:     app.ApplicationUUID,
			CompanyUuid:         app.CompanyUUID,
			DepartmentUuid:      app.DepartmentUUID,
			Title:               app.Title,
			Description:         app.Description,
			Status:              app.Status,
			ResponsibleManager:  app.ResponsibleManager,
			ResponsibleEngineer: app.ResponsibleEngineer,
			CreatedAt:           app.CreatedAt,
			CreatedBy:           app.CreatedBy,
			ClosedAt:            app.ClosedAt,
		})
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Msg("success")
	return &pb.GetApplicationsResponse{Applications: pbApplications}, nil
}

// GetCompanyApplicationStatistic Статистика по заявкам компании
func (s *ApplicationService) GetCompanyApplicationStatistic(ctx context.Context, req *pb.GetCompanyApplicationStatisticRequest) (*pb.GetCompanyApplicationStatisticResponse, error) {
	// Получаем инициатора
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "get company statistic", req.GetCompanyUuid(), req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	if !checkArrayContain([]string{"analytic", "chief"}, initiator.Role) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get company statistic").Err(fmt.Errorf("not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "not allowed to get company statistic")
	}

	// Получаем статистику компании
	statistic, getErr := s.db.ApplicationRepository.GetCompanyApplicationStatistic(ctx, entities.GetCompanyApplicationStatisticDTO{
		CompanyUUID: req.GetCompanyUuid(),
	})
	err = Error.HandleError(getErr, req.GetOperationId(), "get company statistic")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get company statistic").Msg("success")
	return &pb.GetCompanyApplicationStatisticResponse{
		Created:          int64(statistic.Created),
		Assigned:         int64(statistic.Assigned),
		Redirected:       int64(statistic.Redirected),
		Rejected:         int64(statistic.Rejected),
		InProgress:       int64(statistic.InProgress),
		OnHold:           int64(statistic.OnHold),
		AwaitingApproval: int64(statistic.AwaitingApproval),
		Completed:        int64(statistic.Completed),
		Cancelled:        int64(statistic.Cancelled),
		Failed:           int64(statistic.Failed),
		Archived:         int64(statistic.Archived),
	}, nil
}

// GetEmployeeApplicationStatistic Статистика по заявкам конкретного сотрудника
func (s *ApplicationService) GetEmployeeApplicationStatistic(ctx context.Context, req *pb.GetEmployeeApplicationStatisticRequest) (*pb.GetEmployeeApplicationStatisticResponse, error) {
	// Получаем инициатора
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "get employee statistic", req.GetCompanyUuid(), req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	if initiator.Role == "unemployed" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get employee statistic").Err(fmt.Errorf("not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "not allowed to get employee statistic")
	}

	// Получаем сотрудника
	target, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "get employee statistic", req.GetCompanyUuid(), req.GetInitiatorUuid(), req.GetTargetUuid())
	if err != nil {
		return nil, err
	}

	statistic, getErr := s.db.ApplicationRepository.GetEmployeeApplicationStatistic(ctx, entities.GetEmployeeApplicationStatisticDTO{
		CompanyUUID: req.GetCompanyUuid(),
		TargetUUID:  req.GetTargetUuid(),
		TargetRole:  target.Role,
	})
	err = Error.HandleError(getErr, req.GetOperationId(), "get employee statistic")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get employee statistic").Msg("success")
	return &pb.GetEmployeeApplicationStatisticResponse{
		Created:          int64(statistic.Created),
		Assigned:         int64(statistic.Assigned),
		Redirected:       int64(statistic.Redirected),
		Rejected:         int64(statistic.Rejected),
		InProgress:       int64(statistic.InProgress),
		OnHold:           int64(statistic.OnHold),
		AwaitingApproval: int64(statistic.AwaitingApproval),
		Completed:        int64(statistic.Completed),
		Cancelled:        int64(statistic.Cancelled),
		Failed:           int64(statistic.Failed),
		Archived:         int64(statistic.Archived),
	}, nil
}

// UpdateApplicationStatus Обновление статуса заявки
func (s *ApplicationService) UpdateApplicationStatus(ctx context.Context, req *pb.UpdateApplicationStatusRequest) (*emptypb.Empty, error) {
	newStatus := req.GetStatus()

	// Допустимые статусы для инженера и менеджера
	engineerStatuses := []string{"in_progress", "on_hold", "awaiting_approval"}
	managerStatuses := []string{"rejected", "completed", "cancelled", "failed"}
	validEngineerCurrentStatuses := []string{"assigned", "in_progress", "on_hold", "awaiting_approval"}
	validManagerCurrentStatuses := []string{"created", "redirected", "awaiting_approval"}

	// Проверяем, что статус вообще допустим в этом методе
	if !checkArrayContain(engineerStatuses, newStatus) && !checkArrayContain(managerStatuses, newStatus) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").Err(fmt.Errorf("invalid status")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "update application status")
	if err != nil {
		return nil, err
	}

	// Получаем инициатора
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "update application status", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	// Роли без доступа к изменению статусов заявок
	forbiddenRoles := []string{"inspector", "chief", "unemployed", "analytic"}
	if checkArrayContain(forbiddenRoles, initiator.Role) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
			Err(fmt.Errorf("role cannot change application status")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "not enough rights to change application status")
	}

	currentStatus := application.Status

	switch initiator.Role {
	case "engineer":
		// Инженер может устанавливать только свои статусы
		if !checkArrayContain(engineerStatuses, newStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("engineer cannot set status %q", newStatus)).Msg("error")
			return nil, status.Error(codes.PermissionDenied, "engineers can only set in_progress, on_hold or awaiting_approval")
		}
		// Только ответственный инженер
		if application.ResponsibleEngineer != req.GetInitiatorUuid() {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("user is not the responsible engineer")).Msg("error")
			return nil, status.Error(codes.PermissionDenied, "only the responsible engineer can change status")
		}
		// Текущий статус должен допускать переход
		if !checkArrayContain(validEngineerCurrentStatuses, currentStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("cannot transition from status %q", currentStatus)).Msg("error")
			return nil, status.Error(codes.FailedPrecondition, "invalid status transition")
		}

	case "manager":
		// Менеджер может устанавливать только свои статусы
		if !checkArrayContain(managerStatuses, newStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("manager cannot set status %q", newStatus)).Msg("error")
			return nil, status.Error(codes.PermissionDenied, "managers can only set assigned, redirected, rejected, completed, cancelled or failed")
		}
		// Текущий статус должен допускать переход
		if !checkArrayContain(validManagerCurrentStatuses, currentStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("cannot transition from status %q", currentStatus)).Msg("error")
			return nil, status.Error(codes.FailedPrecondition, "invalid status transition")
		}

		// После "create" и "redirect" можно только "assigned", "redirect" или "rejected"
		if checkArrayContain([]string{"created", "redirected"}, currentStatus) {
			if !checkArrayContain([]string{"assigned", "redirected", "rejected"}, req.GetStatus()) {
				log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
					Err(fmt.Errorf("invalid status")).Msg("error")
				return nil, status.Error(codes.PermissionDenied, "after \"create\" or \"redirect\" status can be only \"assigned\" or \"rejected\"")
			}
		}
		
		// После "awaiting_approval" можно только "completed", "cancelled", "failed"
		if currentStatus == "awaiting_approval" {
			if !checkArrayContain([]string{"completed", "cancelled", "failed"}, req.GetStatus()) {
				log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
					Err(fmt.Errorf("invalid status")).Msg("error")
				return nil, status.Error(codes.PermissionDenied, "after \"awaiting_approval\" status can be only \"completed\", \"cancelled\" or \"failed\"")
			}
		}

		// Если текущий статус "redirected"
		if application.Status == "redirected" {
			// Любой менеджер с совпадающим department_uuid может менять статус заявки
			if initiator.DepartmentUUID != application.DepartmentUUID {
				log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
					Err(fmt.Errorf("user is not in responsible department")).Msg("error")
				return nil, status.Error(codes.PermissionDenied, "you are not in responsible department")
			}
		} else { // Если текущий статус не "redirected"
			// Только ответственный менеджер
			if application.ResponsibleManager != req.GetInitiatorUuid() {
				log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
					Err(fmt.Errorf("user is not the responsible manager")).Msg("error")
				return nil, status.Error(codes.PermissionDenied, "only the responsible manager can change status")
			}
		}
	}

	// Обновляем статус заявки
	updateErr := s.db.ApplicationRepository.UpdateApplicationStatus(ctx, entities.UpdateApplicationStatusDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		Status:          newStatus,
		InitiatorUUID:   req.GetInitiatorUuid(),
	})
	err = Error.HandleError(updateErr, req.GetOperationId(), "update application status")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").Msg("success")
	return &emptypb.Empty{}, nil
}

// TransferApplication Передача заявки в другой департамент
func (s *ApplicationService) TransferApplication(ctx context.Context, req *pb.TransferApplicationRequest) (*emptypb.Empty, error) {
	panic("implement me")
}

// AssignApplicationToEmployee Назначение заявки инженеру
func (s *ApplicationService) AssignApplicationToEmployee(ctx context.Context, req *pb.AssignApplicationToEmployeeRequest) (*emptypb.Empty, error) {
	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "assign application")
	if err != nil {
		return nil, err
	}

	// Получаем роль инициатора — назначать заявки могут только manager
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "assign application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	if initiator.Role != "manager" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only managers can assign applications")
	}

	// Получаем роль целевого сотрудника — назначить можно только для engineer
	target, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "assign application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetTargetUuid())
	if err != nil {
		return nil, err
	}

	if target.Role != "engineer" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("target role is not engineer")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "application can only be assigned to an engineer")
	}

	// Назначаем заявку инженеру
	assignErr := s.db.ApplicationRepository.AssignApplicationToEmployee(ctx, entities.AssignApplicationToEmployeeDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
		TargetUUID:      req.GetTargetUuid(),
	})
	err = Error.HandleError(assignErr, req.GetOperationId(), "assign application")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Msg("success")
	return &emptypb.Empty{}, nil
}

// AddApplicationFixLog Добавление новой записи в fix log заявки
func (s *ApplicationService) AddApplicationFixLog(ctx context.Context, req *pb.AddApplicationFixLogRequest) (*emptypb.Empty, error) {
	// Получаем заявку — добавлять записи может только ответственный инженер (executed_by)
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "add fix log")
	if err != nil {
		return nil, err
	}

	if application.ResponsibleEngineer != req.GetInitiatorUuid() {
		log.Info().Str("id", req.GetOperationId()).Str("method", "add fix log").Err(fmt.Errorf("not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only the responsible engineer can add fix logs")
	}

	// Создаем новую запись в fix log
	addErr := s.db.ApplicationRepository.AddApplicationFixLog(ctx, entities.CreateFixLogDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		Text:            req.GetLogText(),
		CreatedBy:       req.GetInitiatorUuid(),
	})
	err = Error.HandleError(addErr, req.GetOperationId(), "add fix log")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "add fix log").Msg("success")
	return &emptypb.Empty{}, nil
}

// DeleteApplication Мягкое удаление заявки
func (s *ApplicationService) DeleteApplication(ctx context.Context, req *pb.DeleteApplicationRequest) (*emptypb.Empty, error) {
	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "delete application")
	if err != nil {
		return nil, err
	}

	// Если статус заявки не created, то ее уже нельзя удалить
	if application.Status != "created" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete application").Err(fmt.Errorf("application is already in use")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only applications with status 'created' can be deleted")
	}

	// Получаем роль инициатора — удалять заявки могут только inspector
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "delete application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	if initiator.Role != "inspector" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete application").
			Err(fmt.Errorf("role is not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only inspectors can delete applications")
	}

	// Мягкое удаление (deleted_at = now(), deleted_by = initiator)
	deleteErr := s.db.ApplicationRepository.DeleteApplicationRequest(ctx, entities.DeleteApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		DeletedBy:       req.GetInitiatorUuid(),
	})
	err = Error.HandleError(deleteErr, req.GetOperationId(), "delete application")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "delete application").Msg("success")
	return &emptypb.Empty{}, nil
}

// ─── Вспомогательные функции ──────────────────────────────────────────────────

// pingStatus Обертка для ответа пинга бд
func pingStatus(err error) string {
	if err != nil {
		return "not connected"
	}
	return "connected"
}

// getEmployeeRole получает роль сотрудника из company сервиса.
// initiatorUUID — кто делает запрос (нужен для проверки прав в company сервисе),
// targetUUID — чью роль мы хотим узнать.
func (s *ApplicationService) getEmployeeInfo(ctx context.Context, opID, method, companyUUID, initiatorUUID, targetUUID string) (*entities.Employee, error) {
	employeeInfo, err := s.companyClient.GetCompanyEmployee(ctx, &company_proto.GetCompanyEmployeeRequest{
		OperationId:   opID,
		CompanyUuid:   companyUUID,
		InitiatorUuid: initiatorUUID,
		TargetUuid:    targetUUID,
	})
	if err != nil {
		log.Error().Str("id", opID).Str("method", method).Err(err).Msg("failed to get employee role from company service")
		return nil, err
	}

	return &entities.Employee{
		UUID:           initiatorUUID,
		Role:           employeeInfo.Role,
		DepartmentUUID: employeeInfo.DepartmentUuid,
	}, nil
}

// checkArrayContain Проверяет наличие строки в массиве строк
func checkArrayContain(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}
