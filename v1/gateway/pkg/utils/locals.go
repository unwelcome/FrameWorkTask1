package utils

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// GetLocal Получение данных конкретного типа из fiber.Locals
func GetLocal[T any](c *fiber.Ctx, key string) T {
	value := c.Locals(key)
	if value == nil {
		var zero T
		return zero
	}

	v, ok := value.(T)
	if !ok {
		var zero T
		return zero
	}

	return v
}

// GetLocalOrError Получение данных конкретного типа из fiber.Locals с ошибкой
func GetLocalOrError[T any](c *fiber.Ctx, key string) (T, error) {
	value := c.Locals(key)
	if value == nil {
		var zero T
		return zero, fmt.Errorf("%s missed", key)
	}

	v, ok := value.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("%s is not of type %T (got %T)", key, zero, value)
	}

	return v, nil
}
