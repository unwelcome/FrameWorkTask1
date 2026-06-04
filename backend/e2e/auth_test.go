package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Register ─────────────────────────────────────────────────────────────────

func TestRegister(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/register", defaultUserPayload(randomEmail(), "Password123"))

		assert.Equal(t, http.StatusCreated, code)

		var resp registerResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.NotEmpty(t, resp.UserUUID)
	})

	t.Run("duplicate_email", func(t *testing.T) {
		c := newClient()
		email := randomEmail()

		code, _ := c.post("/api/register", defaultUserPayload(email, "Password123"))
		require.Equal(t, http.StatusCreated, code)

		code, body := c.post("/api/register", defaultUserPayload(email, "Password123"))
		assert.Equal(t, http.StatusConflict, code, "second registration with same email should return 409 (body: %s)", body)
	})

	t.Run("invalid_email", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/register", defaultUserPayload("not-an-email", "Password123"))

		assert.Equal(t, http.StatusBadRequest, code, "invalid email should return 400 (body: %s)", body)
	})

	t.Run("weak_password", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/register", defaultUserPayload(randomEmail(), "short"))

		assert.Equal(t, http.StatusBadRequest, code, "weak password should return 400 (body: %s)", body)
	})

	t.Run("missing_fields", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/register", map[string]string{
			"email":    randomEmail(),
			"password": "Password123",
		})

		assert.Equal(t, http.StatusBadRequest, code, "missing required fields should return 400 (body: %s)", body)
	})
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email := randomEmail()
		reg := mustRegister(t, c, email, "Password123")
		verificationCode := mustGetVerificationCode(t, c, reg.UserUUID)
		mustVerifyAccount(t, c, email, verificationCode)

		code, body := c.post("/api/login", map[string]string{
			"email":    email,
			"password": "Password123",
		})

		assert.Equal(t, http.StatusOK, code)

		var resp loginResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.NotEmpty(t, resp.UserUUID)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
	})

	t.Run("wrong_password", func(t *testing.T) {
		c := newClient()
		email := randomEmail()
		mustRegister(t, c, email, "Password123")

		code, body := c.post("/api/login", map[string]string{
			"email":    email,
			"password": "WrongPassword1",
		})

		assert.Equal(t, http.StatusBadRequest, code, "wrong password should return 400 (body: %s)", body)
	})

	t.Run("non_existent_user", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/login", map[string]string{
			"email":    randomEmail(),
			"password": "Password123",
		})

		assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
			"non-existent user should return 401 or 404, got %d (body: %s)", code, body)
	})

	t.Run("invalid_email", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/login", map[string]string{
			"email":    "bad-email",
			"password": "Password123",
		})

		assert.Equal(t, http.StatusBadRequest, code, "invalid email format should return 400 (body: %s)", body)
	})
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestRefreshToken(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)

		code, body := c.post("/api/refresh", map[string]string{
			"refresh_token": login.RefreshToken,
		})

		assert.Equal(t, http.StatusOK, code)

		var resp refreshTokenResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
	})

	t.Run("invalid_token", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/refresh", map[string]string{
			"refresh_token": "this.is.not.a.valid.jwt",
		})

		assert.Equal(t, http.StatusBadRequest, code, "invalid token should return 400 (body: %s)", body)
	})
}

// ─── GetUser ──────────────────────────────────────────────────────────────────

func TestGetUser(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.get("/api/auth/user/" + login.UserUUID + "/info")

		assert.Equal(t, http.StatusOK, code)

		var resp getUserResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.Equal(t, login.UserUUID, resp.UserUUID)
		assert.Equal(t, email, resp.Email)
		assert.NotEmpty(t, resp.FirstName)
		assert.NotEmpty(t, resp.CreatedAt)
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)

		code, body := c.get("/api/auth/user/" + login.UserUUID + "/info")
		assert.Equal(t, http.StatusUnauthorized, code, "request without token should return 401 (body: %s)", body)
	})

	t.Run("not_found", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.get("/api/auth/user/00000000-0000-0000-0000-000000000001/info")
		assert.Equal(t, http.StatusNotFound, code, "unknown uuid should return 404 (body: %s)", body)
	})
}

// ─── UpdateUserBio ────────────────────────────────────────────────────────────

func TestUpdateUserBio(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.patch("/api/auth/user/bio", map[string]string{
			"first_name": "Petr",
			"last_name":  "Petrov",
			"patronymic": "Petrovich",
		})
		assert.Equal(t, http.StatusOK, code, "update bio failed (body: %s)", body)

		code, body = auth.get("/api/auth/user/" + login.UserUUID + "/info")
		require.Equal(t, http.StatusOK, code)

		var user getUserResp
		require.NoError(t, json.Unmarshal(body, &user))
		assert.Equal(t, "Petr", user.FirstName)
		assert.Equal(t, "Petrov", user.LastName)
		assert.Equal(t, "Petrovich", user.Patronymic)
	})
}

// ─── ChangePassword ───────────────────────────────────────────────────────────

func TestChangePassword(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.patch("/api/auth/user/password", map[string]string{
			"password": "NewPassword456",
		})
		assert.Equal(t, http.StatusOK, code, "change password failed (body: %s)", body)

		code, _ = c.post("/api/login", map[string]string{
			"email":    email,
			"password": "NewPassword456",
		})
		assert.Equal(t, http.StatusOK, code, "login with new password should succeed")

		code, _ = c.post("/api/login", map[string]string{
			"email":    email,
			"password": "Password123",
		})
		assert.Equal(t, http.StatusBadRequest, code, "login with old password should fail after change")
	})
}

// ─── GetAllActiveTokens ───────────────────────────────────────────────────────

func TestGetAllActiveTokens(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.get("/api/auth/user/tokens")

		assert.Equal(t, http.StatusOK, code)

		var resp tokensResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.GreaterOrEqual(t, len(resp.Tokens), 1, "should have at least 1 active token after login")
	})
}

// ─── RevokeToken ──────────────────────────────────────────────────────────────

func TestRevokeToken(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		// Получаем хеш токена из списка активных токенов
		code, body := auth.get("/api/auth/user/tokens")
		require.Equal(t, http.StatusOK, code)
		var tokens tokensResp
		require.NoError(t, json.Unmarshal(body, &tokens))
		require.NotEmpty(t, tokens.Tokens, "expected at least one active token")
		tokenHash := tokens.Tokens[0].Token

		// Отзываем токен по хешу
		code, body = auth.delete("/api/auth/user/revoke/token", map[string]string{
			"token_hash": tokenHash,
		})
		assert.Equal(t, http.StatusOK, code, "revoke token failed (body: %s)", body)

		// Refresh с отозванным токеном должен завершиться ошибкой
		code, _ = c.post("/api/refresh", map[string]string{
			"refresh_token": login.RefreshToken,
		})
		assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
			"refresh with revoked token should fail, got %d", code)
	})
}

// ─── RevokeAllTokens ──────────────────────────────────────────────────────────

func TestRevokeAllTokens(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		// Создаём вторую сессию
		login2 := mustLogin(t, c, email, "Password123")

		code, body := auth.delete("/api/auth/user/revoke/all", nil)
		assert.Equal(t, http.StatusOK, code, "revoke all tokens failed (body: %s)", body)

		// Оба refresh токена должны стать невалидными
		code, _ = c.post("/api/refresh", map[string]string{"refresh_token": login.RefreshToken})
		assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
			"first token should be revoked, got %d", code)

		code, _ = c.post("/api/refresh", map[string]string{"refresh_token": login2.RefreshToken})
		assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
			"second token should be revoked, got %d", code)
	})
}

// ─── DeleteUser ───────────────────────────────────────────────────────────────

func TestDeleteUser(t *testing.T) {
	t.Run("self_delete", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.delete("/api/auth/user/account", map[string]string{
			"target_uuid": login.UserUUID,
		})
		assert.Equal(t, http.StatusOK, code, "delete user failed (body: %s)", body)

		code, _ = auth.get("/api/auth/user/" + login.UserUUID + "/info")
		assert.True(t, code == http.StatusNotFound || code == http.StatusUnauthorized,
			"deleted user should not be accessible, got %d", code)
	})
}

// ─── Full auth flow ───────────────────────────────────────────────────────────

func TestAuthFullFlow(t *testing.T) {
	c := newClient()
	email := randomEmail()

	// 1. Register
	regCode, regBody := c.post("/api/register", defaultUserPayload(email, "Password123"))
	require.Equal(t, http.StatusCreated, regCode, "register: %s", regBody)
	var reg registerResp
	require.NoError(t, json.Unmarshal(regBody, &reg))

	// 2. Verify account
	verificationCode := mustGetVerificationCode(t, c, reg.UserUUID)
	mustVerifyAccount(t, c, email, verificationCode)

	// 3. Login
	login := mustLogin(t, c, email, "Password123")
	assert.Equal(t, reg.UserUUID, login.UserUUID)
	auth := c.withToken(login.AccessToken)

	// 4. Get user info
	code, body := auth.get("/api/auth/user/" + login.UserUUID + "/info")
	require.Equal(t, http.StatusOK, code, "get user: %s", body)
	var user getUserResp
	require.NoError(t, json.Unmarshal(body, &user))
	assert.Equal(t, email, user.Email)

	// 5. Update bio
	code, body = auth.patch("/api/auth/user/bio", map[string]string{
		"first_name": "UpdatedFirstName",
		"last_name":  "UpdatedLastName",
		"patronymic": "UpdatedPatronymic",
	})
	assert.Equal(t, http.StatusOK, code, "update bio: %s", body)

	// 6. Refresh token
	code, body = c.post("/api/refresh", map[string]string{
		"refresh_token": login.RefreshToken,
	})
	require.Equal(t, http.StatusOK, code, "refresh: %s", body)
	var newTokens refreshTokenResp
	require.NoError(t, json.Unmarshal(body, &newTokens))
	auth = c.withToken(newTokens.AccessToken)

	// 7. New access token работает
	code, _ = auth.get("/api/auth/user/" + login.UserUUID + "/info")
	assert.Equal(t, http.StatusOK, code, "new access token should work")

	// 8. Проверяем активные токены
	code, body = auth.get("/api/auth/user/tokens")
	require.Equal(t, http.StatusOK, code)
	var activeTokens tokensResp
	require.NoError(t, json.Unmarshal(body, &activeTokens))
	assert.GreaterOrEqual(t, len(activeTokens.Tokens), 1)

	// 9. Revoke all
	code, body = auth.delete("/api/auth/user/revoke/all", nil)
	assert.Equal(t, http.StatusOK, code, "revoke all: %s", body)

	// 10. Refresh с новым токеном должен упасть
	code, _ = c.post("/api/refresh", map[string]string{
		"refresh_token": newTokens.RefreshToken,
	})
	assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
		"refresh after revoke-all should fail, got %d", code)
}
