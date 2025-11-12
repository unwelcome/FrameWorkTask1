package middlewares

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	Error "github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/errors"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"
)

func NewAuthMiddleware(secretKey string, userUUIDKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Получаем заголовок авторизации
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(Error.HttpError{Code: 401, Message: "authorization header required"})
		}

		// Проверяем корректность заголовка
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(fiber.StatusUnauthorized).JSON(Error.HttpError{Code: 401, Message: "invalid authorization header format"})
		}

		// Получаем токен из заголовка
		accessToken := authHeader[7:]

		// Парсим токен
		tokenClaims, err := utils.ParseToken(accessToken, secretKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(Error.HttpError{Code: 401, Message: fmt.Errorf("parse token error: %w", err).Error()})
		}

		// Проверяем тип токена
		if tokenClaims.TokenType != utils.AccessTokenType {
			return c.Status(fiber.StatusUnauthorized).JSON(Error.HttpError{Code: 401, Message: "invalid token type"})
		}

		// Устанавливаем UserUUID в контекст
		c.Locals(userUUIDKey, tokenClaims.UserUUID)

		return c.Next()
	}
}
