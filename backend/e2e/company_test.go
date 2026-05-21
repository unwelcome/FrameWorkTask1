package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── CreateCompany ────────────────────────────────────────────────────────────

func TestCreateCompany(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.post("/api/auth/company/create", map[string]string{
			"title": randomTitle(),
		})

		assert.Equal(t, http.StatusCreated, status, "body: %s", body)

		var resp createCompanyResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.NotEmpty(t, resp.CompanyUUID)
	})

	t.Run("empty_title", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.post("/api/auth/company/create", map[string]string{
			"title": "",
		})

		assert.Equal(t, http.StatusBadRequest, status, "empty title should return 400 (body: %s)", body)
	})

	t.Run("title_too_long", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.post("/api/auth/company/create", map[string]string{
			"title": strings.Repeat("a", 256),
		})

		assert.Equal(t, http.StatusBadRequest, status, "title >255 chars should return 400 (body: %s)", body)
	})

	t.Run("invalid_chars_in_title", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.post("/api/auth/company/create", map[string]string{
			"title": "Invalid@Title!",
		})

		assert.Equal(t, http.StatusBadRequest, status, "title with invalid chars should return 400 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()

		status, body := c.post("/api/auth/company/create", map[string]string{
			"title": randomTitle(),
		})

		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── GetCompany ───────────────────────────────────────────────────────────────

func TestGetCompany(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		title := randomTitle()
		companyUUID := mustCreateCompany(t, auth, title)

		status, body := auth.get("/api/auth/company/" + companyUUID)

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		var resp companyResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, companyUUID, resp.CompanyUUID)
		assert.Equal(t, title, resp.Title)
		assert.NotEmpty(t, resp.Status)
	})

	// GetCompany не проверяет членство: любой авторизованный пользователь
	// может получить базовую информацию о компании (title, status).

	t.Run("not_found", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.get("/api/auth/company/00000000-0000-0000-0000-000000000001")
		assert.Equal(t, http.StatusNotFound, status, "unknown uuid should return 404 (body: %s)", body)
	})

	t.Run("invalid_uuid", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.get("/api/auth/company/not-a-uuid")
		assert.Equal(t, http.StatusBadRequest, status, "invalid uuid should return 400 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := c.get("/api/auth/company/" + companyUUID)
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── GetCompaniesList ─────────────────────────────────────────────────────────

func TestGetCompaniesList(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		mustCreateCompany(t, auth, randomTitle())

		// count обязателен: минимум 1, иначе gateway вернёт 400
		status, body := auth.get("/api/auth/company/list?offset=0&count=10")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		var resp companiesResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.NotNil(t, resp.Companies)
	})

	t.Run("pagination", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.get("/api/auth/company/list?offset=0&count=2")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		var resp companiesResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.LessOrEqual(t, len(resp.Companies), 2)
	})

	t.Run("count_below_minimum", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.get("/api/auth/company/list?count=0")
		assert.Equal(t, http.StatusBadRequest, status, "count=0 should return 400 (body: %s)", body)
	})

	t.Run("count_above_maximum", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.get("/api/auth/company/list?count=101")
		assert.Equal(t, http.StatusBadRequest, status, "count=101 should return 400 (body: %s)", body)
	})

	t.Run("negative_offset", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.get("/api/auth/company/list?offset=-1")
		assert.Equal(t, http.StatusBadRequest, status, "offset=-1 should return 400 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()

		status, body := c.get("/api/auth/company/list")
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── GetMyCompanies ───────────────────────────────────────────────────────────

func TestGetMyCompanies(t *testing.T) {
	t.Run("returns_created_company", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.get("/api/auth/company/my")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		var resp companiesResp
		require.NoError(t, json.Unmarshal(body, &resp))
		require.NotEmpty(t, resp.Companies, "should contain at least the created company")

		found := false
		for _, co := range resp.Companies {
			if co.CompanyUUID == companyUUID {
				found = true
				break
			}
		}
		assert.True(t, found, "created company should appear in /my list")
	})

	t.Run("empty_when_no_companies", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.get("/api/auth/company/my")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		var resp companiesResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Empty(t, resp.Companies, "user with no companies should get empty list")
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()

		status, body := c.get("/api/auth/company/my")
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── UpdateCompanyTitle ───────────────────────────────────────────────────────

func TestUpdateCompanyTitle(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.patch("/api/auth/company/"+companyUUID+"/title", map[string]string{
			"title": "Updated Title",
		})

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		// Проверяем, что название действительно изменилось
		status, body = auth.get("/api/auth/company/" + companyUUID)
		require.Equal(t, http.StatusOK, status)
		var resp companyResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, "Updated Title", resp.Title)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.patch("/api/auth/company/"+companyUUID+"/title", map[string]string{
			"title": "Hacked Title",
		})
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("empty_title", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.patch("/api/auth/company/"+companyUUID+"/title", map[string]string{
			"title": "",
		})
		assert.Equal(t, http.StatusBadRequest, status, "empty title should return 400 (body: %s)", body)
	})

	t.Run("title_too_long", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.patch("/api/auth/company/"+companyUUID+"/title", map[string]string{
			"title": strings.Repeat("a", 256),
		})
		assert.Equal(t, http.StatusBadRequest, status, "title >255 chars should return 400 (body: %s)", body)
	})

	t.Run("not_found", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.patch("/api/auth/company/00000000-0000-0000-0000-000000000001/title", map[string]string{
			"title": "New Title",
		})
		assert.Equal(t, http.StatusNotFound, status, "unknown company should return 404 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := c.patch("/api/auth/company/"+companyUUID+"/title", map[string]string{
			"title": "New Title",
		})
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── UpdateCompanyStatus ──────────────────────────────────────────────────────

func TestUpdateCompanyStatus(t *testing.T) {
	t.Run("chief_sets_close", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{
			"status": "close",
		})
		assert.Equal(t, http.StatusOK, status, "body: %s", body)
	})

	t.Run("chief_sets_open", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		// Сначала закрываем
		status, body := auth.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{
			"status": "close",
		})
		require.Equal(t, http.StatusOK, status, "set close failed (body: %s)", body)

		// Потом открываем снова
		status, body = auth.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{
			"status": "open",
		})
		assert.Equal(t, http.StatusOK, status, "body: %s", body)
	})

	t.Run("invalid_status", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{
			"status": "paused",
		})
		assert.Equal(t, http.StatusBadRequest, status, "invalid status should return 400 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{
			"status": "close",
		})
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := c.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{
			"status": "close",
		})
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── DeleteCompany ────────────────────────────────────────────────────────────

func TestDeleteCompany(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.delete("/api/auth/company/"+companyUUID, nil)

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		// После удаления компания должна быть недоступна
		status, body = auth.get("/api/auth/company/" + companyUUID)
		assert.Equal(t, http.StatusNotFound, status, "deleted company should return 404 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.delete("/api/auth/company/"+companyUUID, nil)
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("not_found", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.delete("/api/auth/company/00000000-0000-0000-0000-000000000001", nil)
		assert.Equal(t, http.StatusNotFound, status, "unknown company should return 404 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := c.delete("/api/auth/company/"+companyUUID, nil)
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── CreateJoinCode ───────────────────────────────────────────────────────────

func TestCreateJoinCode(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.post("/api/auth/company/"+companyUUID+"/code", map[string]int{
			"code_ttl": 3600,
		})

		assert.Equal(t, http.StatusCreated, status, "body: %s", body)

		var resp createCodeResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Len(t, resp.Code, 6, "join code should be exactly 6 characters")
		assert.Regexp(t, `^\d{6}$`, resp.Code, "join code should consist of 6 digits")
	})

	t.Run("ttl_below_minimum", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.post("/api/auth/company/"+companyUUID+"/code", map[string]int{
			"code_ttl": 59,
		})
		assert.Equal(t, http.StatusBadRequest, status, "ttl=59 should return 400 (body: %s)", body)
	})

	t.Run("ttl_above_maximum", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.post("/api/auth/company/"+companyUUID+"/code", map[string]int{
			"code_ttl": 604801,
		})
		assert.Equal(t, http.StatusBadRequest, status, "ttl=604801 should return 400 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.post("/api/auth/company/"+companyUUID+"/code", map[string]int{
			"code_ttl": 3600,
		})
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := c.post("/api/auth/company/"+companyUUID+"/code", map[string]int{
			"code_ttl": 3600,
		})
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── GetJoinCodes ─────────────────────────────────────────────────────────────

func TestGetJoinCodes(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())
		mustCreateCode(t, auth, companyUUID, 3600)

		status, body := auth.get("/api/auth/company/" + companyUUID + "/codes")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		var resp codesResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.GreaterOrEqual(t, len(resp.Codes), 1, "should return at least the created code")
	})

	t.Run("empty_when_no_codes", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.get("/api/auth/company/" + companyUUID + "/codes")

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		var resp codesResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Empty(t, resp.Codes, "should return empty list when no codes exist")
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.get("/api/auth/company/" + companyUUID + "/codes")
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := c.get("/api/auth/company/" + companyUUID + "/codes")
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── JoinCompany ──────────────────────────────────────────────────────────────

func TestJoinCompany(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		code := mustCreateCode(t, chief, companyUUID, 3600)

		_, guestLogin := mustRegisterAndLogin(t, c)
		guest := c.withToken(guestLogin.AccessToken)

		status, body := guest.post("/api/auth/company/join", map[string]string{
			"code": code,
		})

		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		var resp joinCompanyResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, companyUUID, resp.CompanyUUID)
		assert.Equal(t, "unemployed", resp.Role, "new member should get unemployed role")
	})

	t.Run("nonexistent_code", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.post("/api/auth/company/join", map[string]string{
			"code": "000000",
		})
		assert.Equal(t, http.StatusNotFound, status, "nonexistent code should return 404 (body: %s)", body)
	})

	t.Run("invalid_code_format", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		status, body := auth.post("/api/auth/company/join", map[string]string{
			"code": "abc",
		})
		assert.Equal(t, http.StatusBadRequest, status, "invalid code format should return 400 (body: %s)", body)
	})

	t.Run("already_member", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		_, guestLogin := mustRegisterAndLogin(t, c)
		guest := c.withToken(guestLogin.AccessToken)

		// Вступаем первый раз
		code := mustCreateCode(t, chief, companyUUID, 3600)
		mustJoinCompany(t, guest, code)

		// Создаём второй код и пытаемся вступить снова
		code2 := mustCreateCode(t, chief, companyUUID, 3600)
		status, body := guest.post("/api/auth/company/join", map[string]string{
			"code": code2,
		})
		assert.Equal(t, http.StatusConflict, status, "already member should return 409 (body: %s)", body)
	})

	t.Run("company_closed", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())

		// Создаём код пока компания открыта
		code := mustCreateCode(t, chief, companyUUID, 3600)

		// Закрываем компанию
		status, body := chief.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{
			"status": "close",
		})
		require.Equal(t, http.StatusOK, status, "set close failed (body: %s)", body)

		// Пытаемся вступить
		_, guestLogin := mustRegisterAndLogin(t, c)
		guest := c.withToken(guestLogin.AccessToken)
		status, body = guest.post("/api/auth/company/join", map[string]string{
			"code": code,
		})
		assert.Equal(t, http.StatusForbidden, status, "joining closed company should return 403 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		code := mustCreateCode(t, chief, companyUUID, 3600)

		status, body := c.post("/api/auth/company/join", map[string]string{
			"code": code,
		})
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── DeleteJoinCode ───────────────────────────────────────────────────────────

func TestDeleteJoinCode(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())
		code := mustCreateCode(t, auth, companyUUID, 3600)

		status, body := auth.delete("/api/auth/company/"+companyUUID+"/code", map[string]string{
			"code": code,
		})
		assert.Equal(t, http.StatusOK, status, "body: %s", body)

		// Код больше не должен быть в списке
		status, body = auth.get("/api/auth/company/" + companyUUID + "/codes")
		require.Equal(t, http.StatusOK, status)
		var resp codesResp
		require.NoError(t, json.Unmarshal(body, &resp))
		for _, existingCode := range resp.Codes {
			assert.NotEqual(t, code, existingCode, "deleted code should not appear in list")
		}
	})

	t.Run("nonexistent_code", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.delete("/api/auth/company/"+companyUUID+"/code", map[string]string{
			"code": "000000",
		})
		assert.Equal(t, http.StatusNotFound, status, "nonexistent code should return 404 (body: %s)", body)
	})

	t.Run("invalid_code_format", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())

		status, body := auth.delete("/api/auth/company/"+companyUUID+"/code", map[string]string{
			"code": "abc",
		})
		assert.Equal(t, http.StatusBadRequest, status, "invalid code format should return 400 (body: %s)", body)
	})

	t.Run("non_chief_forbidden", func(t *testing.T) {
		c := newClient()
		_, chiefLogin := mustRegisterAndLogin(t, c)
		chief := c.withToken(chiefLogin.AccessToken)
		companyUUID := mustCreateCompany(t, chief, randomTitle())
		code := mustCreateCode(t, chief, companyUUID, 3600)

		_, memberLogin := mustRegisterAndLogin(t, c)
		member := c.withToken(memberLogin.AccessToken)
		mustAddMember(t, chief, member, companyUUID)

		status, body := member.delete("/api/auth/company/"+companyUUID+"/code", map[string]string{
			"code": code,
		})
		assert.Equal(t, http.StatusForbidden, status, "non-chief should get 403 (body: %s)", body)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		companyUUID := mustCreateCompany(t, auth, randomTitle())
		code := mustCreateCode(t, auth, companyUUID, 3600)

		status, body := c.delete("/api/auth/company/"+companyUUID+"/code", map[string]string{
			"code": code,
		})
		assert.Equal(t, http.StatusUnauthorized, status, "request without token should return 401 (body: %s)", body)
	})
}

// ─── Full company workflow ────────────────────────────────────────────────────

func TestCompanyFullWorkflow(t *testing.T) {
	c := newClient()

	// 1. Пользователь A создаёт компанию — автоматически становится chief
	_, chiefLogin := mustRegisterAndLogin(t, c)
	chief := c.withToken(chiefLogin.AccessToken)
	companyUUID := mustCreateCompany(t, chief, randomTitle())

	infoStatus, infoBody := chief.get("/api/auth/company/" + companyUUID)
	require.Equal(t, http.StatusOK, infoStatus, "chief should see own company (body: %s)", infoBody)
	var company companyResp
	require.NoError(t, json.Unmarshal(infoBody, &company))
	assert.Equal(t, "open", company.Status, "new company should have open status")

	// 2. Chief генерирует код вступления
	code := mustCreateCode(t, chief, companyUUID, 3600)
	assert.Regexp(t, `^\d{6}$`, code)

	// Код виден в списке кодов
	codesStatus, codesBody := chief.get("/api/auth/company/" + companyUUID + "/codes")
	require.Equal(t, http.StatusOK, codesStatus)
	var codes codesResp
	require.NoError(t, json.Unmarshal(codesBody, &codes))
	assert.Contains(t, codes.Codes, code, "created code should appear in codes list")

	// 3. Пользователь B вступает по коду — получает роль unemployed
	_, memberLogin := mustRegisterAndLogin(t, c)
	member := c.withToken(memberLogin.AccessToken)
	joinResp := mustJoinCompany(t, member, code)
	assert.Equal(t, companyUUID, joinResp.CompanyUUID)
	assert.Equal(t, "unemployed", joinResp.Role)

	// 4. Chief назначает B роль engineer
	roleStatus, roleBody := chief.patch(
		fmt.Sprintf("/api/auth/company/%s/employee/%s/role", companyUUID, memberLogin.UserUUID),
		map[string]string{"role": "engineer"},
	)
	assert.Equal(t, http.StatusOK, roleStatus, "set role failed (body: %s)", roleBody)

	// 5. Chief закрывает компанию
	closeStatus, closeBody := chief.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{
		"status": "close",
	})
	assert.Equal(t, http.StatusOK, closeStatus, "close company failed (body: %s)", closeBody)

	// 6. Пользователь C пытается вступить по новому коду — получает 403
	code2 := mustCreateCode(t, chief, companyUUID, 3600)
	_, guestLogin := mustRegisterAndLogin(t, c)
	guest := c.withToken(guestLogin.AccessToken)
	joinStatus, joinBody := guest.post("/api/auth/company/join", map[string]string{"code": code2})
	assert.Equal(t, http.StatusForbidden, joinStatus, "joining closed company should return 403 (body: %s)", joinBody)

	// 7. Chief открывает компанию обратно
	openStatus, openBody := chief.patch("/api/auth/company/"+companyUUID+"/status", map[string]string{
		"status": "open",
	})
	assert.Equal(t, http.StatusOK, openStatus, "open company failed (body: %s)", openBody)

	// 8. Chief удаляет сотрудника B
	removeStatus, removeBody := chief.delete(
		fmt.Sprintf("/api/auth/company/%s/employee/%s", companyUUID, memberLogin.UserUUID),
		nil,
	)
	assert.Equal(t, http.StatusOK, removeStatus, "remove employee failed (body: %s)", removeBody)

	// 9. Chief удаляет компанию
	deleteStatus, deleteBody := chief.delete("/api/auth/company/"+companyUUID, nil)
	assert.Equal(t, http.StatusOK, deleteStatus, "delete company failed (body: %s)", deleteBody)

	// После удаления компания недоступна
	getStatus, _ := chief.get("/api/auth/company/" + companyUUID)
	assert.Equal(t, http.StatusNotFound, getStatus, "deleted company should return 404")
}
