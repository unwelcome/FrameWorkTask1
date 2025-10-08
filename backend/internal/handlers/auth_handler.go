package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/unwelcome/iqjtest/internal/entities"
	"github.com/unwelcome/iqjtest/internal/services"
)

type AuthHandler interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}

type authHandlerImpl struct {
	authService    services.AuthService
	requestTimeout time.Duration
}

func NewAuthHandler(authService services.AuthService, requestTimeout time.Duration) AuthHandler {
	return &authHandlerImpl{authService: authService, requestTimeout: requestTimeout}
}

// Register
// @Summary Создание пользователя
// @Description Создает нового пользователя в системе и возвращает access и refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Param user body entities.UserCreateRequest true "Данные пользователя"
// @Success 201 {object} entities.AuthResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /register [post]
func (h *authHandlerImpl) Register(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Парсим данные из тела запроса
	userCreateRequest := &entities.UserCreateRequest{}
	if err := c.BodyParser(&userCreateRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Регистрируем пользователя и получаем токены
	authResponse, err := h.authService.RegistrationUser(ctx, userCreateRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(authResponse)
}

// Login
// @Summary Вход в аккаунт пользователя
// @Description Вход в аккаунт пользователя, возвращает access и refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Param user body entities.UserLoginRequest true "Данные пользователя"
// @Success 200 {object} entities.AuthResponse
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /login [post]
func (h *authHandlerImpl) Login(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Парсим тело запроса в структуру
	userLoginRequest := &entities.UserLoginRequest{}
	if err := c.BodyParser(&userLoginRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Авторизуем пользователя
	authResponse, err := h.authService.LoginUser(ctx, userLoginRequest)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(authResponse)
}

// Refresh
// @Summary Обновление токенов
// @Description Обновляет access и refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Param token body entities.RefreshTokenRequest true "Refresh токен"
// @Success 201 {object} entities.TokenPair
// @Failure 400 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /refresh [post]
func (h *authHandlerImpl) Refresh(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	// Получаем refresh токен из тела
	refreshTokenRequest := &entities.RefreshTokenRequest{}
	if err := c.BodyParser(&refreshTokenRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	// Обновляем токены
	tokenPair, err := h.authService.RefreshToken(ctx, refreshTokenRequest.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(tokenPair)
}

// Logout
// @Summary Удаление токена
// @Description Удаляет refresh токен
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param token body entities.LogoutTokenRequest true "Refresh токен"
// @Success 200 {object} string
// @Failure 400 {object} entities.ErrorResponse
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/logout [delete]
func (h *authHandlerImpl) Logout(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	userID := c.Locals("userID").(int)

	// Получаем refresh токен из тела
	logoutTokenRequest := &entities.LogoutTokenRequest{}
	if err := c.BodyParser(&logoutTokenRequest); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid input"})
	}

	// Удаляем refresh токен
	err := h.authService.DeleteRefreshToken(ctx, userID, logoutTokenRequest.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).SendString("Successfully logged out")
}

// DeleteUser
// @Summary Удаление пользователя
// @Description Удаляет пользователя из системы и отзывает все refresh токены
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} string
// @Failure 401 {object} entities.ErrorResponse
// @Failure 500 {object} entities.ErrorResponse
// @Router /auth/user/delete [delete]
func (h *authHandlerImpl) DeleteUser(c *fiber.Ctx) error {

	// Ограничение времени выполнения
	ctx, cancel := context.WithTimeout(context.Background(), h.requestTimeout)
	defer cancel()

	userID := c.Locals("userID").(int)

	// Удаляем пользователя
	err := h.authService.DeleteUser(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).SendString("Successfully deleted user")
}
