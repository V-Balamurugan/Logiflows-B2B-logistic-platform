package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"logiflows/backend/internal/logging"
)

const HeaderXRequestID = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(HeaderXRequestID)
		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), logging.RequestIDKey, reqID)
		w.Header().Set(HeaderXRequestID, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(logging.RequestIDKey).(string); ok {
		return id
	}
	return ""
}
