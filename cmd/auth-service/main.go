package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/nikallow/auth-service/internal/app"
	"github.com/nikallow/auth-service/internal/config"
	"github.com/nikallow/auth-service/internal/logger"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("error loading .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logg := logger.New(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application := app.New(cfg, logg)

	if err := application.Run(ctx); err != nil {
		logg.Error("application stopped with error", "error", err)
	}
}
