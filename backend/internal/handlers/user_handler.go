package handlers

import (
	"backend/internal/entities"
	"backend/internal/services"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type UserHandler struct {
	userService *services.UserService
	logger      *slog.Logger
}

func NewUserHandler(userService *services.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{userService: userService, logger: logger}
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	createUserRequest := &entities.CreateUserRequest{}
	if err := c.BodyParser(&createUserRequest); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Получаем контекст из запроса Fiber
	ctx := c.Context()

	createUserResponse, err := h.userService.CreateUser(ctx, createUserRequest)
	if err != nil {
		h.logger.Warn("Failed to create user", "error", err.Error())
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(createUserResponse)
}
