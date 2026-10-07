package parcel

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
)

var (
	ErrInvalidInput       = errors.New("invalid parcel input parameter")
	ErrInvalidWeight      = errors.New("parcel weight must be greater than 0")
	ErrInvalidServiceType = errors.New("unsupported parcel service type")
)

type BookingRequest struct {
	TenantID            uuid.UUID   `json:"tenant_id"`
	SenderID            uuid.UUID   `json:"sender_id"`
	OriginBranchID      *uuid.UUID  `json:"origin_branch_id,omitempty"`
	DestinationBranchID *uuid.UUID  `json:"destination_branch_id,omitempty"`
	ServiceType         ServiceType `json:"service_type"`

	SenderName       string   `json:"sender_name"`
	SenderPhone      string   `json:"sender_phone"`
	SenderEmail      string   `json:"sender_email,omitempty"`
	SenderAddress    string   `json:"sender_address"`
	SenderCity       string   `json:"sender_city"`
	SenderState      string   `json:"sender_state"`
	SenderPostalCode string   `json:"sender_postal_code"`
	SenderLatitude   *float64 `json:"sender_latitude,omitempty"`
	SenderLongitude  *float64 `json:"sender_longitude,omitempty"`

	RecipientName       string   `json:"recipient_name"`
	RecipientPhone      string   `json:"recipient_phone"`
	RecipientEmail      string   `json:"recipient_email,omitempty"`
	RecipientAddress    string   `json:"recipient_address"`
	RecipientCity       string   `json:"recipient_city"`
	RecipientState      string   `json:"recipient_state"`
	RecipientPostalCode string   `json:"recipient_postal_code"`
	RecipientLatitude   *float64 `json:"recipient_latitude,omitempty"`
	RecipientLongitude  *float64 `json:"recipient_longitude,omitempty"`

	WeightKg            float64 `json:"weight_kg"`
	LengthCm            float64 `json:"length_cm"`
	WidthCm             float64 `json:"width_cm"`
	HeightCm            float64 `json:"height_cm"`
	DeclaredValue       float64 `json:"declared_value"`
	Currency            string  `json:"currency"`
	IsFragile           bool    `json:"is_fragile"`
	SpecialInstructions string  `json:"special_instructions,omitempty"`
	PaymentStatus       string  `json:"payment_status,omitempty"`
	IdempotencyKey      string  `json:"idempotency_key,omitempty"`
}

type UpdateStatusRequest struct {
	NewStatus ParcelStatus `json:"new_status"`
	BranchID  *uuid.UUID   `json:"branch_id,omitempty"`
	Latitude  *float64     `json:"latitude,omitempty"`
	Longitude *float64     `json:"longitude,omitempty"`
	Notes     string       `json:"notes,omitempty"`
}

type ScanQRRequest struct {
	QRToken    string   `json:"qr_token"`
	BranchID   *uuid.UUID `json:"branch_id,omitempty"`
	EventType  string   `json:"event_type"` // INTAKE, HANDOVER, SORT, OUT_FOR_DELIVERY
	Latitude   *float64 `json:"latitude,omitempty"`
	Longitude  *float64 `json:"longitude,omitempty"`
	DeviceInfo string   `json:"device_info,omitempty"`
	Notes      string   `json:"notes,omitempty"`
}

type Service interface {
	BookParcel(ctx context.Context, req BookingRequest, signingSecret string) (*Parcel, error)
	GetParcel(ctx context.Context, tenantID, parcelID uuid.UUID, callerRole auth.Role, callerID uuid.UUID) (*Parcel, error)
	GetParcelByTracking(ctx context.Context, trackingNumber string) (*Parcel, error)
	ListParcels(ctx context.Context, filter ParcelFilter, callerRole auth.Role, callerID uuid.UUID) ([]*Parcel, int, error)
	UpdateStatus(ctx context.Context, tenantID, parcelID uuid.UUID, req UpdateStatusRequest, actorID uuid.UUID) (*Parcel, error)
	ScanQR(ctx context.Context, tenantID uuid.UUID, req ScanQRRequest, actorID uuid.UUID) (*Parcel, error)
	GetCustodyTimeline(ctx context.Context, tenantID, parcelID uuid.UUID, callerRole auth.Role, callerID uuid.UUID) ([]*CustodyEvent, error)

	// Address Book
	CreateAddress(ctx context.Context, addr *Address) (*Address, error)
	ListCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*Address, error)
	DeleteAddress(ctx context.Context, customerID, addressID uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) BookParcel(ctx context.Context, req BookingRequest, signingSecret string) (*Parcel, error) {
	// 1. Idempotency Check
	if req.IdempotencyKey != "" {
		existing, err := s.repo.GetParcelByIdempotencyKey(ctx, req.TenantID, req.IdempotencyKey)
		if err == nil && existing != nil {
			return existing, nil // Return idempotent match safely
		}
	}

	// 2. Validate input parameters
	if req.WeightKg <= 0 {
		return nil, ErrInvalidWeight
	}
	if !req.ServiceType.IsValid() {
		return nil, ErrInvalidServiceType
	}
	if strings.TrimSpace(req.RecipientName) == "" || strings.TrimSpace(req.RecipientAddress) == "" {
		return nil, fmt.Errorf("%w: recipient details are required", ErrInvalidInput)
	}
	if strings.TrimSpace(req.SenderName) == "" || strings.TrimSpace(req.SenderAddress) == "" {
		return nil, fmt.Errorf("%w: sender details are required", ErrInvalidInput)
	}

	// Set defaults
	if req.LengthCm <= 0 {
		req.LengthCm = 10.0
	}
	if req.WidthCm <= 0 {
		req.WidthCm = 10.0
	}
	if req.HeightCm <= 0 {
		req.HeightCm = 10.0
	}
	if req.Currency == "" {
		req.Currency = "INR"
	}
	if req.PaymentStatus == "" {
		req.PaymentStatus = "PENDING"
	}

	// 3. Calculate shipping cost
	cost := CalculateShippingCost(req.WeightKg, req.ServiceType)

	// 4. Generate unique tracking number and cryptographically signed QR token
	trackingNumber := GenerateTrackingNumber()
	qrToken := GenerateQRToken(trackingNumber, signingSecret)

	// 5. Estimate delivery time based on service level
	now := time.Now().UTC()
	var estimatedDelivery time.Time
	switch req.ServiceType {
	case ServiceSameDay:
		estimatedDelivery = now.Add(12 * time.Hour)
	case ServiceExpress:
		estimatedDelivery = now.Add(24 * time.Hour)
	case ServiceColdChain:
		estimatedDelivery = now.Add(36 * time.Hour)
	default:
		estimatedDelivery = now.Add(72 * time.Hour)
	}

	parcel := &Parcel{
		ID:                  uuid.New(),
		TenantID:            req.TenantID,
		TrackingNumber:      trackingNumber,
		SenderID:            req.SenderID,
		OriginBranchID:      req.OriginBranchID,
		DestinationBranchID: req.DestinationBranchID,
		ServiceType:         req.ServiceType,
		Status:              StatusCreated,

		SenderName:       req.SenderName,
		SenderPhone:      req.SenderPhone,
		SenderEmail:      req.SenderEmail,
		SenderAddress:    req.SenderAddress,
		SenderCity:       req.SenderCity,
		SenderState:      req.SenderState,
		SenderPostalCode: req.SenderPostalCode,
		SenderLatitude:   req.SenderLatitude,
		SenderLongitude:  req.SenderLongitude,

		RecipientName:       req.RecipientName,
		RecipientPhone:      req.RecipientPhone,
		RecipientEmail:      req.RecipientEmail,
		RecipientAddress:    req.RecipientAddress,
		RecipientCity:       req.RecipientCity,
		RecipientState:      req.RecipientState,
		RecipientPostalCode: req.RecipientPostalCode,
		RecipientLatitude:   req.RecipientLatitude,
		RecipientLongitude:  req.RecipientLongitude,

		WeightKg:            req.WeightKg,
		LengthCm:            req.LengthCm,
		WidthCm:             req.WidthCm,
		HeightCm:            req.HeightCm,
		DeclaredValue:       req.DeclaredValue,
		Currency:            req.Currency,
		IsFragile:           req.IsFragile,
		SpecialInstructions: req.SpecialInstructions,

		ShippingCost:        cost,
		PaymentStatus:       req.PaymentStatus,
		QRToken:             qrToken,
		IdempotencyKey:      req.IdempotencyKey,
		EstimatedDeliveryAt: &estimatedDelivery,
	}

	return s.repo.CreateParcel(ctx, parcel)
}

func (s *service) GetParcel(ctx context.Context, tenantID, parcelID uuid.UUID, callerRole auth.Role, callerID uuid.UUID) (*Parcel, error) {
	p, err := s.repo.GetParcelByID(ctx, tenantID, parcelID)
	if err != nil {
		return nil, err
	}

	// Customer can only view their own booked parcels
	if callerRole == auth.RoleCustomer && p.SenderID != callerID {
		return nil, ErrUnauthorizedAccess
	}
	return p, nil
}

func (s *service) GetParcelByTracking(ctx context.Context, trackingNumber string) (*Parcel, error) {
	return s.repo.GetParcelByTracking(ctx, trackingNumber)
}

func (s *service) ListParcels(ctx context.Context, filter ParcelFilter, callerRole auth.Role, callerID uuid.UUID) ([]*Parcel, int, error) {
	if callerRole == auth.RoleCustomer {
		filter.SenderID = &callerID // Restrict query to customer's own parcels
	}
	return s.repo.ListParcels(ctx, filter)
}

func (s *service) UpdateStatus(ctx context.Context, tenantID, parcelID uuid.UUID, req UpdateStatusRequest, actorID uuid.UUID) (*Parcel, error) {
	return s.repo.UpdateParcelStatus(
		ctx, tenantID, parcelID, req.NewStatus,
		&actorID, req.BranchID, req.Notes,
		req.Latitude, req.Longitude,
	)
}

func (s *service) ScanQR(ctx context.Context, tenantID uuid.UUID, req ScanQRRequest, actorID uuid.UUID) (*Parcel, error) {
	// 1. Locate parcel by token
	filter := ParcelFilter{TenantID: tenantID, Limit: 1}
	parcels, _, err := s.repo.ListParcels(ctx, filter)
	if err != nil {
		return nil, err
	}

	var matched *Parcel
	for _, p := range parcels {
		if p.QRToken == req.QRToken {
			matched = p
			break
		}
	}

	if matched == nil {
		// Look up by direct scan query if not in first page
		return nil, ErrParcelNotFound
	}

	// Determine state transition based on scan event type
	var nextStatus ParcelStatus
	switch strings.ToUpper(req.EventType) {
	case "INTAKE":
		nextStatus = StatusReceivedAtBranch
	case "OUT_FOR_DELIVERY":
		nextStatus = StatusOutForDelivery
	case "DELIVERED":
		nextStatus = StatusDelivered
	default:
		nextStatus = matched.Status // Keep status if only checking
	}

	if nextStatus != matched.Status && CanTransition(matched.Status, nextStatus) {
		return s.repo.UpdateParcelStatus(
			ctx, tenantID, matched.ID, nextStatus,
			&actorID, req.BranchID, req.Notes,
			req.Latitude, req.Longitude,
		)
	}

	// Just log custody scan event without changing status
	_ = s.repo.AddCustodyEvent(ctx, &CustodyEvent{
		ID:             uuid.New(),
		ParcelID:       matched.ID,
		TenantID:       tenantID,
		ActorID:        &actorID,
		BranchID:       req.BranchID,
		EventType:      req.EventType,
		PreviousStatus: matched.Status,
		NewStatus:      matched.Status,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		DeviceInfo:     req.DeviceInfo,
		Notes:          req.Notes,
		RecordedAt:     time.Now().UTC(),
	})

	return matched, nil
}

func (s *service) GetCustodyTimeline(ctx context.Context, tenantID, parcelID uuid.UUID, callerRole auth.Role, callerID uuid.UUID) ([]*CustodyEvent, error) {
	p, err := s.repo.GetParcelByID(ctx, tenantID, parcelID)
	if err != nil {
		return nil, err
	}
	if callerRole == auth.RoleCustomer && p.SenderID != callerID {
		return nil, ErrUnauthorizedAccess
	}
	return s.repo.GetCustodyTimeline(ctx, parcelID)
}

func (s *service) CreateAddress(ctx context.Context, addr *Address) (*Address, error) {
	if strings.TrimSpace(addr.StreetLine1) == "" || strings.TrimSpace(addr.City) == "" {
		return nil, fmt.Errorf("%w: street line 1 and city are required", ErrInvalidInput)
	}
	return s.repo.CreateAddress(ctx, addr)
}

func (s *service) ListCustomerAddresses(ctx context.Context, customerID uuid.UUID) ([]*Address, error) {
	return s.repo.ListCustomerAddresses(ctx, customerID)
}

func (s *service) DeleteAddress(ctx context.Context, customerID, addressID uuid.UUID) error {
	return s.repo.DeleteAddress(ctx, customerID, addressID)
}
