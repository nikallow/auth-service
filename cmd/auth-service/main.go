package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/nikallow/auth-service/internal/app"
	"github.com/nikallow/auth-service/internal/config"
	"github.com/nikallow/auth-service/internal/logger"
)

func main() {
	os.Exit(run())
}

func run() int {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "load .env: %v\n", err)
		return 1
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 1
	}

	logg := logger.New(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application := app.New(cfg, logg)

	if err := application.Run(ctx); err != nil {
		logg.Error("application stopped with error", "error", err)
		return 1
	}

	logg.Info("application stopped")
	return 0
}
