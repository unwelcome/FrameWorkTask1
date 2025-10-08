package utils

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
)

// Парсинг + валидация int параметра

func ValidateIntParams(c *fiber.Ctx, key string, minValue, maxValue int) (int, error) {

	// Парсим параметр
	param, err := c.ParamsInt(key)
	if err != nil {
		return 0, fmt.Errorf("missing %s", key)
	}

	// Проверяем минимальное значение параметра
	if param < minValue {
		return 0, fmt.Errorf("invalid %s", key)
	}

	// Проверяем максимальное значение параметра (если установлено)
	if maxValue != 0 && param > maxValue {
		return 0, fmt.Errorf("invalid %s", key)
	}

	return param, nil
}
