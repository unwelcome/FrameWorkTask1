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
