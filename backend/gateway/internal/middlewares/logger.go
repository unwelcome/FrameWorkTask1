package middlewares

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
)

// NewRequestLoggerMiddleware logs every HTTP request after it is handled.
//
// Log level is determined by the response status code:
//   - 5xx → Error
//   - 4xx → Warn
//   - response time > 2 s → Warn
//   - otherwise → Info
//
// Fields written for every request: id, method, path, ip, status, duration.
// Optional fields:
//   - user  — user UUID from locals; omitted on unauthenticated routes.
//   - error — error message from the response body; included on 4xx / 5xx only.
func NewRequestLoggerMiddleware(operationIDKey, userUUIDKey string, httpLog zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		statusCode := c.Response().StatusCode()
		duration := time.Since(start).Milliseconds()

		level := zerolog.InfoLevel
		switch {
		case statusCode >= 500:
			level = zerolog.ErrorLevel
		case statusCode >= 400:
			level = zerolog.WarnLevel
		case duration > 2000:
			level = zerolog.WarnLevel
		}

		// Time is added as the first explicit field so it appears right after
		// "level" in the output. Using Timestamp() in the logger context would
		// place it at the end because zerolog runs context hooks after all
		// event fields.
		e := httpLog.WithLevel(level).
			Time("time", time.Now()).
			Str("id", utils.GetLocal[string](c, operationIDKey)).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Str("ip", c.IP()).
			Int("status", statusCode).
			Int64("duration", duration)

		// user UUID is only present on authenticated routes
		if userUUID := utils.GetLocal[string](c, userUUIDKey); userUUID != "" {
			e = e.Str("user", userUUID)
		}

		// on errors, extract the message field from our own JSON error body
		if statusCode >= 400 {
			var errBody struct {
				Message string `json:"message"`
			}
			if jsonErr := json.Unmarshal(c.Response().Body(), &errBody); jsonErr == nil && errBody.Message != "" {
				e = e.Str("error", errBody.Message)
			}
		}

		e.Msg("http")

		return err
	}
}
