package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

// Init инициализирует глобальный логгер. jsonMode=true для production (JSON в stdout для ELK/Loki).
func Init(service string, jsonMode bool) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: levelFromEnv(),
	}
	if jsonMode {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	log = slog.New(handler).With("service", service)
}

func levelFromEnv() slog.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// L возвращает логгер. Вызывать после Init().
func L() *slog.Logger {
	if log == nil {
		log = slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "unknown")
	}
	return log
}
