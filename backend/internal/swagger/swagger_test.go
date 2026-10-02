package swagger

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSpecHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)
	rr := httptest.NewRecorder()

	SpecHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/yaml") {
		t.Errorf("expected Content-Type containing application/yaml, got %s", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "openapi: 3.0.3") {
		t.Errorf("expected openapi spec to contain 'openapi: 3.0.3', got: %s", body[:min(len(body), 50)])
	}
	if !strings.Contains(body, "LogiFlows API Gateway") {
		t.Errorf("expected openapi spec to contain 'LogiFlows API Gateway'")
	}
}

func TestUIHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	rr := httptest.NewRecorder()

	UIHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("expected Content-Type containing text/html, got %s", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "SwaggerUIBundle") {
		t.Errorf("expected Swagger UI HTML to contain 'SwaggerUIBundle'")
	}
	if !strings.Contains(body, "/api/v1/openapi.yaml") {
		t.Errorf("expected Swagger UI to reference '/api/v1/openapi.yaml'")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
