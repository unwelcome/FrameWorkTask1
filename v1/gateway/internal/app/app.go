package app

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/api/auth"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/config"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/handlers"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	// Services
	AuthServiceClient auth_proto.AuthServiceClient

	// Middlewares
	LoggerMiddleware fiber.Handler

	// Handlers
	HealthHandler handlers.HealthHandler
}

func InitApp(cfg *config.Config) *App {
	app := &App{}

	// Init clients to services
	app.AuthServiceClient = auth_proto.NewAuthServiceClient(getGRPCConnection(cfg.AuthService.Host, cfg.AuthService.Port))

	// Init middlewares
	app.LoggerMiddleware = logger.RequestLogger()

	// Init handlers
	app.HealthHandler = handlers.NewHealthHandler(app.AuthServiceClient)

	return app
}

func getGRPCConnection(host string, port int) *grpc.ClientConn {
	endpoint := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to connect to %s:%d", host, port)
	}
	//defer conn.Close()

	return conn
}
