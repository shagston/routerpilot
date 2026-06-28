package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

type CtxKey string

const fieldsKey CtxKey = "log_fields"

type Logger struct {
	*slog.Logger
}

func New() *Logger {
	var h slog.Handler
	format := strings.ToLower(os.Getenv("ROUTERPILOT_LOG_FORMAT"))
	level := parseLevel(os.Getenv("ROUTERPILOT_LOG_LEVEL"))

	opts := &slog.HandlerOptions{Level: level}

	if format == "json" {
		h = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		h = slog.NewTextHandler(os.Stderr, opts)
	}

	l := slog.New(h)
	slog.SetDefault(l)
	return &Logger{l}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
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

func WithFields(ctx context.Context, fields ...any) context.Context {
	existing := ctx.Value(fieldsKey)
	if existing != nil {
		if m, ok := existing.(map[string]any); ok {
			for i := 0; i < len(fields); i += 2 {
				if key, ok := fields[i].(string); ok && i+1 < len(fields) {
					m[key] = fields[i+1]
				}
			}
			return context.WithValue(ctx, fieldsKey, m)
		}
	}
	m := make(map[string]any)
	for i := 0; i < len(fields); i += 2 {
		if key, ok := fields[i].(string); ok && i+1 < len(fields) {
			m[key] = fields[i+1]
		}
	}
	return context.WithValue(ctx, fieldsKey, m)
}

func CtxFields(ctx context.Context) []any {
	var result []any
	if m, ok := ctx.Value(fieldsKey).(map[string]any); ok {
		for k, v := range m {
			result = append(result, k, v)
		}
	}
	return result
}
