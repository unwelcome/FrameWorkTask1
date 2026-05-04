package services

import (
	"context"
	"fmt"

	company_proto "github.com/unwelcome/FrameWorkTask1/backend/company/api/generated"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ─── Mock: ApplicationRepository ─────────────────────────────────────────────

type mockApplicationRepo struct {
	createApplication               func(ctx context.Context, dto entities.CreateApplicationDTO) Error.CodeError
	addApplicationFixLog            func(ctx context.Context, dto entities.CreateFixLogDTO) Error.CodeError
	getApplication                  func(ctx context.Context, dto entities.GetApplicationDTO) (*entities.Application, Error.CodeError)
	getApplicationFixLogs           func(ctx context.Context, dto entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError)
	getApplications                 func(ctx context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError)
	getCompanyApplicationStatistic  func(ctx context.Context, dto entities.GetCompanyApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError)
	getEmployeeApplicationStatistic func(ctx context.Context, dto entities.GetEmployeeApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError)
	updateApplicationData           func(ctx context.Context, dto entities.UpdateApplicationDataDTO) Error.CodeError
	updateApplicationStatus         func(ctx context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError
	assignApplicationToEmployee     func(ctx context.Context, dto entities.AssignApplicationToEmployeeDTO) Error.CodeError
	deleteApplicationRequest        func(ctx context.Context, dto entities.DeleteApplicationDTO) Error.CodeError
}

func (m *mockApplicationRepo) CreateApplication(ctx context.Context, dto entities.CreateApplicationDTO) Error.CodeError {
	return m.createApplication(ctx, dto)
}
func (m *mockApplicationRepo) AddApplicationFixLog(ctx context.Context, dto entities.CreateFixLogDTO) Error.CodeError {
	return m.addApplicationFixLog(ctx, dto)
}
func (m *mockApplicationRepo) GetApplication(ctx context.Context, dto entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
	return m.getApplication(ctx, dto)
}
func (m *mockApplicationRepo) GetApplicationFixLogs(ctx context.Context, dto entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError) {
	return m.getApplicationFixLogs(ctx, dto)
}
func (m *mockApplicationRepo) GetApplications(ctx context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError) {
	return m.getApplications(ctx, dto)
}
func (m *mockApplicationRepo) GetCompanyApplicationStatistic(ctx context.Context, dto entities.GetCompanyApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError) {
	return m.getCompanyApplicationStatistic(ctx, dto)
}
func (m *mockApplicationRepo) GetEmployeeApplicationStatistic(ctx context.Context, dto entities.GetEmployeeApplicationStatisticDTO) (*entities.ApplicationStatistic, Error.CodeError) {
	return m.getEmployeeApplicationStatistic(ctx, dto)
}
func (m *mockApplicationRepo) UpdateApplicationData(ctx context.Context, dto entities.UpdateApplicationDataDTO) Error.CodeError {
	return m.updateApplicationData(ctx, dto)
}
func (m *mockApplicationRepo) UpdateApplicationStatus(ctx context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError {
	return m.updateApplicationStatus(ctx, dto)
}
func (m *mockApplicationRepo) AssignApplicationToEmployee(ctx context.Context, dto entities.AssignApplicationToEmployeeDTO) Error.CodeError {
	return m.assignApplicationToEmployee(ctx, dto)
}
func (m *mockApplicationRepo) DeleteApplicationRequest(ctx context.Context, dto entities.DeleteApplicationDTO) Error.CodeError {
	return m.deleteApplicationRequest(ctx, dto)
}

// ─── Mock: CompanyServiceClient ───────────────────────────────────────────────

// mockCompanyClient реализует company_proto.CompanyServiceClient.
// Только GetCompanyEmployee настраивается через функциональное поле —
// остальные методы в application сервисе не вызываются.
type mockCompanyClient struct {
	getCompanyEmployee func(ctx context.Context, in *company_proto.GetCompanyEmployeeRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error)
}

func (m *mockCompanyClient) GetCompanyEmployee(ctx context.Context, in *company_proto.GetCompanyEmployeeRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
	return m.getCompanyEmployee(ctx, in, opts...)
}

func (m *mockCompanyClient) Health(_ context.Context, _ *company_proto.HealthRequest, _ ...grpc.CallOption) (*company_proto.HealthResponse, error) {
	panic("unexpected call to Health")
}
func (m *mockCompanyClient) CreateCompany(_ context.Context, _ *company_proto.CreateCompanyRequest, _ ...grpc.CallOption) (*company_proto.CreateCompanyResponse, error) {
	panic("unexpected call to CreateCompany")
}
func (m *mockCompanyClient) GetCompany(_ context.Context, _ *company_proto.GetCompanyRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyResponse, error) {
	panic("unexpected call to GetCompany")
}
func (m *mockCompanyClient) GetCompanies(_ context.Context, _ *company_proto.GetCompaniesRequest, _ ...grpc.CallOption) (*company_proto.GetCompaniesResponse, error) {
	panic("unexpected call to GetCompanies")
}
func (m *mockCompanyClient) GetUserCompanies(_ context.Context, _ *company_proto.GetUserCompaniesRequest, _ ...grpc.CallOption) (*company_proto.GetUserCompaniesResponse, error) {
	panic("unexpected call to GetUserCompanies")
}
func (m *mockCompanyClient) UpdateCompanyTitle(_ context.Context, _ *company_proto.UpdateCompanyTitleRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to UpdateCompanyTitle")
}
func (m *mockCompanyClient) UpdateCompanyStatus(_ context.Context, _ *company_proto.UpdateCompanyStatusRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to UpdateCompanyStatus")
}
func (m *mockCompanyClient) DeleteCompany(_ context.Context, _ *company_proto.DeleteCompanyRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to DeleteCompany")
}
func (m *mockCompanyClient) CreateCompanyJoinCode(_ context.Context, _ *company_proto.CreateCompanyJoinCodeRequest, _ ...grpc.CallOption) (*company_proto.CreateCompanyJoinCodeResponse, error) {
	panic("unexpected call to CreateCompanyJoinCode")
}
func (m *mockCompanyClient) GetCompanyJoinCodes(_ context.Context, _ *company_proto.GetCompanyJoinCodesRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyJoinCodesResponse, error) {
	panic("unexpected call to GetCompanyJoinCodes")
}
func (m *mockCompanyClient) DeleteCompanyJoinCode(_ context.Context, _ *company_proto.DeleteCompanyJoinCodeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to DeleteCompanyJoinCode")
}
func (m *mockCompanyClient) JoinCompany(_ context.Context, _ *company_proto.JoinCompanyRequest, _ ...grpc.CallOption) (*company_proto.JoinCompanyResponse, error) {
	panic("unexpected call to JoinCompany")
}
func (m *mockCompanyClient) GetCompanyEmployees(_ context.Context, _ *company_proto.GetCompanyEmployeesRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeesResponse, error) {
	panic("unexpected call to GetCompanyEmployees")
}
func (m *mockCompanyClient) GetCompanyEmployeesSummary(_ context.Context, _ *company_proto.GetCompanyEmployeesSummaryRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeesSummaryResponse, error) {
	panic("unexpected call to GetCompanyEmployeesSummary")
}
func (m *mockCompanyClient) UpdateEmployeeRole(_ context.Context, _ *company_proto.UpdateEmployeeRoleRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to UpdateEmployeeRole")
}
func (m *mockCompanyClient) RemoveCompanyEmployee(_ context.Context, _ *company_proto.RemoveCompanyEmployeeRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to RemoveCompanyEmployee")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// emptyRepo — пустая заглушка репозитория (поля не инициализированы — паника при вызове)
func emptyRepo() *mockApplicationRepo { return &mockApplicationRepo{} }

// newAppTestService создаёт ApplicationService с подменёнными зависимостями
func newAppTestService(repo postgresDB.ApplicationRepository, client company_proto.CompanyServiceClient) *ApplicationService {
	db := &postgresDB.DatabaseRepository{ApplicationRepository: repo}
	return NewApplicationService(db, client)
}

// ok — успешный CodeError
func ok() Error.CodeError { return Error.CodeError{Code: -1} }

// notFound — CodeError с кодом NotFound
func notFound() Error.CodeError {
	return Error.CodeError{Code: int(codes.NotFound), Err: fmt.Errorf("not found")}
}

// internalErr — CodeError с кодом Internal
func internalErr() Error.CodeError {
	return Error.CodeError{Code: 0, Err: fmt.Errorf("db error")}
}

// roleClient возвращает мок company-клиента, всегда отдающий заданную роль
func roleClient(role string) *mockCompanyClient {
	return &mockCompanyClient{
		getCompanyEmployee: func(_ context.Context, _ *company_proto.GetCompanyEmployeeRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
			return &company_proto.GetCompanyEmployeeResponse{Role: role}, nil
		},
	}
}

// errCompanyClient возвращает мок company-клиента, всегда возвращающий Internal ошибку
func errCompanyClient() *mockCompanyClient {
	return &mockCompanyClient{
		getCompanyEmployee: func(_ context.Context, _ *company_proto.GetCompanyEmployeeRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
			return nil, status.Error(codes.Internal, "company service unavailable")
		},
	}
}

// roleByTargetClient возвращает мок, отдающий разные роли в зависимости от TargetUuid в запросе.
// Если TargetUuid нет в карте — возвращается NotFound.
func roleByTargetClient(roles map[string]string) *mockCompanyClient {
	return &mockCompanyClient{
		getCompanyEmployee: func(_ context.Context, in *company_proto.GetCompanyEmployeeRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
			role, found := roles[in.GetTargetUuid()]
			if !found {
				return nil, status.Error(codes.NotFound, "employee not found in company")
			}
			return &company_proto.GetCompanyEmployeeResponse{Role: role}, nil
		},
	}
}

// ─── Тестовые сущности ────────────────────────────────────────────────────────

// testApp возвращает базовый объект заявки со статусом "created"
func testApp() *entities.Application {
	return &entities.Application{
		ApplicationUUID: appID,
		CompanyUUID:     companyID,
		Title:           "Test Application",
		Description:     "Test description",
		Status:          "created",
		CreatedAt:       "2024-01-01 00:00:00",
		CreatedBy:       initiatorID,
	}
}

// assignedApp возвращает заявку со статусом "assigned" и назначенными ответственными
func assignedApp() *entities.Application {
	app := testApp()
	app.Status = "assigned"
	app.ResponsibleManager = initiatorID
	app.ResponsibleEngineer = targetID
	return app
}

// awaitingApprovalApp возвращает заявку со статусом "awaiting_approval"
func awaitingApprovalApp() *entities.Application {
	app := testApp()
	app.Status = "awaiting_approval"
	app.ResponsibleManager = initiatorID
	app.ResponsibleEngineer = targetID
	return app
}

// repoWithApp настраивает мок репозитория так, чтобы GetApplication возвращал заданную заявку
func repoWithApp(app *entities.Application) *mockApplicationRepo {
	repo := emptyRepo()
	repo.getApplication = func(_ context.Context, _ entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
		return app, ok()
	}
	return repo
}

// testStatistic возвращает тестовую статистику по заявкам
func testStatistic() *entities.ApplicationStatistic {
	return &entities.ApplicationStatistic{
		Created:    2,
		Assigned:   1,
		InProgress: 3,
		Completed:  5,
	}
}
