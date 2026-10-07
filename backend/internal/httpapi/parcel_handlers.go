package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/middleware"
	"logiflows/backend/internal/parcel"
)

type ParcelHandler struct {
	service       parcel.Service
	signingSecret string
}

func NewParcelHandler(service parcel.Service, signingSecret string) *ParcelHandler {
	return &ParcelHandler{
		service:       service,
		signingSecret: signingSecret,
	}
}

// BookParcel handles booking a new consignment
func (h *ParcelHandler) BookParcel(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req parcel.BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Malformed request payload", nil)
		return
	}

	// Auto-assign tenant from claims if not provided or caller is tenant-scoped
	if claims.TenantID != nil && *claims.TenantID != uuid.Nil {
		req.TenantID = *claims.TenantID
	} else if req.TenantID == uuid.Nil {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TENANT", "Tenant ID is required for parcel booking", nil)
		return
	}

	// If customer, sender ID is strictly the caller's ID
	if claims.Role == auth.RoleCustomer {
		req.SenderID = claims.UserID
	} else if req.SenderID == uuid.Nil {
		req.SenderID = claims.UserID
	}

	p, err := h.service.BookParcel(r.Context(), req, h.signingSecret)
	if err != nil {
		if errors.Is(err, parcel.ErrInvalidWeight) || errors.Is(err, parcel.ErrInvalidServiceType) || errors.Is(err, parcel.ErrInvalidInput) {
			RespondError(w, r, http.StatusBadRequest, "INVALID_BOOKING_DATA", err.Error(), nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "BOOKING_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusCreated, p, map[string]interface{}{
		"tracking_number": p.TrackingNumber,
		"qr_token":        p.QRToken,
	})
}

// ListParcels returns filtered parcels
func (h *ParcelHandler) ListParcels(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var tenantID uuid.UUID
	if claims.TenantID != nil {
		tenantID = *claims.TenantID
	} else if qTenant := r.URL.Query().Get("tenant_id"); qTenant != "" {
		parsed, err := uuid.Parse(qTenant)
		if err == nil {
			tenantID = parsed
		}
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	filter := parcel.ParcelFilter{
		TenantID:    tenantID,
		SearchQuery: r.URL.Query().Get("q"),
		Limit:       limit,
		Offset:      offset,
	}

	if st := r.URL.Query().Get("status"); st != "" {
		status := parcel.ParcelStatus(st)
		filter.Status = &status
	}
	if svc := r.URL.Query().Get("service_type"); svc != "" {
		serviceType := parcel.ServiceType(svc)
		filter.ServiceType = &serviceType
	}

	parcels, total, err := h.service.ListParcels(r.Context(), filter, claims.Role, claims.UserID)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, "QUERY_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, parcels, map[string]interface{}{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetParcel retrieves parcel by ID with ownership enforcement
func (h *ParcelHandler) GetParcel(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	parcelID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid parcel UUID", nil)
		return
	}

	var tenantID uuid.UUID
	if claims.TenantID != nil {
		tenantID = *claims.TenantID
	}

	p, err := h.service.GetParcel(r.Context(), tenantID, parcelID, claims.Role, claims.UserID)
	if err != nil {
		if errors.Is(err, parcel.ErrParcelNotFound) {
			RespondError(w, r, http.StatusNotFound, "PARCEL_NOT_FOUND", "Parcel not found", nil)
			return
		}
		if errors.Is(err, parcel.ErrUnauthorizedAccess) {
			RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "You do not have permission to view this parcel", nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, p, nil)
}

// TrackParcel returns public/customer tracking data by tracking number
func (h *ParcelHandler) TrackParcel(w http.ResponseWriter, r *http.Request) {
	trackingNumber := chi.URLParam(r, "tracking_number")
	if trackingNumber == "" {
		RespondError(w, r, http.StatusBadRequest, "MISSING_TRACKING_NUMBER", "Tracking number is required", nil)
		return
	}

	p, err := h.service.GetParcelByTracking(r.Context(), trackingNumber)
	if err != nil {
		if errors.Is(err, parcel.ErrParcelNotFound) {
			RespondError(w, r, http.StatusNotFound, "PARCEL_NOT_FOUND", "No consignment found with tracking number: "+trackingNumber, nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "TRACKING_ERROR", err.Error(), nil)
		return
	}

	// Fetch timeline
	timeline, _ := h.service.GetCustodyTimeline(r.Context(), p.TenantID, p.ID, auth.RolePlatformAdmin, uuid.Nil)

	RespondJSON(w, r, http.StatusOK, map[string]interface{}{
		"tracking_number":       p.TrackingNumber,
		"status":                p.Status,
		"service_type":          p.ServiceType,
		"origin_city":           p.SenderCity,
		"destination_city":      p.RecipientCity,
		"estimated_delivery_at": p.EstimatedDeliveryAt,
		"delivered_at":          p.DeliveredAt,
		"timeline":              timeline,
	}, nil)
}

// UpdateStatus updates parcel status and appends custody event
func (h *ParcelHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	parcelID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid parcel UUID", nil)
		return
	}

	var req parcel.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Malformed status update payload", nil)
		return
	}

	var tenantID uuid.UUID
	if claims.TenantID != nil {
		tenantID = *claims.TenantID
	}

	updated, err := h.service.UpdateStatus(r.Context(), tenantID, parcelID, req, claims.UserID)
	if err != nil {
		if errors.Is(err, parcel.ErrParcelNotFound) {
			RespondError(w, r, http.StatusNotFound, "PARCEL_NOT_FOUND", "Parcel not found", nil)
			return
		}
		if errors.Is(err, parcel.ErrInvalidStateTransition) {
			RespondError(w, r, http.StatusConflict, "INVALID_STATE_TRANSITION", err.Error(), nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "UPDATE_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, updated, nil)
}

// ScanQR handles intake and verification via signed QR token
func (h *ParcelHandler) ScanQR(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var req parcel.ScanQRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Malformed scan payload", nil)
		return
	}

	if req.QRToken == "" {
		RespondError(w, r, http.StatusBadRequest, "MISSING_QR_TOKEN", "QR token is required", nil)
		return
	}

	var tenantID uuid.UUID
	if claims.TenantID != nil {
		tenantID = *claims.TenantID
	}

	scanned, err := h.service.ScanQR(r.Context(), tenantID, req, claims.UserID)
	if err != nil {
		if errors.Is(err, parcel.ErrParcelNotFound) {
			RespondError(w, r, http.StatusNotFound, "PARCEL_NOT_FOUND", "Invalid QR token or parcel not found", nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "SCAN_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, scanned, map[string]interface{}{
		"scanned_event": req.EventType,
	})
}

// GetCustodyTimeline retrieves custody history
func (h *ParcelHandler) GetCustodyTimeline(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	parcelID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid parcel UUID", nil)
		return
	}

	var tenantID uuid.UUID
	if claims.TenantID != nil {
		tenantID = *claims.TenantID
	}

	timeline, err := h.service.GetCustodyTimeline(r.Context(), tenantID, parcelID, claims.Role, claims.UserID)
	if err != nil {
		if errors.Is(err, parcel.ErrParcelNotFound) {
			RespondError(w, r, http.StatusNotFound, "PARCEL_NOT_FOUND", "Parcel not found", nil)
			return
		}
		if errors.Is(err, parcel.ErrUnauthorizedAccess) {
			RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "Access denied", nil)
			return
		}
		RespondError(w, r, http.StatusInternalServerError, "QUERY_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, timeline, map[string]interface{}{
		"events_count": len(timeline),
	})
}

// Address Book Handlers
func (h *ParcelHandler) CreateAddress(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	var addr parcel.Address
	if err := json.NewDecoder(r.Body).Decode(&addr); err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_JSON", "Malformed address payload", nil)
		return
	}
	addr.CustomerID = claims.UserID
	addr.TenantID = claims.TenantID

	created, err := h.service.CreateAddress(r.Context(), &addr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ADDRESS", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusCreated, created, nil)
}

func (h *ParcelHandler) ListAddresses(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	addresses, err := h.service.ListCustomerAddresses(r.Context(), claims.UserID)
	if err != nil {
		RespondError(w, r, http.StatusInternalServerError, "QUERY_FAILED", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, addresses, map[string]interface{}{
		"count": len(addresses),
	})
}

func (h *ParcelHandler) DeleteAddress(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserClaims(r.Context())
	if claims == nil {
		RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	idStr := chi.URLParam(r, "id")
	addrID, err := uuid.Parse(idStr)
	if err != nil {
		RespondError(w, r, http.StatusBadRequest, "INVALID_ID", "Invalid address UUID", nil)
		return
	}

	if err := h.service.DeleteAddress(r.Context(), claims.UserID, addrID); err != nil {
		RespondError(w, r, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		return
	}

	RespondJSON(w, r, http.StatusOK, map[string]string{"message": "Address deleted"}, nil)
}
