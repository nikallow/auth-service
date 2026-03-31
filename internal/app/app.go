package app

import (
	"context"
	"log/slog"

	"github.com/nikallow/auth-service/internal/config"
)

type App struct {
	cfg *config.Config
	log *slog.Logger
}

func New(cfg *config.Config, log *slog.Logger) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("starting application")

	<-ctx.Done()

	a.log.Info("stopping application")

	return nil
}
