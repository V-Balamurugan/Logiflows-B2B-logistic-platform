package tenancy

import (
	"encoding/json"
	"testing"
)

func TestGeoPointValidation(t *testing.T) {
	tests := []struct {
		name    string
		point   GeoPoint
		wantErr bool
	}{
		{
			name:    "valid coordinates Bangalore",
			point:   GeoPoint{Latitude: 12.9716, Longitude: 77.5946},
			wantErr: false,
		},
		{
			name:    "invalid latitude too high",
			point:   GeoPoint{Latitude: 95.0, Longitude: 77.5946},
			wantErr: true,
		},
		{
			name:    "invalid latitude too low",
			point:   GeoPoint{Latitude: -95.0, Longitude: 77.5946},
			wantErr: true,
		},
		{
			name:    "invalid longitude too high",
			point:   GeoPoint{Latitude: 12.9716, Longitude: 185.0},
			wantErr: true,
		},
		{
			name:    "invalid longitude too low",
			point:   GeoPoint{Latitude: 12.9716, Longitude: -190.0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.point.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAndFormatPolygonGeoJSON(t *testing.T) {
	t.Run("valid closed polygon", func(t *testing.T) {
		raw := json.RawMessage(`{
			"type": "Polygon",
			"coordinates": [
				[
					[77.5, 12.9],
					[77.6, 12.9],
					[77.6, 13.0],
					[77.5, 13.0],
					[77.5, 12.9]
				]
			]
		}`)
		formatted, err := validateAndFormatPolygonGeoJSON(raw)
		if err != nil {
			t.Fatalf("expected valid polygon, got error: %v", err)
		}
		if len(formatted) == 0 {
			t.Fatal("expected formatted GeoJSON")
		}
	})

	t.Run("auto-closes open polygon ring", func(t *testing.T) {
		raw := json.RawMessage(`{
			"type": "Polygon",
			"coordinates": [
				[
					[77.5, 12.9],
					[77.6, 12.9],
					[77.6, 13.0],
					[77.5, 13.0]
				]
			]
		}`)
		formatted, err := validateAndFormatPolygonGeoJSON(raw)
		if err != nil {
			t.Fatalf("expected auto-closed polygon, got error: %v", err)
		}

		var parsed map[string]interface{}
		_ = json.Unmarshal(formatted, &parsed)
		coords := parsed["coordinates"].([]interface{})[0].([]interface{})
		if len(coords) != 5 {
			t.Fatalf("expected 5 coordinates after auto-closing, got %d", len(coords))
		}
	})

	t.Run("rejects non-polygon type", func(t *testing.T) {
		raw := json.RawMessage(`{
			"type": "Point",
			"coordinates": [77.5, 12.9]
		}`)
		_, err := validateAndFormatPolygonGeoJSON(raw)
		if err == nil {
			t.Fatal("expected error for non-polygon geometry type")
		}
	})

	t.Run("rejects polygon with insufficient vertices", func(t *testing.T) {
		raw := json.RawMessage(`{
			"type": "Polygon",
			"coordinates": [
				[
					[77.5, 12.9],
					[77.6, 12.9]
				]
			]
		}`)
		_, err := validateAndFormatPolygonGeoJSON(raw)
		if err == nil {
			t.Fatal("expected error for polygon with fewer than 4 vertices")
		}
	})
}

func TestTenantCreateRequestValidation(t *testing.T) {
	req := TenantCreateRequest{
		Name: "Swift Logistics",
		Code: "SWIFT-LOG",
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid tenant create request, got %v", err)
	}
	if req.Slug != "swift-log" {
		t.Errorf("expected slug 'swift-log', got '%s'", req.Slug)
	}
	if req.Tier != "STANDARD" {
		t.Errorf("expected default tier STANDARD, got '%s'", req.Tier)
	}
	if req.QuotaBranches != 50 {
		t.Errorf("expected default quota branches 50, got %d", req.QuotaBranches)
	}
}

func TestBranchCreateRequestValidation(t *testing.T) {
	req := BranchCreateRequest{
		Code:        "BLR-HUB-01",
		Name:        "Central Sorting Hub",
		BranchType:  BranchTypeSortingHub,
		Address:     "123 Industrial Area, Bangalore",
		Location:    GeoPoint{Latitude: 12.9716, Longitude: 77.5946},
		ServiceArea: json.RawMessage(`{"type":"Polygon","coordinates":[[[77.5,12.9],[77.6,12.9],[77.6,13.0],[77.5,13.0],[77.5,12.9]]]}`),
	}

	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid branch create request, got %v", err)
	}
	if req.Status != BranchStatusActive {
		t.Errorf("expected default status ACTIVE, got '%s'", req.Status)
	}
	if req.DailyCapacity != 1000 {
		t.Errorf("expected default daily capacity 1000, got %d", req.DailyCapacity)
	}
}
