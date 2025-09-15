package middlewares

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"time"
)

func RequestLog(l *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()

		err := c.Next()

		l.Info(
			"request",
			"method", c.Method(),
			"ip", c.IP(),
			"path", c.Path(),
			"duration", fmt.Sprintf("%dms", time.Since(startTime).Milliseconds()),
			"status", c.Response().StatusCode())

		return err
	}
}
