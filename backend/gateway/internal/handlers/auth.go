package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/errors"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/session"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/format"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/interceptors"
	"google.golang.org/grpc/metadata"
)

type AuthHandler interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
	ChangePassword(c *fiber.Ctx) error
	UpdateUserBio(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
	GetAllActiveSessions(c *fiber.Ctx) error
	RefreshToken(c *fiber.Ctx) error
	RevokeSession(c *fiber.Ctx) error
	RevokeAllSessions(c *fiber.Ctx) error
	VerifyAccount(c *fiber.Ctx) error
	ResendVerificationCode(c *fiber.Ctx) error
	GetVerificationToken(c *fiber.Ctx) error
	GetRecoveryCode(c *fiber.Ctx) error
	Get2FACode(c *fiber.Ctx) error
	ForgotPassword(c *fiber.Ctx) error
	ResetPassword(c *fiber.Ctx) error
	Verify2FA(c *fiber.Ctx) error
	UpdateUser2FA(c *fiber.Ctx) error
	RestoreAccount(c *fiber.Ctx) error
}

type authHandler struct {
	AuthServiceClient    auth_proto.AuthServiceClient
	CompanyServiceClient company_proto.CompanyServiceClient
	operationIDKey       string
	userUUIDKey          string
	session              *session.Provider
}

func NewAuthHandler(authServiceClient auth_proto.AuthServiceClient, companyServiceClient company_proto.CompanyServiceClient, operationIDKey, userUUIDKey string, sessionProvider *session.Provider) AuthHandler {
	return &authHandler{
		AuthServiceClient:    authServiceClient,
		CompanyServiceClient: companyServiceClient,
		operationIDKey:       operationIDKey,
		userUUIDKey:          userUUIDKey,
		session:              sessionProvider,
	}
}

// Register
//
//	@Summary      Register
//	@Description  Register new user
//	@Tags         User
//	@Accept 			json
//	@Produce 			json
//	@Param 				data body entities.RegisterRequest true "Данные пользователя"
//	@Success      201
//	@Failure      400  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /register [post]
func (h *authHandler) Register(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.RegisterRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.AuthServiceClient.Register(ctx, &auth_proto.RegisterRequest{
		Email:      httpReq.Email,
		Password:   httpReq.Password,
		FirstName:  httpReq.FirstName,
		LastName:   httpReq.LastName,
		Patronymic: httpReq.Patronymic,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.SendStatus(fiber.StatusCreated)
}

// Login
//
//	@Summary      Login
//	@Description  Авторизация пользователя. Если включена 2FA - возвращает session_uuid и отправляет письмо на почту, данные необходимо передать в Verify2FA; Если 2FA выключена - возвращает user_uuid и пару токенов
//	@Tags         Auth
//	@Accept 			json
//	@Produce 			json
//	@Param 				data body entities.LoginRequest true "Данные пользователя"
//	@Success      200  {object}  entities.LoginResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      403  {object}  Error.HttpError
//	@Failure      429  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /login [post]
func (h *authHandler) Login(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Парсит тело запроса
	httpReq := &entities.LoginRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.LoginRequest{
		Email:    httpReq.Email,
		Password: httpReq.Password,
		Session:  h.session.Extract(c),
	}

	// Запрос в auth сервис
	res, err := h.AuthServiceClient.Login(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.LoginResponse{
		UserUUID:     res.GetUserUuid(),
		AccessToken:  res.GetAccessToken(),
		RefreshToken: res.GetRefreshToken(),
		SessionUUID:  res.GetSessionUuid(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// GetUser
//
//	@Summary      GetUser
//	@Description  Get colleague's public profile. Only accessible if the requester and target share at least one company.
//	@Tags         User
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Param 				user_uuid path string true "User UUID"
//	@Success      200  {object}  entities.GetUserResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      403  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/{user_uuid}/info [get]
func (h *authHandler) GetUser(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Получаем UserUUID из параметров
	httpReq := &entities.GetUserRequest{
		UserUUID: c.Params("user_uuid", ""),
	}

	// Валидация
	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	requesterUUID := utils.GetLocal[string](c, h.userUUIDKey)

	// Проверяем, что requester и target — коллеги в одной компании
	// (пропускаем проверку если пользователь запрашивает свой собственный профиль)
	if requesterUUID != httpReq.UserUUID {
		colleagueRes, err := h.CompanyServiceClient.CheckColleagues(ctx, &company_proto.CheckColleaguesRequest{
			InitiatorUuid: requesterUUID,
			TargetUuid:    httpReq.UserUUID,
		})
		if err != nil {
			return Error.GRPCErrorToHTTP(err, c)
		}
		if !colleagueRes.GetAreColleagues() {
			return c.Status(fiber.StatusForbidden).JSON(Error.HttpError{Code: 403, Message: "access denied"})
		}
	}

	// Запрос в auth сервис
	res, err := h.AuthServiceClient.GetUser(ctx, &auth_proto.GetUserRequest{
		UserUuid: httpReq.UserUUID,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.GetUserResponse{
		UserUUID:    res.GetUserUuid(),
		FirstName:   res.GetFirstName(),
		LastName:    res.GetLastName(),
		Patronymic:  res.GetPatronymic(),
		Description: res.GetDescription(),
		CreatedAt:   res.GetCreatedAt(),
		DeletedAt:   res.GetDeletedAt(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// ChangePassword
//
//	@Summary      ChangePassword
//	@Description  Change user password (requires current password verification)
//	@Tags         User
//	@Accept 			json
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Param 				data body entities.ChangePasswordRequest true "Новый пароль"
//	@Success      200  {object}  entities.ChangePasswordResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      403  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/password [patch]
func (h *authHandler) ChangePassword(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Парсит тело запроса
	httpReq := &entities.ChangePasswordRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.ChangePasswordRequest{
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		OldPassword: httpReq.OldPassword,
		Password:    httpReq.Password,
	}

	// Запрос в auth сервис
	_, err = h.AuthServiceClient.ChangePassword(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.ChangePasswordResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// UpdateUserBio
//
//	@Summary      UpdateUserBio
//	@Description  Update user firstName, lastName, patronymic
//	@Tags         User
//	@Accept 			json
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Param 				data body entities.UpdateUserBioRequest true "ФИО пользователя"
//	@Success      200  {object}  entities.UpdateUserBioResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      403  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/bio [patch]
func (h *authHandler) UpdateUserBio(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Парсит тело запроса
	httpReq := &entities.UpdateUserBioRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.UpdateUserBioRequest{
		UserUuid:    utils.GetLocal[string](c, h.userUUIDKey),
		FirstName:   httpReq.FirstName,
		LastName:    httpReq.LastName,
		Patronymic:  httpReq.Patronymic,
		Description: httpReq.Description,
	}

	// Запрос в auth сервис
	_, err = h.AuthServiceClient.UpdateUserBio(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.UpdateUserBioResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// DeleteUser
//
//	@Summary      DeleteUser
//	@Description  Delete user
//	@Tags         Auth
//	@Accept 			json
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Success      200  {object}  entities.DeleteUserResponse
//	@Failure      401  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/account [delete]
func (h *authHandler) DeleteUser(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	req := &auth_proto.DeleteUserRequest{
		UserUuid: utils.GetLocal[string](c, h.userUUIDKey),
	}

	_, err := h.AuthServiceClient.DeleteUser(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.DeleteUserResponse{})
}

// GetAllActiveSessions
//
//	@Summary      GetAllActiveSessions
//	@Description  Get all active sessions
//	@Tags         Auth
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Success      200  {object}  entities.GetAllActiveSessionsResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/sessions [get]
func (h *authHandler) GetAllActiveSessions(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Формируем тело запроса
	req := &auth_proto.GetAllActiveSessionsRequest{
		UserUuid: utils.GetLocal[string](c, h.userUUIDKey),
	}

	// Запрос в auth сервис
	res, err := h.AuthServiceClient.GetAllActiveSessions(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	tokens := make([]*entities.TokenInfo, 0, len(res.GetTokens()))
	for _, token := range res.GetTokens() {
		s := token.GetSession()
		tokens = append(tokens, &entities.TokenInfo{
			SessionUUID:    token.GetSessionUuid(),
			IP:             s.GetIp(),
			LastIP:         s.GetLastIp(),
			ISP:            s.GetIsp(),
			CountryCode:    s.GetCountryCode(),
			CountryName:    s.GetCountryName(),
			City:           s.GetCity(),
			Timezone:       s.GetTimezone(),
			DeviceType:     s.GetDeviceType(),
			OS:             s.GetOs(),
			OSVersion:      s.GetOsVersion(),
			Browser:        s.GetBrowser(),
			BrowserVersion: s.GetBrowserVersion(),
			CreatedAt:      format.UnixTimestamp(s.GetCreatedAt()),
			LastActiveAt:   format.UnixTimestamp(s.GetLastActiveAt()),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&entities.GetAllActiveSessionsResponse{Tokens: tokens})
}

// RefreshToken
//
//	@Summary      RefreshToken
//	@Description  Refresh tokens
//	@Tags         Auth
//	@Produce 			json
//	@Param 				data body entities.RefreshTokenRequest true "Refresh token"
//	@Success      200  {object}  entities.RefreshTokenResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      403  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /refresh [post]
func (h *authHandler) RefreshToken(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Парсит тело запроса
	httpReq := &entities.RefreshTokenRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.RefreshTokenRequest{
		RefreshToken: httpReq.RefreshToken,
		Ip:           session.ClientIP(c),
	}

	// Запрос в auth сервис
	res, err := h.AuthServiceClient.RefreshToken(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.RefreshTokenResponse{
		AccessToken:  res.GetAccessToken(),
		RefreshToken: res.GetRefreshToken(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// RevokeSession
//
//	@Summary      RevokeSession
//	@Description  Revoke session by session UUID
//	@Tags         Auth
//	@Accept 			json
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Param 				data body entities.RevokeSessionRequest true "Refresh token hash to revoke"
//	@Success      200  {object}  entities.RevokeSessionResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/session [delete]
func (h *authHandler) RevokeSession(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Парсит тело запроса
	httpReq := &entities.RevokeSessionRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	userUUID := utils.GetLocal[string](c, h.userUUIDKey)

	// Формируем тело запроса
	req := &auth_proto.RevokeSessionRequest{
		UserUuid:    userUUID,
		SessionUuid: httpReq.SessionUUID,
	}

	// Запрос в auth сервис
	_, err = h.AuthServiceClient.RevokeSession(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.RevokeSessionResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// VerifyAccount
//
//	@Summary      VerifyAccount
//	@Description  Verify user account with code from email
//	@Tags         User
//	@Accept 			json
//	@Produce 			json
//	@Param 				data body entities.VerifyAccountRequest true "Email и код верификации"
//	@Success      200  {object}  entities.VerifyAccountResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      409  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /user/verify [post]
func (h *authHandler) VerifyAccount(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.VerifyAccountRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.VerifyAccountRequest{
		VerificationToken: httpReq.VerificationToken,
	}

	_, err := h.AuthServiceClient.VerifyAccount(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	return c.Status(fiber.StatusOK).JSON(&entities.VerifyAccountResponse{})
}

// ResendVerificationCode
//
//	@Summary      ResendVerificationCode
//	@Description  Resend verification code to user's email
//	@Tags         User
//	@Produce 			json
//	@Param 				data body entities.ResendVerificationCodeRequest true "Email пользователя"
//	@Success      200  {object}  entities.ResendVerificationCodeResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /user/verify/resend [post]
func (h *authHandler) ResendVerificationCode(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.ResendVerificationCodeRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.ResendVerificationCodeRequest{
		Email: httpReq.Email,
	}

	_, err := h.AuthServiceClient.ResendVerificationCode(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	return c.Status(fiber.StatusOK).JSON(&entities.ResendVerificationCodeResponse{})
}

// GetVerificationToken
//
//	@Summary      GetVerificationToken
//	@Description  Debug endpoint: generates and returns a verification token by email. Available only when APP_ENV=test.
//	@Tags         Debug
//	@Produce 			json
//	@Param 				email path string true "Email"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      501  {object}  Error.HttpError
//	@Router       /debug/user/email/{email}/verification-token [get]
func (h *authHandler) GetVerificationToken(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	res, err := h.AuthServiceClient.GetVerificationToken(ctx, &auth_proto.GetVerificationTokenRequest{
		Email: c.Params("email", ""),
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": res.GetToken()})
}

// GetRecoveryCode
//
//	@Summary      GetResetPasswordToken
//	@Description  Debug endpoint: generates and returns a reset password token by email. Available only when APP_ENV=test.
//	@Tags         Debug
//	@Produce 			json
//	@Param 				email path string true "Email"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      501  {object}  Error.HttpError
//	@Router       /debug/user/email/{email}/reset-password-token [get]
func (h *authHandler) GetRecoveryCode(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	res, err := h.AuthServiceClient.GetResetPasswordToken(ctx, &auth_proto.GetResetPasswordTokenRequest{
		Email: c.Params("email", ""),
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": res.GetToken()})
}

// Get2FACode
//
//	@Summary      Get2FACode
//	@Description  Debug endpoint: returns the active 2FA code by session UUID. Available only when APP_ENV=test.
//	@Tags         Debug
//	@Produce 			json
//	@Param 				session_uuid path string true "Session UUID"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      501  {object}  Error.HttpError
//	@Router       /debug/2fa/{session_uuid}/code [get]
func (h *authHandler) Get2FACode(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	req := &auth_proto.Get2FACodeRequest{
		SessionUuid: c.Params("session_uuid", ""),
	}

	res, err := h.AuthServiceClient.Get2FACode(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"code": res.GetCode()})
}

// ForgotPassword
//
//	@Summary      ForgotPassword
//	@Description  Request password recovery — sends a one-time code to the user's email
//	@Tags         User
//	@Accept 			json
//	@Produce 			json
//	@Param 				data body entities.ForgotPasswordRequest true "Email пользователя"
//	@Success      200  {object}  entities.ForgotPasswordResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /forgot-password [post]
func (h *authHandler) ForgotPassword(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.ForgotPasswordRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.AuthServiceClient.ForgotPassword(ctx, &auth_proto.ForgotPasswordRequest{
		Email: httpReq.Email,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.ForgotPasswordResponse{})
}

// ResetPassword
//
//	@Summary      ResetPassword
//	@Description  Reset password using a one-time JWT reset token from email
//	@Tags         User
//	@Accept 			json
//	@Produce 			json
//	@Param 				data body entities.ResetPasswordRequest true "JWT токен и новый пароль"
//	@Success      200  {object}  entities.ResetPasswordResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /reset-password [post]
func (h *authHandler) ResetPassword(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.ResetPasswordRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.AuthServiceClient.ResetPassword(ctx, &auth_proto.ResetPasswordRequest{
		ResetToken:  httpReq.ResetToken,
		NewPassword: httpReq.NewPassword,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.ResetPasswordResponse{})
}

// RevokeAllSessions
//
//	@Summary      RevokeAllSessions
//	@Description  Revoke all user sessions
//	@Tags         Auth
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Success      200  {object}  entities.RevokeAllSessionsResponse
//	@Failure      401  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/sessions [delete]
func (h *authHandler) RevokeAllSessions(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Формируем тело запроса
	req := &auth_proto.RevokeAllSessionsRequest{
		UserUuid: utils.GetLocal[string](c, h.userUUIDKey),
	}

	// Запрос в auth сервис
	_, err := h.AuthServiceClient.RevokeAllSessions(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.RevokeAllSessionsResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// Verify2FA
//
//	@Summary      Verify2FA
//	@Description  Verification confirmation with 2FA enabled
//	@Tags         Auth
//	@Produce 			json
//	@Param 				data body entities.Verify2FARequest true "SessionUUID and email code"
//	@Success      200  {object}  entities.Verify2FAResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      429  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /verify-2fa [post]
func (h *authHandler) Verify2FA(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.Verify2FARequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	res, err := h.AuthServiceClient.Verify2FA(ctx, &auth_proto.Verify2FARequest{
		SessionUuid: httpReq.SessionUUID,
		Code:        httpReq.Code,
		Session:     h.session.Extract(c),
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.Verify2FAResponse{
		UserUUID:     res.UserUuid,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

// UpdateUser2FA
//
//	@Summary      UpdateUser2FA
//	@Description  Enable / disable 2FA
//	@Tags         User
//	@Produce 			json
//	@Param 				data body entities.UpdateUser2FARequest true "Body"
//	@Success      200  {object}  entities.UpdateUser2FAResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      403  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/2fa [patch]
func (h *authHandler) UpdateUser2FA(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.UpdateUser2FARequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	httpReq.UserUUID = utils.GetLocal[string](c, h.userUUIDKey)
	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	_, err := h.AuthServiceClient.UpdateUser2FA(ctx, &auth_proto.UpdateUser2FARequest{
		UserUuid:   httpReq.UserUUID,
		Enable_2Fa: httpReq.Enable2FA,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.UpdateUser2FAResponse{})
}

// RestoreAccount
//
//	@Summary      RestoreAccount
//	@Description  Restore a soft-deleted account within 30 days of deletion
//	@Tags         User
//	@Accept       json
//	@Produce      json
//	@Param        data body entities.RestoreAccountRequest true "Email и пароль"
//	@Success      200  {object}  entities.RestoreAccountResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      403  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /restore-account [post]
func (h *authHandler) RestoreAccount(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	httpReq := &entities.RestoreAccountRequest{}
	if err := c.BodyParser(httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}
	if err := httpReq.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	_, err := h.AuthServiceClient.RestoreAccount(ctx, &auth_proto.RestoreAccountRequest{
		Email:    httpReq.Email,
		Password: httpReq.Password,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.RestoreAccountResponse{})
}
