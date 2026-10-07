package parcel

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type ParcelStatus string

const (
	StatusCreated          ParcelStatus = "CREATED"
	StatusConfirmed        ParcelStatus = "CONFIRMED"
	StatusReceivedAtBranch ParcelStatus = "RECEIVED_AT_BRANCH"
	StatusAssigned         ParcelStatus = "ASSIGNED"
	StatusPickedUp         ParcelStatus = "PICKED_UP"
	StatusInTransit        ParcelStatus = "IN_TRANSIT"
	StatusOutForDelivery   ParcelStatus = "OUT_FOR_DELIVERY"
	StatusDelivered        ParcelStatus = "DELIVERED"
	StatusCancelled        ParcelStatus = "CANCELLED"
	StatusFailed           ParcelStatus = "FAILED"
	StatusReturned         ParcelStatus = "RETURNED"
	StatusOnHold           ParcelStatus = "ON_HOLD"
	StatusException        ParcelStatus = "EXCEPTION"
)

// AllStatuses returns the list of all valid parcel statuses
var AllStatuses = []ParcelStatus{
	StatusCreated, StatusConfirmed, StatusReceivedAtBranch, StatusAssigned,
	StatusPickedUp, StatusInTransit, StatusOutForDelivery, StatusDelivered,
	StatusCancelled, StatusFailed, StatusReturned, StatusOnHold, StatusException,
}

// ValidTransitions defines the authoritative table-driven state machine
var ValidTransitions = map[ParcelStatus]map[ParcelStatus]bool{
	StatusCreated: {
		StatusConfirmed: true,
		StatusCancelled: true,
	},
	StatusConfirmed: {
		StatusReceivedAtBranch: true,
		StatusAssigned:         true,
		StatusCancelled:        true,
		StatusOnHold:           true,
	},
	StatusReceivedAtBranch: {
		StatusAssigned:   true,
		StatusInTransit:  true,
		StatusOnHold:     true,
		StatusException:  true,
		StatusCancelled:  true,
	},
	StatusAssigned: {
		StatusPickedUp:   true,
		StatusConfirmed:  true, // Courier rejected/reassigned
		StatusOnHold:     true,
		StatusCancelled:  true,
	},
	StatusPickedUp: {
		StatusReceivedAtBranch: true,
		StatusInTransit:        true,
		StatusOutForDelivery:   true,
		StatusException:        true,
		StatusOnHold:           true,
	},
	StatusInTransit: {
		StatusReceivedAtBranch: true,
		StatusOutForDelivery:   true,
		StatusException:        true,
		StatusOnHold:           true,
	},
	StatusOutForDelivery: {
		StatusDelivered: true,
		StatusFailed:    true,
		StatusReturned:  true,
		StatusException: true,
		StatusOnHold:    true,
	},
	StatusFailed: {
		StatusOutForDelivery:   true, // Retry attempt
		StatusReturned:         true,
		StatusReceivedAtBranch: true,
	},
	StatusOnHold: {
		StatusConfirmed:        true,
		StatusReceivedAtBranch: true,
		StatusAssigned:         true,
		StatusPickedUp:         true,
		StatusInTransit:        true,
		StatusOutForDelivery:   true,
		StatusCancelled:        true,
	},
	StatusException: {
		StatusReceivedAtBranch: true,
		StatusOutForDelivery:   true,
		StatusReturned:         true,
		StatusOnHold:           true,
	},
	StatusDelivered: {}, // Terminal
	StatusReturned:  {}, // Terminal
	StatusCancelled: {}, // Terminal
}

// CanTransition checks if moving from 'current' to 'next' is legal
func CanTransition(current, next ParcelStatus) bool {
	if current == next {
		return false
	}
	allowed, ok := ValidTransitions[current]
	if !ok {
		return false
	}
	return allowed[next]
}

type ServiceType string

const (
	ServiceStandard  ServiceType = "STANDARD"
	ServiceExpress   ServiceType = "EXPRESS"
	ServiceSameDay   ServiceType = "SAME_DAY"
	ServiceColdChain ServiceType = "COLD_CHAIN"
	ServiceFreight   ServiceType = "FREIGHT"
)

func (s ServiceType) IsValid() bool {
	switch s {
	case ServiceStandard, ServiceExpress, ServiceSameDay, ServiceColdChain, ServiceFreight:
		return true
	default:
		return false
	}
}

type Parcel struct {
	ID                  uuid.UUID    `json:"id"`
	TenantID            uuid.UUID    `json:"tenant_id"`
	TrackingNumber      string       `json:"tracking_number"`
	SenderID            uuid.UUID    `json:"sender_id"`
	OriginBranchID      *uuid.UUID   `json:"origin_branch_id,omitempty"`
	DestinationBranchID *uuid.UUID   `json:"destination_branch_id,omitempty"`
	CurrentBranchID     *uuid.UUID   `json:"current_branch_id,omitempty"`
	ServiceType         ServiceType  `json:"service_type"`
	Status              ParcelStatus `json:"status"`

	// Sender Details
	SenderName       string   `json:"sender_name"`
	SenderPhone      string   `json:"sender_phone"`
	SenderEmail      string   `json:"sender_email,omitempty"`
	SenderAddress    string   `json:"sender_address"`
	SenderCity       string   `json:"sender_city"`
	SenderState      string   `json:"sender_state"`
	SenderPostalCode string   `json:"sender_postal_code"`
	SenderLatitude   *float64 `json:"sender_latitude,omitempty"`
	SenderLongitude  *float64 `json:"sender_longitude,omitempty"`

	// Recipient Details
	RecipientName       string   `json:"recipient_name"`
	RecipientPhone      string   `json:"recipient_phone"`
	RecipientEmail      string   `json:"recipient_email,omitempty"`
	RecipientAddress    string   `json:"recipient_address"`
	RecipientCity       string   `json:"recipient_city"`
	RecipientState      string   `json:"recipient_state"`
	RecipientPostalCode string   `json:"recipient_postal_code"`
	RecipientLatitude   *float64 `json:"recipient_latitude,omitempty"`
	RecipientLongitude  *float64 `json:"recipient_longitude,omitempty"`

	// Dimensions & Weight
	WeightKg            float64 `json:"weight_kg"`
	LengthCm            float64 `json:"length_cm"`
	WidthCm             float64 `json:"width_cm"`
	HeightCm            float64 `json:"height_cm"`
	DeclaredValue       float64 `json:"declared_value"`
	Currency            string  `json:"currency"`
	IsFragile           bool    `json:"is_fragile"`
	SpecialInstructions string  `json:"special_instructions,omitempty"`

	// Commercial & Identification
	ShippingCost   float64    `json:"shipping_cost"`
	PaymentStatus  string     `json:"payment_status"`
	QRToken        string     `json:"qr_token"`
	IdempotencyKey string     `json:"idempotency_key,omitempty"`

	// Lifecycle Timestamps
	EstimatedDeliveryAt *time.Time `json:"estimated_delivery_at,omitempty"`
	DeliveredAt         *time.Time `json:"delivered_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type CustodyEvent struct {
	ID             uuid.UUID    `json:"id"`
	ParcelID       uuid.UUID    `json:"parcel_id"`
	TenantID       uuid.UUID    `json:"tenant_id"`
	ActorID        *uuid.UUID   `json:"actor_id,omitempty"`
	ActorName      string       `json:"actor_name,omitempty"`
	BranchID       *uuid.UUID   `json:"branch_id,omitempty"`
	BranchName     string       `json:"branch_name,omitempty"`
	EventType      string       `json:"event_type"`
	PreviousStatus ParcelStatus `json:"previous_status,omitempty"`
	NewStatus      ParcelStatus `json:"new_status,omitempty"`
	Latitude       *float64     `json:"latitude,omitempty"`
	Longitude      *float64     `json:"longitude,omitempty"`
	LocationName   string       `json:"location_name,omitempty"`
	DeviceInfo     string       `json:"device_info,omitempty"`
	Notes          string       `json:"notes,omitempty"`
	RecordedAt     time.Time    `json:"recorded_at"`
}

type Address struct {
	ID                uuid.UUID  `json:"id"`
	CustomerID        uuid.UUID  `json:"customer_id"`
	TenantID          *uuid.UUID `json:"tenant_id,omitempty"`
	Label             string     `json:"label"`
	ContactName       string     `json:"contact_name"`
	ContactPhone      string     `json:"contact_phone"`
	ContactEmail      string     `json:"contact_email,omitempty"`
	StreetLine1       string     `json:"street_line1"`
	StreetLine2       string     `json:"street_line2,omitempty"`
	City              string     `json:"city"`
	State             string     `json:"state"`
	PostalCode        string     `json:"postal_code"`
	Country           string     `json:"country"`
	Latitude          *float64   `json:"latitude,omitempty"`
	Longitude         *float64   `json:"longitude,omitempty"`
	IsDefaultPickup   bool       `json:"is_default_pickup"`
	IsDefaultDelivery bool       `json:"is_default_delivery"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// GenerateTrackingNumber creates an authoritative unique tracking identifier: LF-YYYYMMDD-XXXXXX
func GenerateTrackingNumber() string {
	dateStr := time.Now().UTC().Format("20060102")
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude confusing chars O, 0, 1, I
	result := make([]byte, 6)
	for i := 0; i < 6; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[num.Int64()]
	}
	return fmt.Sprintf("LF-%s-%s", dateStr, string(result))
}

// GenerateQRToken produces a cryptographically signed unguessable opaque token (no raw PII)
func GenerateQRToken(trackingNumber string, secret string) string {
	if secret == "" {
		secret = "logiflows-secure-qr-signing-key"
	}
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)
	nonceHex := hex.EncodeToString(nonce)

	payload := fmt.Sprintf("%s:%s:%d", trackingNumber, nonceHex, time.Now().UnixNano())
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	return fmt.Sprintf("LFTK-%s-%s", nonceHex[:8], signature[:32])
}

// CalculateShippingCost evaluates estimated shipping charge based on weight and service level
func CalculateShippingCost(weightKg float64, service ServiceType) float64 {
	baseRate := 80.0
	switch service {
	case ServiceSameDay:
		baseRate = 250.0
	case ServiceExpress:
		baseRate = 150.0
	case ServiceColdChain:
		baseRate = 300.0
	case ServiceFreight:
		baseRate = 500.0
	}

	weightMultiplier := 40.0
	if weightKg > 1.0 {
		baseRate += (weightKg - 1.0) * weightMultiplier
	}
	return baseRate
}
