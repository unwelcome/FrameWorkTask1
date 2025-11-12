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

type HealthHandler interface {
	Health(c *fiber.Ctx) error
}

type healthHandler struct {
	AuthServiceClient auth_proto.AuthServiceClient
	operationIDKey    string
}

func NewHealthHandler(authServiceClient auth_proto.AuthServiceClient, operationIDKey string) HealthHandler {
	return &healthHandler{AuthServiceClient: authServiceClient, operationIDKey: operationIDKey}
}

// Health
//
//	@Summary      Health check
//	@Description  Get all services health status
//	@Tags         Health
//	@Produce      json
//	@Success      200  {object}  entities.HealthResponse
//	@Failure      500  {object}  Error.HttpError
//	@Router       /health [get]
func (h *healthHandler) Health(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Health запрос в auth сервис
	res, err := h.AuthServiceClient.Health(ctx, &auth_proto.HealthRequest{OperationId: operationID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(Error.HttpError{Code: 500, Message: err.Error()})
	}

	// Сборка ответа
	return c.Status(200).JSON(&entities.HealthResponse{
		Gateway:     "health",
		Auth:        res.GetHealth(),
		Application: "not implemented",
	})
}
