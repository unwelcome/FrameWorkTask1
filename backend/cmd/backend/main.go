package main

import (
	"backend/database/postgresql"
	"backend/internal/config"
	"backend/internal/entities"
	"backend/internal/repositories"
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	fmt.Println(cfg)

	// Подключаемся к БД
	db, err := postgresql.Connect(cfg)
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}
	defer db.Close()

	// Создаем репозиторий
	userRepo := repositories.NewUserRepository(db.DB)

	ctx := context.Background()

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		log.Println(c.Method(), "request:", c.IP(), "at", time.Now().Format("2006-01-02 15:04:05"))
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Go!")
	})
	app.Get("/download", func(c *fiber.Ctx) error {
		return c.Download("./main.go")
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		// Пример создания пользователя
		newUser := &entities.User{
			Login:        "test_login",
			PasswordHash: "123123123",
			PasswordSalt: "123",
			FirstName:    "Name",
			SecondName:   "SecondName",
			ThirdName:    "ThirdName",
			Email:        "john@example.com",
			Role:         "руководитель",
		}

		if err := userRepo.CreateUser(ctx, newUser); err != nil {
			log.Print("Failed to create user:", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		log.Printf("Created user with ID: %d", newUser.ID)

		return c.SendString(fmt.Sprintf("User successfully created, id: %d", newUser.ID))
	})

	app.Listen(":8080")

	// test 2
}
