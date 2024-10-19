package grpcapp

import (
	"fmt"
	"net"
	authgrpc "vieo/auth/internal/grpc/auth"
	"vieo/auth/internal/lib/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// the application in which we wrap the grpc server

type App struct {
	log        *logger.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *logger.Logger, auth authgrpc.Auth, port int) *App {
	gRPCServer := grpc.NewServer()
	authgrpc.Register(gRPCServer, auth)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustStart() {
	if err := a.Start(); err != nil {
		panic(err)
	}
}

func (a *App) Start() error {
	const op = "grpcapp.Start"
	log := a.log.With(zap.String("op", op), zap.Int("port", a.port))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running")

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.log.With(zap.String("op", op)).
		Info("stopping gRPC server", zap.Int("port", a.port))
	// stops accepting new requests and finalizes old ones
	a.gRPCServer.GracefulStop()
}
