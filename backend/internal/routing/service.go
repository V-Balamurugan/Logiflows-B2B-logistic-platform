package routing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"logiflows/backend/internal/config"
)

// Service defines routing operations for LogiFlows.
type Service interface {
	GetDirections(ctx context.Context, req RouteRequest) (*RouteResponse, error)
	GetMatrix(ctx context.Context, req MatrixRequest) (*MatrixResponse, error)
	GetStatus() StatusResponse
}

// Client implements the routing Service.
type Client struct {
	apiKey     string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewRoutingService creates an authoritative routing service instance.
func NewRoutingService(cfg *config.Config, logger *slog.Logger) Service {
	if logger == nil {
		logger = slog.Default()
	}

	apiKey := strings.TrimSpace(cfg.RoutingAPIKey)
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) GetStatus() StatusResponse {
	hasKey := c.apiKey != ""
	mode := "SIMULATED_FALLBACK"
	if hasKey {
		mode = "LIVE_CLOUD"
	}

	var preview string
	if len(c.apiKey) > 12 {
		preview = c.apiKey[:6] + "..." + c.apiKey[len(c.apiKey)-4:]
	} else if hasKey {
		preview = "***"
	}

	return StatusResponse{
		Status:     "OPERATIONAL",
		Provider:   "OpenRouteService",
		HasAPIKey:  hasKey,
		KeyPreview: preview,
		Mode:       mode,
	}
}

// GetDirections computes route geometry, metrics and steps.
func (c *Client) GetDirections(ctx context.Context, req RouteRequest) (*RouteResponse, error) {
	// If API key is available, attempt OpenRouteService API call
	if c.apiKey != "" {
		resp, err := c.callOpenRouteService(ctx, req)
		if err == nil && resp != nil {
			return resp, nil
		}
		c.logger.Warn("OpenRouteService API call failed; falling back to simulated road geometry",
			slog.String("error", err.Error()),
		)
	}

	// Graceful high-fidelity fallback
	return c.generateSimulatedRoute(req), nil
}

// GetMatrix computes pairwise travel distance and duration.
func (c *Client) GetMatrix(ctx context.Context, req MatrixRequest) (*MatrixResponse, error) {
	n := len(req.Locations)
	if n == 0 {
		return &MatrixResponse{
			Durations: [][]float64{},
			Distances: [][]float64{},
			Source:    "SIMULATED_GEODESIC",
		}, nil
	}

	durations := make([][]float64, n)
	distances := make([][]float64, n)

	for i := 0; i < n; i++ {
		durations[i] = make([]float64, n)
		distances[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i == j {
				durations[i][j] = 0
				distances[i][j] = 0
				continue
			}
			dist := haversineDistanceMeters(
				req.Locations[i].Latitude, req.Locations[i].Longitude,
				req.Locations[j].Latitude, req.Locations[j].Longitude,
			) * 1.28 // city road factor

			// Assume 35 km/h urban speed = 9.72 m/s
			dur := dist / 9.72
			distances[i][j] = math.Round(dist)
			durations[i][j] = math.Round(dur)
		}
	}

	return &MatrixResponse{
		Durations: durations,
		Distances: distances,
		Source:    "SIMULATED_GEODESIC",
	}, nil
}

// callOpenRouteService invokes the OpenRouteService Directions API.
func (c *Client) callOpenRouteService(ctx context.Context, req RouteRequest) (*RouteResponse, error) {
	profile := "driving-car"
	if req.Profile != "" {
		profile = req.Profile
	}

	url := fmt.Sprintf("https://api.openrouteservice.org/v2/directions/%s/geojson", profile)

	// Coordinates format: [[lon, lat], [lon, lat]]
	coords := [][]float64{
		{req.Origin.Longitude, req.Origin.Latitude},
	}
	for _, wp := range req.Waypoints {
		coords = append(coords, []float64{wp.Longitude, wp.Latitude})
	}
	coords = append(coords, []float64{req.Destination.Longitude, req.Destination.Latitude})

	payload := map[string]any{
		"coordinates": coords,
		"instructions": true,
		"units":        "m",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", c.apiKey)
	httpReq.Header.Set("Accept", "application/json, application/geo+json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request error: %w", err)
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openrouteservice returned status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var orsResp struct {
		Features []struct {
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"` // [lon, lat]
			} `json:"geometry"`
			Properties struct {
				Summary struct {
					Distance float64 `json:"distance"`
					Duration float64 `json:"duration"`
				} `json:"summary"`
				Segments []struct {
					Steps []struct {
						Instruction string  `json:"instruction"`
						Distance    float64 `json:"distance"`
						Duration    float64 `json:"duration"`
						Name        string  `json:"name"`
					} `json:"steps"`
				} `json:"segments"`
			} `json:"properties"`
		} `json:"features"`
	}

	if err := json.Unmarshal(bodyBytes, &orsResp); err != nil {
		return nil, fmt.Errorf("failed to decode ors response: %w", err)
	}

	if len(orsResp.Features) == 0 {
		return nil, fmt.Errorf("no route features found in response")
	}

	feature := orsResp.Features[0]

	// Convert [lon, lat] coordinates to Leaflet [lat, lon]
	leafletCoords := make([][]float64, 0, len(feature.Geometry.Coordinates))
	for _, pt := range feature.Geometry.Coordinates {
		if len(pt) >= 2 {
			leafletCoords = append(leafletCoords, []float64{pt[1], pt[0]})
		}
	}

	distMeters := feature.Properties.Summary.Distance
	durSeconds := feature.Properties.Summary.Duration
	distKm := math.Round((distMeters/1000)*10) / 10
	durMinutes := math.Round(durSeconds / 60)

	var steps []RouteStep
	for _, seg := range feature.Properties.Segments {
		for _, st := range seg.Steps {
			steps = append(steps, RouteStep{
				Instruction:     st.Instruction,
				DistanceMeters:  st.Distance,
				DurationSeconds: st.Duration,
				Name:            st.Name,
			})
		}
	}

	return &RouteResponse{
		DistanceMeters:  distMeters,
		DurationSeconds: durSeconds,
		DistanceKm:      distKm,
		DurationMinutes: durMinutes,
		ETAFormatted:    formatETA(durSeconds),
		Geometry:        leafletCoords,
		Steps:           steps,
		Source:          "LIVE_OPENROUTESERVICE",
		Provider:        "OpenRouteService Live API",
	}, nil
}

// generateSimulatedRoute generates realistic curved road coordinates and ETA.
func (c *Client) generateSimulatedRoute(req RouteRequest) *RouteResponse {
	// Points list starting with origin, then waypoints, then destination
	allPoints := []LatLng{req.Origin}
	allPoints = append(allPoints, req.Waypoints...)
	allPoints = append(allPoints, req.Destination)

	var totalDistance float64
	var polyline [][]float64

	for i := 0; i < len(allPoints)-1; i++ {
		p1 := allPoints[i]
		p2 := allPoints[i+1]

		segDist := haversineDistanceMeters(p1.Latitude, p1.Longitude, p2.Latitude, p2.Longitude) * 1.26
		totalDistance += segDist

		// Interpolate intermediate curved road points
		numIntermediate := 6
		for step := 0; step <= numIntermediate; step++ {
			t := float64(step) / float64(numIntermediate)
			lat := p1.Latitude + t*(p2.Latitude-p1.Latitude)
			lon := p1.Longitude + t*(p2.Longitude-p1.Longitude)

			// Subtle sinusoidal deviation simulating road turns
			deviation := math.Sin(t*math.Pi) * 0.0028
			lat += deviation * 0.6
			lon += deviation * 0.8

			// Avoid repeating joining nodes
			if step > 0 || i == 0 {
				polyline = append(polyline, []float64{
					math.Round(lat*100000) / 100000,
					math.Round(lon*100000) / 100000,
				})
			}
		}
	}

	// Speed: 36 km/h = 10 m/s
	durationSeconds := totalDistance / 10.0
	distanceKm := math.Round((totalDistance/1000)*10) / 10
	durationMinutes := math.Round(durationSeconds / 60)

	steps := []RouteStep{
		{
			Instruction:     "Depart origin facility along outbound logistics route",
			DistanceMeters:  totalDistance * 0.15,
			DurationSeconds: durationSeconds * 0.15,
			Name:            "Arterial Ring Road",
		},
		{
			Instruction:     "Merge onto express delivery corridor",
			DistanceMeters:  totalDistance * 0.65,
			DurationSeconds: durationSeconds * 0.65,
			Name:            "Transit Expressway",
		},
		{
			Instruction:     "Arrive at destination delivery point",
			DistanceMeters:  totalDistance * 0.20,
			DurationSeconds: durationSeconds * 0.20,
			Name:            "Local Service Lane",
		},
	}

	return &RouteResponse{
		DistanceMeters:  math.Round(totalDistance),
		DurationSeconds: math.Round(durationSeconds),
		DistanceKm:      distanceKm,
		DurationMinutes: durationMinutes,
		ETAFormatted:    formatETA(durationSeconds),
		Geometry:        polyline,
		Steps:           steps,
		Source:          "SIMULATED_GEODESIC",
		Provider:        "LogiFlows Spatial Fallback Engine",
	}
}

func haversineDistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000.0 // meters
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func formatETA(seconds float64) string {
	mins := int(math.Round(seconds / 60))
	if mins < 1 {
		return "Under 1 min"
	}
	if mins < 60 {
		return fmt.Sprintf("%d mins", mins)
	}
	hrs := mins / 60
	remMins := mins % 60
	if remMins == 0 {
		return fmt.Sprintf("%d hr", hrs)
	}
	return fmt.Sprintf("%d hr %d mins", hrs, remMins)
}
