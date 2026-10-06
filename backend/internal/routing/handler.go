package routing

import (
	"encoding/json"
	"net/http"
)

// ResponseWriter helper type definitions matching LogiFlows standards
type ErrorResponder func(w http.ResponseWriter, r *http.Request, statusCode int, code string, message string, details map[string]interface{})
type JSONResponder func(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}, meta map[string]interface{})

// Handler handles HTTP requests for routing and navigation.
type Handler struct {
	service      Service
	respondJSON  JSONResponder
	respondError ErrorResponder
}

// NewHandler creates a routing HTTP handler.
func NewHandler(service Service, jsonFn JSONResponder, errFn ErrorResponder) *Handler {
	return &Handler{
		service:      service,
		respondJSON:  jsonFn,
		respondError: errFn,
	}
}

// GetStatus returns the operational status and provider of the routing engine.
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.service.GetStatus()
	h.respondJSON(w, r, http.StatusOK, status, nil)
}

// CalculateDirections handles POST /api/v1/routing/directions
func (h *Handler) CalculateDirections(w http.ResponseWriter, r *http.Request) {
	var req RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload for directions", nil)
		return
	}

	if req.Origin.Latitude == 0 && req.Origin.Longitude == 0 {
		h.respondError(w, r, http.StatusBadRequest, "INVALID_ORIGIN", "Origin coordinates are required", nil)
		return
	}

	if req.Destination.Latitude == 0 && req.Destination.Longitude == 0 {
		h.respondError(w, r, http.StatusBadRequest, "INVALID_DESTINATION", "Destination coordinates are required", nil)
		return
	}

	route, err := h.service.GetDirections(r.Context(), req)
	if err != nil {
		h.respondError(w, r, http.StatusInternalServerError, "ROUTING_ERROR", err.Error(), nil)
		return
	}

	h.respondJSON(w, r, http.StatusOK, route, nil)
}

// CalculateMatrix handles POST /api/v1/routing/matrix
func (h *Handler) CalculateMatrix(w http.ResponseWriter, r *http.Request) {
	var req MatrixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, r, http.StatusBadRequest, "BAD_REQUEST", "Invalid JSON payload for distance matrix", nil)
		return
	}

	if len(req.Locations) < 2 {
		h.respondError(w, r, http.StatusBadRequest, "INSUFFICIENT_LOCATIONS", "At least 2 locations are required", nil)
		return
	}

	matrix, err := h.service.GetMatrix(r.Context(), req)
	if err != nil {
		h.respondError(w, r, http.StatusInternalServerError, "MATRIX_ERROR", err.Error(), nil)
		return
	}

	h.respondJSON(w, r, http.StatusOK, matrix, nil)
}
