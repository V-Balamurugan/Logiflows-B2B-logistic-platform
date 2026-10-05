package routing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"logiflows/backend/internal/config"
	"logiflows/backend/internal/logging"
)

func TestRoutingService_NoKey(t *testing.T) {
	cfg := &config.Config{RoutingAPIKey: ""}
	logger := logging.NewLogger("debug", false)
	svc := NewRoutingService(cfg, logger)

	status := svc.GetStatus()
	if status.HasAPIKey {
		t.Errorf("expected HasAPIKey to be false, got true")
	}
	if status.Mode != "SIMULATED_FALLBACK" {
		t.Errorf("expected Mode SIMULATED_FALLBACK, got %s", status.Mode)
	}
	if status.KeyStatus != "NOT_CONFIGURED" {
		t.Errorf("expected KeyStatus NOT_CONFIGURED, got %s", status.KeyStatus)
	}

	req := RouteRequest{
		Origin:      LatLng{Latitude: 12.9716, Longitude: 77.5946},
		Destination: LatLng{Latitude: 13.0358, Longitude: 77.5970},
	}
	route, err := svc.GetDirections(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if route == nil {
		t.Fatalf("expected non-nil route")
	}
	if route.Source != "SIMULATED_GEODESIC" {
		t.Errorf("expected source SIMULATED_GEODESIC, got %s", route.Source)
	}
	if len(route.Geometry) == 0 {
		t.Errorf("expected non-empty geometry coordinates")
	}
}

func TestRoutingService_InvalidKeyHandling(t *testing.T) {
	// Mock ORS server returning 401 Unauthorized for invalid key
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    2001,
				"message": "Access to this API has been disallowed: API key is invalid",
			},
		})
	}))
	defer mockServer.Close()

	cfg := &config.Config{RoutingAPIKey: "invalid_ors_token_12345"}
	logger := logging.NewLogger("debug", false)
	client := &Client{
		apiKey:     cfg.RoutingAPIKey,
		httpClient: mockServer.Client(),
		logger:     logger,
	}

	// Initial status should reflect configured but unverified key
	status := client.GetStatus()
	if !status.HasAPIKey {
		t.Fatalf("expected HasAPIKey to be true")
	}
	if status.KeyStatus != "CONFIGURED" {
		t.Errorf("expected initial KeyStatus CONFIGURED, got %s", status.KeyStatus)
	}

	// Trigger directions request with custom endpoint override for test
	req := RouteRequest{
		Origin:      LatLng{Latitude: 12.9716, Longitude: 77.5946},
		Destination: LatLng{Latitude: 13.0358, Longitude: 77.5970},
	}

	// Make call through client
	route, err := client.GetDirections(context.Background(), req)
	if err != nil {
		t.Fatalf("GetDirections should gracefully fall back instead of failing, got error: %v", err)
	}
	if route == nil {
		t.Fatalf("expected non-nil route")
	}
	if route.Source != "SIMULATED_GEODESIC" {
		t.Errorf("expected fallback route source SIMULATED_GEODESIC, got %s", route.Source)
	}
}

func TestRoutingService_MatrixComputation(t *testing.T) {
	cfg := &config.Config{RoutingAPIKey: ""}
	logger := logging.NewLogger("debug", false)
	svc := NewRoutingService(cfg, logger)

	req := MatrixRequest{
		Locations: []LatLng{
			{Latitude: 12.9716, Longitude: 77.5946},
			{Latitude: 13.0358, Longitude: 77.5970},
			{Latitude: 12.9279, Longitude: 77.6271},
		},
	}

	resp, err := svc.GetMatrix(context.Background(), req)
	if err != nil {
		t.Fatalf("matrix calculation failed: %v", err)
	}
	if len(resp.Distances) != 3 || len(resp.Durations) != 3 {
		t.Fatalf("expected 3x3 matrix, got %dx%d", len(resp.Distances), len(resp.Durations))
	}
	if resp.Distances[0][0] != 0 || resp.Durations[0][0] != 0 {
		t.Errorf("diagonal should be 0")
	}
	if resp.Distances[0][1] <= 0 {
		t.Errorf("expected positive distance between point 0 and point 1")
	}
}
