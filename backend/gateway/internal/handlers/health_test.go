package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/application/api/generated"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/auth/api/generated"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/company/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/entities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// newHealthApp создаёт Fiber-приложение с маршрутом /health и middleware для locals
func newHealthApp(h HealthHandler) *fiber.App {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(opKey, opID)
		return c.Next()
	})
	app.Get("/health", h.Health)
	return app
}

// ─── Health ───────────────────────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	// Успешный клиент auth-сервиса
	healthyAuth := &mockAuthClient{
		health: func(_ context.Context, _ *auth_proto.HealthRequest, _ ...grpc.CallOption) (*auth_proto.HealthResponse, error) {
			return &auth_proto.HealthResponse{Health: "healthy"}, nil
		},
	}
	// Успешный клиент company-сервиса
	healthyCompany := &mockCompanyClient{
		health: func(_ context.Context, _ *company_proto.HealthRequest, _ ...grpc.CallOption) (*company_proto.HealthResponse, error) {
			return &company_proto.HealthResponse{Health: "healthy"}, nil
		},
	}
	// Успешный клиент application-сервиса
	healthyApp := &mockApplicationClient{
		health: func(_ context.Context, _ *application_proto.HealthRequest, _ ...grpc.CallOption) (*application_proto.HealthResponse, error) {
			return &application_proto.HealthResponse{Health: "healthy"}, nil
		},
	}

	t.Run("all services healthy", func(t *testing.T) {
		h := newHealthHandler(healthyAuth, healthyCompany, healthyApp)
		resp, err := newHealthApp(h).Test(httptest.NewRequest("GET", "/health", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}

		assertStatus(t, resp, fiber.StatusOK)

		var res entities.HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if res.Gateway != "healthy" {
			t.Errorf("gateway: expected %q, got %q", "healthy", res.Gateway)
		}
		if res.Auth != "healthy" {
			t.Errorf("auth: expected %q, got %q", "healthy", res.Auth)
		}
		if res.Company != "healthy" {
			t.Errorf("company: expected %q, got %q", "healthy", res.Company)
		}
		if res.Application != "healthy" {
			t.Errorf("application: expected %q, got %q", "healthy", res.Application)
		}
	})

	t.Run("auth service error - marked unhealthy", func(t *testing.T) {
		failAuth := &mockAuthClient{
			health: func(_ context.Context, _ *auth_proto.HealthRequest, _ ...grpc.CallOption) (*auth_proto.HealthResponse, error) {
				return nil, grpcErr(codes.Unavailable)
			},
		}

		h := newHealthHandler(failAuth, healthyCompany, healthyApp)
		resp, err := newHealthApp(h).Test(httptest.NewRequest("GET", "/health", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}

		// Health endpoint всегда возвращает 200 — статусы пишутся в тело
		assertStatus(t, resp, fiber.StatusOK)

		var res entities.HealthResponse
		json.NewDecoder(resp.Body).Decode(&res)
		if res.Auth != "unhealthy" {
			t.Errorf("auth: expected %q, got %q", "unhealthy", res.Auth)
		}
		if res.Company != "healthy" {
			t.Errorf("company: expected %q, got %q", "healthy", res.Company)
		}
		if res.Application != "healthy" {
			t.Errorf("application: expected %q, got %q", "healthy", res.Application)
		}
	})

	t.Run("company service error - marked unhealthy", func(t *testing.T) {
		failCompany := &mockCompanyClient{
			health: func(_ context.Context, _ *company_proto.HealthRequest, _ ...grpc.CallOption) (*company_proto.HealthResponse, error) {
				return nil, grpcErr(codes.Internal)
			},
		}

		h := newHealthHandler(healthyAuth, failCompany, healthyApp)
		resp, err := newHealthApp(h).Test(httptest.NewRequest("GET", "/health", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}

		assertStatus(t, resp, fiber.StatusOK)

		var res entities.HealthResponse
		json.NewDecoder(resp.Body).Decode(&res)
		if res.Company != "unhealthy" {
			t.Errorf("company: expected %q, got %q", "unhealthy", res.Company)
		}
	})

	t.Run("all services fail - all marked unhealthy", func(t *testing.T) {
		failAuth := &mockAuthClient{
			health: func(_ context.Context, _ *auth_proto.HealthRequest, _ ...grpc.CallOption) (*auth_proto.HealthResponse, error) {
				return nil, grpcErr(codes.Internal)
			},
		}
		failCompany := &mockCompanyClient{
			health: func(_ context.Context, _ *company_proto.HealthRequest, _ ...grpc.CallOption) (*company_proto.HealthResponse, error) {
				return nil, grpcErr(codes.Internal)
			},
		}
		failApp := &mockApplicationClient{
			health: func(_ context.Context, _ *application_proto.HealthRequest, _ ...grpc.CallOption) (*application_proto.HealthResponse, error) {
				return nil, grpcErr(codes.Internal)
			},
		}

		h := newHealthHandler(failAuth, failCompany, failApp)
		resp, err := newHealthApp(h).Test(httptest.NewRequest("GET", "/health", nil))
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}

		assertStatus(t, resp, fiber.StatusOK)

		var res entities.HealthResponse
		json.NewDecoder(resp.Body).Decode(&res)
		if res.Gateway != "healthy" {
			t.Errorf("gateway should always be healthy, got %q", res.Gateway)
		}
		if res.Auth != "unhealthy" {
			t.Errorf("auth: expected %q, got %q", "unhealthy", res.Auth)
		}
		if res.Company != "unhealthy" {
			t.Errorf("company: expected %q, got %q", "unhealthy", res.Company)
		}
		if res.Application != "unhealthy" {
			t.Errorf("application: expected %q, got %q", "unhealthy", res.Application)
		}
	})
}
