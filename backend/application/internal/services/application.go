package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/application/generated"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/helpers"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/validate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var AllApplicationStatuses = []string{"created", "assigned", "in_progress", "on_hold", "completed", "failed", "redirected", "rejected", "recalled", "pending_verification", "on_verification", "on_revision"}

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
		Postgres: helpers.PingStatus(s.db.Ping(ctx)),
		Redis:    "not implemented",
		Minio:    "not implemented",
		Mongo:    "not implemented",
	}, nil
}

// CreateApplication Создание новой заявки (только inspector)
func (s *ApplicationService) CreateApplication(ctx context.Context, req *pb.CreateApplicationRequest) (*pb.CreateApplicationResponse, error) {
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get department").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get department").Err(fmt.Errorf("invalid company uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.ApplicationTitle(req.GetApplicationData().GetTitle()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get department").Err(fmt.Errorf("invalid application title")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application title")
	}
	if err := validate.ApplicationDescription(req.GetApplicationData().GetDescription()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get department").Err(fmt.Errorf("invalid application description")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application description")
	}

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
	applicationUUID := uuid.Must(uuid.NewV7()).String()

	// Создаем заявку
	createErr := s.db.ApplicationRepository.CreateApplication(ctx, entities.CreateApplicationDTO{
		ApplicationUUID: applicationUUID,
		CompanyUUID:     req.GetCompanyUuid(),
		DepartmentUUID:  initiator.DepartmentUUID,
		Title:           req.GetApplicationData().GetTitle(),
		Description:     req.GetApplicationData().GetDescription(),
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
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get application").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get application").Err(fmt.Errorf("invalid application uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}

	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "get application")
	if err != nil {
		return nil, err
	}

	// Получаем инициатора (проверяем что он числится в компании)
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "get application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	// Работники, чьи uuid не фигурируют в заявке не имеют к ней доступа (исключая chief и analytic)
	if !helpers.Contains([]string{"chief", "analytic"}, initiator.Role) &&
		!helpers.Contains([]string{application.CreatedBy, application.ManagedBy, application.ExecutedBy, application.InspectedBy}, req.GetInitiatorUuid()) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get application").Err(fmt.Errorf("not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to get application")
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
			ApplicationUuid: application.ApplicationUUID,
			CompanyUuid:     application.CompanyUUID,
			DepartmentUuid:  application.DepartmentUUID,
			Version:         int64(application.Version),
			Title:           application.Title,
			Description:     application.Description,
			RevisionCount:   int64(application.RevisionCount),
			Status:          application.Status,
			CreatedAt:       application.CreatedAt,
			CreatedBy:       application.CreatedBy,
			UpdatedAt:       application.UpdatedAt,
			UpdatedBy:       application.UpdatedBy,
			ManagedBy:       application.ManagedBy,
			ExecutedBy:      application.ExecutedBy,
			InspectedBy:     application.InspectedBy,
			ClosedAt:        application.ClosedAt,
			DeletedAt:       application.DeletedAt,
			DeletedBy:       application.DeletedBy,
			FixLogs:         pbFixLogs,
		},
	}, nil
}

// GetApplications Получение списка заявок
func (s *ApplicationService) GetApplications(ctx context.Context, req *pb.GetApplicationsRequest) (*pb.GetApplicationsResponse, error) {
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid company uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	// Валидация count
	if req.GetCount() <= 0 || req.GetCount() > 100 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid count")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid count (1..100)")
	}
	// Валидация offset
	if req.GetOffset() < 0 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid offset")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}

	// Получаем инициатора
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "get applications", req.GetCompanyUuid(), req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	var applications []*entities.Application
	var getErr Error.CodeError

	switch initiator.Role {

	// Если инициатор "chief" или "analytic" - используем department_uuid и status из запроса
	case "chief", "analytic":
		if err := validate.UUID(req.GetDepartmentUuid()); err != nil && req.GetDepartmentUuid() != "" {
			log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid department uuid")).Msg("error")
			return nil, status.Errorf(codes.InvalidArgument, "invalid department uuid")
		}
		// Валидация statuses
		if !helpers.ContainsAll(AllApplicationStatuses, req.GetStatuses()) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid statuses")).Msg("error")
			return nil, status.Errorf(codes.InvalidArgument, "invalid statuses")
		}

		applications, getErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
			CompanyUUID:    req.GetCompanyUuid(),
			DepartmentUUID: req.GetDepartmentUuid(),
			Statuses:       req.GetStatuses(),
			Offset:         int(req.GetOffset()),
			Count:          int(req.GetCount()),
			IsDeleted:      req.GetIsDeleted(),
		})

	// Если инициатор "inspector" - department_uuid инициатора, statuses: ["pending_verification", "on_verification"]
	case "inspector":
		if req.GetFromPool() {
			applications, getErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
				CompanyUUID:    req.GetCompanyUuid(),
				DepartmentUUID: initiator.DepartmentUUID,
				Statuses:       []string{"pending_verification"},
				Offset:         int(req.GetOffset()),
				Count:          int(req.GetCount()),
			})
		} else {
			createdBy := req.GetInitiatorUuid()
			inspectedBy := req.GetInitiatorUuid()

			// Если статусов в фильтре нет - выдаем все заявки, которые инспектор создал
			if len(req.GetStatuses()) == 0 {
				inspectedBy = ""
			} else { // Иначе - выдаем заявки, с которыми он работает
				createdBy = ""

				// Проверяем статусы в фильтре
				if !helpers.ContainsAll([]string{"on_verification"}, req.GetStatuses()) {
					log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid statuses")).Msg("error")
					return nil, status.Errorf(codes.InvalidArgument, "invalid statuses")
				}
			}

			applications, getErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
				CompanyUUID:    req.GetCompanyUuid(),
				DepartmentUUID: initiator.DepartmentUUID,
				Statuses:       req.GetStatuses(),
				CreatedBy:      createdBy,
				InspectedBy:    inspectedBy,
				Offset:         int(req.GetOffset()),
				Count:          int(req.GetCount()),
				IsDeleted:      req.GetIsDeleted(),
			})
		}

	// Если инициатор "manager" - department_uuid инициатора, statuses: ["created", "redirected", "recalled", "on_revision"]
	case "manager":
		if req.GetFromPool() {
			applications, getErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
				CompanyUUID:      req.GetCompanyUuid(),
				DepartmentUUID:   initiator.DepartmentUUID,
				Statuses:         []string{"created", "redirected", "recalled", "on_revision"},
				ExecutedByIsNull: true,
				Offset:           int(req.GetOffset()),
				Count:            int(req.GetCount()),
			})
		} else {
			applications, getErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
				CompanyUUID:    req.GetCompanyUuid(),
				DepartmentUUID: initiator.DepartmentUUID,
				ManagedBy:      req.GetInitiatorUuid(),
				Offset:         int(req.GetOffset()),
				Count:          int(req.GetCount()),
			})
		}

	// Если инициатор "engineer" - department_uuid инициатора, statuses: ["assigned", "on_revision", "in_progress", "on_hold"]
	case "engineer":
		// Валидация statuses
		if !helpers.ContainsAll([]string{"assigned", "on_revision", "in_progress", "on_hold"}, req.GetStatuses()) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("invalid statuses")).Msg("error")
			return nil, status.Errorf(codes.InvalidArgument, "invalid statuses")
		}

		applications, getErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
			CompanyUUID:    req.GetCompanyUuid(),
			DepartmentUUID: initiator.DepartmentUUID,
			Statuses:       req.GetStatuses(),
			ExecutedBy:     req.GetInitiatorUuid(),
			Offset:         int(req.GetOffset()),
			Count:          int(req.GetCount()),
		})

	default:
		log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Err(fmt.Errorf("not allowed")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "not allowed to get applications")

	}
	err = Error.HandleError(getErr, req.GetOperationId(), "get applications")
	if err != nil {
		return nil, err
	}

	// Маппинг ответа
	pbApplications := make([]*pb.Application, 0, len(applications))
	for _, app := range applications {
		pbApplications = append(pbApplications, &pb.Application{
			ApplicationUuid: app.ApplicationUUID,
			//CompanyUuid:     app.CompanyUUID,
			//DepartmentUuid:  app.DepartmentUUID,
			//Version:         int64(app.Version),
			Title: app.Title,
			//Description:     app.Description,
			Status: app.Status,
			//RevisionCount:   int64(app.RevisionCount),
			CreatedAt: app.CreatedAt,
			//CreatedBy:       app.CreatedBy,
			UpdatedAt: app.UpdatedAt,
			//UpdatedBy:       app.UpdatedBy,
			//ManagedBy:       app.ManagedBy,
			//ExecutedBy:      app.ExecutedBy,
			//InspectedBy:     app.InspectedBy,
			//ClosedAt:        app.ClosedAt,
			//DeletedAt:       app.DeletedAt,
			//DeletedBy:       app.DeletedBy,
		})
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "get applications").Msg("success")
	return &pb.GetApplicationsResponse{Applications: pbApplications}, nil
}

// UpdateApplicationStatus Обновление статуса заявки
func (s *ApplicationService) UpdateApplicationStatus(ctx context.Context, req *pb.UpdateApplicationStatusRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").Err(fmt.Errorf("invalid application uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}

	newStatus := req.GetStatus()

	// Проверяем, что статус допустим в данном методе
	if !helpers.Contains([]string{"rejected", "in_progress", "on_hold", "pending_verification", "completed", "failed", "on_revision"}, newStatus) {
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

	currentStatus := application.Status
	dropManagedBy := false
	dropExecutedBy := false

	switch initiator.Role {

	case "inspector":
		// Только ответственный инспектор
		if application.InspectedBy != req.GetInitiatorUuid() {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("user is not the responsible inspector")).Msg("error")
			return nil, status.Error(codes.PermissionDenied, "only the responsible inspector can change status")
		}

		// Допустимые статусы для инспектора
		if !helpers.Contains([]string{"completed", "failed", "on_revision"}, newStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("invalid inspector status")).Msg("error")
			return nil, status.Error(codes.PermissionDenied, "inspectors can only set \"completed\", \"failed\" or \"on_revision\"")
		}

		// Текущий статус должен допускать переход
		if !helpers.Contains([]string{"on_verification"}, currentStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("invalid status transition")).Msg("error")
			return nil, status.Error(codes.FailedPrecondition, "invalid status transition")
		}

		// Если новый статус on_revision и revision_count % 5 == 0, то отдаем заявку в общий пул менеджеров
		if newStatus == "on_revision" && (application.RevisionCount+1)%5 == 0 {
			dropManagedBy = true
			dropExecutedBy = true
		}

	case "manager":
		// Департамент должен совпадать
		if initiator.DepartmentUUID != application.DepartmentUUID {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("unresponsible department")).Msg("error")
			return nil, status.Error(codes.PermissionDenied, "your department is not responsible")
		}

		// Допустимые статусы для менеджера
		if !helpers.Contains([]string{"rejected"}, newStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("invalid manager status")).Msg("error")
			return nil, status.Error(codes.PermissionDenied, "managers can only set \"rejected\"")
		}

		// Текущий статус должен допускать переход
		if !helpers.Contains([]string{"created", "redirected", "recalled", "on_revision"}, currentStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("invalid status transition")).Msg("error")
			return nil, status.Error(codes.FailedPrecondition, "invalid status transition")
		}

	case "engineer":
		// Только ответственный инженер
		if application.ExecutedBy != req.GetInitiatorUuid() {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("user is not the responsible engineer")).Msg("error")
			return nil, status.Error(codes.PermissionDenied, "only the responsible engineer can change status")
		}

		// Допустимые статусы для инженера
		if !helpers.Contains([]string{"in_progress", "on_hold", "pending_verification"}, newStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("invalid engineer status")).Msg("error")
			return nil, status.Error(codes.PermissionDenied, "engineers can only set \"in_progress\", \"on_hold\" or \"pending_verification\"")
		}

		// Текущий статус должен допускать переход
		if !helpers.Contains([]string{"assigned", "in_progress", "on_hold", "on_revision"}, currentStatus) {
			log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
				Err(fmt.Errorf("invalid status transition")).Msg("error")
			return nil, status.Error(codes.FailedPrecondition, "invalid status transition")
		}

	default:
		log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").
			Err(fmt.Errorf("unallowed role")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "unallowed role")

	}

	// Обновляем статус заявки
	updateErr := s.db.ApplicationRepository.UpdateApplicationStatus(ctx, entities.UpdateApplicationStatusDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
		Status:          newStatus,
		DropManagedBy:   dropManagedBy,
		DropExecutedBy:  dropExecutedBy,
	})
	err = Error.HandleError(updateErr, req.GetOperationId(), "update application status")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "update application status").Msg("success")
	return &emptypb.Empty{}, nil
}

// AssignApplication Назначение инженера на выполнение заявки
func (s *ApplicationService) AssignApplication(ctx context.Context, req *pb.AssignApplicationRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("invalid application uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	if err := validate.UUID(req.GetTargetUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("invalid target uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid target uuid")
	}

	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "assign application")
	if err != nil {
		return nil, err
	}

	// Проверяем статус заявки
	if !helpers.Contains([]string{"created", "redirected", "recalled", "on_revision"}, application.Status) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("invalid status")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "can't assign application with current status")
	}
	// Проверяем, чтобы заявка со статусом on_revision имела кратное 5 кол-во пересмотров
	if application.Status == "on_revision" && application.RevisionCount%5 != 0 {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("revision is not")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "revision has not reached the boiling point")
	}

	// Получаем менеджера
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "assign application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "manager" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("not a manager")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only managers can assign applications")
	}
	if application.DepartmentUUID != initiator.DepartmentUUID {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("unresponsible department")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "your department is not responsible")
	}

	// Получаем инженера
	target, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "assign application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetTargetUuid())
	if err != nil {
		return nil, err
	}
	if target.Role != "engineer" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("target role is not engineer")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "application can only be assigned to an engineer")
	}
	if target.DepartmentUUID != initiator.DepartmentUUID {
		log.Info().Str("id", req.GetOperationId()).Str("method", "assign application").Err(fmt.Errorf("invalid engineer department")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "engineer is not from your department")
	}

	// Назначаем заявку инженеру
	assignErr := s.db.ApplicationRepository.AssignApplicationToEmployee(ctx, entities.AssignApplicationDTO{
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

// RedirectApplication Передача заявки в другой департамент
func (s *ApplicationService) RedirectApplication(ctx context.Context, req *pb.RedirectApplicationRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "redirect application").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "redirect application").Err(fmt.Errorf("invalid application uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	if err := validate.UUID(req.GetTargetDepartmentUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "redirect application").Err(fmt.Errorf("invalid target department uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid target depatment uuid")
	}
	// Валидация message
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "redirect application").Err(fmt.Errorf("invalid message")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	// Получаем заявку
	application, gerErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(gerErr, req.GetOperationId(), "redirect application")
	if err != nil {
		return nil, err
	}

	// Проверяем статус заявки
	if !helpers.Contains([]string{"created", "redirected", "recalled", "on_revision"}, application.Status) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "redirect application").Err(fmt.Errorf("invalid status")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "invalid application status")
	}

	// Получаем менеджера
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "redirect application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "manager" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "redirect application").Err(fmt.Errorf("not a manager")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only managers can redirect applications")
	}
	if initiator.DepartmentUUID != application.DepartmentUUID {
		log.Info().Str("id", req.GetOperationId()).Str("method", "redirect application").Err(fmt.Errorf("unresponsible department")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "your department is not responsible")
	}

	// Проверяем, что департамент принадлежит компании
	department, err := s.getDepartmentInfo(ctx, req.GetOperationId(), "redirect application", req.GetInitiatorUuid(), req.GetTargetDepartmentUuid())
	if err != nil {
		return nil, err
	}
	if department.CompanyUUID != application.CompanyUUID {
		log.Info().Str("id", req.GetOperationId()).Str("method", "redirect application").Err(fmt.Errorf("invalid department")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "department is from another company")
	}

	// Пишем fix log
	addFixLogErr := s.db.ApplicationRepository.AddApplicationFixLog(ctx, entities.AddFixLogDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		Text:            message,
		CreatedBy:       req.GetInitiatorUuid(),
	})
	err = Error.HandleError(addFixLogErr, req.GetOperationId(), "redirect application")
	if err != nil {
		return nil, err
	}

	// Передаем заявку в другой департамент
	redirectErr := s.db.ApplicationRepository.RedirectApplication(ctx, entities.RedirectApplicationDTO{
		ApplicationUUID:      req.GetApplicationUuid(),
		InitiatorUUID:        req.GetInitiatorUuid(),
		TargetDepartmentUUID: req.GetTargetDepartmentUuid(),
	})
	err = Error.HandleError(redirectErr, req.GetOperationId(), "redirect application")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "redirect application").Msg("success")
	return &emptypb.Empty{}, nil
}

// RecallApplication Отзыв заявки у инженера
func (s *ApplicationService) RecallApplication(ctx context.Context, req *pb.RecallApplicationRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "recall application").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "recall application").Err(fmt.Errorf("invalid application uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	// Валидация message
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "recall application").Err(fmt.Errorf("invalid message")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "recall application")
	if err != nil {
		return nil, err
	}

	// Проверяем статус заявки
	if !helpers.Contains([]string{"assigned", "in_progress", "on_hold", "on_revision"}, application.Status) {
		log.Info().Str("id", req.GetOperationId()).Str("method", "recall application").Err(fmt.Errorf("invalid status")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "invalid application status")
	}
	if application.ManagedBy != req.GetInitiatorUuid() {
		log.Info().Str("id", req.GetOperationId()).Str("method", "recall application").Err(fmt.Errorf("not a responsible manager")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only responsible manager can recall applications")
	}

	// Получаем менеджера, чтобы проверить что он еще в компании
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "recall application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "manager" {
		log.Warn().Str("id", req.GetOperationId()).Str("method", "recall application").Err(fmt.Errorf("user is a responsible manager but no longer a manager")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only responsible manager can recall applications")
	}

	// Пишем fix log
	addFixLogErr := s.db.ApplicationRepository.AddApplicationFixLog(ctx, entities.AddFixLogDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		Text:            message,
		CreatedBy:       req.GetInitiatorUuid(),
	})
	err = Error.HandleError(addFixLogErr, req.GetOperationId(), "recall application")
	if err != nil {
		return nil, err
	}

	// Отзываем заявку
	recallErr := s.db.ApplicationRepository.RecallApplication(ctx, entities.RecallApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
	})
	err = Error.HandleError(recallErr, req.GetOperationId(), "recall application")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "recall application").Msg("success")
	return &emptypb.Empty{}, nil
}

// TakeApplicationToVerification Взятие заявки на проверку
func (s *ApplicationService) TakeApplicationToVerification(ctx context.Context, req *pb.TakeApplicationToVerificationRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "take application").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "take application").Err(fmt.Errorf("invalid application uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}

	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "take application")
	if err != nil {
		return nil, err
	}

	// Проверяем статус заявки
	if application.Status != "pending_verification" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "take application").Err(fmt.Errorf("invalid status")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "invalid application status")
	}

	// Получаем инспектора
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "take application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "inspector" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "take application").Err(fmt.Errorf("invalid role")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only inspector can take application on verification")
	}
	if initiator.DepartmentUUID != application.DepartmentUUID {
		log.Info().Str("id", req.GetOperationId()).Str("method", "take application").Err(fmt.Errorf("unresponsible department")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "your department is not responsible")
	}

	// Назначаем заявку инспектору
	takeErr := s.db.ApplicationRepository.TakeApplicationToVerification(ctx, entities.TakeApplicationToVerificationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
	})
	err = Error.HandleError(takeErr, req.GetOperationId(), "take application")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "take application").Msg("success")
	return &emptypb.Empty{}, nil
}

// ReleaseApplicationVerification Отмена взятия заявки на проверку
func (s *ApplicationService) ReleaseApplicationVerification(ctx context.Context, req *pb.ReleaseApplicationVerificationRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "release application").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "release application").Err(fmt.Errorf("invalid application uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	// Валидация message
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "release application").Err(fmt.Errorf("invalid message")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "release application")
	if err != nil {
		return nil, err
	}

	// Проверяем статус заявки
	if application.Status != "on_verification" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "release application").Err(fmt.Errorf("invalid status")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "invalid application status")
	}

	// Проверяем, что пользователь держит заявку в данный момент
	if application.InspectedBy != req.GetInitiatorUuid() {
		log.Info().Str("id", req.GetOperationId()).Str("method", "release application").Err(fmt.Errorf("unresponsible inspector")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only responsible inspector can release application")
	}

	// Получаем инспектора, чтобы проверить, что он еще состоит в компании
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "release application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "inspector" {
		log.Warn().Str("id", req.GetOperationId()).Str("method", "release application").Err(fmt.Errorf("user is a responsible inspector but no longer an inspector")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only responsible inspector can release application")
	}

	// Пишем fix log
	addFixLogErr := s.db.ApplicationRepository.AddApplicationFixLog(ctx, entities.AddFixLogDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		Text:            message,
		CreatedBy:       req.GetInitiatorUuid(),
	})
	err = Error.HandleError(addFixLogErr, req.GetOperationId(), "release application")
	if err != nil {
		return nil, err
	}

	// Возвращаем заявку на проверку
	releaseErr := s.db.ApplicationRepository.ReleaseApplicationVerification(ctx, entities.ReleaseApplicationVerificationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
	})
	err = Error.HandleError(releaseErr, req.GetOperationId(), "release application")
	if err != nil {
		return nil, err
	}

	log.Info().Str("id", req.GetOperationId()).Str("method", "release application").Msg("success")
	return &emptypb.Empty{}, nil
}

// AddApplicationFixLog Добавление новой записи в fix log заявки
func (s *ApplicationService) AddApplicationFixLog(ctx context.Context, req *pb.AddApplicationFixLogRequest) (*emptypb.Empty, error) {
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "add fix log").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "add fix log").Err(fmt.Errorf("invalid application uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	// Валидация message
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "add fix log").Err(fmt.Errorf("invalid message")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	// Получаем заявку — добавлять записи может только ответственный инженер (executed_by)
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "add fix log")
	if err != nil {
		return nil, err
	}

	if application.ExecutedBy != req.GetInitiatorUuid() {
		log.Info().Str("id", req.GetOperationId()).Str("method", "add fix log").Err(fmt.Errorf("unresponsible engineer")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only the responsible engineer can add fix logs")
	}

	// Создаем новую запись в fix log
	addErr := s.db.ApplicationRepository.AddApplicationFixLog(ctx, entities.AddFixLogDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		Text:            message,
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
	// Валидации
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete application").Err(fmt.Errorf("invalid initiator uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete application").Err(fmt.Errorf("invalid application uuid")).Msg("error")
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	// Валидация message
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete application").Err(fmt.Errorf("invalid message")).Msg("error")
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	// Получаем заявку
	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	err := Error.HandleError(getErr, req.GetOperationId(), "delete application")
	if err != nil {
		return nil, err
	}

	// Удалить заявку может только ее создатель
	if application.CreatedBy != req.GetInitiatorUuid() {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete application").Err(fmt.Errorf("not a creator")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only a creator can delete application")
	}

	// Если статус заявки не created, то ее уже нельзя удалить
	if application.Status != "created" {
		log.Info().Str("id", req.GetOperationId()).Str("method", "delete application").Err(fmt.Errorf("application is already in use")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only applications with status 'created' can be deleted")
	}

	// Получаем инициатора, чтобы проверить, что он еще состоит в компании
	initiator, err := s.getEmployeeInfo(ctx, req.GetOperationId(), "delete application", application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "inspector" {
		log.Warn().Str("id", req.GetOperationId()).Str("method", "delete application").Err(fmt.Errorf("user is a creator, but is no longer an inspector")).Msg("error")
		return nil, status.Error(codes.PermissionDenied, "only a creator can delete application")
	}

	// Создаем новую запись в fix log
	addErr := s.db.ApplicationRepository.AddApplicationFixLog(ctx, entities.AddFixLogDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		Text:            message,
		CreatedBy:       req.GetInitiatorUuid(),
	})
	err = Error.HandleError(addErr, req.GetOperationId(), "delete application")
	if err != nil {
		return nil, err
	}

	// Мягкое удаление (deleted_at = now(), deleted_by = initiator)
	deleteErr := s.db.ApplicationRepository.DeleteApplication(ctx, entities.DeleteApplicationDTO{
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

// getEmployeeInfo - Получает роль сотрудника из company сервиса
func (s *ApplicationService) getEmployeeInfo(ctx context.Context, opID, method, companyUUID, initiatorUUID, targetUUID string) (*entities.Employee, error) {
	employeeInfo, err := s.companyClient.GetCompanyEmployee(ctx, &company_proto.GetCompanyEmployeeRequest{
		OperationId:   opID,
		CompanyUuid:   companyUUID,
		InitiatorUuid: initiatorUUID,
		TargetUuid:    targetUUID,
	})
	if err != nil {
		log.Error().Str("id", opID).Str("method", method).Err(err).Msg("failed to get employee from company service")
		return nil, err
	}

	return &entities.Employee{
		UUID:           targetUUID,
		Role:           employeeInfo.Role,
		DepartmentUUID: employeeInfo.DepartmentUuid,
	}, nil
}

// getDepartmentInfo - Получает данные о департаменте из company сервиса
func (s *ApplicationService) getDepartmentInfo(ctx context.Context, opID, method, initiatorUUID, departmentUUID string) (*entities.Department, error) {
	departmentInfo, err := s.companyClient.GetDepartment(ctx, &company_proto.GetDepartmentRequest{
		InitiatorUuid:  initiatorUUID,
		DepartmentUuid: departmentUUID,
	})
	if err != nil {
		log.Error().Str("id", opID).Str("method", method).Err(err).Msg("failed to get department from company service")
		return nil, err
	}

	return &entities.Department{
		DepartmentUUID: departmentUUID,
		CompanyUUID:    departmentInfo.GetCompanyUuid(),
	}, nil
}
