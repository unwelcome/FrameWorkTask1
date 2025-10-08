package handlers

import "github.com/gofiber/fiber/v2"

type HealthHandler interface {
	Health(c *fiber.Ctx) error
}

type healthHandlerImpl struct {
}

func NewHealthHandler() HealthHandler {
	return &healthHandlerImpl{}
}

func (h *healthHandlerImpl) Health(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).SendString("Healthy")
}
