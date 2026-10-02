package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/middleware"
	"logiflows/backend/internal/tenancy"
)

type TenancyHandler struct {
	service tenancy.Service
}

func NewTenancyHandler(service tenancy.Service) *TenancyHandler {
	return &TenancyHandler{service: service}
}

// POST /api/v1/tenants (Platform Admin only)
func (h *TenancyHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req tenancy.TenantCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to decode tenant payload", map[string]interface{}{"error": err.Error()})
		return
	}

	tenant, err := h.service.CreateTenant(r.Context(), &req)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "CREATE_TENANT_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusCreated, tenant, nil)
}

// GET /api/v1/tenants
func (h *TenancyHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authentication claims", nil)
		return
	}

	// Platform Admin sees all tenants; others see only their assigned tenant
	if claims.Role == auth.RolePlatformAdmin {
		tenants, err := h.service.ListTenants(r.Context())
		if err != nil {
			RespondError(w, r, http.StatusInternalServerError, "LIST_TENANTS_FAILED", err.Error(), nil)
			return
		}
		RespondJSON(w, r, http.StatusOK, tenants, nil)
		return
	}

	if claims.TenantID == nil {
		RespondJSON(w, r, http.StatusOK, []tenancy.Tenant{}, nil)
		return
	}

	tenant, err := h.service.GetTenantByID(r.Context(), *claims.TenantID)
	if err != nil {
		RespondError(w, r, http.StatusNotFound, "TENANT_NOT_FOUND", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, []tenancy.Tenant{*tenant}, nil)
}

// GET /api/v1/tenants/{tenant_id}
func (h *TenancyHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	tenant, err := h.service.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		RespondError(w, r, http.StatusNotFound, "TENANT_NOT_FOUND", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, tenant, nil)
}

// PUT /api/v1/tenants/{tenant_id}
func (h *TenancyHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	var req tenancy.TenantUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to decode tenant update payload", map[string]interface{}{"error": err.Error()})
		return
	}

	tenant, err := h.service.UpdateTenant(r.Context(), tenantID, &req)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "UPDATE_TENANT_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, tenant, nil)
}

// Branch Operations

// POST /api/v1/tenants/{tenant_id}/branches
func (h *TenancyHandler) CreateBranch(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	var req tenancy.BranchCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to decode branch payload", map[string]interface{}{"error": err.Error()})
		return
	}

	branch, err := h.service.CreateBranch(r.Context(), tenantID, &req)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "CREATE_BRANCH_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusCreated, branch, nil)
}

// GET /api/v1/tenants/{tenant_id}/branches
func (h *TenancyHandler) ListBranches(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	branches, err := h.service.ListBranches(r.Context(), tenantID)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, "LIST_BRANCHES_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, branches, nil)
}

// GET /api/v1/tenants/{tenant_id}/branches/{branch_id}
func (h *TenancyHandler) GetBranch(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	branchIDStr := chi.URLParam(r, "branch_id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_BRANCH_ID", "Invalid branch UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	branch, err := h.service.GetBranchByID(r.Context(), tenantID, branchID)
	if err != nil {
		RespondError(w, r, http.StatusNotFound, "BRANCH_NOT_FOUND", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, branch, nil)
}

// PUT /api/v1/tenants/{tenant_id}/branches/{branch_id}
func (h *TenancyHandler) UpdateBranch(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	branchIDStr := chi.URLParam(r, "branch_id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_BRANCH_ID", "Invalid branch UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	var req tenancy.BranchUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to decode branch update payload", map[string]interface{}{"error": err.Error()})
		return
	}

	branch, err := h.service.UpdateBranch(r.Context(), tenantID, branchID, &req)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "UPDATE_BRANCH_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, branch, nil)
}

// DELETE /api/v1/tenants/{tenant_id}/branches/{branch_id}
func (h *TenancyHandler) DeleteBranch(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	branchIDStr := chi.URLParam(r, "branch_id")
	branchID, err := uuid.Parse(branchIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_BRANCH_ID", "Invalid branch UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := h.service.DeleteBranch(r.Context(), tenantID, branchID); err != nil {
		RespondError(w, r, http.StatusNotFound, "DELETE_BRANCH_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, map[string]string{
		"message": "Branch deleted successfully",
	}, nil)
}

// Geospatial Serviceability Handlers

// POST /api/v1/serviceability/check
func (h *TenancyHandler) CheckServiceability(w http.ResponseWriter, r *http.Request) {
	var req tenancy.ServiceabilityCheckRequest

	// Support query parameters for easy GET testing if body empty
	if r.Method == http.MethodGet {
		tIDStr := r.URL.Query().Get("tenant_id")
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")
		if tIDStr != "" && latStr != "" && lonStr != "" {
			req.TenantID, _ = uuid.Parse(tIDStr)
			req.Latitude, _ = strconv.ParseFloat(latStr, 64)
			req.Longitude, _ = strconv.ParseFloat(lonStr, 64)
			req.Address = r.URL.Query().Get("address")
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to decode serviceability check payload", map[string]interface{}{"error": err.Error()})
			return
		}
	}

	res, err := h.service.CheckServiceability(r.Context(), &req)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "SERVICEABILITY_CHECK_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, res, nil)
}

// GET /api/v1/tenants/{tenant_id}/serviceability/coverage
func (h *TenancyHandler) GetCoverage(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := chi.URLParam(r, "tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_TENANT_ID", "Invalid tenant UUID", map[string]interface{}{"error": err.Error()})
		return
	}

	coverage, err := h.service.GetCoverageGeoJSON(r.Context(), tenantID)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, "COVERAGE_QUERY_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, coverage, nil)
}
