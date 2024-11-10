package main

import (
	"os"
	"os/signal"
	"syscall"
	"vieo/auth/internal/app"
	"vieo/auth/internal/config"
	"vieo/auth/internal/lib/logger"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.GetConfig()
	log := logger.NewLogger(cfg.Env)

	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.GRPC.TokenTTL, cfg.GRPC.SecretKey)

	go func() {
		application.GRPCSrv.MustStart()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	application.GRPCSrv.Stop()
	log.Info("Gracefully stopped")

}
