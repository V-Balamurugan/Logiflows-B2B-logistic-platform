package middleware

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware_Generated(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	var capturedReqID string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReqID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestID(testHandler)
	handler.ServeHTTP(rr, req)

	if capturedReqID == "" {
		t.Fatal("expected request id to be injected into context, got empty")
	}

	resHeader := rr.Header().Get(HeaderXRequestID)
	if resHeader != capturedReqID {
		t.Fatalf("expected header %s to match context %s", resHeader, capturedReqID)
	}
}

func TestRequestIDMiddleware_Propagated(t *testing.T) {
	existingID := "custom-client-trace-12345"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderXRequestID, existingID)
	rr := httptest.NewRecorder()

	var capturedReqID string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReqID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestID(testHandler)
	handler.ServeHTTP(rr, req)

	if capturedReqID != existingID {
		t.Fatalf("expected request id %s, got %s", existingID, capturedReqID)
	}
}

func TestRecovererMiddleware(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("deliberate test failure")
	})

	handler := RequestID(Recoverer(logger)(panicHandler))
	req := httptest.NewRequest("GET", "/panic", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if resp.Error.Code != "INTERNAL_SERVER_ERROR" {
		t.Errorf("expected code INTERNAL_SERVER_ERROR, got %s", resp.Error.Code)
	}
	if resp.Error.RequestID == "" {
		t.Errorf("expected request_id to be populated in error response")
	}
}
