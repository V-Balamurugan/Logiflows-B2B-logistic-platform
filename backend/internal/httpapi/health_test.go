package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_Healthz(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "http://localhost:8000", "0.1.0-alpha")

	req := httptest.NewRequest("GET", "/api/v1/healthz", nil)
	rr := httptest.NewRecorder()

	handler.Healthz(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp SuccessEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data map, got %T", resp.Data)
	}

	if dataMap["status"] != "OK" {
		t.Errorf("expected status OK, got %v", dataMap["status"])
	}
	if dataMap["version"] != "0.1.0-alpha" {
		t.Errorf("expected version 0.1.0-alpha, got %v", dataMap["version"])
	}
}

func TestHealthHandler_Readyz(t *testing.T) {
	handler := NewHealthHandler(nil, nil, "http://localhost:8000", "0.1.0-alpha")

	req := httptest.NewRequest("GET", "/api/v1/readyz", nil)
	rr := httptest.NewRecorder()

	handler.Readyz(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp SuccessEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data map, got %T", resp.Data)
	}

	if dataMap["status"] != "READY" {
		t.Errorf("expected status READY, got %v", dataMap["status"])
	}
}
