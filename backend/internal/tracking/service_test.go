package tracking

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"logiflows/backend/internal/logging"
)

type mockTrackingService struct {
	locations map[string]LocationUpdate
}

func newMockTrackingService() *mockTrackingService {
	return &mockTrackingService{
		locations: make(map[string]LocationUpdate),
	}
}

func (m *mockTrackingService) RecordLocation(ctx context.Context, loc LocationUpdate) error {
	m.locations[loc.AssignmentID] = loc
	return nil
}

func (m *mockTrackingService) GetLatestLocation(ctx context.Context, assignmentID string) (*LocationUpdate, error) {
	loc, ok := m.locations[assignmentID]
	if !ok {
		return nil, nil
	}
	return &loc, nil
}

func (m *mockTrackingService) Subscribe(ctx context.Context, assignmentID string) *redis.PubSub {
	return nil
}

func TestTrackingHandler_CoordinateValidation(t *testing.T) {
	svc := newMockTrackingService()
	logger := logging.NewLogger("debug", false)
	handler := NewHandler(svc, logger)

	t.Run("Valid coordinates are accepted", func(t *testing.T) {
		payload := LocationUpdate{
			AssignmentID: "asg-101",
			EmployeeID:   "emp-55",
			Latitude:     12.9716,
			Longitude:    77.5946,
			Speed:        35.0,
			Heading:      120.0,
			Timestamp:    time.Now().UTC(),
		}
		data, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/tracking/location", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateLocation(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		if saved, _ := svc.GetLatestLocation(context.Background(), "asg-101"); saved == nil || saved.Latitude != 12.9716 {
			t.Fatalf("expected location to be saved in service, got %+v", saved)
		}
	})

	t.Run("Invalid latitude > 90 rejected with 400", func(t *testing.T) {
		payload := LocationUpdate{
			AssignmentID: "asg-101",
			Latitude:     95.5,
			Longitude:    77.5946,
		}
		data, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/tracking/location", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateLocation(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request for latitude > 90, got %d", rr.Code)
		}
	})

	t.Run("Invalid longitude < -180 rejected with 400", func(t *testing.T) {
		payload := LocationUpdate{
			AssignmentID: "asg-101",
			Latitude:     12.9716,
			Longitude:    -185.0,
		}
		data, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/tracking/location", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateLocation(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request for longitude < -180, got %d", rr.Code)
		}
	})

	t.Run("Missing assignment_id rejected with 400", func(t *testing.T) {
		payload := LocationUpdate{
			Latitude:  12.9716,
			Longitude: 77.5946,
		}
		data, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/api/v1/tracking/location", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.UpdateLocation(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request for missing assignment_id, got %d", rr.Code)
		}
	})
}
