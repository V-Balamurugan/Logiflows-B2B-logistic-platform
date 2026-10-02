package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type ErrorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					reqID := GetRequestID(r.Context())
					stack := string(debug.Stack())

					logger.ErrorContext(r.Context(), "Unhandled panic recovered",
						slog.Any("panic", rvr),
						slog.String("stack", stack),
						slog.String("path", r.URL.Path),
						slog.String("method", r.Method),
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					var resp ErrorResponse
					resp.Error.Code = "INTERNAL_SERVER_ERROR"
					resp.Error.Message = "An unexpected server error occurred."
					resp.Error.RequestID = reqID

					_ = json.NewEncoder(w).Encode(resp)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
