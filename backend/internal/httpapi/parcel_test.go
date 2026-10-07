package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/parcel"
)

type mockParcelRepo struct {
	parcels      map[uuid.UUID]*parcel.Parcel
	byTracking   map[string]*parcel.Parcel
	byIdemp      map[string]*parcel.Parcel
	custodyEvents []*parcel.CustodyEvent
	addresses    map[uuid.UUID]*parcel.Address
}

func newMockParcelRepo() *mockParcelRepo {
	return &mockParcelRepo{
		parcels:    make(map[uuid.UUID]*parcel.Parcel),
		byTracking: make(map[string]*parcel.Parcel),
		byIdemp:    make(map[string]*parcel.Parcel),
		addresses:  make(map[uuid.UUID]*parcel.Address),
	}
}

func (m *mockParcelRepo) CreateParcel(ctx context.Context, p *parcel.Parcel) (*parcel.Parcel, error) {
	m.parcels[p.ID] = p
	m.byTracking[p.TrackingNumber] = p
	if p.IdempotencyKey != "" {
		m.byIdemp[p.IdempotencyKey] = p
	}
	return p, nil
}

func (m *mockParcelRepo) GetParcelByID(ctx context.Context, tenantID, id uuid.UUID) (*parcel.Parcel, error) {
	p, ok := m.parcels[id]
	if !ok || p.TenantID != tenantID {
		return nil, parcel.ErrParcelNotFound
	}
	return p, nil
}

func (m *mockParcelRepo) GetParcelByTracking(ctx context.Context, trackingNumber string) (*parcel.Parcel, error) {
	p, ok := m.byTracking[trackingNumber]
	if !ok {
		return nil, parcel.ErrParcelNotFound
	}
	return p, nil
}

func (m *mockParcelRepo) GetParcelByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (*parcel.Parcel, error) {
	p, ok := m.byIdemp[key]
	if !ok || p.TenantID != tenantID {
		return nil, nil
	}
	return p, nil
}

func (m *mockParcelRepo) ListParcels(ctx context.Context, filter parcel.ParcelFilter) ([]*parcel.Parcel, int, error) {
	var list []*parcel.Parcel
	for _, p := range m.parcels {
		if p.TenantID != filter.TenantID {
			continue
		}
		if filter.SenderID != nil && p.SenderID != *filter.SenderID {
			continue
		}
		list = append(list, p)
	}
	return list, len(list), nil
}

func (m *mockParcelRepo) UpdateParcelStatus(ctx context.Context, tenantID, parcelID uuid.UUID, newStatus parcel.ParcelStatus, actorID *uuid.UUID, branchID *uuid.UUID, notes string, lat, lon *float64) (*parcel.Parcel, error) {
	p, ok := m.parcels[parcelID]
	if !ok || p.TenantID != tenantID {
		return nil, parcel.ErrParcelNotFound
	}
	if !parcel.CanTransition(p.Status, newStatus) {
		return nil, parcel.ErrInvalidStateTransition
	}
	p.Status = newStatus
	return p, nil
}

func (m *mockParcelRepo) CreateAddress(ctx context.Context, a *parcel.Address) (*parcel.Address, error) {
	m.addresses[a.ID] = a
	return a, nil
}

func (m *mockParcelRepo) GetAddressByID(ctx context.Context, customerID, id uuid.UUID) (*parcel.Address, error) {
	a, ok := m.addresses[id]
	if !ok || a.CustomerID != customerID {
		return nil, assert.AnError
	}
	return a, nil
}

func (m *mockParcelRepo) ListCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*parcel.Address, error) {
	var res []*parcel.Address
	for _, a := range m.addresses {
		if a.CustomerID == customerID {
			res = append(res, a)
		}
	}
	return res, nil
}

func (m *mockParcelRepo) DeleteAddress(ctx context.Context, customerID, id uuid.UUID) error {
	delete(m.addresses, id)
	return nil
}

func (m *mockParcelRepo) AddCustodyEvent(ctx context.Context, e *parcel.CustodyEvent) error {
	m.custodyEvents = append(m.custodyEvents, e)
	return nil
}

func (m *mockParcelRepo) GetCustodyTimeline(ctx context.Context, parcelID uuid.UUID) ([]*parcel.CustodyEvent, error) {
	var res []*parcel.CustodyEvent
	for _, e := range m.custodyEvents {
		if e.ParcelID == parcelID {
			res = append(res, e)
		}
	}
	return res, nil
}

func setupTestParcelRouter() (http.Handler, *mockParcelRepo, *config.Config) {
	repo := newMockParcelRepo()
	svc := parcel.NewService(repo)
	cfg := &config.Config{
		JWTSecret: "test-secret-at-least-32-bytes-long-super-safe",
	}

	router := BuildRouter(ServerDeps{
		Config:        cfg,
		ParcelService: svc,
	})
	return router, repo, cfg
}

// 1. Booking endpoint (positive + validation)
func TestParcelAPI_BookParcel(t *testing.T) {
	router, _, cfg := setupTestParcelRouter()

	tenantID := uuid.New()
	customerID := uuid.New()
	token := createTestToken(t, customerID, auth.RoleCustomer, &tenantID, cfg.JWTSecret, false)

	bookingPayload := parcel.BookingRequest{
		ServiceType:         parcel.ServiceExpress,
		SenderName:          "Sender Enterprise",
		SenderPhone:         "+919876543210",
		SenderAddress:       "Warehouse 1",
		SenderCity:          "Bengaluru",
		SenderState:         "Karnataka",
		SenderPostalCode:    "560001",
		RecipientName:       "Recipient Corp",
		RecipientPhone:      "+919876543211",
		RecipientAddress:    "Retail 2",
		RecipientCity:       "Mumbai",
		RecipientState:      "Maharashtra",
		RecipientPostalCode: "400001",
		WeightKg:            2.0,
	}

	body, _ := json.Marshal(bookingPayload)
	req := httptest.NewRequest("POST", "/api/v1/parcels/book", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp SuccessEnvelope
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	dataMap := resp.Data.(map[string]interface{})
	assert.NotEmpty(t, dataMap["tracking_number"])
	assert.Equal(t, "CREATED", dataMap["status"])
}

// 2. Negative Test: Invalid weight rejected with 400 Bad Request
func TestParcelAPI_BookParcel_InvalidWeight(t *testing.T) {
	router, _, cfg := setupTestParcelRouter()

	tenantID := uuid.New()
	customerID := uuid.New()
	token := createTestToken(t, customerID, auth.RoleCustomer, &tenantID, cfg.JWTSecret, false)

	bookingPayload := parcel.BookingRequest{
		ServiceType:      parcel.ServiceStandard,
		SenderName:       "Sender",
		SenderAddress:    "Road 1",
		RecipientName:    "Recipient",
		RecipientAddress: "Road 2",
		WeightKg:         -5.0, // Invalid weight
	}

	body, _ := json.Marshal(bookingPayload)
	req := httptest.NewRequest("POST", "/api/v1/parcels/book", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// 3. Security: Unauthenticated request rejected with 401
func TestParcelAPI_UnauthenticatedRejected(t *testing.T) {
	router, _, _ := setupTestParcelRouter()

	req := httptest.NewRequest("GET", "/api/v1/parcels", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// 4. Public Tracking by Tracking Number
func TestParcelAPI_PublicTracking(t *testing.T) {
	router, repo, _ := setupTestParcelRouter()

	tenantID := uuid.New()
	senderID := uuid.New()
	trackingID := "LF-20261006-PUBLIC"

	p := &parcel.Parcel{
		ID:             uuid.New(),
		TenantID:       tenantID,
		TrackingNumber: trackingID,
		SenderID:       senderID,
		Status:         parcel.StatusInTransit,
		ServiceType:    parcel.ServiceExpress,
		SenderCity:     "Bengaluru",
		RecipientCity:  "Hyderabad",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, _ = repo.CreateParcel(context.Background(), p)

	// Public GET without any Authorization header
	req := httptest.NewRequest("GET", "/api/v1/parcels/track/"+trackingID, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp SuccessEnvelope
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	dataMap := resp.Data.(map[string]interface{})
	assert.Equal(t, trackingID, dataMap["tracking_number"])
	assert.Equal(t, "IN_TRANSIT", dataMap["status"])
	assert.Nil(t, dataMap["sender_address"], "Public tracking must never leak raw PII addresses")
}

// 5. State Machine Transition via PATCH /api/v1/parcels/{id}/status
func TestParcelAPI_StatusTransition_ValidAndInvalid(t *testing.T) {
	router, repo, cfg := setupTestParcelRouter()

	tenantID := uuid.New()
	adminID := uuid.New()
	token := createTestToken(t, adminID, auth.RoleTenantAdmin, &tenantID, cfg.JWTSecret, false)

	parcelID := uuid.New()
	p := &parcel.Parcel{
		ID:             parcelID,
		TenantID:       tenantID,
		TrackingNumber: "LF-20261006-STATE",
		SenderID:       uuid.New(),
		Status:         parcel.StatusCreated,
		ServiceType:    parcel.ServiceStandard,
	}
	_, _ = repo.CreateParcel(context.Background(), p)

	// 1. Legal transition: CREATED -> CONFIRMED (200 OK)
	validBody, _ := json.Marshal(parcel.UpdateStatusRequest{
		NewStatus: parcel.StatusConfirmed,
		Notes:     "Order confirmed by dispatch",
	})
	req := httptest.NewRequest("PATCH", "/api/v1/parcels/"+parcelID.String()+"/status", bytes.NewReader(validBody))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// 2. Illegal transition: Jump from CONFIRMED -> DELIVERED (409 Conflict)
	invalidBody, _ := json.Marshal(parcel.UpdateStatusRequest{
		NewStatus: parcel.StatusDelivered,
	})
	req2 := httptest.NewRequest("PATCH", "/api/v1/parcels/"+parcelID.String()+"/status", bytes.NewReader(invalidBody))
	req2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusConflict, rec2.Code)
}
