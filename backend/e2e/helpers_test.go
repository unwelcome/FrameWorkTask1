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
