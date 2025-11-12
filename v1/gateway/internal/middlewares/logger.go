package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/pkg/utils"
)

func NewRequestLoggerMiddleware(operationIDKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Получаем id запроса
		operationID := utils.GetLocal[string](c, operationIDKey)

		startTime := time.Now()
		err := c.Next()

		logLevel := zerolog.InfoLevel
		if time.Since(startTime) > time.Second*2 {
			logLevel = zerolog.WarnLevel
		}

		log.WithLevel(logLevel).
			Str("id", operationID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("duration", int(time.Since(startTime).Milliseconds())).
			Int("status", c.Response().StatusCode()).
			Msg("request")

		return err
	}
}
