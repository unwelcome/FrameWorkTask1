package middlewares

import (
	"context"
	"github.com/unwelcome/iqjtest/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/unwelcome/iqjtest/internal/services"
)

func CatOwnershipMiddleware(catService services.CatService, middlewareRequestTimeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Ограничение времени выполнения
		ctx, cansel := context.WithTimeout(context.Background(), middlewareRequestTimeout)
		defer cansel()

		// Получаем catID из параметров
		catID, err := utils.ValidateIntParams(c, "id", 1, 0)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		// Получаем userID
		userID := c.Locals("userID").(int)

		// Проверяем права на изменения кота
		hasRights, err := catService.CheckOwnershipRight(ctx, userID, catID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		} else if !hasRights {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not enough right for this operation"})
		}

		// Устанавливаем catID в Locals
		c.Locals("catID", catID)

		return c.Next()
	}
}
