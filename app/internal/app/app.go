package app

import (
	"github.com/rs/zerolog"
	grpcapp "microapi/internal/app/grpc"
	"microapi/internal/config"
)

type App struct {
	GRPCServer *grpcapp.App
}

func NewApp(log zerolog.Logger, cfg *config.Config) *App {

	grpcApp := grpcapp.NewApp(log, cfg.GRPC.Port)

	return &App{GRPCServer: grpcApp}
}
