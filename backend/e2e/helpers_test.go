package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ─── HTTP client ──────────────────────────────────────────────────────────────

type apiClient struct {
	base        string
	accessToken string
	http        *http.Client
}

func newClient() *apiClient {
	return &apiClient{
		base: gatewayBaseURL,
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *apiClient) withToken(token string) *apiClient {
	return &apiClient{base: c.base, accessToken: token, http: c.http}
}

func (c *apiClient) do(method, path string, body any) (int, []byte) {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, _ := http.NewRequest(method, c.base+path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody
}

func (c *apiClient) post(path string, body any) (int, []byte) {
	return c.do(http.MethodPost, path, body)
}

func (c *apiClient) get(path string) (int, []byte) {
	return c.do(http.MethodGet, path, nil)
}

func (c *apiClient) patch(path string, body any) (int, []byte) {
	return c.do(http.MethodPatch, path, body)
}

func (c *apiClient) delete(path string, body any) (int, []byte) {
	return c.do(http.MethodDelete, path, body)
}

// ─── Response types ───────────────────────────────────────────────────────────

type registerResp struct {
	UserUUID string `json:"user_uuid"`
}

type loginResp struct {
	UserUUID     string `json:"user_uuid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type getUserResp struct {
	UserUUID   string `json:"user_uuid"`
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Patronymic string `json:"patronymic"`
	CreatedAt  string `json:"created_at"`
}

type refreshTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type tokensResp struct {
	Tokens []struct {
		Token string `json:"token"`
	} `json:"tokens"`
}

type httpErrResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ─── Data helpers ─────────────────────────────────────────────────────────────

func randomEmail() string {
	return fmt.Sprintf("user_%d@test.com", rand.Int63())
}

func defaultUserPayload(email, password string) map[string]string {
	return map[string]string{
		"email":      email,
		"password":   password,
		"first_name": "Ivan",
		"last_name":  "Ivanov",
		"patronymic": "Ivanovich",
	}
}

// ─── Flow helpers ─────────────────────────────────────────────────────────────

// mustRegister registers a user and fails the test immediately if registration fails.
func mustRegister(t *testing.T, c *apiClient, email, password string) registerResp {
	t.Helper()
	code, body := c.post("/api/register", defaultUserPayload(email, password))
	require.Equalf(t, http.StatusCreated, code, "register failed (body: %s)", body)

	var resp registerResp
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.UserUUID, "register returned empty user_uuid")
	return resp
}

// mustLogin logs in a user and fails the test immediately if login fails.
func mustLogin(t *testing.T, c *apiClient, email, password string) loginResp {
	t.Helper()
	code, body := c.post("/api/login", map[string]string{
		"email":    email,
		"password": password,
	})
	require.Equalf(t, http.StatusOK, code, "login failed (body: %s)", body)

	var resp loginResp
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.AccessToken, "login returned empty access_token")
	require.NotEmpty(t, resp.RefreshToken, "login returned empty refresh_token")
	return resp
}

// mustRegisterAndLogin registers a user and immediately logs in, returning the login response.
func mustRegisterAndLogin(t *testing.T, c *apiClient) (string, loginResp) {
	t.Helper()
	email := randomEmail()
	mustRegister(t, c, email, "Password123")
	login := mustLogin(t, c, email, "Password123")
	return email, login
}

// ─── Department response types ────────────────────────────────────────────────

type createDepartmentResp struct {
	DepartmentUUID string `json:"department_uuid"`
}

type departmentResp struct {
	DepartmentUUID string `json:"department_uuid"`
	CompanyUUID    string `json:"company_uuid"`
	Title          string `json:"title"`
	CreatedAt      string `json:"created_at"`
	CreatedBy      string `json:"created_by"`
}

type departmentListItem struct {
	DepartmentUUID string `json:"department_uuid"`
	Title          string `json:"title"`
}

type departmentsListResp struct {
	Departments []departmentListItem `json:"departments"`
}

type employeeInfoResp struct {
	UserUUID       string `json:"user_uuid"`
	Role           string `json:"role"`
	DepartmentUUID string `json:"department_uuid"`
	JoinedAt       string `json:"joined_at"`
}

type employeesListResp struct {
	Employees []employeeInfoResp `json:"employees"`
}

type employeesSummaryResp struct {
	ChiefCount      int64 `json:"chief_count"`
	AnalyticsCount  int64 `json:"analytics_count"`
	ManagerCount    int64 `json:"manager_count"`
	EngineerCount   int64 `json:"engineer_count"`
	InspectorCount  int64 `json:"inspector_count"`
	UnemployedCount int64 `json:"unemployed_count"`
}

// ─── Company response types ───────────────────────────────────────────────────

type createCompanyResp struct {
	CompanyUUID string `json:"company_uuid"`
}

type companyResp struct {
	CompanyUUID string `json:"company_uuid"`
	Title       string `json:"title"`
	Status      string `json:"status"`
}

type companiesResp struct {
	Companies []companyResp `json:"companies"`
}

type createCodeResp struct {
	Code string `json:"code"`
}

type codesResp struct {
	Codes []string `json:"codes"`
}

type joinCompanyResp struct {
	CompanyUUID string `json:"company_uuid"`
	Role        string `json:"role"`
}

// ─── Company data helpers ─────────────────────────────────────────────────────

func randomTitle() string {
	return fmt.Sprintf("TestCompany_%d", rand.Int63())
}

// ─── Company flow helpers ─────────────────────────────────────────────────────

// mustCreateCompany creates a company with the given title and returns its UUID.
func mustCreateCompany(t *testing.T, c *apiClient, title string) string {
	t.Helper()
	status, body := c.post("/api/auth/company/create", map[string]string{"title": title})
	require.Equalf(t, http.StatusCreated, status, "create company failed (body: %s)", body)
	var resp createCompanyResp
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.CompanyUUID, "create company returned empty company_uuid")
	return resp.CompanyUUID
}

// mustCreateCode creates a join code for the given company and returns the 6-digit code string.
func mustCreateCode(t *testing.T, c *apiClient, companyUUID string, ttl int) string {
	t.Helper()
	status, body := c.post("/api/auth/company/"+companyUUID+"/code", map[string]int{"code_ttl": ttl})
	require.Equalf(t, http.StatusCreated, status, "create join code failed (body: %s)", body)
	var resp createCodeResp
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.Code, "create join code returned empty code")
	return resp.Code
}

// mustJoinCompany joins a company using the given 6-digit code.
func mustJoinCompany(t *testing.T, c *apiClient, code string) joinCompanyResp {
	t.Helper()
	status, body := c.post("/api/auth/company/join", map[string]string{"code": code})
	require.Equalf(t, http.StatusOK, status, "join company failed (body: %s)", body)
	var resp joinCompanyResp
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.CompanyUUID)
	return resp
}

// mustCreateDepartment создаёт департамент и возвращает его UUID.
func mustCreateDepartment(t *testing.T, chief *apiClient, companyUUID, title string) string {
	t.Helper()
	status, body := chief.post("/api/auth/company/"+companyUUID+"/department", map[string]string{"title": title})
	require.Equalf(t, http.StatusCreated, status, "create department failed (body: %s)", body)
	var resp createDepartmentResp
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.DepartmentUUID, "create department returned empty department_uuid")
	return resp.DepartmentUUID
}

// mustSetEmployeeRole устанавливает роль сотрудника в компании.
func mustSetEmployeeRole(t *testing.T, chief *apiClient, companyUUID, targetUUID, role string) {
	t.Helper()
	status, body := chief.patch(
		fmt.Sprintf("/api/auth/company/%s/employee/%s/role", companyUUID, targetUUID),
		map[string]string{"role": role},
	)
	require.Equalf(t, http.StatusOK, status, "set employee role failed (body: %s)", body)
}

// mustAddEmployeeToDepartment добавляет сотрудника в департамент.
func mustAddEmployeeToDepartment(t *testing.T, chief *apiClient, companyUUID, deptUUID, targetUUID string) {
	t.Helper()
	status, body := chief.post(
		fmt.Sprintf("/api/auth/company/%s/department/%s/employee/%s", companyUUID, deptUUID, targetUUID),
		nil,
	)
	require.Equalf(t, http.StatusOK, status, "add employee to department failed (body: %s)", body)
}

// mustOpenCompany открывает компанию (устанавливает статус "open") от имени chief.
func mustOpenCompany(t *testing.T, chief *apiClient, companyUUID string) {
	t.Helper()
	status, body := chief.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{"status": "open"})
	require.Equalf(t, http.StatusOK, status, "open company failed (body: %s)", body)
}

// mustAddMember opens the company, creates a join code as chief and has the guest client join the company.
func mustAddMember(t *testing.T, chief *apiClient, guest *apiClient, companyUUID string) {
	t.Helper()
	mustOpenCompany(t, chief, companyUUID)
	code := mustCreateCode(t, chief, companyUUID, 3600)
	mustJoinCompany(t, guest, code)
}

// ─── Application response types ───────────────────────────────────────────────

type createApplicationResp struct {
	ApplicationUUID string `json:"application_uuid"`
}

type applicationDetailResp struct {
	Application applicationDetail `json:"application"`
}

type applicationDetail struct {
	ApplicationUUID string           `json:"application_uuid"`
	Status          string           `json:"status"`
	RevisionCount   int64            `json:"revision_count"`
	ManagedBy       string           `json:"managed_by"`
	ExecutedBy      string           `json:"executed_by"`
	InspectedBy     string           `json:"inspected_by"`
	FixLogs         []fixLogRespItem `json:"fix_logs"`
}

type fixLogRespItem struct {
	UUID      string `json:"uuid"`
	Text      string `json:"text"`
	CreatedBy string `json:"created_by"`
}

type applicationListResp struct {
	Applications []applicationListItem `json:"applications"`
}

type applicationListItem struct {
	ApplicationUUID string `json:"application_uuid"`
	Status          string `json:"status"`
}

// ─── Application environment ──────────────────────────────────────────────────

// appEnv holds all clients and identifiers required for application e2e tests.
type appEnv struct {
	CompanyUUID string

	// Department 1 — primary department used for main test scenarios.
	DeptUUID string

	Chief        *apiClient
	Analytic     *apiClient
	AnalyticUUID string

	Inspector      *apiClient
	InspectorUUID  string
	Inspector2     *apiClient // second inspector — used for InspectedBy tests
	Inspector2UUID string

	Manager      *apiClient
	ManagerUUID  string
	Engineer     *apiClient
	EngineerUUID string

	// Department 2 — separate department for cross-department isolation tests.
	Dept2UUID      string
	Inspector3     *apiClient
	Inspector3UUID string
	Manager2       *apiClient
	Manager2UUID   string
	Engineer2      *apiClient
	Engineer2UUID  string
}

// mustSetupAppEnv creates a full company environment:
//
//	chief → company → dept1 + dept2
//	dept1: inspector, inspector2, manager, engineer, analytic
//	dept2: inspector3, manager2, engineer2
//
// A single join code is created once and reused for all members.
func mustSetupAppEnv(t *testing.T, c *apiClient) appEnv {
	t.Helper()

	_, chiefLogin := mustRegisterAndLogin(t, c)
	chief := c.withToken(chiefLogin.AccessToken)
	companyUUID := mustCreateCompany(t, chief, randomTitle())
	mustOpenCompany(t, chief, companyUUID)

	deptUUID := mustCreateDepartment(t, chief, companyUUID, "Main Department")
	dept2UUID := mustCreateDepartment(t, chief, companyUUID, "Other Department")

	// One shared code is enough — codes are multi-use until their TTL expires.
	sharedCode := mustCreateCode(t, chief, companyUUID, 3600)

	// joinMember registers a new user, joins via the shared code, sets the role,
	// and adds the user to the given department.
	joinMember := func(deptID, role string) (*apiClient, string) {
		_, login := mustRegisterAndLogin(t, c)
		client := c.withToken(login.AccessToken)
		mustJoinCompany(t, client, sharedCode)
		mustSetEmployeeRole(t, chief, companyUUID, login.UserUUID, role)
		if deptID != "" {
			mustAddEmployeeToDepartment(t, chief, companyUUID, deptID, login.UserUUID)
		}
		return client, login.UserUUID
	}

	analytic, analyticUUID := joinMember("", "analytic")

	// Dept 1 members.
	inspector, inspectorUUID := joinMember(deptUUID, "inspector")
	inspector2, inspector2UUID := joinMember(deptUUID, "inspector")
	manager, managerUUID := joinMember(deptUUID, "manager")
	engineer, engineerUUID := joinMember(deptUUID, "engineer")

	// Dept 2 members — used for cross-department isolation tests.
	inspector3, inspector3UUID := joinMember(dept2UUID, "inspector")
	manager2, manager2UUID := joinMember(dept2UUID, "manager")
	engineer2, engineer2UUID := joinMember(dept2UUID, "engineer")

	return appEnv{
		CompanyUUID:    companyUUID,
		DeptUUID:       deptUUID,
		Chief:          chief,
		Analytic:       analytic,
		AnalyticUUID:   analyticUUID,
		Inspector:      inspector,
		InspectorUUID:  inspectorUUID,
		Inspector2:     inspector2,
		Inspector2UUID: inspector2UUID,
		Manager:        manager,
		ManagerUUID:    managerUUID,
		Engineer:       engineer,
		EngineerUUID:   engineerUUID,
		Dept2UUID:      dept2UUID,
		Inspector3:     inspector3,
		Inspector3UUID: inspector3UUID,
		Manager2:       manager2,
		Manager2UUID:   manager2UUID,
		Engineer2:      engineer2,
		Engineer2UUID:  engineer2UUID,
	}
}

// ─── Application flow helpers ─────────────────────────────────────────────────

// mustCreateApplication creates an application on behalf of an inspector and returns its UUID.
func mustCreateApplication(t *testing.T, inspector *apiClient, companyUUID, title, description string) string {
	t.Helper()
	code, body := inspector.post("/api/auth/application/create", map[string]string{
		"company_uuid": companyUUID,
		"title":        title,
		"description":  description,
	})
	require.Equalf(t, http.StatusCreated, code, "create application failed (body: %s)", body)
	var resp createApplicationResp
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.ApplicationUUID, "create application returned empty uuid")
	return resp.ApplicationUUID
}

// mustAssignApplication assigns an application to an engineer on behalf of a manager.
func mustAssignApplication(t *testing.T, manager *apiClient, appUUID, engineerUUID string) {
	t.Helper()
	code, body := manager.patch("/api/auth/application/"+appUUID+"/assign", map[string]string{
		"target_uuid": engineerUUID,
	})
	require.Equalf(t, http.StatusOK, code, "assign application failed (body: %s)", body)
}

// mustSetAppStatus updates the application status via PATCH /status.
func mustSetAppStatus(t *testing.T, client *apiClient, appUUID, newStatus string) {
	t.Helper()
	code, body := client.patch("/api/auth/application/"+appUUID+"/status", map[string]string{
		"status": newStatus,
	})
	require.Equalf(t, http.StatusOK, code, "set app status %q failed (body: %s)", newStatus, body)
}

// mustTakeToVerification moves an application from pending_verification to on_verification.
func mustTakeToVerification(t *testing.T, inspector *apiClient, appUUID string) {
	t.Helper()
	code, body := inspector.patch("/api/auth/application/"+appUUID+"/take-verification", nil)
	require.Equalf(t, http.StatusOK, code, "take to verification failed (body: %s)", body)
}
