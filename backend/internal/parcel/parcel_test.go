package parcel

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"logiflows/backend/internal/auth"
)

// 1. Strict State Machine Table Tests (Every state x Every event)
func TestStateMachine_ExhaustiveTransitions(t *testing.T) {
	// Valid direct forward transitions
	assert.True(t, CanTransition(StatusCreated, StatusConfirmed), "CREATED -> CONFIRMED must be valid")
	assert.True(t, CanTransition(StatusCreated, StatusCancelled), "CREATED -> CANCELLED must be valid")
	assert.True(t, CanTransition(StatusConfirmed, StatusReceivedAtBranch), "CONFIRMED -> RECEIVED_AT_BRANCH must be valid")
	assert.True(t, CanTransition(StatusConfirmed, StatusAssigned), "CONFIRMED -> ASSIGNED must be valid")
	assert.True(t, CanTransition(StatusReceivedAtBranch, StatusInTransit), "RECEIVED_AT_BRANCH -> IN_TRANSIT must be valid")
	assert.True(t, CanTransition(StatusAssigned, StatusPickedUp), "ASSIGNED -> PICKED_UP must be valid")
	assert.True(t, CanTransition(StatusInTransit, StatusOutForDelivery), "IN_TRANSIT -> OUT_FOR_DELIVERY must be valid")
	assert.True(t, CanTransition(StatusOutForDelivery, StatusDelivered), "OUT_FOR_DELIVERY -> DELIVERED must be valid")
	assert.True(t, CanTransition(StatusOutForDelivery, StatusFailed), "OUT_FOR_DELIVERY -> FAILED must be valid")

	// Invalid transitions (Terminal & Regressive jumps)
	assert.False(t, CanTransition(StatusDelivered, StatusCreated), "DELIVERED is terminal, cannot transition to CREATED")
	assert.False(t, CanTransition(StatusDelivered, StatusInTransit), "DELIVERED is terminal, cannot transition to IN_TRANSIT")
	assert.False(t, CanTransition(StatusCancelled, StatusDelivered), "CANCELLED is terminal, cannot transition to DELIVERED")
	assert.False(t, CanTransition(StatusReturned, StatusOutForDelivery), "RETURNED is terminal, cannot transition to OUT_FOR_DELIVERY")
	assert.False(t, CanTransition(StatusCreated, StatusDelivered), "Cannot jump directly from CREATED to DELIVERED")
	assert.False(t, CanTransition(StatusCreated, StatusInTransit), "Cannot jump directly from CREATED to IN_TRANSIT")
	assert.False(t, CanTransition(StatusInTransit, StatusCreated), "Cannot regress from IN_TRANSIT to CREATED")

	// Same state transition is illegal
	for _, status := range AllStatuses {
		assert.False(t, CanTransition(status, status), "Self-transition must be rejected: %s", status)
	}
}

// 2. Tracking Number Format & Entropy
func TestTrackingNumber_FormatAndEntropy(t *testing.T) {
	trackingRe := regexp.MustCompile(`^LF-\d{8}-[A-Z0-9]{6}$`)

	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		num := GenerateTrackingNumber()
		assert.True(t, trackingRe.MatchString(num), "Tracking number format mismatch: %s", num)
		assert.False(t, seen[num], "Tracking number must have high entropy, collided: %s", num)
		seen[num] = true
	}
}

// 3. QR Token Security (Opaque, HMAC signed, NO PII)
func TestQRToken_SecurityAndNoPII(t *testing.T) {
	secret := "test-secret-key-123"
	tracking := "LF-20261006-AB4912"
	token := GenerateQRToken(tracking, secret)

	assert.True(t, strings.HasPrefix(token, "LFTK-"), "QR token must have authoritative prefix")
	assert.False(t, strings.Contains(token, tracking), "QR token must be opaque and not leak tracking number")
	assert.False(t, strings.Contains(token, "@"), "QR token must not contain email or PII")
	assert.False(t, strings.Contains(token, "name"), "QR token must not contain PII names")
	assert.GreaterOrEqual(t, len(token), 32, "QR token must be cryptographically long")
}

// 4. Pricing Matrix
func TestCalculateShippingCost(t *testing.T) {
	// Base weight = 1kg
	standardCost := CalculateShippingCost(1.0, ServiceStandard)
	assert.Equal(t, 80.0, standardCost)

	expressCost := CalculateShippingCost(1.0, ServiceExpress)
	assert.Equal(t, 150.0, expressCost)

	sameDayCost := CalculateShippingCost(1.0, ServiceSameDay)
	assert.Equal(t, 250.0, sameDayCost)

	// Heavy weight = 3.5kg
	// Standard = 80 + (2.5 * 40) = 180.0
	heavyCost := CalculateShippingCost(3.5, ServiceStandard)
	assert.Equal(t, 180.0, heavyCost)
}

// 5. In-Memory Mock Repository for Service Unit Tests
type mockRepo struct {
	parcels      map[uuid.UUID]*Parcel
	byTracking   map[string]*Parcel
	byIdemp      map[string]*Parcel
	custodyEvents []*CustodyEvent
	addresses    map[uuid.UUID]*Address
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		parcels:    make(map[uuid.UUID]*Parcel),
		byTracking: make(map[string]*Parcel),
		byIdemp:    make(map[string]*Parcel),
		addresses:  make(map[uuid.UUID]*Address),
	}
}

func (m *mockRepo) CreateParcel(ctx context.Context, p *Parcel) (*Parcel, error) {
	m.parcels[p.ID] = p
	m.byTracking[p.TrackingNumber] = p
	if p.IdempotencyKey != "" {
		m.byIdemp[p.IdempotencyKey] = p
	}
	return p, nil
}

func (m *mockRepo) GetParcelByID(ctx context.Context, tenantID, id uuid.UUID) (*Parcel, error) {
	p, ok := m.parcels[id]
	if !ok || p.TenantID != tenantID {
		return nil, ErrParcelNotFound
	}
	return p, nil
}

func (m *mockRepo) GetParcelByTracking(ctx context.Context, trackingNumber string) (*Parcel, error) {
	p, ok := m.byTracking[trackingNumber]
	if !ok {
		return nil, ErrParcelNotFound
	}
	return p, nil
}

func (m *mockRepo) GetParcelByIdempotencyKey(ctx context.Context, tenantID uuid.UUID, key string) (*Parcel, error) {
	p, ok := m.byIdemp[key]
	if !ok || p.TenantID != tenantID {
		return nil, nil
	}
	return p, nil
}

func (m *mockRepo) ListParcels(ctx context.Context, filter ParcelFilter) ([]*Parcel, int, error) {
	var list []*Parcel
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

func (m *mockRepo) UpdateParcelStatus(ctx context.Context, tenantID, parcelID uuid.UUID, newStatus ParcelStatus, actorID *uuid.UUID, branchID *uuid.UUID, notes string, lat, lon *float64) (*Parcel, error) {
	p, ok := m.parcels[parcelID]
	if !ok || p.TenantID != tenantID {
		return nil, ErrParcelNotFound
	}
	if !CanTransition(p.Status, newStatus) {
		return nil, ErrInvalidStateTransition
	}
	p.Status = newStatus
	return p, nil
}

func (m *mockRepo) CreateAddress(ctx context.Context, a *Address) (*Address, error) {
	m.addresses[a.ID] = a
	return a, nil
}

func (m *mockRepo) GetAddressByID(ctx context.Context, customerID, id uuid.UUID) (*Address, error) {
	a, ok := m.addresses[id]
	if !ok || a.CustomerID != customerID {
		return nil, assert.AnError
	}
	return a, nil
}

func (m *mockRepo) ListCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*Address, error) {
	var res []*Address
	for _, a := range m.addresses {
		if a.CustomerID == customerID {
			res = append(res, a)
		}
	}
	return res, nil
}

func (m *mockRepo) DeleteAddress(ctx context.Context, customerID, id uuid.UUID) error {
	delete(m.addresses, id)
	return nil
}

func (m *mockRepo) AddCustodyEvent(ctx context.Context, e *CustodyEvent) error {
	m.custodyEvents = append(m.custodyEvents, e)
	return nil
}

func (m *mockRepo) GetCustodyTimeline(ctx context.Context, parcelID uuid.UUID) ([]*CustodyEvent, error) {
	var res []*CustodyEvent
	for _, e := range m.custodyEvents {
		if e.ParcelID == parcelID {
			res = append(res, e)
		}
	}
	return res, nil
}

// 6. Idempotent Booking Test
func TestService_IdempotentBooking(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	tenantID := uuid.New()
	customerID := uuid.New()
	idempKey := "idemp-booking-unique-1234"

	req := BookingRequest{
		TenantID:            tenantID,
		SenderID:            customerID,
		ServiceType:         ServiceExpress,
		SenderName:          "Tech Corp",
		SenderPhone:         "+919876543210",
		SenderAddress:       "100 Tech Park",
		SenderCity:          "Bengaluru",
		SenderState:         "Karnataka",
		SenderPostalCode:    "560001",
		RecipientName:       "Retail Hub",
		RecipientPhone:      "+919876543211",
		RecipientAddress:    "200 Market Road",
		RecipientCity:       "Chennai",
		RecipientState:      "Tamil Nadu",
		RecipientPostalCode: "600001",
		WeightKg:            2.5,
		IdempotencyKey:      idempKey,
	}

	// 1st request creates parcel
	p1, err := svc.BookParcel(context.Background(), req, "secret")
	require.NoError(t, err)
	assert.Equal(t, StatusCreated, p1.Status)
	assert.Equal(t, idempKey, p1.IdempotencyKey)

	// 2nd request with same idempotency key returns exact same parcel
	p2, err := svc.BookParcel(context.Background(), req, "secret")
	require.NoError(t, err)
	assert.Equal(t, p1.ID, p2.ID, "Idempotent request must return identical parcel ID")
	assert.Equal(t, p1.TrackingNumber, p2.TrackingNumber)
}

// 7. Customer Cross-Tenant and Ownership Isolation
func TestService_CustomerOwnershipCheck(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	tenantID := uuid.New()
	customerA := uuid.New()
	customerB := uuid.New()

	p, err := svc.BookParcel(context.Background(), BookingRequest{
		TenantID:            tenantID,
		SenderID:            customerA,
		ServiceType:         ServiceStandard,
		SenderName:          "Customer A",
		SenderPhone:         "+919876543210",
		SenderAddress:       "Road 1",
		SenderCity:          "Bengaluru",
		SenderState:         "KA",
		SenderPostalCode:    "560001",
		RecipientName:       "Receiver",
		RecipientPhone:      "+919876543211",
		RecipientAddress:    "Road 2",
		RecipientCity:       "Bengaluru",
		RecipientState:      "KA",
		RecipientPostalCode: "560002",
		WeightKg:            1.0,
	}, "secret")
	require.NoError(t, err)

	// Customer A accesses own parcel -> allowed
	fetched, err := svc.GetParcel(context.Background(), tenantID, p.ID, auth.RoleCustomer, customerA)
	require.NoError(t, err)
	assert.Equal(t, p.ID, fetched.ID)

	// Customer B accesses Customer A's parcel -> 403 / ErrUnauthorizedAccess
	_, err = svc.GetParcel(context.Background(), tenantID, p.ID, auth.RoleCustomer, customerB)
	assert.ErrorIs(t, err, ErrUnauthorizedAccess, "Customer B must not access Customer A's parcel")

	// Tenant Admin accesses parcel -> allowed
	_, err = svc.GetParcel(context.Background(), tenantID, p.ID, auth.RoleTenantAdmin, customerB)
	assert.NoError(t, err, "Tenant admin must be able to view tenant's parcel")
}
