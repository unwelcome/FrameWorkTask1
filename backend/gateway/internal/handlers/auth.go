package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/errors"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
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
	GetAllActiveTokens(c *fiber.Ctx) error
	RefreshToken(c *fiber.Ctx) error
	RevokeToken(c *fiber.Ctx) error
	RevokeAllTokens(c *fiber.Ctx) error
	VerifyAccount(c *fiber.Ctx) error
	ResendVerificationCode(c *fiber.Ctx) error
	GetVerificationCode(c *fiber.Ctx) error
	GetRecoveryCode(c *fiber.Ctx) error
	Get2FACode(c *fiber.Ctx) error
	ForgotPassword(c *fiber.Ctx) error
	ResetPassword(c *fiber.Ctx) error
	Verify2FA(c *fiber.Ctx) error
	UpdateUser2FA(c *fiber.Ctx) error
}

type authHandler struct {
	AuthServiceClient auth_proto.AuthServiceClient
	operationIDKey    string
	userUUIDKey       string
}

func NewAuthHandler(authServiceClient auth_proto.AuthServiceClient, operationIDKey, userUUIDKey string) AuthHandler {
	return &authHandler{AuthServiceClient: authServiceClient, operationIDKey: operationIDKey, userUUIDKey: userUUIDKey}
}

// Register
//
//	@Summary      Register
//	@Description  Register new user
//	@Tags         User
//	@Accept 			json
//	@Produce 			json
//	@Param 				data body entities.RegisterRequest true "Данные пользователя"
//	@Success      201  {object}  entities.RegisterResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      409  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /register [post]
func (h *authHandler) Register(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Парсит тело запроса
	httpReq := &entities.RegisterRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.RegisterRequest{
		Email:      httpReq.Email,
		Password:   httpReq.Password,
		FirstName:  httpReq.FirstName,
		LastName:   httpReq.LastName,
		Patronymic: httpReq.Patronymic,
	}

	// Запрос в auth сервис
	res, err := h.AuthServiceClient.Register(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.RegisterResponse{
		UserUUID: res.GetUserUuid(),
	}

	return c.Status(fiber.StatusCreated).JSON(httpRes)
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
//	@Failure      404  {object}  Error.HttpError
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
//	@Description  Get user info
//	@Tags         User
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Param 				user_uuid path string true "User UUID"
//	@Success      200  {object}  entities.GetUserResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
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
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.GetUserRequest{
		UserUuid: httpReq.UserUUID,
	}

	// Запрос в auth сервис
	res, err := h.AuthServiceClient.GetUser(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.GetUserResponse{
		UserUUID:   res.GetUserUuid(),
		Email:      res.GetEmail(),
		FirstName:  res.GetFirstName(),
		LastName:   res.GetLastName(),
		Patronymic: res.GetPatronymic(),
		CreatedAt:  res.GetCreatedAt(),
	}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// ChangePassword
//
//	@Summary      ChangePassword
//	@Description  Change user password
//	@Tags         User
//	@Accept 			json
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Param 				data body entities.ChangePasswordRequest true "Новый пароль"
//	@Success      200  {object}  entities.ChangePasswordResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
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
		UserUuid: utils.GetLocal[string](c, h.userUUIDKey),
		Password: httpReq.Password,
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
		UserUuid:   utils.GetLocal[string](c, h.userUUIDKey),
		FirstName:  httpReq.FirstName,
		LastName:   httpReq.LastName,
		Patronymic: httpReq.Patronymic,
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
//	@Param 				data body entities.DeleteUserRequest true "Target UUID"
//	@Success      200  {object}  entities.DeleteUserResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/account [delete]
func (h *authHandler) DeleteUser(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Парсит тело запроса
	httpReq := &entities.DeleteUserRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.DeleteUserRequest{
		InitiatorUserUuid: utils.GetLocal[string](c, h.userUUIDKey),
		TargetUserUuid:    httpReq.TargetUUID,
	}

	// Запрос в auth сервис
	_, err = h.AuthServiceClient.DeleteUser(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.DeleteUserResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}

// GetAllActiveTokens
//
//	@Summary      GetAllActiveTokens
//	@Description  Get all active refresh tokens
//	@Tags         Auth
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Success      200  {object}  entities.GetAllActiveTokensResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/tokens [get]
func (h *authHandler) GetAllActiveTokens(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Формируем тело запроса
	req := &auth_proto.GetAllActiveTokensRequest{
		UserUuid: utils.GetLocal[string](c, h.userUUIDKey),
	}

	// Запрос в auth сервис
	res, err := h.AuthServiceClient.GetAllActiveTokens(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.GetAllActiveTokensResponse{}
	tokens := make([]*entities.TokenInfo, 0)

	for _, token := range res.GetTokens() {
		tokens = append(tokens, &entities.TokenInfo{
			Token: token.GetToken(),
		})
	}

	httpRes.Tokens = tokens

	return c.Status(fiber.StatusOK).JSON(httpRes)
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

// RevokeToken
//
//	@Summary      RevokeToken
//	@Description  Revoke refresh token
//	@Tags         Auth
//	@Accept 			json
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Param 				data body entities.RevokeTokenRequest true "Refresh token for revoke"
//	@Success      200  {object}  entities.RevokeTokenResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/revoke/token [delete]
func (h *authHandler) RevokeToken(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Парсит тело запроса
	httpReq := &entities.RevokeTokenRequest{}
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
	req := &auth_proto.RevokeTokenRequest{
		UserUuid:  userUUID,
		TokenHash: httpReq.TokenHash,
	}

	// Запрос в auth сервис
	_, err = h.AuthServiceClient.RevokeToken(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.RevokeTokenResponse{}

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
//	@Failure      404  {object}  Error.HttpError
//	@Failure      409  {object}  Error.HttpError
//	@Failure      429  {object}  Error.HttpError
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
		Email: httpReq.Email,
		Code:  httpReq.Code,
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
//	@Failure      409  {object}  Error.HttpError
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

// GetVerificationCode
//
//	@Summary      GetVerificationCode
//	@Description  Debug endpoint: returns the active verification code. Available only when APP_ENV=test.
//	@Tags         Debug
//	@Produce 			json
//	@Param 				user_uuid path string true "User UUID"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      501  {object}  Error.HttpError
//	@Router       /debug/user/{user_uuid}/verification-code [get]
func (h *authHandler) GetVerificationCode(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	req := &auth_proto.GetVerificationCodeRequest{
		UserUuid: c.Params("user_uuid", ""),
	}

	res, err := h.AuthServiceClient.GetVerificationCode(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"code": res.GetCode()})
}

// GetRecoveryCode
//
//	@Summary      GetRecoveryCode
//	@Description  Debug endpoint: returns the active recovery code. Available only when APP_ENV=test.
//	@Tags         Debug
//	@Produce 			json
//	@Param 				user_uuid path string true "User UUID"
//	@Success      200  {object}  map[string]string
//	@Failure      400  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      501  {object}  Error.HttpError
//	@Router       /debug/user/{user_uuid}/recovery-code [get]
func (h *authHandler) GetRecoveryCode(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	req := &auth_proto.GetRecoveryCodeRequest{
		UserUuid: c.Params("user_uuid", ""),
	}

	res, err := h.AuthServiceClient.GetRecoveryCode(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"code": res.GetCode()})
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
//	@Description  Reset password using a recovery code from email
//	@Tags         User
//	@Accept 			json
//	@Produce 			json
//	@Param 				data body entities.ResetPasswordRequest true "Email, код и новый пароль"
//	@Success      200  {object}  entities.ResetPasswordResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      429  {object}  Error.HttpError
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
		Email:       httpReq.Email,
		Code:        httpReq.Code,
		NewPassword: httpReq.NewPassword,
	})
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	return c.Status(fiber.StatusOK).JSON(&entities.ResetPasswordResponse{})
}

// RevokeAllTokens
//
//	@Summary      RevokeAllTokens
//	@Description  Revoke all user refresh tokens
//	@Tags         Auth
//	@Produce 			json
//	@Security 		ApiKeyAuth
//	@Success      200  {object}  entities.RevokeAllTokensResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      401  {object}  Error.HttpError
//	@Failure      404  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /auth/user/revoke/all [delete]
func (h *authHandler) RevokeAllTokens(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(interceptors.OperationIDMetaKey, operationID))

	// Формируем тело запроса
	req := &auth_proto.RevokeAllTokensRequest{
		UserUuid: utils.GetLocal[string](c, h.userUUIDKey),
	}

	// Запрос в auth сервис
	_, err := h.AuthServiceClient.RevokeAllTokens(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.RevokeAllTokensResponse{}

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
