package app

import (
	"time"
	grpcapp "vieo/auth/internal/app/grpc"
	"vieo/auth/internal/lib/logger"
	"vieo/auth/internal/services/auth"
	postgre "vieo/auth/internal/storage/postgres"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *logger.Logger,
	grpcPort int,
	storagePath string,
	tokenTTL time.Duration,
	secretKey string,
) *App {

	storage, err := postgre.New(storagePath)
	if err != nil {
		panic(err)
	}
	authService := auth.New(log, storage, storage, storage, storage, tokenTTL, secretKey)
	// secret key on two levels transport and service!
	grpcApp := grpcapp.New(log, authService, grpcPort, secretKey)

	return &App{
		GRPCSrv: grpcApp,
	}
}
