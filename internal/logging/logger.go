package logging

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const (
	loggerKey    contextKey = "logger"
	requestIDKey contextKey = "request_id"
)

var globalLogger *slog.Logger

// Init initializes the global structured logger.
// level can be "debug", "info", "warn", "error" (defaults to "info").
// Use "debug" in development, "info" in production.
func Init(level string) {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	globalLogger = slog.New(handler)
	slog.SetDefault(globalLogger)
}

// Logger returns the global logger instance.
// If Init was never called, returns the default slog logger.
func Logger() *slog.Logger {
	if globalLogger == nil {
		return slog.Default()
	}
	return globalLogger
}

// WithRequestID returns a new context with the request ID attached.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext extracts the request ID from context.
// Returns empty string if not found.
func RequestIDFromContext(ctx context.Context) string {
	if rid, ok := ctx.Value(requestIDKey).(string); ok {
		return rid
	}
	return ""
}

// WithLogger returns a new context with the logger attached.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// LoggerFromContext retrieves the logger from context.
// Falls back to the global logger if not found.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return Logger()
}
