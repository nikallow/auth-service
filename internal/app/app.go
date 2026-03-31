package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/nikallow/auth-service/internal/config"
	httptransport "github.com/nikallow/auth-service/internal/transport/http"
)

const shutdownTimeout = 10 * time.Second

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

	router := httptransport.NewRouter()
	server := &http.Server{
		Addr:    a.cfg.HTTP.Address(),
		Handler: router,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		a.log.Info("starting http server", slog.String("addr", server.Addr))

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		serverErrCh <- nil
	}()

	select {
	case err := <-serverErrCh:
		return err
	case <-ctx.Done():
		a.log.Info("stopping application")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}

		if err := <-serverErrCh; err != nil {
			return err
		}

		return nil
	}
}
