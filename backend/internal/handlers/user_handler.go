package handlers

import (
	"context"
	"github.com/unwelcome/iqjtest/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/unwelcome/iqjtest/internal/entities"
	"github.com/unwelcome/iqjtest/internal/services"
)

type UserHandler interface {
	GetUserByID(c *fiber.Ctx) error
	GetAllUsers(c *fiber.Ctx) error
	UpdateUserPassword(c *fiber.Ctx) error
}

type userHandlerImpl struct {
	userService    services.UserService
	requestTimeout time.Duration
}

func NewUserHandler(userService services.UserService, requestTimeout time.Duration) UserHandler {
	return &userHandlerImpl{userService: userService, requestTimeout: requestTimeout}
}

// GetUserByID
// @Summary получение пользователя по ID
// @Description получаем пользователя по ID
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "User ID"
// @Success 200 {object} entities.UserGet
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 404 {object} entities.ErrorResponse
// @Router /auth/user/{id} [get]
func (h *userHandlerImpl) GetUserByID(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Получаем id из параметров
	userID, err := utils.ValidateIntParams(c, "id", 1, 0)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Получаем пользователя
	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// GetAllUsers
// @Summary получение всех пользователей
// @Description получение всех пользователей
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} []entities.UserGet
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/user/all [get]
func (h *userHandlerImpl) GetAllUsers(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Получаем всех пользователей
	users, err := h.userService.GetAllUsers(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(users)
}

// UpdateUserPassword
// @Summary обновление пароля пользователя
// @Description обновление пароля пользователя
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user body entities.UserUpdatePasswordRequest true "Данные пользователя"
// @Success 200 {object} entities.UserUpdatePasswordResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/user/password [patch]
func (h *userHandlerImpl) UpdateUserPassword(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Парсим тело запроса в структуру
	userUpdatePasswordRequest := &entities.UserUpdatePasswordRequest{}
	if err := c.BodyParser(&userUpdatePasswordRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	userID := c.Locals("userID").(int)

	// Обновляем пароль пользователя
	err := h.userService.UpdateUserPassword(ctx, userID, userUpdatePasswordRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(&entities.UserUpdatePasswordResponse{ID: userID})
}
