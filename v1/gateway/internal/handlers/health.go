package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	auth_proto "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/auth"
	company_proto "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/company"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"
	"golang.org/x/sync/errgroup"
)

type HealthHandler interface {
	Health(c *fiber.Ctx) error
}

type healthHandler struct {
	AuthServiceClient    auth_proto.AuthServiceClient
	CompanyServiceClient company_proto.CompanyServiceClient
	operationIDKey       string
}

func NewHealthHandler(authServiceClient auth_proto.AuthServiceClient, companyServiceClient company_proto.CompanyServiceClient, operationIDKey string) HealthHandler {
	return &healthHandler{
		AuthServiceClient:    authServiceClient,
		CompanyServiceClient: companyServiceClient,
		operationIDKey:       operationIDKey,
	}
}

// Health
//
//	@Summary      Health check
//	@Description  Get all services health status
//	@Tags         Health
//	@Produce      json
//	@Success      200  {object}  entities.HealthResponse
//	@Router       /health [get]
func (h *healthHandler) Health(c *fiber.Ctx) error {
	operationID := utils.GetLocal[string](c, h.operationIDKey)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// Результаты
	var authHealth, companyHealth string

	// Auth health check
	g.Go(func() error {
		res, err := h.AuthServiceClient.Health(ctx, &auth_proto.HealthRequest{OperationId: operationID})
		if err != nil {
			authHealth = "unhealthy"
			return fmt.Errorf("auth service: %w", err)
		}
		authHealth = res.GetHealth()
		return nil
	})

	// Company health check
	g.Go(func() error {
		res, err := h.CompanyServiceClient.Health(ctx, &company_proto.HealthRequest{OperationId: operationID})
		if err != nil {
			companyHealth = "unhealthy"
			return fmt.Errorf("company service: %w", err)
		}
		companyHealth = res.GetHealth()
		return nil
	})

	// Ждем завершения всех горутин
	_ = g.Wait()

	// Заполняем неизвестные статусы
	if authHealth == "" {
		authHealth = "timeout"
	}
	if companyHealth == "" {
		companyHealth = "timeout"
	}

	// Сборка ответа
	return c.Status(200).JSON(&entities.HealthResponse{
		Gateway:     "healthy",
		Auth:        authHealth,
		Company:     companyHealth,
		Application: "not implemented",
	})
}
