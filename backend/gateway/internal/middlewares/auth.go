package middlewares

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/gofiber/fiber/v2"
	Error "github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/errors"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
)

func NewAuthMiddleware(publicKey *ecdsa.PublicKey, userUUIDKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Получаем заголовок авторизации
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return Error.Unauthorized(c, "authorization header required")
		}

		// Проверяем корректность заголовка
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return Error.Unauthorized(c, "invalid authorization header format")
		}

		// Получаем токен из заголовка
		accessToken := authHeader[7:]

		// Парсим токен
		tokenClaims, err := utils.ParseToken(accessToken, publicKey)
		if err != nil {
			return Error.Unauthorized(c, fmt.Errorf("parse token error: %w", err).Error())
		}

		// Проверяем тип токена
		if tokenClaims.TokenType != utils.AccessTokenType {
			return Error.Unauthorized(c, "invalid token type")
		}

		// Устанавливаем UserUUID в контекст
		c.Locals(userUUIDKey, tokenClaims.UserUUID)

		return c.Next()
	}
}
