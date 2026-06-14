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


type loginResp struct {
	UserUUID     string `json:"user_uuid"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionUUID  string `json:"session_uuid"`
}

type getUserResp struct {
	UserUUID    string `json:"user_uuid"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	DeletedAt   string `json:"deleted_at"`
}

type refreshTokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type tokenInfoResp struct {
	SessionUUID    string `json:"session_uuid"`
	IP             string `json:"ip"`
	LastIP         string `json:"last_ip"`
	ISP            string `json:"isp"`
	CountryCode    string `json:"country_code"`
	CountryName    string `json:"country_name"`
	City           string `json:"city"`
	Timezone       string `json:"timezone"`
	DeviceType     string `json:"device_type"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
	CreatedAt      string `json:"created_at"`
	LastActiveAt   string `json:"last_active_at"`
}

type tokensResp struct {
	Tokens []tokenInfoResp `json:"tokens"`
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
func mustRegister(t *testing.T, c *apiClient, email, password string) {
	t.Helper()
	code, body := c.post("/api/register", defaultUserPayload(email, password))
	require.Equalf(t, http.StatusCreated, code, "register failed (body: %s)", body)
}

// mustGetVerificationToken fetches a verification token for a user via debug endpoint (by email).
// Only works when APP_ENV=test.
func mustGetVerificationToken(t *testing.T, c *apiClient, email string) string {
	t.Helper()
	code, body := c.get(fmt.Sprintf("/api/debug/user/email/%s/verification-token", email))
	require.Equalf(t, http.StatusOK, code, "get verification token by email failed (body: %s)", body)

	var resp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.Token, "verification token is empty")
	return resp.Token
}

// mustVerifyAccount verifies a user account using a JWT verification token (magic link).
func mustVerifyAccount(t *testing.T, c *apiClient, verificationToken string) {
	t.Helper()
	code, body := c.post("/api/user/verify", map[string]string{
		"verification_token": verificationToken,
	})
	require.Equalf(t, http.StatusOK, code, "verify account failed (body: %s)", body)
}

// mustRegisterVerifyAndLogin registers a user, verifies the account via the debug endpoint,
// and logs in — returning the email and login response.
func mustRegisterVerifyAndLogin(t *testing.T, c *apiClient) (string, loginResp) {
	t.Helper()
	email := randomEmail()
	mustRegister(t, c, email, "Password123")
	verificationToken := mustGetVerificationToken(t, c, email)
	mustVerifyAccount(t, c, verificationToken)
	login := mustLogin(t, c, email, "Password123")
	return email, login
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

// mustRegisterAndLogin registers, verifies, and logs in a user.
// Alias for mustRegisterVerifyAndLogin — kept for backward compatibility.
func mustRegisterAndLogin(t *testing.T, c *apiClient) (string, loginResp) {
	t.Helper()
	return mustRegisterVerifyAndLogin(t, c)
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
	DepartmentUUID  string           `json:"department_uuid"`
	Status          string           `json:"status"`
	RevisionCount   int64            `json:"revision_count"`
	ManagedBy       string           `json:"managed_by"`
	ExecutedBy      string           `json:"executed_by"`
	InspectedBy     string           `json:"inspected_by"`
	ClosedAt        string           `json:"closed_at"`
	DeletedAt       string           `json:"deleted_at"`
	DeletedBy       string           `json:"deleted_by"`
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

// mustRecallApplication recalls an application (manager takes it back from the engineer).
func mustRecallApplication(t *testing.T, manager *apiClient, appUUID string) {
	t.Helper()
	code, body := manager.patch("/api/auth/application/"+appUUID+"/recall", map[string]string{
		"message": "Recalled for reassignment.",
	})
	require.Equalf(t, http.StatusOK, code, "recall application failed (body: %s)", body)
}

// mustAddFixLog adds a fix log entry to an application on behalf of the responsible engineer.
func mustAddFixLog(t *testing.T, engineer *apiClient, appUUID, message string) {
	t.Helper()
	code, body := engineer.post("/api/auth/application/"+appUUID+"/fix-log", map[string]string{
		"message": message,
	})
	require.Equalf(t, http.StatusCreated, code, "add fix log failed (body: %s)", body)
}

// mustReleaseApplicationVerification releases an application from on_verification back to pending_verification.
func mustReleaseApplicationVerification(t *testing.T, inspector *apiClient, appUUID, message string) {
	t.Helper()
	code, body := inspector.patch("/api/auth/application/"+appUUID+"/release-verification", map[string]string{
		"message": message,
	})
	require.Equalf(t, http.StatusOK, code, "release application verification failed (body: %s)", body)
}

// mustRedirectApplication redirects an application to another department.
func mustRedirectApplication(t *testing.T, manager *apiClient, appUUID, targetDeptUUID string) {
	t.Helper()
	code, body := manager.patch("/api/auth/application/"+appUUID+"/redirect", map[string]string{
		"target_department_uuid": targetDeptUUID,
		"message":                "Redirected to another department.",
	})
	require.Equalf(t, http.StatusOK, code, "redirect application failed (body: %s)", body)
}

// mustAdvanceToOnVerification creates an application and advances it to on_verification status.
// Returns the application UUID.
// Flow: created → assigned → in_progress → pending_verification → on_verification (held by inspector2).
func mustAdvanceToOnVerification(t *testing.T, env appEnv, title string) string {
	t.Helper()
	appUUID := mustCreateApplication(t, env.Inspector, env.CompanyUUID, title, "Advance to on_verification.")
	mustAssignApplication(t, env.Manager, appUUID, env.EngineerUUID)
	mustSetAppStatus(t, env.Engineer, appUUID, "in_progress")
	mustSetAppStatus(t, env.Engineer, appUUID, "pending_verification")
	mustTakeToVerification(t, env.Inspector2, appUUID)
	return appUUID
}

// mustAdvanceToOnRevision creates an application and advances it to on_revision status.
// Returns the application UUID.
// Flow: …on_verification → on_revision (set by inspector2).
func mustAdvanceToOnRevision(t *testing.T, env appEnv, title string) string {
	t.Helper()
	appUUID := mustAdvanceToOnVerification(t, env, title)
	mustSetAppStatus(t, env.Inspector2, appUUID, "on_revision")
	return appUUID
}

type applicationHistoryItem struct {
	ApplicationUUID string `json:"application_uuid"`
	Version         int64  `json:"version"`
	Status          string `json:"status"`
	ManagedBy       string `json:"managed_by"`
	ExecutedBy      string `json:"executed_by"`
	InspectedBy     string `json:"inspected_by"`
}

type applicationHistoryResp struct {
	History []applicationHistoryItem `json:"history"`
}

// getApplicationHistory fetches application history and returns the raw status code plus parsed response.
// On non-200 the response struct is empty.
func getApplicationHistory(t *testing.T, client *apiClient, appUUID string, count, offset int) (int, applicationHistoryResp) {
	t.Helper()
	code, body := client.get(fmt.Sprintf("/api/auth/application/%s/history?count=%d&offset=%d", appUUID, count, offset))
	if code != http.StatusOK {
		return code, applicationHistoryResp{}
	}
	var resp applicationHistoryResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return code, resp
}

// mustGetApplicationDetail fetches and unmarshals the full application detail.
func mustGetApplicationDetail(t *testing.T, client *apiClient, appUUID string) applicationDetail {
	t.Helper()
	code, body := client.get("/api/auth/application/" + appUUID)
	require.Equalf(t, http.StatusOK, code, "get application failed (body: %s)", body)
	var resp applicationDetailResp
	require.NoError(t, json.Unmarshal(body, &resp))
	return resp.Application
}

// ─── Recovery / password-reset helpers ───────────────────────────────────────

// mustGetResetPasswordToken fetches a reset-password JWT token for the given email via debug endpoint.
// Only works when APP_ENV=test.
func mustGetResetPasswordToken(t *testing.T, c *apiClient, email string) string {
	t.Helper()
	code, body := c.get(fmt.Sprintf("/api/debug/user/email/%s/reset-password-token", email))
	require.Equalf(t, http.StatusOK, code, "get reset password token failed (body: %s)", body)
	var resp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.Token, "reset password token is empty")
	return resp.Token
}

// mustForgotPassword triggers the password-recovery flow for the given email.
func mustForgotPassword(t *testing.T, c *apiClient, email string) {
	t.Helper()
	code, body := c.post("/api/forgot-password", map[string]string{"email": email})
	require.Equalf(t, http.StatusOK, code, "forgot password failed (body: %s)", body)
}

// mustChangePassword changes the authenticated user's password, verifying the old one first.
func mustChangePassword(t *testing.T, auth *apiClient, oldPassword, newPassword string) {
	t.Helper()
	code, body := auth.patch("/api/auth/user/password", map[string]string{
		"old_password": oldPassword,
		"password":     newPassword,
	})
	require.Equalf(t, http.StatusOK, code, "change password failed (body: %s)", body)
}

// mustDeleteAccount soft-deletes the authenticated user's own account.
func mustDeleteAccount(t *testing.T, auth *apiClient) {
	t.Helper()
	code, body := auth.delete("/api/auth/user/account", nil)
	require.Equalf(t, http.StatusOK, code, "delete account failed (body: %s)", body)
}

// mustRestoreAccount restores a soft-deleted account using email and password.
func mustRestoreAccount(t *testing.T, c *apiClient, email, password string) {
	t.Helper()
	code, body := c.post("/api/restore-account", map[string]string{
		"email":    email,
		"password": password,
	})
	require.Equalf(t, http.StatusOK, code, "restore account failed (body: %s)", body)
}

// mustResetPassword completes the password-reset using a JWT reset-password token.
func mustResetPassword(t *testing.T, c *apiClient, resetToken, newPassword string) {
	t.Helper()
	code, body := c.post("/api/reset-password", map[string]string{
		"reset_token":  resetToken,
		"new_password": newPassword,
	})
	require.Equalf(t, http.StatusOK, code, "reset password failed (body: %s)", body)
}

// ─── 2FA helpers ──────────────────────────────────────────────────────────────

// mustGet2FACode fetches the active 2FA code for a given session via debug endpoint.
// Only works when APP_ENV=test.
func mustGet2FACode(t *testing.T, c *apiClient, sessionUUID string) string {
	t.Helper()
	code, body := c.get(fmt.Sprintf("/api/debug/2fa/%s/code", sessionUUID))
	require.Equalf(t, http.StatusOK, code, "get 2FA code failed (body: %s)", body)
	var resp struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.Code, "2FA code is empty")
	return resp.Code
}

// mustEnable2FA enables two-factor authentication for the authenticated user.
func mustEnable2FA(t *testing.T, auth *apiClient) {
	t.Helper()
	code, body := auth.patch("/api/auth/user/2fa", map[string]bool{"enable_2fa": true})
	require.Equalf(t, http.StatusOK, code, "enable 2FA failed (body: %s)", body)
}

// mustDisable2FA disables two-factor authentication for the authenticated user.
func mustDisable2FA(t *testing.T, auth *apiClient) {
	t.Helper()
	code, body := auth.patch("/api/auth/user/2fa", map[string]bool{"enable_2fa": false})
	require.Equalf(t, http.StatusOK, code, "disable 2FA failed (body: %s)", body)
}

// mustLoginWith2FA performs step 1 of a 2FA login (password check) and returns the session_uuid.
// Fails the test if the response does not contain a session_uuid or unexpectedly returns tokens.
func mustLoginWith2FA(t *testing.T, c *apiClient, email, password string) string {
	t.Helper()
	httpCode, body := c.post("/api/login", map[string]string{
		"email":    email,
		"password": password,
	})
	require.Equalf(t, http.StatusOK, httpCode, "login step-1 (2FA) failed (body: %s)", body)

	var resp loginResp
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.SessionUUID, "expected session_uuid in 2FA login response")
	require.Empty(t, resp.AccessToken, "access_token must be empty during 2FA step 1")
	return resp.SessionUUID
}

// mustVerify2FA completes step 2 of the 2FA login and returns the resulting token pair.
func mustVerify2FA(t *testing.T, c *apiClient, sessionUUID, code string) loginResp {
	t.Helper()
	httpCode, body := c.post("/api/verify-2fa", map[string]string{
		"session_uuid": sessionUUID,
		"code":         code,
	})
	require.Equalf(t, http.StatusOK, httpCode, "verify 2FA failed (body: %s)", body)

	var resp loginResp
	require.NoError(t, json.Unmarshal(body, &resp))
	require.NotEmpty(t, resp.AccessToken, "verify 2FA returned empty access_token")
	require.NotEmpty(t, resp.RefreshToken, "verify 2FA returned empty refresh_token")
	return resp
}
