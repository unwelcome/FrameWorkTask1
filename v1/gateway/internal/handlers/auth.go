package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	auth_proto "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/auth"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/errors"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"
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
		OperationId: operationID,
		Email:       httpReq.Email,
		Password:    httpReq.Password,
		FirstName:   httpReq.FirstName,
		LastName:    httpReq.LastName,
		Patronymic:  httpReq.Patronymic,
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
//	@Description  Login user
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
		OperationId: operationID,
		Email:       httpReq.Email,
		Password:    httpReq.Password,
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
		OperationId: operationID,
		UserUuid:    httpReq.UserUUID,
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

	// Парсит тело запроса
	httpReq := &entities.ChangePasswordRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Получаем UserUUID из локал и собираем целиком тело запроса
	httpReqFull := &entities.ChangePasswordRequestFull{
		UserUUID: utils.GetLocal[string](c, h.userUUIDKey),
		Password: httpReq.Password,
	}

	// Валидация
	err := httpReqFull.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.ChangePasswordRequest{
		OperationId: operationID,
		UserUuid:    httpReqFull.UserUUID,
		Password:    httpReqFull.Password,
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

	// Парсит тело запроса
	httpReq := &entities.UpdateUserBioRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Получаем UserUUID из локал и собираем целиком тело запроса
	httpReqFull := &entities.UpdateUserBioRequestFull{
		UserUUID:   utils.GetLocal[string](c, h.userUUIDKey),
		FirstName:  httpReq.FirstName,
		LastName:   httpReq.LastName,
		Patronymic: httpReq.Patronymic,
	}

	// Валидация
	err := httpReqFull.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.UpdateUserBioRequest{
		OperationId: operationID,
		UserUuid:    httpReqFull.UserUUID,
		FirstName:   httpReqFull.FirstName,
		LastName:    httpReqFull.LastName,
		Patronymic:  httpReqFull.Patronymic,
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

	// Парсит тело запроса
	httpReq := &entities.DeleteUserRequest{}
	if err := c.BodyParser(&httpReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: "invalid input"})
	}

	// Получаем InitiatorUUID из локал и собираем целиком тело запроса
	httpReqFull := &entities.DeleteUserRequestFull{
		InitiatorUUID: utils.GetLocal[string](c, h.userUUIDKey),
		TargetUUID:    httpReq.TargetUUID,
	}

	// Валидация
	err := httpReqFull.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.DeleteUserRequest{
		OperationId:       operationID,
		InitiatorUserUuid: httpReqFull.InitiatorUUID,
		TargetUserUuid:    httpReqFull.TargetUUID,
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

	// Получаем UserUUID из локал
	httpReq := &entities.GetAllActiveTokensRequest{
		UserUUID: utils.GetLocal[string](c, h.userUUIDKey),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.GetAllActiveTokensRequest{
		OperationId: operationID,
		UserUuid:    httpReq.UserUUID,
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
		OperationId:  operationID,
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

	// Формируем тело запроса
	req := &auth_proto.RevokeTokenRequest{
		OperationId:  operationID,
		RefreshToken: httpReq.RefreshToken,
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

	// Получаем UserUUID из параметров
	httpReq := &entities.RevokeAllTokensRequest{
		UserUUID: utils.GetLocal[string](c, h.userUUIDKey),
	}

	// Валидация
	err := httpReq.Validate()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(Error.HttpError{Code: 400, Message: err.Error()})
	}

	// Формируем тело запроса
	req := &auth_proto.RevokeAllTokensRequest{
		OperationId: operationID,
		UserUuid:    httpReq.UserUUID,
	}

	// Запрос в auth сервис
	_, err = h.AuthServiceClient.RevokeAllTokens(ctx, req)
	if err != nil {
		return Error.GRPCErrorToHTTP(err, c)
	}

	// Формируем тело ответа
	httpRes := &entities.RevokeAllTokensResponse{}

	return c.Status(fiber.StatusOK).JSON(httpRes)
}
