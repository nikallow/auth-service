package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/nikallow/auth-service/internal/config"
)

func New(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Logger.Level),
		AddSource: cfg.Env == "local",
	}

	var handler slog.Handler

	switch strings.ToLower(cfg.Logger.Format) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(
		slog.String("service", cfg.Service.Name),
		slog.String("env", cfg.Env),
	)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
