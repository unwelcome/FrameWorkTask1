package grpcapp

import (
	"fmt"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	authgrpc "microapi/internal/grpc/auth"
	"net"
)

type App struct {
	log        zerolog.Logger
	gRPCServer *grpc.Server
	port       int
}

func NewApp(log zerolog.Logger, port int) *App {
	gRPCServer := grpc.NewServer()

	authgrpc.Register(gRPCServer)

	return &App{log: log, gRPCServer: gRPCServer, port: port}
}

func (a *App) MustRun() {
	const op = "grpcapp.Run"

	a.log.Info().Msgf("Starting tcp listener on port %d", a.port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		a.log.Fatal().Err(fmt.Errorf("%s: %w", op, err))
	}

	a.log.Info().Str("address", lis.Addr().String()).Msg("gRPC server in running")
	if err = a.gRPCServer.Serve(lis); err != nil {
		a.log.Fatal().Err(fmt.Errorf("%s: %w", op, err))
	}

}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.Info().Msg("Stopping gRPC server")

	a.gRPCServer.GracefulStop()
}
