package routing

// LatLng represents a geographic coordinate [latitude, longitude].
type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// RouteRequest defines the input payload for directions calculation.
type RouteRequest struct {
	Origin      LatLng   `json:"origin"`
	Destination LatLng   `json:"destination"`
	Waypoints   []LatLng `json:"waypoints,omitempty"`
	Profile     string   `json:"profile,omitempty"` // driving-car, driving-hgv, cycling
}

// RouteStep represents a single turn-by-turn instruction.
type RouteStep struct {
	Instruction     string  `json:"instruction"`
	DistanceMeters  float64 `json:"distance_meters"`
	DurationSeconds float64 `json:"duration_seconds"`
	Name            string  `json:"name,omitempty"`
}

// RouteResponse defines the output route geometry and metrics.
type RouteResponse struct {
	DistanceMeters  float64     `json:"distance_meters"`
	DurationSeconds float64     `json:"duration_seconds"`
	DistanceKm      float64     `json:"distance_km"`
	DurationMinutes float64     `json:"duration_minutes"`
	ETAFormatted    string      `json:"eta_formatted"`
	Geometry        [][]float64 `json:"geometry"` // Array of [lat, lon] coordinates for Leaflet polyline
	Steps           []RouteStep `json:"steps"`
	Source          string      `json:"source"`   // LIVE_OPENROUTESERVICE or SIMULATED_GEODESIC
	Provider        string      `json:"provider"`
}

// MatrixRequest defines multiple origins and destinations for distance/duration calculation.
type MatrixRequest struct {
	Locations []LatLng `json:"locations"`
	Metrics   []string `json:"metrics,omitempty"` // distance, duration
	Profile   string   `json:"profile,omitempty"`
}

// MatrixResponse contains the computed NxM distance and duration tables.
type MatrixResponse struct {
	Durations [][]float64 `json:"durations"` // in seconds
	Distances [][]float64 `json:"distances"` // in meters
	Source    string      `json:"source"`
}

// StatusResponse describes the routing engine readiness and API key state.
type StatusResponse struct {
	Status     string `json:"status"`
	Provider   string `json:"provider"`
	HasAPIKey  bool   `json:"has_api_key"`
	KeyPreview string `json:"key_preview,omitempty"`
	Mode       string `json:"mode"` // LIVE_CLOUD or SIMULATED_FALLBACK
}
