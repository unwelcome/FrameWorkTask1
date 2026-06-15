package Error

import "github.com/gofiber/fiber/v2"

type HttpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Unauthorized отвечает 401 и ставит заголовок WWW-Authenticate (RFC 7235),
// сообщающий клиенту требуемую схему аутентификации — Bearer (RFC 6750).
func Unauthorized(c *fiber.Ctx, message string) error {
	c.Set(fiber.HeaderWWWAuthenticate, "Bearer")
	return c.Status(fiber.StatusUnauthorized).JSON(HttpError{Code: fiber.StatusUnauthorized, Message: message})
}
