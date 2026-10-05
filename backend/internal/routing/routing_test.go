package routing

import (
	"context"
	"testing"

	"logiflows/backend/internal/config"
)

func TestRoutingService_FallbackRoute(t *testing.T) {
	cfg := &config.Config{
		RoutingAPIKey: "", // empty to test simulated fallback
	}

	service := NewRoutingService(cfg, nil)
	status := service.GetStatus()

	if status.Mode != "SIMULATED_FALLBACK" {
		t.Fatalf("expected SIMULATED_FALLBACK, got %s", status.Mode)
	}
	if status.HasAPIKey {
		t.Fatalf("expected HasAPIKey to be false")
	}

	req := RouteRequest{
		Origin: LatLng{
			Latitude:  12.9716,
			Longitude: 77.5946,
		},
		Destination: LatLng{
			Latitude:  13.0358,
			Longitude: 77.5970,
		},
	}

	resp, err := service.GetDirections(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	if resp.DistanceMeters <= 0 {
		t.Fatalf("expected positive distance, got %f", resp.DistanceMeters)
	}

	if len(resp.Geometry) < 2 {
		t.Fatalf("expected at least 2 coordinate points in geometry, got %d", len(resp.Geometry))
	}

	if resp.ETAFormatted == "" {
		t.Fatalf("expected non-empty ETA formatted string")
	}

	if resp.Source != "SIMULATED_GEODESIC" {
		t.Fatalf("expected SIMULATED_GEODESIC source, got %s", resp.Source)
	}
}

func TestRoutingService_Matrix(t *testing.T) {
	service := NewRoutingService(&config.Config{}, nil)

	req := MatrixRequest{
		Locations: []LatLng{
			{Latitude: 12.9716, Longitude: 77.5946},
			{Latitude: 12.9352, Longitude: 77.6245},
			{Latitude: 13.0358, Longitude: 77.5970},
		},
	}

	resp, err := service.GetMatrix(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Distances) != 3 || len(resp.Durations) != 3 {
		t.Fatalf("expected 3x3 matrix, got %dx%d", len(resp.Distances), len(resp.Durations))
	}

	// Diagonal should be zero
	for i := 0; i < 3; i++ {
		if resp.Distances[i][i] != 0 || resp.Durations[i][i] != 0 {
			t.Fatalf("expected diagonal to be zero at %d", i)
		}
	}

	// Non-diagonal should be positive
	if resp.Distances[0][1] <= 0 || resp.Durations[0][1] <= 0 {
		t.Fatalf("expected non-diagonal distances to be positive")
	}
}
