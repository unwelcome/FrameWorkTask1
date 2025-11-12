package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewRequestLoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Получаем id запроса
		operationID := c.Locals("operationID")
		if operationID == nil {
			operationID = ""
		}

		startTime := time.Now()
		err := c.Next()

		logLevel := zerolog.InfoLevel
		if time.Since(startTime) > time.Second*2 {
			logLevel = zerolog.WarnLevel
		}

		log.WithLevel(logLevel).
			Str("id", operationID.(string)).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("duration", int(time.Since(startTime).Milliseconds())).
			Int("status", c.Response().StatusCode()).
			Msg("request")

		return err
	}
}
