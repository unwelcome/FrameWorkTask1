package main

import (
	"backend/database/postgresql"
	"backend/internal/config"
	"backend/internal/di"
	"backend/internal/routes"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"log/slog"
	"os"
)

func main() {
	// Создание логгера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Загрузка конфигурации
	cfg := config.LoadConfig(logger)

	// Подключение к БД
	postgres := postgresql.Connect(cfg, logger)
	defer postgres.Close()

	// Инициализация di
	container := di.NewContainer(postgres.DB, logger)

	// Создание сервера
	app := fiber.New()

	// Установка роутера
	routes.SetupRoutes(app, container)

	// Включение сервера
	if err := app.Listen(":8080"); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
