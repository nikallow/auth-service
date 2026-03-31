package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"

	"github.com/nikallow/auth-service/internal/config"
	postgresstorage "github.com/nikallow/auth-service/internal/storage/postgres"
	redisstorage "github.com/nikallow/auth-service/internal/storage/redis"
	httptransport "github.com/nikallow/auth-service/internal/transport/http"
)

const shutdownTimeout = 10 * time.Second

type App struct {
	cfg *config.Config
	log *slog.Logger

	postgres *pgxpool.Pool
	redis    *goredis.Client
}

func New(cfg *config.Config, log *slog.Logger) *App {
	return &App{
		cfg: cfg,
		log: log,
	}
}

func (a *App) Run(ctx context.Context) error {
	a.log.Info("starting application")

	if err := a.initStorage(ctx); err != nil {
		return err
	}
	defer a.closeStorage()

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
			return fmt.Errorf("shutdown http server: %w", err)
		}

		if err := <-serverErrCh; err != nil {
			return err
		}

		return nil
	}
}

func (a *App) initStorage(ctx context.Context) error {
	postgresLog := a.log.With(slog.String("component", "postgres"))
	postgresLog.Info("connecting to postgres")

	pg, err := postgresstorage.New(ctx, a.cfg.PG.DSN())
	if err != nil {
		return fmt.Errorf("init postgres: %w", err)
	}
	a.postgres = pg

	redisLog := a.log.With(slog.String("component", "redis"))
	redisLog.Info("connecting to redis")

	redisClient, err := redisstorage.New(ctx, redisstorage.Config{
		Addr:     a.cfg.Redis.Addr,
		Password: a.cfg.Redis.Password,
		DB:       a.cfg.Redis.DB,
	})
	if err != nil {
		a.postgres.Close()
		a.postgres = nil

		return fmt.Errorf("init redis: %w", err)
	}
	a.redis = redisClient

	a.log.Info("storage initialized")
	return nil
}

func (a *App) closeStorage() {
	if a.postgres != nil {
		a.postgres.Close()
	}

	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			a.log.Warn("close redis client", slog.Any("error", err))
		}
	}
}
