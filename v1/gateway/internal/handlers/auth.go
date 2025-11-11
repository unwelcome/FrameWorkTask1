package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	auth_proto "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/auth"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/entities"
	Error "github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/errors"
)

type AuthHandler interface {
	Register(c *fiber.Ctx) error
}

type authHandler struct {
	AuthServiceClient auth_proto.AuthServiceClient
}

func NewAuthHandler(authServiceClient auth_proto.AuthServiceClient) AuthHandler {
	return &authHandler{AuthServiceClient: authServiceClient}
}

// Register
//
//	@Summary      Register
//	@Description  Register new user
//	@Tags         Auth
//	@Accept 			json
//	@Produce 			json
//	@Param 				user body entities.RegisterRequest true "Данные пользователя"
//	@Success      201  {object}  entities.RegisterResponse
//	@Failure      400  {object}  Error.HttpError
//	@Failure      409  {object}  Error.HttpError
//	@Failure      500  {object}  Error.HttpError
//	@Router       /register [post]
func (h *authHandler) Register(c *fiber.Ctx) error {
	operationID := c.Locals("operationID").(string)

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
		UserUUID: res.UserUuid,
	}

	return c.Status(fiber.StatusCreated).JSON(httpRes)
}
