package middleware

import (
	"encoding/json"
	"net/http"
)

type MiddlewareAPIError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type MiddlewareErrorEnvelope struct {
	Error MiddlewareAPIError `json:"error"`
}

func respondJSONError(w http.ResponseWriter, r *http.Request, statusCode int, code, message string, details map[string]interface{}) {
	reqID := GetRequestID(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	env := MiddlewareErrorEnvelope{
		Error: MiddlewareAPIError{
			Code:      code,
			Message:   message,
			RequestID: reqID,
			Details:   details,
		},
	}

	_ = json.NewEncoder(w).Encode(env)
}
