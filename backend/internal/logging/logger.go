package logging

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	TenantIDKey  contextKey = "tenant_id"
	UserIDKey    contextKey = "user_id"
)

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx != nil {
		if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
			r.AddAttrs(slog.String("request_id", reqID))
		}
		if tenantID, ok := ctx.Value(TenantIDKey).(string); ok && tenantID != "" {
			r.AddAttrs(slog.String("tenant_id", tenantID))
		}
		if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
			r.AddAttrs(slog.String("user_id", userID))
		}
	}
	return h.Handler.Handle(ctx, r)
}

func NewLogger(level string, isJSON bool) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug", "DEBUG":
		logLevel = slog.LevelDebug
	case "warn", "WARN":
		logLevel = slog.LevelWarn
	case "error", "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if isJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(&ContextHandler{Handler: handler})
}
