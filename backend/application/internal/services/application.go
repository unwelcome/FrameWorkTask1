package services

import (
	"context"
	"strings"

	"github.com/google/uuid"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	pb "github.com/unwelcome/FrameWorkTask1/backend/contracts/application/generated"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	sharedErrors "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
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
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if err := validate.ApplicationTitle(req.GetApplicationData().GetTitle()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application title")
	}
	if err := validate.ApplicationDescription(req.GetApplicationData().GetDescription()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application description")
	}

	initiator, err := s.getEmployeeInfo(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	if initiator.Role != "inspector" {
		return nil, status.Error(codes.PermissionDenied, "only inspectors can create applications")
	}

	applicationUUID := uuid.Must(uuid.NewV7()).String()

	if err := s.db.ApplicationRepository.CreateApplication(ctx, entities.CreateApplicationDTO{
		ApplicationUUID: applicationUUID,
		CompanyUUID:     req.GetCompanyUuid(),
		DepartmentUUID:  initiator.DepartmentUUID,
		Title:           req.GetApplicationData().GetTitle(),
		Description:     req.GetApplicationData().GetDescription(),
		CreatedBy:       req.GetInitiatorUuid(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &pb.CreateApplicationResponse{ApplicationUuid: applicationUUID}, nil
}

// GetApplication Получение полной информации о заявке
func (s *ApplicationService) GetApplication(ctx context.Context, req *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}

	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	initiator, err := s.getEmployeeInfo(ctx, application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	if !helpers.Contains([]string{"chief", "analytic"}, initiator.Role) &&
		!helpers.Contains([]string{application.CreatedBy, application.ManagedBy, application.ExecutedBy, application.InspectedBy}, req.GetInitiatorUuid()) {
		return nil, status.Error(codes.PermissionDenied, "you are not allowed to get application")
	}

	fixLogs, getLogsErr := s.db.ApplicationRepository.GetApplicationFixLogs(ctx, entities.GetApplicationFixLogsDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getLogsErr.GRPCError(); err != nil {
		return nil, err
	}

	pbFixLogs := make([]*pb.FixLog, 0, len(fixLogs))
	for _, fl := range fixLogs {
		pbFixLogs = append(pbFixLogs, &pb.FixLog{
			Uuid:      fl.UUID,
			Text:      fl.Text,
			CreatedAt: fl.CreatedAt,
			CreatedBy: fl.CreatedBy,
		})
	}

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
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetCompanyUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid company uuid")
	}
	if req.GetCount() <= 0 || req.GetCount() > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid count (1..100)")
	}
	if req.GetOffset() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset")
	}

	initiator, err := s.getEmployeeInfo(ctx, req.GetCompanyUuid(), req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	var applications []*entities.Application
	var dbErr sharedErrors.CodeError

	switch initiator.Role {

	// Если инициатор "chief" или "analytic" - используем department_uuid и status из запроса
	case "chief", "analytic":
		if err := validate.UUID(req.GetDepartmentUuid()); err != nil && req.GetDepartmentUuid() != "" {
			return nil, status.Errorf(codes.InvalidArgument, "invalid department uuid")
		}
		if !helpers.ContainsAll(AllApplicationStatuses, req.GetStatuses()) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid statuses")
		}

		applications, dbErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
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
			applications, dbErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
				CompanyUUID:    req.GetCompanyUuid(),
				DepartmentUUID: initiator.DepartmentUUID,
				Statuses:       []string{"pending_verification"},
				Offset:         int(req.GetOffset()),
				Count:          int(req.GetCount()),
			})
		} else {
			createdBy := req.GetInitiatorUuid()
			inspectedBy := req.GetInitiatorUuid()

			if len(req.GetStatuses()) == 0 {
				inspectedBy = ""
			} else {
				createdBy = ""
				if !helpers.ContainsAll([]string{"on_verification"}, req.GetStatuses()) {
					return nil, status.Errorf(codes.InvalidArgument, "invalid statuses")
				}
			}

			applications, dbErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
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
			applications, dbErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
				CompanyUUID:      req.GetCompanyUuid(),
				DepartmentUUID:   initiator.DepartmentUUID,
				Statuses:         []string{"created", "redirected", "recalled", "on_revision"},
				ExecutedByIsNull: true,
				Offset:           int(req.GetOffset()),
				Count:            int(req.GetCount()),
			})
		} else {
			applications, dbErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
				CompanyUUID:    req.GetCompanyUuid(),
				DepartmentUUID: initiator.DepartmentUUID,
				ManagedBy:      req.GetInitiatorUuid(),
				Offset:         int(req.GetOffset()),
				Count:          int(req.GetCount()),
			})
		}

	// Если инициатор "engineer" - department_uuid инициатора, statuses: ["assigned", "on_revision", "in_progress", "on_hold"]
	case "engineer":
		if !helpers.ContainsAll([]string{"assigned", "on_revision", "in_progress", "on_hold"}, req.GetStatuses()) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid statuses")
		}

		applications, dbErr = s.db.ApplicationRepository.GetApplications(ctx, entities.GetApplicationsDTO{
			CompanyUUID:    req.GetCompanyUuid(),
			DepartmentUUID: initiator.DepartmentUUID,
			Statuses:       req.GetStatuses(),
			ExecutedBy:     req.GetInitiatorUuid(),
			Offset:         int(req.GetOffset()),
			Count:          int(req.GetCount()),
		})

	default:
		return nil, status.Error(codes.PermissionDenied, "not allowed to get applications")
	}

	if err := dbErr.GRPCError(); err != nil {
		return nil, err
	}

	pbApplications := make([]*pb.Application, 0, len(applications))
	for _, app := range applications {
		pbApplications = append(pbApplications, &pb.Application{
			ApplicationUuid: app.ApplicationUUID,
			Title:           app.Title,
			Status:          app.Status,
			CreatedAt:       app.CreatedAt,
			UpdatedAt:       app.UpdatedAt,
		})
	}

	return &pb.GetApplicationsResponse{Applications: pbApplications}, nil
}

// UpdateApplicationStatus Обновление статуса заявки
func (s *ApplicationService) UpdateApplicationStatus(ctx context.Context, req *pb.UpdateApplicationStatusRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}

	newStatus := req.GetStatus()

	if !helpers.Contains([]string{"rejected", "in_progress", "on_hold", "pending_verification", "completed", "failed", "on_revision"}, newStatus) {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	initiator, err := s.getEmployeeInfo(ctx, application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}

	currentStatus := application.Status

	switch initiator.Role {

	case "inspector":
		if application.InspectedBy != req.GetInitiatorUuid() {
			return nil, status.Error(codes.PermissionDenied, "only the responsible inspector can change status")
		}
		if !helpers.Contains([]string{"completed", "failed", "on_revision"}, newStatus) {
			return nil, status.Error(codes.PermissionDenied, "inspectors can only set \"completed\", \"failed\" or \"on_revision\"")
		}
		if !helpers.Contains([]string{"on_verification"}, currentStatus) {
			return nil, status.Error(codes.FailedPrecondition, "invalid status transition")
		}

	case "manager":
		if initiator.DepartmentUUID != application.DepartmentUUID {
			return nil, status.Error(codes.PermissionDenied, "your department is not responsible")
		}
		if !helpers.Contains([]string{"rejected"}, newStatus) {
			return nil, status.Error(codes.PermissionDenied, "managers can only set \"rejected\"")
		}
		if !helpers.Contains([]string{"created", "redirected", "recalled", "on_revision"}, currentStatus) {
			return nil, status.Error(codes.FailedPrecondition, "invalid status transition")
		}

	case "engineer":
		if application.ExecutedBy != req.GetInitiatorUuid() {
			return nil, status.Error(codes.PermissionDenied, "only the responsible engineer can change status")
		}
		if !helpers.Contains([]string{"in_progress", "on_hold", "pending_verification"}, newStatus) {
			return nil, status.Error(codes.PermissionDenied, "engineers can only set \"in_progress\", \"on_hold\" or \"pending_verification\"")
		}
		if !helpers.Contains([]string{"assigned", "in_progress", "on_hold", "on_revision"}, currentStatus) {
			return nil, status.Error(codes.FailedPrecondition, "invalid status transition")
		}

	default:
		return nil, status.Error(codes.PermissionDenied, "unallowed role")
	}

	if err := s.db.ApplicationRepository.UpdateApplicationStatus(ctx, entities.UpdateApplicationStatusDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
		Status:          newStatus,
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// AssignApplication Назначение инженера на выполнение заявки
func (s *ApplicationService) AssignApplication(ctx context.Context, req *pb.AssignApplicationRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	if err := validate.UUID(req.GetTargetUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target uuid")
	}

	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if !helpers.Contains([]string{"created", "redirected", "recalled", "on_revision"}, application.Status) {
		return nil, status.Error(codes.InvalidArgument, "can't assign application with current status")
	}

	initiator, err := s.getEmployeeInfo(ctx, application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "manager" {
		return nil, status.Error(codes.PermissionDenied, "only managers can assign applications")
	}
	if application.DepartmentUUID != initiator.DepartmentUUID {
		return nil, status.Error(codes.PermissionDenied, "your department is not responsible")
	}

	target, err := s.getEmployeeInfo(ctx, application.CompanyUUID, req.GetInitiatorUuid(), req.GetTargetUuid())
	if err != nil {
		return nil, err
	}
	if target.Role != "engineer" {
		return nil, status.Error(codes.InvalidArgument, "application can only be assigned to an engineer")
	}
	if target.DepartmentUUID != initiator.DepartmentUUID {
		return nil, status.Error(codes.InvalidArgument, "engineer is not from your department")
	}

	if err := s.db.ApplicationRepository.AssignApplicationToEmployee(ctx, entities.AssignApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
		TargetUUID:      req.GetTargetUuid(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RedirectApplication Передача заявки в другой департамент
func (s *ApplicationService) RedirectApplication(ctx context.Context, req *pb.RedirectApplicationRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	if err := validate.UUID(req.GetTargetDepartmentUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid target depatment uuid")
	}
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if !helpers.Contains([]string{"created", "redirected", "recalled", "on_revision"}, application.Status) {
		return nil, status.Error(codes.PermissionDenied, "invalid application status")
	}

	initiator, err := s.getEmployeeInfo(ctx, application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "manager" {
		return nil, status.Error(codes.PermissionDenied, "only managers can redirect applications")
	}
	if initiator.DepartmentUUID != application.DepartmentUUID {
		return nil, status.Error(codes.PermissionDenied, "your department is not responsible")
	}

	department, err := s.getDepartmentInfo(ctx, req.GetInitiatorUuid(), req.GetTargetDepartmentUuid())
	if err != nil {
		return nil, err
	}
	if department.CompanyUUID != application.CompanyUUID {
		return nil, status.Error(codes.PermissionDenied, "department is from another company")
	}

	if err := s.db.ApplicationRepository.RedirectApplication(ctx, entities.RedirectApplicationDTO{
		ApplicationUUID:      req.GetApplicationUuid(),
		InitiatorUUID:        req.GetInitiatorUuid(),
		TargetDepartmentUUID: req.GetTargetDepartmentUuid(),
		FixLogText:           message,
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RecallApplication Отзыв заявки у инженера
func (s *ApplicationService) RecallApplication(ctx context.Context, req *pb.RecallApplicationRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if !helpers.Contains([]string{"assigned", "in_progress", "on_hold", "on_revision"}, application.Status) {
		return nil, status.Error(codes.PermissionDenied, "invalid application status")
	}
	if application.ManagedBy != req.GetInitiatorUuid() {
		return nil, status.Error(codes.PermissionDenied, "only responsible manager can recall applications")
	}

	initiator, err := s.getEmployeeInfo(ctx, application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "manager" {
		return nil, status.Error(codes.PermissionDenied, "only responsible manager can recall applications")
	}

	if err := s.db.ApplicationRepository.RecallApplication(ctx, entities.RecallApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
		FixLogText:      message,
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// TakeApplicationToVerification Взятие заявки на проверку
func (s *ApplicationService) TakeApplicationToVerification(ctx context.Context, req *pb.TakeApplicationToVerificationRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}

	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if application.Status != "pending_verification" {
		return nil, status.Error(codes.PermissionDenied, "invalid application status")
	}

	initiator, err := s.getEmployeeInfo(ctx, application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "inspector" {
		return nil, status.Error(codes.PermissionDenied, "only inspector can take application on verification")
	}
	if initiator.DepartmentUUID != application.DepartmentUUID {
		return nil, status.Error(codes.PermissionDenied, "your department is not responsible")
	}

	if err := s.db.ApplicationRepository.TakeApplicationToVerification(ctx, entities.TakeApplicationToVerificationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ReleaseApplicationVerification Отмена взятия заявки на проверку
func (s *ApplicationService) ReleaseApplicationVerification(ctx context.Context, req *pb.ReleaseApplicationVerificationRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if application.Status != "on_verification" {
		return nil, status.Error(codes.PermissionDenied, "invalid application status")
	}
	if application.InspectedBy != req.GetInitiatorUuid() {
		return nil, status.Error(codes.PermissionDenied, "only responsible inspector can release application")
	}

	initiator, err := s.getEmployeeInfo(ctx, application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "inspector" {
		return nil, status.Error(codes.PermissionDenied, "only responsible inspector can release application")
	}

	if err := s.db.ApplicationRepository.ReleaseApplicationVerification(ctx, entities.ReleaseApplicationVerificationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		InitiatorUUID:   req.GetInitiatorUuid(),
		FixLogText:      message,
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// AddApplicationFixLog Добавление новой записи в fix log заявки
func (s *ApplicationService) AddApplicationFixLog(ctx context.Context, req *pb.AddApplicationFixLogRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if application.ExecutedBy != req.GetInitiatorUuid() {
		return nil, status.Error(codes.PermissionDenied, "only the responsible engineer can add fix logs")
	}

	if err := s.db.ApplicationRepository.AddApplicationFixLog(ctx, entities.AddFixLogDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		Text:            message,
		CreatedBy:       req.GetInitiatorUuid(),
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// DeleteApplication Мягкое удаление заявки
func (s *ApplicationService) DeleteApplication(ctx context.Context, req *pb.DeleteApplicationRequest) (*emptypb.Empty, error) {
	if err := validate.UUID(req.GetInitiatorUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initiator uuid")
	}
	if err := validate.UUID(req.GetApplicationUuid()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid application uuid")
	}
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		return nil, status.Error(codes.InvalidArgument, "message is empty")
	}

	application, getErr := s.db.ApplicationRepository.GetApplication(ctx, entities.GetApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
	})
	if err := getErr.GRPCError(); err != nil {
		return nil, err
	}

	if application.CreatedBy != req.GetInitiatorUuid() {
		return nil, status.Error(codes.PermissionDenied, "only a creator can delete application")
	}
	if application.Status != "created" {
		return nil, status.Error(codes.PermissionDenied, "only applications with status 'created' can be deleted")
	}

	initiator, err := s.getEmployeeInfo(ctx, application.CompanyUUID, req.GetInitiatorUuid(), req.GetInitiatorUuid())
	if err != nil {
		return nil, err
	}
	if initiator.Role != "inspector" {
		return nil, status.Error(codes.PermissionDenied, "only a creator can delete application")
	}

	if err := s.db.ApplicationRepository.DeleteApplication(ctx, entities.DeleteApplicationDTO{
		ApplicationUUID: req.GetApplicationUuid(),
		DeletedBy:       req.GetInitiatorUuid(),
		FixLogText:      message,
	}).GRPCError(); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ─── Вспомогательные функции ──────────────────────────────────────────────────

// getEmployeeInfo Получает роль сотрудника из company сервиса
func (s *ApplicationService) getEmployeeInfo(ctx context.Context, companyUUID, initiatorUUID, targetUUID string) (*entities.Employee, error) {
	employeeInfo, err := s.companyClient.GetCompanyEmployee(ctx, &company_proto.GetCompanyEmployeeRequest{
		CompanyUuid:   companyUUID,
		InitiatorUuid: initiatorUUID,
		TargetUuid:    targetUUID,
	})
	if err != nil {
		return nil, err
	}

	return &entities.Employee{
		UUID:           targetUUID,
		Role:           employeeInfo.Role,
		DepartmentUUID: employeeInfo.DepartmentUuid,
	}, nil
}

// getDepartmentInfo Получает данные о департаменте из company сервиса
func (s *ApplicationService) getDepartmentInfo(ctx context.Context, initiatorUUID, departmentUUID string) (*entities.Department, error) {
	departmentInfo, err := s.companyClient.GetDepartment(ctx, &company_proto.GetDepartmentRequest{
		InitiatorUuid:  initiatorUUID,
		DepartmentUuid: departmentUUID,
	})
	if err != nil {
		return nil, err
	}

	return &entities.Department{
		DepartmentUUID: departmentUUID,
		CompanyUUID:    departmentInfo.GetCompanyUuid(),
	}, nil
}
