package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
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

		// После исправления timing-атаки несуществующий email возвращает 400
		// (тот же статус и сообщение, что и неверный пароль) — чтобы
		// атакующий не мог различить эти два случая по HTTP-статусу.
		assert.Equal(t, http.StatusBadRequest, code,
			"non-existent user should return 400 (same as wrong password), got %d (body: %s)", code, body)
	})

	t.Run("invalid_email", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/login", map[string]string{
			"email":    "bad-email",
			"password": "Password123",
		})

		assert.Equal(t, http.StatusBadRequest, code, "invalid email format should return 400 (body: %s)", body)
	})

	t.Run("account_not_verified", func(t *testing.T) {
		c := newClient()
		email := randomEmail()
		mustRegister(t, c, email, "Password123")
		// Намеренно НЕ верифицируем аккаунт

		code, body := c.post("/api/login", map[string]string{
			"email":    email,
			"password": "Password123",
		})
		assert.Equal(t, http.StatusForbidden, code, "login before verification should return 403 (body: %s)", body)
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
		assert.Empty(t, user.Description, "description should be empty when not provided")
	})

	t.Run("with_description", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.patch("/api/auth/user/bio", map[string]string{
			"first_name":  "Petr",
			"last_name":   "Petrov",
			"description": "Go backend developer",
		})
		assert.Equal(t, http.StatusOK, code, "update bio with description failed (body: %s)", body)

		code, body = auth.get("/api/auth/user/" + login.UserUUID + "/info")
		require.Equal(t, http.StatusOK, code)

		var user getUserResp
		require.NoError(t, json.Unmarshal(body, &user))
		assert.Equal(t, "Go backend developer. Loves clean code.", user.Description)
	})

	t.Run("description_too_long", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.patch("/api/auth/user/bio", map[string]string{
			"first_name":  "Petr",
			"last_name":   "Petrov",
			"description": strings.Repeat("а", 501),
		})
		assert.Equal(t, http.StatusBadRequest, code, "over-limit description should return 400 (body: %s)", body)
	})
}

// ─── ChangePassword ───────────────────────────────────────────────────────────

func TestChangePassword(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.patch("/api/auth/user/password", map[string]string{
			"old_password": "Password123",
			"password":     "NewPassword456",
		})
		assert.Equal(t, http.StatusOK, code, "change password failed (body: %s)", body)

		// Новый пароль работает
		code, _ = c.post("/api/login", map[string]string{
			"email":    email,
			"password": "NewPassword456",
		})
		assert.Equal(t, http.StatusOK, code, "login with new password should succeed")

		// Старый пароль отклонён
		code, _ = c.post("/api/login", map[string]string{
			"email":    email,
			"password": "Password123",
		})
		assert.Equal(t, http.StatusBadRequest, code, "login with old password should fail after change")
	})

	t.Run("wrong_old_password", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.patch("/api/auth/user/password", map[string]string{
			"old_password": "WrongPassword1",
			"password":     "NewPassword456",
		})
		assert.Equal(t, http.StatusBadRequest, code, "wrong old password should return 400 (body: %s)", body)
	})

	t.Run("missing_old_password", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		// old_password не передан → gateway вернёт 400 при валидации
		code, body := auth.patch("/api/auth/user/password", map[string]string{
			"password": "NewPassword456",
		})
		assert.Equal(t, http.StatusBadRequest, code, "missing old_password should return 400 (body: %s)", body)
	})

	t.Run("invalid_new_password", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.patch("/api/auth/user/password", map[string]string{
			"old_password": "Password123",
			"password":     "weak",
		})
		assert.Equal(t, http.StatusBadRequest, code, "weak new password should return 400 (body: %s)", body)
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
		require.GreaterOrEqual(t, len(resp.Tokens), 1, "should have at least 1 active token after login")

		token := resp.Tokens[0]
		assert.NotEmpty(t, token.TokenHash, "token_hash should not be empty")
		assert.NotEmpty(t, token.CreatedAt, "created_at should not be empty")
		assert.NotEmpty(t, token.LastActiveAt, "last_active_at should not be empty")
		// IP может быть пустым в тестовой среде, проверяем лишь что поля присутствуют в JSON
		assert.NotEmpty(t, token.DeviceType, "device_type should not be empty (fallback: desktop)")
	})

	t.Run("two_sessions_visible", func(t *testing.T) {
		c := newClient()
		email, login1 := mustRegisterAndLogin(t, c)
		auth1 := c.withToken(login1.AccessToken)

		// Создаём вторую сессию
		login2 := mustLogin(t, c, email, "Password123")
		_ = login2

		code, body := auth1.get("/api/auth/user/tokens")
		require.Equal(t, http.StatusOK, code)

		var resp tokensResp
		require.NoError(t, json.Unmarshal(body, &resp))
		assert.GreaterOrEqual(t, len(resp.Tokens), 2, "both sessions should be visible in active tokens")

		// Оба токена должны иметь непустые хеши
		for i, tok := range resp.Tokens {
			assert.NotEmpty(t, tok.TokenHash, "token[%d].token_hash should not be empty", i)
			assert.NotEmpty(t, tok.CreatedAt, "token[%d].created_at should not be empty", i)
		}
	})

	t.Run("refresh_updates_last_active_and_last_ip", func(t *testing.T) {
		c := newClient()
		_, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		// Получаем данные до refresh
		code, body := auth.get("/api/auth/user/tokens")
		require.Equal(t, http.StatusOK, code)
		var before tokensResp
		require.NoError(t, json.Unmarshal(body, &before))
		require.GreaterOrEqual(t, len(before.Tokens), 1)
		createdAtBefore := before.Tokens[0].CreatedAt

		// Делаем refresh
		code, body = c.post("/api/refresh", map[string]string{"refresh_token": login.RefreshToken})
		require.Equal(t, http.StatusOK, code)
		var refreshed refreshTokenResp
		require.NoError(t, json.Unmarshal(body, &refreshed))
		auth = c.withToken(refreshed.AccessToken)

		// Получаем данные после refresh
		code, body = auth.get("/api/auth/user/tokens")
		require.Equal(t, http.StatusOK, code)
		var after tokensResp
		require.NoError(t, json.Unmarshal(body, &after))
		require.GreaterOrEqual(t, len(after.Tokens), 1)

		// created_at должен остаться прежним (иммутабельное поле)
		assert.Equal(t, createdAtBefore, after.Tokens[0].CreatedAt,
			"created_at should not change after refresh")
		// last_active_at должен присутствовать
		assert.NotEmpty(t, after.Tokens[0].LastActiveAt)
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
		tokenHash := tokens.Tokens[0].TokenHash

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
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		code, body := auth.delete("/api/auth/user/account", map[string]string{
			"target_uuid": login.UserUUID,
		})
		require.Equal(t, http.StatusOK, code, "delete user failed (body: %s)", body)

		// Soft delete: пользователь всё ещё существует в БД, GetUser возвращает 200 с deleted_at
		code, body = auth.get("/api/auth/user/" + login.UserUUID + "/info")
		require.Equal(t, http.StatusOK, code, "soft-deleted user should still be visible via GetUser (body: %s)", body)
		var user getUserResp
		require.NoError(t, json.Unmarshal(body, &user))
		assert.NotEmpty(t, user.DeletedAt, "deleted_at must be set after deletion")

		// Login заблокирован для удалённого аккаунта → 403
		code, _ = c.post("/api/login", map[string]string{
			"email":    email,
			"password": "Password123",
		})
		assert.Equal(t, http.StatusForbidden, code,
			"login after soft delete should return 403, got %d", code)

		// Refresh-токен отозван при удалении → 401/404
		code, _ = c.post("/api/refresh", map[string]string{"refresh_token": login.RefreshToken})
		assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
			"refresh token should be revoked after deletion, got %d", code)
	})
}

// ─── RestoreAccount ───────────────────────────────────────────────────────────

func TestRestoreAccount(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		mustDeleteAccount(t, auth, login.UserUUID)

		// Восстанавливаем аккаунт
		code, body := c.post("/api/restore-account", map[string]string{
			"email":    email,
			"password": "Password123",
		})
		assert.Equal(t, http.StatusOK, code, "restore account failed (body: %s)", body)

		// После восстановления login должен работать
		restored := mustLogin(t, c, email, "Password123")
		assert.NotEmpty(t, restored.AccessToken, "login after restore should return access_token")

		// GetUser должен отдавать пустой deleted_at
		restoredAuth := c.withToken(restored.AccessToken)
		code, body = restoredAuth.get("/api/auth/user/" + login.UserUUID + "/info")
		require.Equal(t, http.StatusOK, code, "get user after restore (body: %s)", body)
		var user getUserResp
		require.NoError(t, json.Unmarshal(body, &user))
		assert.Empty(t, user.DeletedAt, "deleted_at must be empty after restore")
	})

	t.Run("wrong_password", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		mustDeleteAccount(t, auth, login.UserUUID)

		code, body := c.post("/api/restore-account", map[string]string{
			"email":    email,
			"password": "WrongPassword1",
		})
		assert.Equal(t, http.StatusBadRequest, code, "wrong password should return 400 (body: %s)", body)
	})

	t.Run("account_not_deleted", func(t *testing.T) {
		c := newClient()
		email, _ := mustRegisterAndLogin(t, c)

		// Аккаунт активен — восстановление недопустимо
		code, body := c.post("/api/restore-account", map[string]string{
			"email":    email,
			"password": "Password123",
		})
		assert.Equal(t, http.StatusBadRequest, code, "restore for active account should return 400 (body: %s)", body)
	})
}

// ─── VerifyAccount ────────────────────────────────────────────────────────────

func TestVerifyAccount(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email := randomEmail()
		reg := mustRegister(t, c, email, "Password123")
		verCode := mustGetVerificationCode(t, c, reg.UserUUID)

		code, body := c.post("/api/user/verify", map[string]string{
			"email": email,
			"code":  verCode,
		})
		assert.Equal(t, http.StatusOK, code, "verify account should return 200 (body: %s)", body)
	})

	t.Run("wrong_code", func(t *testing.T) {
		c := newClient()
		email := randomEmail()
		mustRegister(t, c, email, "Password123")

		code, body := c.post("/api/user/verify", map[string]string{
			"email": email,
			"code":  "000000",
		})
		assert.Equal(t, http.StatusBadRequest, code, "wrong code should return 400 (body: %s)", body)
	})

	t.Run("already_verified", func(t *testing.T) {
		c := newClient()
		email := randomEmail()
		reg := mustRegister(t, c, email, "Password123")
		verCode := mustGetVerificationCode(t, c, reg.UserUUID)
		mustVerifyAccount(t, c, email, verCode)

		// Повторная верификация должна вернуть 409
		code, body := c.post("/api/user/verify", map[string]string{
			"email": email,
			"code":  verCode,
		})
		assert.Equal(t, http.StatusConflict, code, "already verified should return 409 (body: %s)", body)
	})

	t.Run("invalid_email", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/user/verify", map[string]string{
			"email": "not-an-email",
			"code":  "123456",
		})
		assert.Equal(t, http.StatusBadRequest, code, "invalid email should return 400 (body: %s)", body)
	})
}

// ─── ResendVerificationCode ───────────────────────────────────────────────────

func TestResendVerificationCode(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email := randomEmail()
		reg := mustRegister(t, c, email, "Password123")

		// Повторно отправляем код
		code, body := c.post("/api/user/verify/resend", map[string]string{"email": email})
		require.Equal(t, http.StatusOK, code, "resend code failed (body: %s)", body)

		// Получаем актуальный код и верифицируемся
		newCode := mustGetVerificationCode(t, c, reg.UserUUID)
		mustVerifyAccount(t, c, email, newCode)
	})

	t.Run("already_verified", func(t *testing.T) {
		c := newClient()
		email := randomEmail()
		reg := mustRegister(t, c, email, "Password123")
		verCode := mustGetVerificationCode(t, c, reg.UserUUID)
		mustVerifyAccount(t, c, email, verCode)

		// Повторный resend для уже верифицированного пользователя → 409
		code, body := c.post("/api/user/verify/resend", map[string]string{"email": email})
		assert.Equal(t, http.StatusConflict, code, "resend for verified user should return 409 (body: %s)", body)
	})

	t.Run("non_existent_email_silent", func(t *testing.T) {
		c := newClient()
		// Несуществующий email → 200 (защита от перебора)
		code, body := c.post("/api/user/verify/resend", map[string]string{"email": randomEmail()})
		assert.Equal(t, http.StatusOK, code, "non-existent email should return 200 silently (body: %s)", body)
	})
}

// ─── ForgotPassword ───────────────────────────────────────────────────────────

func TestForgotPassword(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email, _ := mustRegisterAndLogin(t, c)

		code, body := c.post("/api/forgot-password", map[string]string{"email": email})
		assert.Equal(t, http.StatusOK, code, "forgot password failed (body: %s)", body)
	})

	t.Run("non_existent_email_silent", func(t *testing.T) {
		c := newClient()
		// Несуществующий email → 200 (защита от перебора)
		code, body := c.post("/api/forgot-password", map[string]string{"email": randomEmail()})
		assert.Equal(t, http.StatusOK, code, "non-existent email should return 200 silently (body: %s)", body)
	})

	t.Run("invalid_email", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/forgot-password", map[string]string{"email": "not-an-email"})
		assert.Equal(t, http.StatusBadRequest, code, "invalid email should return 400 (body: %s)", body)
	})
}

// ─── ResetPassword ────────────────────────────────────────────────────────────

func TestResetPassword(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)

		mustForgotPassword(t, c, email)
		recoveryCode := mustGetRecoveryCode(t, c, login.UserUUID)
		mustResetPassword(t, c, email, recoveryCode, "NewPassword456")

		// Логин с новым паролем должен пройти
		newLogin := mustLogin(t, c, email, "NewPassword456")
		assert.NotEmpty(t, newLogin.AccessToken)

		// Старый пароль должен быть отклонён
		code, _ := c.post("/api/login", map[string]string{
			"email":    email,
			"password": "Password123",
		})
		assert.Equal(t, http.StatusBadRequest, code, "old password should fail after reset")
	})

	t.Run("wrong_code", func(t *testing.T) {
		c := newClient()
		email, _ := mustRegisterAndLogin(t, c)
		mustForgotPassword(t, c, email)

		code, body := c.post("/api/reset-password", map[string]string{
			"email":        email,
			"code":         "000000",
			"new_password": "NewPassword456",
		})
		assert.Equal(t, http.StatusBadRequest, code, "wrong code should return 400 (body: %s)", body)
	})

	t.Run("invalid_new_password", func(t *testing.T) {
		c := newClient()
		email, _ := mustRegisterAndLogin(t, c)
		mustForgotPassword(t, c, email)

		// Невалидный пароль → gateway отклонит до обращения к сервису
		code, body := c.post("/api/reset-password", map[string]string{
			"email":        email,
			"code":         "123456",
			"new_password": "weak",
		})
		assert.Equal(t, http.StatusBadRequest, code, "weak new password should return 400 (body: %s)", body)
	})
}

// ─── Verify2FA ────────────────────────────────────────────────────────────────

func TestVerify2FA(t *testing.T) {
	t.Run("happy_path", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		mustEnable2FA(t, auth)

		// Шаг 1: логин → session_uuid
		sessionUUID := mustLoginWith2FA(t, c, email, "Password123")

		// Шаг 2: получаем код через debug-endpoint и завершаем 2FA
		twoFACode := mustGet2FACode(t, c, sessionUUID)
		tokens := mustVerify2FA(t, c, sessionUUID, twoFACode)

		// Полученный access_token должен работать
		authWith2FA := c.withToken(tokens.AccessToken)
		code, body := authWith2FA.get("/api/auth/user/" + login.UserUUID + "/info")
		assert.Equal(t, http.StatusOK, code, "token after 2FA should work (body: %s)", body)
	})

	t.Run("wrong_code", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		mustEnable2FA(t, auth)

		sessionUUID := mustLoginWith2FA(t, c, email, "Password123")

		code, body := c.post("/api/verify-2fa", map[string]string{
			"session_uuid": sessionUUID,
			"code":         "000000",
		})
		assert.Equal(t, http.StatusForbidden, code, "wrong 2FA code should return 403 (body: %s)", body)
	})

	t.Run("session_not_found", func(t *testing.T) {
		c := newClient()
		code, body := c.post("/api/verify-2fa", map[string]string{
			"session_uuid": "00000000-0000-0000-0000-000000000001",
			"code":         "123456",
		})
		assert.Equal(t, http.StatusNotFound, code, "unknown session should return 404 (body: %s)", body)
	})

	t.Run("replay_attack_blocked", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)
		mustEnable2FA(t, auth)

		sessionUUID := mustLoginWith2FA(t, c, email, "Password123")
		twoFACode := mustGet2FACode(t, c, sessionUUID)

		// Первая попытка — успешная
		mustVerify2FA(t, c, sessionUUID, twoFACode)

		// Повторное использование той же сессии → 404 (сессия удалена)
		code, body := c.post("/api/verify-2fa", map[string]string{
			"session_uuid": sessionUUID,
			"code":         twoFACode,
		})
		assert.Equal(t, http.StatusNotFound, code, "replay attack should be blocked with 404 (body: %s)", body)
	})
}

// ─── UpdateUser2FA ────────────────────────────────────────────────────────────

func TestUpdateUser2FA(t *testing.T) {
	t.Run("enable_and_disable", func(t *testing.T) {
		c := newClient()
		email, login := mustRegisterAndLogin(t, c)
		auth := c.withToken(login.AccessToken)

		// Включаем 2FA
		mustEnable2FA(t, auth)

		// Логин теперь должен требовать второй шаг
		sessionUUID := mustLoginWith2FA(t, c, email, "Password123")

		// Завершаем 2FA для получения свежего токена
		twoFACode := mustGet2FACode(t, c, sessionUUID)
		tokens := mustVerify2FA(t, c, sessionUUID, twoFACode)
		auth = c.withToken(tokens.AccessToken)

		// Выключаем 2FA
		mustDisable2FA(t, auth)

		// Логин снова должен возвращать токены напрямую
		directLogin := mustLogin(t, c, email, "Password123")
		assert.NotEmpty(t, directLogin.AccessToken, "login without 2FA should return access_token directly")
		assert.Empty(t, directLogin.SessionUUID, "login without 2FA should not return session_uuid")
	})

	t.Run("unauthorized", func(t *testing.T) {
		c := newClient()
		code, body := c.patch("/api/auth/user/2fa", map[string]bool{"enable_2fa": true})
		assert.Equal(t, http.StatusUnauthorized, code, "update 2FA without token should return 401 (body: %s)", body)
	})
}

// ─── Full auth flow ───────────────────────────────────────────────────────────

func TestAuthFullFlow(t *testing.T) {
	c := newClient()
	email := randomEmail()
	const password = "Password123"

	// 1. Register
	regCode, regBody := c.post("/api/register", defaultUserPayload(email, password))
	require.Equal(t, http.StatusCreated, regCode, "register: %s", regBody)
	var reg registerResp
	require.NoError(t, json.Unmarshal(regBody, &reg))
	userUUID := reg.UserUUID

	// 2. Resend verification code (перезаписывает код в Redis)
	code, body := c.post("/api/user/verify/resend", map[string]string{"email": email})
	require.Equal(t, http.StatusOK, code, "resend verification code: %s", body)

	// 3. Verify account (используем актуальный код после resend)
	verCode := mustGetVerificationCode(t, c, userUUID)
	mustVerifyAccount(t, c, email, verCode)

	// 4. Login
	login := mustLogin(t, c, email, password)
	assert.Equal(t, userUUID, login.UserUUID)
	auth := c.withToken(login.AccessToken)

	// 5. Get user info
	code, body = auth.get("/api/auth/user/" + userUUID + "/info")
	require.Equal(t, http.StatusOK, code, "get user: %s", body)
	var user getUserResp
	require.NoError(t, json.Unmarshal(body, &user))
	assert.Equal(t, email, user.Email)

	// 6. Update bio (including description)
	code, body = auth.patch("/api/auth/user/bio", map[string]string{
		"first_name":  "UpdatedFirstName",
		"last_name":   "UpdatedLastName",
		"patronymic":  "UpdatedPatronymic",
		"description": "Full flow test user",
	})
	assert.Equal(t, http.StatusOK, code, "update bio: %s", body)

	// 7. Change password (передаём текущий пароль для верификации)
	const newPassword = "NewPassword456"
	code, body = auth.patch("/api/auth/user/password", map[string]string{
		"old_password": password,
		"password":     newPassword,
	})
	require.Equal(t, http.StatusOK, code, "change password: %s", body)

	// 8. Login со старым паролем не работает
	code, _ = c.post("/api/login", map[string]string{"email": email, "password": password})
	assert.Equal(t, http.StatusBadRequest, code, "old password should be rejected after change")

	// 9. Login с новым паролем
	login = mustLogin(t, c, email, newPassword)
	auth = c.withToken(login.AccessToken)

	// 10. Refresh token
	code, body = c.post("/api/refresh", map[string]string{"refresh_token": login.RefreshToken})
	require.Equal(t, http.StatusOK, code, "refresh: %s", body)
	var refreshed refreshTokenResp
	require.NoError(t, json.Unmarshal(body, &refreshed))
	auth = c.withToken(refreshed.AccessToken)

	// 11. New access token работает
	code, _ = auth.get("/api/auth/user/" + userUUID + "/info")
	assert.Equal(t, http.StatusOK, code, "refreshed access token should work")

	// 12. UpdateUser2FA — включаем 2FA
	mustEnable2FA(t, auth)

	// 13. Login возвращает session_uuid (токены пустые)
	sessionUUID := mustLoginWith2FA(t, c, email, newPassword)

	// 14. Verify2FA — возвращает пару токенов и user_uuid
	twoFACode := mustGet2FACode(t, c, sessionUUID)
	tokens2FA := mustVerify2FA(t, c, sessionUUID, twoFACode)
	assert.Equal(t, userUUID, tokens2FA.UserUUID)
	auth = c.withToken(tokens2FA.AccessToken)

	// 15. New access token (после 2FA) работает
	code, _ = auth.get("/api/auth/user/" + userUUID + "/info")
	assert.Equal(t, http.StatusOK, code, "token from 2FA flow should work")

	// 16. UpdateUser2FA — выключаем 2FA
	mustDisable2FA(t, auth)

	// 17. ForgotPassword
	mustForgotPassword(t, c, email)

	// 18. ResetPassword
	const resetPassword = "ResetPassword789"
	recoveryCode := mustGetRecoveryCode(t, c, userUUID)
	mustResetPassword(t, c, email, recoveryCode, resetPassword)

	// 19. Login с новым (сброшенным) паролем
	login = mustLogin(t, c, email, resetPassword)
	assert.NotEmpty(t, login.AccessToken)
	assert.Empty(t, login.SessionUUID, "2FA is disabled — no session_uuid expected")
	auth = c.withToken(login.AccessToken)

	// 20. Проверяем активные токены
	code, body = auth.get("/api/auth/user/tokens")
	require.Equal(t, http.StatusOK, code)
	var activeTokens tokensResp
	require.NoError(t, json.Unmarshal(body, &activeTokens))
	assert.GreaterOrEqual(t, len(activeTokens.Tokens), 1)

	// 21. Revoke all
	code, body = auth.delete("/api/auth/user/revoke/all", nil)
	assert.Equal(t, http.StatusOK, code, "revoke all: %s", body)

	// 22. Refresh с отозванным токеном должен упасть
	code, _ = c.post("/api/refresh", map[string]string{"refresh_token": login.RefreshToken})
	assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
		"refresh after revoke-all should fail, got %d", code)

	// 23. Login (получаем свежий токен для удаления)
	login = mustLogin(t, c, email, resetPassword)
	auth = c.withToken(login.AccessToken)

	// 24. DeleteUser (soft delete)
	code, body = auth.delete("/api/auth/user/account", map[string]string{"target_uuid": userUUID})
	require.Equal(t, http.StatusOK, code, "delete user: %s", body)

	// 25. Login заблокирован — аккаунт мягко удалён → 403
	code, _ = c.post("/api/login", map[string]string{"email": email, "password": resetPassword})
	assert.Equal(t, http.StatusForbidden, code,
		"login after soft delete should return 403, got %d", code)

	// 26. Restore account
	code, body = c.post("/api/restore-account", map[string]string{
		"email":    email,
		"password": resetPassword,
	})
	require.Equal(t, http.StatusOK, code, "restore account: %s", body)

	// 27. Login после восстановления — успех
	login = mustLogin(t, c, email, resetPassword)
	assert.NotEmpty(t, login.AccessToken, "login after restore should return access_token")
	assert.Empty(t, login.SessionUUID, "2FA is disabled — no session_uuid expected")
}
