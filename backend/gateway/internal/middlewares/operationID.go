package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func NewOperationIDMiddleware(operationIDKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		operationID := uuid.NewString()
		c.Locals(operationIDKey, operationID)

		return c.Next()
	}
}
