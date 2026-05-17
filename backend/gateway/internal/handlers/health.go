package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/application/generated"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/entities"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
	"golang.org/x/sync/errgroup"
)

type HealthHandler interface {
	Health(c *fiber.Ctx) error
}

type healthHandler struct {
	AuthServiceClient        auth_proto.AuthServiceClient
	CompanyServiceClient     company_proto.CompanyServiceClient
	ApplicationServiceClient application_proto.ApplicationServiceClient
	operationIDKey           string
}

func NewHealthHandler(authServiceClient auth_proto.AuthServiceClient, companyServiceClient company_proto.CompanyServiceClient, applicationServiceClient application_proto.ApplicationServiceClient, operationIDKey string) HealthHandler {
	return &healthHandler{
		AuthServiceClient:        authServiceClient,
		CompanyServiceClient:     companyServiceClient,
		ApplicationServiceClient: applicationServiceClient,
		operationIDKey:           operationIDKey,
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
	var authHealth, companyHealth, applicationHealth entities.ServiceHealth

	// Auth health check
	g.Go(func() error {
		res, err := h.AuthServiceClient.Health(ctx, &auth_proto.HealthRequest{OperationId: operationID})
		if err != nil {
			authHealth = entities.ServiceHealth{Service: "unhealthy"}
			return fmt.Errorf("auth service: %w", err)
		}
		authHealth = entities.ServiceHealth{
			Service:  res.GetService(),
			Postgres: res.GetPostgres(),
			Redis:    res.GetRedis(),
			Minio:    res.GetMinio(),
			Mongo:    res.GetMongo(),
		}
		return nil
	})

	// Company health check
	g.Go(func() error {
		res, err := h.CompanyServiceClient.Health(ctx, &company_proto.HealthRequest{OperationId: operationID})
		if err != nil {
			companyHealth = entities.ServiceHealth{Service: "unhealthy"}
			return fmt.Errorf("company service: %w", err)
		}
		companyHealth = entities.ServiceHealth{
			Service:  res.GetService(),
			Postgres: res.GetPostgres(),
			Redis:    res.GetRedis(),
			Minio:    res.GetMinio(),
			Mongo:    res.GetMongo(),
		}
		return nil
	})

	// Application health check
	g.Go(func() error {
		res, err := h.ApplicationServiceClient.Health(ctx, &application_proto.HealthRequest{OperationId: operationID})
		if err != nil {
			applicationHealth = entities.ServiceHealth{Service: "unhealthy"}
			return fmt.Errorf("application service: %w", err)
		}
		applicationHealth = entities.ServiceHealth{
			Service:  res.GetService(),
			Postgres: res.GetPostgres(),
			Redis:    res.GetRedis(),
			Minio:    res.GetMinio(),
			Mongo:    res.GetMongo(),
		}
		return nil
	})

	// Ждем завершения всех горутин
	_ = g.Wait()

	// Заполняем статус при таймауте
	if authHealth.Service == "" {
		authHealth = entities.ServiceHealth{Service: "timeout"}
	}
	if companyHealth.Service == "" {
		companyHealth = entities.ServiceHealth{Service: "timeout"}
	}
	if applicationHealth.Service == "" {
		applicationHealth = entities.ServiceHealth{Service: "timeout"}
	}

	// Сборка ответа
	return c.Status(200).JSON(&entities.HealthResponse{
		Gateway:     "healthy",
		Auth:        authHealth,
		Company:     companyHealth,
		Application: applicationHealth,
	})
}
