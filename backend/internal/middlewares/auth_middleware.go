package middlewares

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/unwelcome/iqjtest/internal/entities"
	"github.com/unwelcome/iqjtest/pkg/utils"
)

func AuthMiddleware(secretKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Получаем заголовок авторизации
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authorization header required"})
		}

		// Проверяем корректность заголовка
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization header format"})
		}

		// Получаем токен из заголовка
		accessToken := authHeader[7:]

		// Парсим токен
		tokenClaims, err := utils.ParseToken(accessToken, secretKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": fmt.Errorf("parse token error: %w", err).Error()})
		}

		// Проверяем тип токена
		if tokenClaims.Type != entities.AccessTokenType {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token type"})
		}

		// Устанавливаем userID в контекст
		c.Locals("userID", tokenClaims.UserID)

		return c.Next()
	}
}
