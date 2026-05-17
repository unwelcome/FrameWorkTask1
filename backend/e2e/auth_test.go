package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Register ---

func TestRegister_HappyPath(t *testing.T) {
	c := newClient()
	code, body := c.post("/api/register", defaultUserPayload(randomEmail(), "Password123"))

	assert.Equal(t, http.StatusCreated, code)

	var resp registerResp
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.NotEmpty(t, resp.UserUUID)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	c := newClient()
	email := randomEmail()

	code, _ := c.post("/api/register", defaultUserPayload(email, "Password123"))
	require.Equal(t, http.StatusCreated, code)

	code, body := c.post("/api/register", defaultUserPayload(email, "Password123"))
	assert.Equal(t, http.StatusConflict, code, "second registration with same email should return 409 (body: %s)", body)
}

func TestRegister_InvalidEmail(t *testing.T) {
	c := newClient()
	code, body := c.post("/api/register", defaultUserPayload("not-an-email", "Password123"))

	assert.Equal(t, http.StatusBadRequest, code, "invalid email should return 400 (body: %s)", body)
}

func TestRegister_WeakPassword(t *testing.T) {
	c := newClient()
	code, body := c.post("/api/register", defaultUserPayload(randomEmail(), "short"))

	assert.Equal(t, http.StatusBadRequest, code, "weak password should return 400 (body: %s)", body)
}

func TestRegister_MissingFields(t *testing.T) {
	c := newClient()
	code, body := c.post("/api/register", map[string]string{
		"email":    randomEmail(),
		"password": "Password123",
		// first_name, last_name, patronymic omitted
	})

	assert.Equal(t, http.StatusBadRequest, code, "missing required fields should return 400 (body: %s)", body)
}

// --- Login ---

func TestLogin_HappyPath(t *testing.T) {
	c := newClient()
	email := randomEmail()
	mustRegister(t, c, email, "Password123")

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
}

func TestLogin_WrongPassword(t *testing.T) {
	c := newClient()
	email := randomEmail()
	mustRegister(t, c, email, "Password123")

	code, body := c.post("/api/login", map[string]string{
		"email":    email,
		"password": "WrongPassword1",
	})

	assert.True(t, code == http.StatusBadRequest, "wrong password should return 401 or 404, got %d (body: %s)", code, body)
}

func TestLogin_NonExistentUser(t *testing.T) {
	c := newClient()
	code, body := c.post("/api/login", map[string]string{
		"email":    randomEmail(),
		"password": "Password123",
	})

	assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
		"non-existent user should return 401 or 404, got %d (body: %s)", code, body)
}

func TestLogin_InvalidEmail(t *testing.T) {
	c := newClient()
	code, body := c.post("/api/login", map[string]string{
		"email":    "bad-email",
		"password": "Password123",
	})

	assert.Equal(t, http.StatusBadRequest, code, "invalid email format should return 400 (body: %s)", body)
}

// --- Refresh Token ---

func TestRefreshToken_HappyPath(t *testing.T) {
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
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	c := newClient()
	code, body := c.post("/api/refresh", map[string]string{
		"refresh_token": "this.is.not.a.valid.jwt",
	})

	assert.Equal(t, http.StatusBadRequest, code, "invalid token should return 400 (body: %s)", body)
}

// --- GetUser ---

func TestGetUser_HappyPath(t *testing.T) {
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
}

func TestGetUser_Unauthorized(t *testing.T) {
	c := newClient()
	_, login := mustRegisterAndLogin(t, c)

	// No token
	code, body := c.get("/api/auth/user/" + login.UserUUID + "/info")
	assert.Equal(t, http.StatusUnauthorized, code, "request without token should return 401 (body: %s)", body)
}

func TestGetUser_NotFound(t *testing.T) {
	c := newClient()
	_, login := mustRegisterAndLogin(t, c)
	auth := c.withToken(login.AccessToken)

	code, body := auth.get("/api/auth/user/00000000-0000-0000-0000-000000000001/info")
	assert.Equal(t, http.StatusNotFound, code, "unknown uuid should return 404 (body: %s)", body)
}

// --- UpdateUserBio ---

func TestUpdateUserBio_HappyPath(t *testing.T) {
	c := newClient()
	_, login := mustRegisterAndLogin(t, c)
	auth := c.withToken(login.AccessToken)

	code, body := auth.patch("/api/auth/user/bio", map[string]string{
		"first_name": "Petr",
		"last_name":  "Petrov",
		"patronymic": "Petrovich",
	})

	assert.Equal(t, http.StatusOK, code, "update bio failed (body: %s)", body)

	// Verify the change was applied
	code, body = auth.get("/api/auth/user/" + login.UserUUID + "/info")
	require.Equal(t, http.StatusOK, code)

	var user getUserResp
	require.NoError(t, json.Unmarshal(body, &user))
	assert.Equal(t, "Petr", user.FirstName)
	assert.Equal(t, "Petrov", user.LastName)
	assert.Equal(t, "Petrovich", user.Patronymic)
}

// --- ChangePassword ---

func TestChangePassword_HappyPath(t *testing.T) {
	c := newClient()
	email, login := mustRegisterAndLogin(t, c)
	auth := c.withToken(login.AccessToken)

	code, body := auth.patch("/api/auth/user/password", map[string]string{
		"password": "NewPassword456",
	})
	assert.Equal(t, http.StatusOK, code, "change password failed (body: %s)", body)

	// Can login with new password
	code, _ = c.post("/api/login", map[string]string{
		"email":    email,
		"password": "NewPassword456",
	})
	assert.Equal(t, http.StatusOK, code, "login with new password should succeed")

	// Cannot login with old password
	code, _ = c.post("/api/login", map[string]string{
		"email":    email,
		"password": "Password123",
	})
	assert.True(t, code == http.StatusBadRequest, "login with old password should fail after change")
}

// --- GetAllActiveTokens ---

func TestGetAllActiveTokens_HappyPath(t *testing.T) {
	c := newClient()
	_, login := mustRegisterAndLogin(t, c)
	auth := c.withToken(login.AccessToken)

	code, body := auth.get("/api/auth/user/tokens")

	assert.Equal(t, http.StatusOK, code)

	var resp tokensResp
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.GreaterOrEqual(t, len(resp.Tokens), 1, "should have at least 1 active token after login")
}

// --- RevokeToken ---

func TestRevokeToken_HappyPath(t *testing.T) {
	c := newClient()
	_, login := mustRegisterAndLogin(t, c)
	auth := c.withToken(login.AccessToken)

	code, body := auth.delete("/api/auth/user/revoke/token", map[string]string{
		"refresh_token": login.RefreshToken,
	})
	assert.Equal(t, http.StatusOK, code, "revoke token failed (body: %s)", body)

	// Refreshing with the revoked token should now fail
	code, _ = c.post("/api/refresh", map[string]string{
		"refresh_token": login.RefreshToken,
	})
	assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
		"refresh with revoked token should fail, got %d", code)
}

// --- RevokeAllTokens ---

func TestRevokeAllTokens_HappyPath(t *testing.T) {
	c := newClient()
	email, login := mustRegisterAndLogin(t, c)
	auth := c.withToken(login.AccessToken)

	// Create a second session
	login2 := mustLogin(t, c, email, "Password123")

	code, body := auth.delete("/api/auth/user/revoke/all", nil)
	assert.Equal(t, http.StatusOK, code, "revoke all tokens failed (body: %s)", body)

	// Both refresh tokens should now be invalid
	code, _ = c.post("/api/refresh", map[string]string{"refresh_token": login.RefreshToken})
	assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
		"first token should be revoked, got %d", code)

	code, _ = c.post("/api/refresh", map[string]string{"refresh_token": login2.RefreshToken})
	assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
		"second token should be revoked, got %d", code)
}

// --- DeleteUser ---

func TestDeleteUser_SelfDelete(t *testing.T) {
	c := newClient()
	_, login := mustRegisterAndLogin(t, c)
	auth := c.withToken(login.AccessToken)

	code, body := auth.delete("/api/auth/user/account", map[string]string{
		"target_uuid": login.UserUUID,
	})
	assert.Equal(t, http.StatusOK, code, "delete user failed (body: %s)", body)

	// User no longer exists
	code, _ = auth.get("/api/auth/user/" + login.UserUUID + "/info")
	assert.True(t, code == http.StatusNotFound || code == http.StatusUnauthorized,
		"deleted user should not be accessible, got %d", code)
}

// --- Full auth flow ---

func TestAuthFullFlow(t *testing.T) {
	c := newClient()
	email := randomEmail()

	// 1. Register
	regCode, regBody := c.post("/api/register", defaultUserPayload(email, "Password123"))
	require.Equal(t, http.StatusCreated, regCode, "register: %s", regBody)
	var reg registerResp
	require.NoError(t, json.Unmarshal(regBody, &reg))

	// 2. Login
	login := mustLogin(t, c, email, "Password123")
	assert.Equal(t, reg.UserUUID, login.UserUUID)
	auth := c.withToken(login.AccessToken)

	// 3. Get user info
	code, body := auth.get("/api/auth/user/" + login.UserUUID + "/info")
	require.Equal(t, http.StatusOK, code, "get user: %s", body)
	var user getUserResp
	require.NoError(t, json.Unmarshal(body, &user))
	assert.Equal(t, email, user.Email)

	// 4. Update bio
	code, body = auth.patch("/api/auth/user/bio", map[string]string{
		"first_name": "UpdatedFirstName",
		"last_name":  "UpdatedLastName",
		"patronymic": "UpdatedPatronymic",
	})
	assert.Equal(t, http.StatusOK, code, "update bio: %s", body)

	// 5. Refresh token — получаем новую пару токенов
	code, body = c.post("/api/refresh", map[string]string{
		"refresh_token": login.RefreshToken,
	})
	require.Equal(t, http.StatusOK, code, "refresh: %s", body)
	var newTokens refreshTokenResp
	require.NoError(t, json.Unmarshal(body, &newTokens))
	auth = c.withToken(newTokens.AccessToken)

	// 6. Get user info - Работаем с новым access token
	code, _ = auth.get("/api/auth/user/" + login.UserUUID + "/info")
	assert.Equal(t, http.StatusOK, code, "new access token should work")

	// 7. Get active tokens
	code, body = auth.get("/api/auth/user/tokens")
	assert.Equal(t, http.StatusOK, code)
	var activeTokens tokensResp
	require.NoError(t, json.Unmarshal(body, &activeTokens))

	// 8. Revoke all tokens
	code, body = auth.delete("/api/auth/user/revoke/all", nil)
	assert.Equal(t, http.StatusOK, code, "revoke all: %s", body)

	// 9. Refresh with the new refresh token should now fail
	code, _ = c.post("/api/refresh", map[string]string{
		"refresh_token": newTokens.RefreshToken,
	})
	assert.True(t, code == http.StatusUnauthorized || code == http.StatusNotFound,
		"refresh after revoke-all should fail, got %d", code)
}
