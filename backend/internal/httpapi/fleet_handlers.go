package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"logiflows/backend/internal/fleet"
	"logiflows/backend/internal/middleware"
)

type FleetHandler struct {
	service fleet.Service
}

func NewFleetHandler(service fleet.Service) *FleetHandler {
	return &FleetHandler{service: service}
}

func getTenantID(r *http.Request) (uuid.UUID, error) {
	// First check route URL param
	routeParam := chi.URLParam(r, "tenant_id")
	if routeParam != "" {
		tid, err := uuid.Parse(routeParam)
		if err == nil && tid != uuid.Nil {
			return tid, nil
		}
	}

	// Check claims
	claims := middleware.GetUserClaims(r.Context())
	if claims != nil && claims.TenantID != nil && *claims.TenantID != uuid.Nil {
		return *claims.TenantID, nil
	}

	return uuid.Nil, errors.New("missing or invalid tenant_id")
}

// POST /api/v1/tenants/{tenant_id}/vehicles or /api/v1/vehicles
func (h *FleetHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	var dto fleet.CreateVehicleDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to decode vehicle payload", map[string]any{"error": err.Error()})
		return
	}

	v, err := h.service.RegisterVehicle(r.Context(), tenantID, dto)
	if err != nil {
		if errors.Is(err, fleet.ErrAlreadyExists) {
			RespondError(w, r, http.StatusConflict, "VEHICLE_EXISTS", "Vehicle with this license plate already registered for tenant", nil)
			return
		}
		RespondError(w, r, http.StatusBadRequest, "CREATE_VEHICLE_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusCreated, v, nil)
}

// GET /api/v1/tenants/{tenant_id}/vehicles or /api/v1/vehicles
func (h *FleetHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	filter := fleet.VehicleFilter{}
	if branchStr := r.URL.Query().Get("branch_id"); branchStr != "" {
		if bid, err := uuid.Parse(branchStr); err == nil {
			filter.BranchID = &bid
		}
	}
	if driverStr := r.URL.Query().Get("driver_id"); driverStr != "" {
		if did, err := uuid.Parse(driverStr); err == nil {
			filter.DriverID = &did
		}
	}
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		st := fleet.VehicleStatus(statusStr)
		if st.IsValid() {
			filter.Status = &st
		}
	}
	if typeStr := r.URL.Query().Get("type"); typeStr != "" {
		vt := fleet.VehicleType(typeStr)
		if vt.IsValid() {
			filter.VehicleType = &vt
		}
	}

	vehicles, err := h.service.ListVehicles(r.Context(), tenantID, filter)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, "LIST_VEHICLES_FAILED", err.Error(), nil)
		return
	}

	if vehicles == nil {
		vehicles = []fleet.Vehicle{}
	}

	RespondJSON(w, r, http.StatusOK, vehicles, nil)
}

// GET /api/v1/tenants/{tenant_id}/vehicles/{vehicle_id} or /api/v1/vehicles/{vehicle_id}
func (h *FleetHandler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicle_id"))
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle UUID", nil)
		return
	}

	v, err := h.service.GetVehicle(r.Context(), tenantID, vehicleID)
	if err != nil {
		if errors.Is(err, fleet.ErrNotFound) {
			RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "Vehicle not found", nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "GET_VEHICLE_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, v, nil)
}

// PUT /api/v1/tenants/{tenant_id}/vehicles/{vehicle_id} or /api/v1/vehicles/{vehicle_id}
func (h *FleetHandler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicle_id"))
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle UUID", nil)
		return
	}

	var dto fleet.UpdateVehicleDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to decode vehicle update payload", map[string]any{"error": err.Error()})
		return
	}

	v, err := h.service.UpdateVehicle(r.Context(), tenantID, vehicleID, dto)
	if err != nil {
		if errors.Is(err, fleet.ErrNotFound) {
			RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "Vehicle not found", nil)
			return
		}
		RespondError(w, r, http.StatusBadRequest, "UPDATE_VEHICLE_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, v, nil)
}

// DELETE /api/v1/tenants/{tenant_id}/vehicles/{vehicle_id} or /api/v1/vehicles/{vehicle_id}
func (h *FleetHandler) DeleteVehicle(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicle_id"))
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle UUID", nil)
		return
	}

	if err := h.service.DeleteVehicle(r.Context(), tenantID, vehicleID); err != nil {
		if errors.Is(err, fleet.ErrNotFound) {
			RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "Vehicle not found", nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "DELETE_VEHICLE_FAILED", err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/vehicles/{vehicle_id}/telematics
func (h *FleetHandler) IngestTelematics(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicle_id"))
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle UUID", nil)
		return
	}

	var dto fleet.IngestTelematicsDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to decode telematics payload", map[string]any{"error": err.Error()})
		return
	}

	t, err := h.service.RecordTelematics(r.Context(), tenantID, vehicleID, dto)
	if err != nil {
		if errors.Is(err, fleet.ErrNotFound) {
			RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "Vehicle not found", nil)
			return
		}
		RespondError(w, r, http.StatusBadRequest, "INGEST_TELEMATICS_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusCreated, t, nil)
}

// GET /api/v1/vehicles/{vehicle_id}/telematics/latest
func (h *FleetHandler) GetLatestTelematics(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicle_id"))
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle UUID", nil)
		return
	}

	t, err := h.service.GetLatestTelematics(r.Context(), tenantID, vehicleID)
	if err != nil {
		if errors.Is(err, fleet.ErrNotFound) {
			RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "No telematics recorded for vehicle", nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "GET_TELEMATICS_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, t, nil)
}

// GET /api/v1/vehicles/{vehicle_id}/telematics/history
func (h *FleetHandler) GetTelematicsHistory(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicle_id"))
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle UUID", nil)
		return
	}

	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	history, err := h.service.GetTelematicsHistory(r.Context(), tenantID, vehicleID, limit)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, "GET_HISTORY_FAILED", err.Error(), nil)
		return
	}

	if history == nil {
		history = []fleet.VehicleTelematics{}
	}

	RespondJSON(w, r, http.StatusOK, history, nil)
}

// GET /api/v1/fleet/telematics/live or /api/v1/tenants/{tenant_id}/fleet/live
func (h *FleetHandler) GetLiveFleetPositions(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	positions, err := h.service.GetLiveFleetPositions(r.Context(), tenantID)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, "GET_LIVE_FLEET_FAILED", err.Error(), nil)
		return
	}

	if positions == nil {
		positions = []fleet.LiveVehiclePosition{}
	}

	RespondJSON(w, r, http.StatusOK, positions, nil)
}

// POST /api/v1/vehicles/{vehicle_id}/maintenance
func (h *FleetHandler) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicle_id"))
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle UUID", nil)
		return
	}

	var dto fleet.CreateMaintenanceDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to decode maintenance payload", map[string]any{"error": err.Error()})
		return
	}

	m, err := h.service.RecordMaintenance(r.Context(), tenantID, vehicleID, dto)
	if err != nil {
		if errors.Is(err, fleet.ErrNotFound) {
			RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "Vehicle not found", nil)
			return
		}
		RespondError(w, r, http.StatusBadRequest, "CREATE_MAINTENANCE_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusCreated, m, nil)
}

// GET /api/v1/vehicles/{vehicle_id}/maintenance
func (h *FleetHandler) ListMaintenance(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	vehicleID, err := uuid.Parse(chi.URLParam(r, "vehicle_id"))
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid vehicle UUID", nil)
		return
	}

	records, err := h.service.ListMaintenanceRecords(r.Context(), tenantID, vehicleID)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, "LIST_MAINTENANCE_FAILED", err.Error(), nil)
		return
	}

	if records == nil {
		records = []fleet.MaintenanceRecord{}
	}

	RespondJSON(w, r, http.StatusOK, records, nil)
}

// GET /api/v1/fleet/maintenance/upcoming or /api/v1/tenants/{tenant_id}/fleet/maintenance/upcoming
func (h *FleetHandler) ListUpcomingMaintenance(w http.ResponseWriter, r *http.Request) {
	tenantID, err := getTenantID(r)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", err.Error(), nil)
		return
	}

	vehicles, err := h.service.ListUpcomingMaintenance(r.Context(), tenantID)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, "LIST_UPCOMING_MAINTENANCE_FAILED", err.Error(), nil)
		return
	}

	if vehicles == nil {
		vehicles = []fleet.Vehicle{}
	}

	RespondJSON(w, r, http.StatusOK, vehicles, nil)
}
