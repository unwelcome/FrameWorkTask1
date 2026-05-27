package services

import (
	"context"
	"fmt"

	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/entities"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	Error "github.com/unwelcome/FrameWorkTask1/backend/shared/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ─── Mock: ApplicationRepository ─────────────────────────────────────────────

type mockApplicationRepo struct {
	createApplication              func(ctx context.Context, dto entities.CreateApplicationDTO) Error.CodeError
	addApplicationFixLog           func(ctx context.Context, dto entities.AddFixLogDTO) Error.CodeError
	getApplication                 func(ctx context.Context, dto entities.GetApplicationDTO) (*entities.Application, Error.CodeError)
	getApplicationFixLogs          func(ctx context.Context, dto entities.GetApplicationFixLogsDTO) ([]*entities.FixLog, Error.CodeError)
	getApplications                func(ctx context.Context, dto entities.GetApplicationsDTO) ([]*entities.Application, Error.CodeError)
	updateApplicationStatus        func(ctx context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError
	assignApplicationToEmployee    func(ctx context.Context, dto entities.AssignApplicationDTO) Error.CodeError
	redirectApplication            func(ctx context.Context, dto entities.RedirectApplicationDTO) Error.CodeError
	recallApplication              func(ctx context.Context, dto entities.RecallApplicationDTO) Error.CodeError
	takeApplicationToVerification  func(ctx context.Context, dto entities.TakeApplicationToVerificationDTO) Error.CodeError
	releaseApplicationVerification func(ctx context.Context, dto entities.ReleaseApplicationVerificationDTO) Error.CodeError
	deleteApplication              func(ctx context.Context, dto entities.DeleteApplicationDTO) Error.CodeError
}

func (m *mockApplicationRepo) CreateApplication(ctx context.Context, dto entities.CreateApplicationDTO) Error.CodeError {
	return m.createApplication(ctx, dto)
}
func (m *mockApplicationRepo) AddApplicationFixLog(ctx context.Context, dto entities.AddFixLogDTO) Error.CodeError {
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
func (m *mockApplicationRepo) UpdateApplicationStatus(ctx context.Context, dto entities.UpdateApplicationStatusDTO) Error.CodeError {
	return m.updateApplicationStatus(ctx, dto)
}
func (m *mockApplicationRepo) AssignApplicationToEmployee(ctx context.Context, dto entities.AssignApplicationDTO) Error.CodeError {
	return m.assignApplicationToEmployee(ctx, dto)
}
func (m *mockApplicationRepo) RedirectApplication(ctx context.Context, dto entities.RedirectApplicationDTO) Error.CodeError {
	return m.redirectApplication(ctx, dto)
}
func (m *mockApplicationRepo) RecallApplication(ctx context.Context, dto entities.RecallApplicationDTO) Error.CodeError {
	return m.recallApplication(ctx, dto)
}
func (m *mockApplicationRepo) TakeApplicationToVerification(ctx context.Context, dto entities.TakeApplicationToVerificationDTO) Error.CodeError {
	return m.takeApplicationToVerification(ctx, dto)
}
func (m *mockApplicationRepo) ReleaseApplicationVerification(ctx context.Context, dto entities.ReleaseApplicationVerificationDTO) Error.CodeError {
	return m.releaseApplicationVerification(ctx, dto)
}
func (m *mockApplicationRepo) DeleteApplication(ctx context.Context, dto entities.DeleteApplicationDTO) Error.CodeError {
	return m.deleteApplication(ctx, dto)
}

// ─── Mock: CompanyServiceClient ───────────────────────────────────────────────

type mockCompanyClient struct {
	getCompanyEmployee func(ctx context.Context, in *company_proto.GetCompanyEmployeeRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error)
	getDepartment      func(ctx context.Context, in *company_proto.GetDepartmentRequest, opts ...grpc.CallOption) (*company_proto.GetDepartmentResponse, error)
}

func (m *mockCompanyClient) GetCompanyEmployee(ctx context.Context, in *company_proto.GetCompanyEmployeeRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
	return m.getCompanyEmployee(ctx, in, opts...)
}
func (m *mockCompanyClient) GetDepartment(ctx context.Context, in *company_proto.GetDepartmentRequest, opts ...grpc.CallOption) (*company_proto.GetDepartmentResponse, error) {
	if m.getDepartment != nil {
		return m.getDepartment(ctx, in, opts...)
	}
	panic("unexpected call to GetDepartment")
}
func (m *mockCompanyClient) Health(_ context.Context, _ *emptypb.Empty, _ ...grpc.CallOption) (*company_proto.HealthResponse, error) {
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
func (m *mockCompanyClient) CreateDepartment(_ context.Context, _ *company_proto.CreateDepartmentRequest, _ ...grpc.CallOption) (*company_proto.CreateDepartmentResponse, error) {
	panic("unexpected call to CreateDepartment")
}
func (m *mockCompanyClient) AddEmployeeToDepartment(_ context.Context, _ *company_proto.AddEmployeeToDepartmentRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to AddEmployeeToDepartment")
}
func (m *mockCompanyClient) GetCompanyDepartments(_ context.Context, _ *company_proto.GetCompanyDepartmentsRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyDepartmentsResponse, error) {
	panic("unexpected call to GetCompanyDepartments")
}
func (m *mockCompanyClient) UpdateDepartmentTitle(_ context.Context, _ *company_proto.UpdateDepartmentTitleRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to UpdateDepartmentTitle")
}
func (m *mockCompanyClient) DeleteDepartment(_ context.Context, _ *company_proto.DeleteDepartmentRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to DeleteDepartment")
}
func (m *mockCompanyClient) RemoveEmployeeFromDepartment(_ context.Context, _ *company_proto.RemoveEmployeeFromDepartmentRequest, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	panic("unexpected call to RemoveEmployeeFromDepartment")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// emptyRepo — пустая заглушка репозитория (паника при любом вызове)
func emptyRepo() *mockApplicationRepo { return &mockApplicationRepo{} }

// newAppTestService создаёт ApplicationService с подменёнными зависимостями
func newAppTestService(repo postgresDB.ApplicationRepository, client company_proto.CompanyServiceClient) *ApplicationService {
	db := &postgresDB.DatabaseRepository{ApplicationRepository: repo}
	return NewApplicationService(db, client)
}

// ok — успешный CodeError (Code == 0 означает «нет ошибки» в HandleError)
func ok() Error.CodeError { return Error.CodeError{} }

// notFound — CodeError с кодом NotFound
func notFound() Error.CodeError {
	return Error.Public(codes.NotFound, "not found")
}

// internalErr — CodeError с кодом Internal
func internalErr() Error.CodeError {
	return Error.Internal(fmt.Errorf("db error"))
}

// roleClient — мок company-клиента, всегда возвращающий заданную роль и deptID
func roleClient(role string) *mockCompanyClient {
	return &mockCompanyClient{
		getCompanyEmployee: func(_ context.Context, _ *company_proto.GetCompanyEmployeeRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
			return &company_proto.GetCompanyEmployeeResponse{Role: role, DepartmentUuid: deptID}, nil
		},
	}
}

// errCompanyClient — мок company-клиента, всегда возвращающий Internal-ошибку
func errCompanyClient() *mockCompanyClient {
	return &mockCompanyClient{
		getCompanyEmployee: func(_ context.Context, _ *company_proto.GetCompanyEmployeeRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
			return nil, status.Error(codes.Internal, "company service unavailable")
		},
	}
}

// roleByTargetClient — мок, возвращающий разные роли в зависимости от TargetUuid
func roleByTargetClient(roles map[string]string) *mockCompanyClient {
	return &mockCompanyClient{
		getCompanyEmployee: func(_ context.Context, in *company_proto.GetCompanyEmployeeRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
			role, found := roles[in.GetTargetUuid()]
			if !found {
				return nil, status.Error(codes.NotFound, "employee not found in company")
			}
			return &company_proto.GetCompanyEmployeeResponse{Role: role, DepartmentUuid: deptID}, nil
		},
	}
}

// redirectClient — мок, поддерживающий и GetCompanyEmployee, и GetDepartment
func redirectClient(role, deptCompanyUUID string) *mockCompanyClient {
	return &mockCompanyClient{
		getCompanyEmployee: func(_ context.Context, _ *company_proto.GetCompanyEmployeeRequest, _ ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
			return &company_proto.GetCompanyEmployeeResponse{Role: role, DepartmentUuid: deptID}, nil
		},
		getDepartment: func(_ context.Context, _ *company_proto.GetDepartmentRequest, _ ...grpc.CallOption) (*company_proto.GetDepartmentResponse, error) {
			return &company_proto.GetDepartmentResponse{CompanyUuid: deptCompanyUUID}, nil
		},
	}
}

// repoWithApp — мок репозитория, возвращающий заданную заявку через GetApplication
func repoWithApp(app *entities.Application) *mockApplicationRepo {
	repo := emptyRepo()
	repo.getApplication = func(_ context.Context, _ entities.GetApplicationDTO) (*entities.Application, Error.CodeError) {
		return app, ok()
	}
	return repo
}

// ─── Тестовые сущности ────────────────────────────────────────────────────────

func testApp() *entities.Application {
	return &entities.Application{
		ApplicationUUID: appID,
		CompanyUUID:     companyID,
		DepartmentUUID:  deptID,
		Title:           "Test Application",
		Description:     "Test description",
		Status:          "created",
		RevisionCount:   0,
		CreatedAt:       "2024-01-01 00:00:00",
		CreatedBy:       initiatorID,
	}
}

func assignedApp() *entities.Application {
	app := testApp()
	app.Status = "assigned"
	app.ManagedBy = initiatorID
	app.ExecutedBy = targetID
	return app
}

func inProgressApp() *entities.Application {
	app := testApp()
	app.Status = "in_progress"
	app.ManagedBy = initiatorID
	app.ExecutedBy = initiatorID
	return app
}

func pendingVerificationApp() *entities.Application {
	app := testApp()
	app.Status = "pending_verification"
	app.ManagedBy = initiatorID
	app.ExecutedBy = targetID
	return app
}

func onVerificationApp() *entities.Application {
	app := testApp()
	app.Status = "on_verification"
	app.ManagedBy = initiatorID
	app.ExecutedBy = targetID
	app.InspectedBy = initiatorID
	return app
}

func onRevisionApp() *entities.Application {
	app := testApp()
	app.Status = "on_revision"
	app.ManagedBy = initiatorID
	app.ExecutedBy = initiatorID
	app.RevisionCount = 3
	return app
}
