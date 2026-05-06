package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/application/api/generated"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/auth/api/generated"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/company/api/generated"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ─── Тестовые константы ───────────────────────────────────────────────────────

const (
	opKey     = "operation_id"
	uuidKey   = "user_uuid"
	opID      = "op-test-1"
	userID    = "11111111-1111-1111-1111-111111111111"
	companyID = "22222222-2222-2222-2222-222222222222"
	appID     = "33333333-3333-3333-3333-333333333333"
	targetID  = "44444444-4444-4444-4444-444444444444"

	// Синтаксически корректный JWT (формат валидируется только по паттерну)
	validJWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
)

// ─── Утилиты утверждений ──────────────────────────────────────────────────────

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		t.Errorf("expected HTTP %d, got %d", expected, resp.StatusCode)
	}
}

func decodeBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return v
}

// ─── Вспомогательные функции для запросов ────────────────────────────────────

// newApp создаёт Fiber-приложение с route и middleware для locals
func newApp(method, route string, h fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(opKey, opID)
		c.Locals(uuidKey, userID)
		return c.Next()
	})
	switch method {
	case http.MethodGet:
		app.Get(route, h)
	case http.MethodPost:
		app.Post(route, h)
	case http.MethodPatch:
		app.Patch(route, h)
	case http.MethodDelete:
		app.Delete(route, h)
	}
	return app
}

// jsonReq создаёт HTTP-запрос с JSON-телом
func jsonReq(method, url, body string) *http.Request {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// getReq создаёт GET-запрос без тела
func getReq(url string) *http.Request {
	return httptest.NewRequest(http.MethodGet, url, nil)
}

// deleteReq создаёт DELETE-запрос без тела
func deleteReq(url string) *http.Request {
	return httptest.NewRequest(http.MethodDelete, url, nil)
}

// grpcErr создаёт gRPC-ошибку с заданным кодом
func grpcErr(code codes.Code) error {
	return status.Error(code, code.String())
}

// ─── Mock: AuthServiceClient ──────────────────────────────────────────────────

type mockAuthClient struct {
	health            func(ctx context.Context, in *auth_proto.HealthRequest, opts ...grpc.CallOption) (*auth_proto.HealthResponse, error)
	register          func(ctx context.Context, in *auth_proto.RegisterRequest, opts ...grpc.CallOption) (*auth_proto.RegisterResponse, error)
	login             func(ctx context.Context, in *auth_proto.LoginRequest, opts ...grpc.CallOption) (*auth_proto.LoginResponse, error)
	getUser           func(ctx context.Context, in *auth_proto.GetUserRequest, opts ...grpc.CallOption) (*auth_proto.GetUserResponse, error)
	changePassword    func(ctx context.Context, in *auth_proto.ChangePasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	updateUserBio     func(ctx context.Context, in *auth_proto.UpdateUserBioRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	deleteUser        func(ctx context.Context, in *auth_proto.DeleteUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	refreshToken      func(ctx context.Context, in *auth_proto.RefreshTokenRequest, opts ...grpc.CallOption) (*auth_proto.RefreshTokenResponse, error)
	getAllActiveTokens func(ctx context.Context, in *auth_proto.GetAllActiveTokensRequest, opts ...grpc.CallOption) (*auth_proto.GetAllActiveTokensResponse, error)
	revokeToken       func(ctx context.Context, in *auth_proto.RevokeTokenRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	revokeAllTokens   func(ctx context.Context, in *auth_proto.RevokeAllTokensRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func (m *mockAuthClient) Health(ctx context.Context, in *auth_proto.HealthRequest, opts ...grpc.CallOption) (*auth_proto.HealthResponse, error) {
	return m.health(ctx, in, opts...)
}
func (m *mockAuthClient) Register(ctx context.Context, in *auth_proto.RegisterRequest, opts ...grpc.CallOption) (*auth_proto.RegisterResponse, error) {
	return m.register(ctx, in, opts...)
}
func (m *mockAuthClient) Login(ctx context.Context, in *auth_proto.LoginRequest, opts ...grpc.CallOption) (*auth_proto.LoginResponse, error) {
	return m.login(ctx, in, opts...)
}
func (m *mockAuthClient) GetUser(ctx context.Context, in *auth_proto.GetUserRequest, opts ...grpc.CallOption) (*auth_proto.GetUserResponse, error) {
	return m.getUser(ctx, in, opts...)
}
func (m *mockAuthClient) ChangePassword(ctx context.Context, in *auth_proto.ChangePasswordRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.changePassword(ctx, in, opts...)
}
func (m *mockAuthClient) UpdateUserBio(ctx context.Context, in *auth_proto.UpdateUserBioRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.updateUserBio(ctx, in, opts...)
}
func (m *mockAuthClient) DeleteUser(ctx context.Context, in *auth_proto.DeleteUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.deleteUser(ctx, in, opts...)
}
func (m *mockAuthClient) RefreshToken(ctx context.Context, in *auth_proto.RefreshTokenRequest, opts ...grpc.CallOption) (*auth_proto.RefreshTokenResponse, error) {
	return m.refreshToken(ctx, in, opts...)
}
func (m *mockAuthClient) GetAllActiveTokens(ctx context.Context, in *auth_proto.GetAllActiveTokensRequest, opts ...grpc.CallOption) (*auth_proto.GetAllActiveTokensResponse, error) {
	return m.getAllActiveTokens(ctx, in, opts...)
}
func (m *mockAuthClient) RevokeToken(ctx context.Context, in *auth_proto.RevokeTokenRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.revokeToken(ctx, in, opts...)
}
func (m *mockAuthClient) RevokeAllTokens(ctx context.Context, in *auth_proto.RevokeAllTokensRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.revokeAllTokens(ctx, in, opts...)
}

// ─── Mock: CompanyServiceClient ───────────────────────────────────────────────

type mockCompanyClient struct {
	health                  func(ctx context.Context, in *company_proto.HealthRequest, opts ...grpc.CallOption) (*company_proto.HealthResponse, error)
	createCompany           func(ctx context.Context, in *company_proto.CreateCompanyRequest, opts ...grpc.CallOption) (*company_proto.CreateCompanyResponse, error)
	getCompany              func(ctx context.Context, in *company_proto.GetCompanyRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyResponse, error)
	getCompanies            func(ctx context.Context, in *company_proto.GetCompaniesRequest, opts ...grpc.CallOption) (*company_proto.GetCompaniesResponse, error)
	getUserCompanies        func(ctx context.Context, in *company_proto.GetUserCompaniesRequest, opts ...grpc.CallOption) (*company_proto.GetUserCompaniesResponse, error)
	updateCompanyTitle      func(ctx context.Context, in *company_proto.UpdateCompanyTitleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	updateCompanyStatus     func(ctx context.Context, in *company_proto.UpdateCompanyStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	deleteCompany           func(ctx context.Context, in *company_proto.DeleteCompanyRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	createCompanyJoinCode   func(ctx context.Context, in *company_proto.CreateCompanyJoinCodeRequest, opts ...grpc.CallOption) (*company_proto.CreateCompanyJoinCodeResponse, error)
	getCompanyJoinCodes     func(ctx context.Context, in *company_proto.GetCompanyJoinCodesRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyJoinCodesResponse, error)
	deleteCompanyJoinCode   func(ctx context.Context, in *company_proto.DeleteCompanyJoinCodeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	joinCompany             func(ctx context.Context, in *company_proto.JoinCompanyRequest, opts ...grpc.CallOption) (*company_proto.JoinCompanyResponse, error)
	getCompanyEmployee      func(ctx context.Context, in *company_proto.GetCompanyEmployeeRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error)
	getCompanyEmployees     func(ctx context.Context, in *company_proto.GetCompanyEmployeesRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeesResponse, error)
	getCompanyEmployeesSummary func(ctx context.Context, in *company_proto.GetCompanyEmployeesSummaryRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeesSummaryResponse, error)
	updateEmployeeRole      func(ctx context.Context, in *company_proto.UpdateEmployeeRoleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	removeCompanyEmployee   func(ctx context.Context, in *company_proto.RemoveCompanyEmployeeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func (m *mockCompanyClient) Health(ctx context.Context, in *company_proto.HealthRequest, opts ...grpc.CallOption) (*company_proto.HealthResponse, error) {
	return m.health(ctx, in, opts...)
}
func (m *mockCompanyClient) CreateCompany(ctx context.Context, in *company_proto.CreateCompanyRequest, opts ...grpc.CallOption) (*company_proto.CreateCompanyResponse, error) {
	return m.createCompany(ctx, in, opts...)
}
func (m *mockCompanyClient) GetCompany(ctx context.Context, in *company_proto.GetCompanyRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyResponse, error) {
	return m.getCompany(ctx, in, opts...)
}
func (m *mockCompanyClient) GetCompanies(ctx context.Context, in *company_proto.GetCompaniesRequest, opts ...grpc.CallOption) (*company_proto.GetCompaniesResponse, error) {
	return m.getCompanies(ctx, in, opts...)
}
func (m *mockCompanyClient) GetUserCompanies(ctx context.Context, in *company_proto.GetUserCompaniesRequest, opts ...grpc.CallOption) (*company_proto.GetUserCompaniesResponse, error) {
	if m.getUserCompanies != nil {
		return m.getUserCompanies(ctx, in, opts...)
	}
	panic("unexpected call to GetUserCompanies")
}
func (m *mockCompanyClient) UpdateCompanyTitle(ctx context.Context, in *company_proto.UpdateCompanyTitleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.updateCompanyTitle(ctx, in, opts...)
}
func (m *mockCompanyClient) UpdateCompanyStatus(ctx context.Context, in *company_proto.UpdateCompanyStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.updateCompanyStatus(ctx, in, opts...)
}
func (m *mockCompanyClient) DeleteCompany(ctx context.Context, in *company_proto.DeleteCompanyRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.deleteCompany(ctx, in, opts...)
}
func (m *mockCompanyClient) CreateCompanyJoinCode(ctx context.Context, in *company_proto.CreateCompanyJoinCodeRequest, opts ...grpc.CallOption) (*company_proto.CreateCompanyJoinCodeResponse, error) {
	return m.createCompanyJoinCode(ctx, in, opts...)
}
func (m *mockCompanyClient) GetCompanyJoinCodes(ctx context.Context, in *company_proto.GetCompanyJoinCodesRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyJoinCodesResponse, error) {
	return m.getCompanyJoinCodes(ctx, in, opts...)
}
func (m *mockCompanyClient) DeleteCompanyJoinCode(ctx context.Context, in *company_proto.DeleteCompanyJoinCodeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.deleteCompanyJoinCode(ctx, in, opts...)
}
func (m *mockCompanyClient) JoinCompany(ctx context.Context, in *company_proto.JoinCompanyRequest, opts ...grpc.CallOption) (*company_proto.JoinCompanyResponse, error) {
	return m.joinCompany(ctx, in, opts...)
}
func (m *mockCompanyClient) GetCompanyEmployee(ctx context.Context, in *company_proto.GetCompanyEmployeeRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeeResponse, error) {
	return m.getCompanyEmployee(ctx, in, opts...)
}
func (m *mockCompanyClient) GetCompanyEmployees(ctx context.Context, in *company_proto.GetCompanyEmployeesRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeesResponse, error) {
	return m.getCompanyEmployees(ctx, in, opts...)
}
func (m *mockCompanyClient) GetCompanyEmployeesSummary(ctx context.Context, in *company_proto.GetCompanyEmployeesSummaryRequest, opts ...grpc.CallOption) (*company_proto.GetCompanyEmployeesSummaryResponse, error) {
	return m.getCompanyEmployeesSummary(ctx, in, opts...)
}
func (m *mockCompanyClient) UpdateEmployeeRole(ctx context.Context, in *company_proto.UpdateEmployeeRoleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.updateEmployeeRole(ctx, in, opts...)
}
func (m *mockCompanyClient) RemoveCompanyEmployee(ctx context.Context, in *company_proto.RemoveCompanyEmployeeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.removeCompanyEmployee(ctx, in, opts...)
}

// ─── Mock: ApplicationServiceClient ──────────────────────────────────────────

type mockApplicationClient struct {
	health                         func(ctx context.Context, in *application_proto.HealthRequest, opts ...grpc.CallOption) (*application_proto.HealthResponse, error)
	createApplication              func(ctx context.Context, in *application_proto.CreateApplicationRequest, opts ...grpc.CallOption) (*application_proto.CreateApplicationResponse, error)
	getApplication                 func(ctx context.Context, in *application_proto.GetApplicationRequest, opts ...grpc.CallOption) (*application_proto.GetApplicationResponse, error)
	getApplications                func(ctx context.Context, in *application_proto.GetApplicationsRequest, opts ...grpc.CallOption) (*application_proto.GetApplicationsResponse, error)
	getCompanyApplicationStatistic func(ctx context.Context, in *application_proto.GetCompanyApplicationStatisticRequest, opts ...grpc.CallOption) (*application_proto.GetCompanyApplicationStatisticResponse, error)
	getEmployeeApplicationStatistic func(ctx context.Context, in *application_proto.GetEmployeeApplicationStatisticRequest, opts ...grpc.CallOption) (*application_proto.GetEmployeeApplicationStatisticResponse, error)
	updateApplicationStatus        func(ctx context.Context, in *application_proto.UpdateApplicationStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	assignApplicationToEmployee    func(ctx context.Context, in *application_proto.AssignApplicationToEmployeeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	addApplicationFixLog           func(ctx context.Context, in *application_proto.AddApplicationFixLogRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	deleteApplication              func(ctx context.Context, in *application_proto.DeleteApplicationRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

func (m *mockApplicationClient) Health(ctx context.Context, in *application_proto.HealthRequest, opts ...grpc.CallOption) (*application_proto.HealthResponse, error) {
	return m.health(ctx, in, opts...)
}
func (m *mockApplicationClient) CreateApplication(ctx context.Context, in *application_proto.CreateApplicationRequest, opts ...grpc.CallOption) (*application_proto.CreateApplicationResponse, error) {
	return m.createApplication(ctx, in, opts...)
}
func (m *mockApplicationClient) GetApplication(ctx context.Context, in *application_proto.GetApplicationRequest, opts ...grpc.CallOption) (*application_proto.GetApplicationResponse, error) {
	return m.getApplication(ctx, in, opts...)
}
func (m *mockApplicationClient) GetApplications(ctx context.Context, in *application_proto.GetApplicationsRequest, opts ...grpc.CallOption) (*application_proto.GetApplicationsResponse, error) {
	return m.getApplications(ctx, in, opts...)
}
func (m *mockApplicationClient) GetCompanyApplicationStatistic(ctx context.Context, in *application_proto.GetCompanyApplicationStatisticRequest, opts ...grpc.CallOption) (*application_proto.GetCompanyApplicationStatisticResponse, error) {
	return m.getCompanyApplicationStatistic(ctx, in, opts...)
}
func (m *mockApplicationClient) GetEmployeeApplicationStatistic(ctx context.Context, in *application_proto.GetEmployeeApplicationStatisticRequest, opts ...grpc.CallOption) (*application_proto.GetEmployeeApplicationStatisticResponse, error) {
	return m.getEmployeeApplicationStatistic(ctx, in, opts...)
}
func (m *mockApplicationClient) UpdateApplicationStatus(ctx context.Context, in *application_proto.UpdateApplicationStatusRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.updateApplicationStatus(ctx, in, opts...)
}
func (m *mockApplicationClient) AssignApplicationToEmployee(ctx context.Context, in *application_proto.AssignApplicationToEmployeeRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.assignApplicationToEmployee(ctx, in, opts...)
}
func (m *mockApplicationClient) AddApplicationFixLog(ctx context.Context, in *application_proto.AddApplicationFixLogRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.addApplicationFixLog(ctx, in, opts...)
}
func (m *mockApplicationClient) DeleteApplication(ctx context.Context, in *application_proto.DeleteApplicationRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.deleteApplication(ctx, in, opts...)
}

// ─── Фабрики обработчиков ─────────────────────────────────────────────────────

func newAuthHandler(client auth_proto.AuthServiceClient) AuthHandler {
	return NewAuthHandler(client, opKey, uuidKey)
}

func newCompanyHandler(client company_proto.CompanyServiceClient) CompanyHandler {
	return NewCompanyHandler(client, opKey, uuidKey)
}

func newApplicationHandler(client application_proto.ApplicationServiceClient) ApplicationHandler {
	return NewApplicationHandler(client, opKey, uuidKey)
}

func newHealthHandler(auth auth_proto.AuthServiceClient, company company_proto.CompanyServiceClient, app application_proto.ApplicationServiceClient) HealthHandler {
	return NewHealthHandler(auth, company, app, opKey)
}
