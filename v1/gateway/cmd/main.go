package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/config"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/logger"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	pb "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Инициализация конфига
	cfg := config.NewConfig()
	cfg.Print()

	// Инициализация логгера
	loggerConf := logger.Setup(cfg.App.LogPath, cfg.App.LogConsoleOut)
	log.Logger = *loggerConf

	app := fiber.New()

	conn, err := grpc.NewClient("auth_service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("Failed to connect to auth service")
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Gateway")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		authServiceHealth := "unknown"

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		authOperationUUID := uuid.NewString()
		fmt.Printf("Method: get Path: health\n")
		fmt.Printf("Service: auth UUID: %s Status: send\n", authOperationUUID)

		res, err := client.Health(ctx, &pb.HealthRequest{OperationId: authOperationUUID})
		if err != nil {
			fmt.Printf("Service: auth UUID: %s Status: error Error: %v", authOperationUUID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		fmt.Printf("Service: auth UUID: %s Status: success\n", authOperationUUID)

		authServiceHealth = res.Health

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"gateway": "healthy", "auth": authServiceHealth})
	})

	fmt.Printf("Server listen on http://%s", "localhost:8080")

	if err := app.Listen(":8080"); err != nil {
		fmt.Printf("Failed to start http server: %v", err)
		os.Exit(1)
	}
}
