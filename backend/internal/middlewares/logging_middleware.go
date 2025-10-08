package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"time"
)

func LoggingRequest(l zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()

		err := c.Next()

		l.Info().
			Str("ip", c.IP()).
			Str("method", c.Method()).
			Str("path", c.Path()).
			//RawJSON("body", c.Body()).
			Int("duration", int(time.Since(startTime).Milliseconds())).
			Int("status", c.Response().StatusCode()).
			Msg("request")

		return err
	}
}
