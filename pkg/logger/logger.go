package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Init инициализирует глобальный JSON логгер с уровнем из ENV (LOG_LEVEL).
func SetupLogger() *slog.Logger {
	level := os.Getenv("LOG_LEVEL")
	lvl := parseLogLevel(level)

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})

	return slog.New(handler)
}

// parseLogLevel преобразует строку в slog.Level
func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
