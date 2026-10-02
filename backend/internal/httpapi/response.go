package httpapi

import (
	"encoding/json"
	"net/http"

	"logiflows/backend/internal/middleware"
)

type APIError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Error APIError `json:"error"`
}

type SuccessEnvelope struct {
	Data      interface{}            `json:"data"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
}

func RespondJSON(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}, meta map[string]interface{}) {
	reqID := middleware.GetRequestID(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	env := SuccessEnvelope{
		Data:      data,
		Meta:      meta,
		RequestID: reqID,
	}

	_ = json.NewEncoder(w).Encode(env)
}

func RespondError(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string, details map[string]interface{}) {
	reqID := middleware.GetRequestID(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	env := ErrorEnvelope{
		Error: APIError{
			Code:      code,
			Message:   message,
			RequestID: reqID,
			Details:   details,
		},
	}

	_ = json.NewEncoder(w).Encode(env)
}
