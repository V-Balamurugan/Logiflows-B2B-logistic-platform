package tenancy

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BranchType string

const (
	BranchTypeSortingHub          BranchType = "SORTING_HUB"
	BranchTypeDistributionCenter BranchType = "DISTRIBUTION_CENTER"
	BranchTypeLocalOffice         BranchType = "LOCAL_OFFICE"
)

func (b BranchType) IsValid() bool {
	switch b {
	case BranchTypeSortingHub, BranchTypeDistributionCenter, BranchTypeLocalOffice:
		return true
	default:
		return false
	}
}

type BranchStatus string

const (
	BranchStatusActive      BranchStatus = "ACTIVE"
	BranchStatusInactive    BranchStatus = "INACTIVE"
	BranchStatusMaintenance BranchStatus = "MAINTENANCE"
)

func (s BranchStatus) IsValid() bool {
	switch s {
	case BranchStatusActive, BranchStatusInactive, BranchStatusMaintenance:
		return true
	default:
		return false
	}
}

type TenantBranding struct {
	PrimaryColor string `json:"primary_color"`
	AccentColor  string `json:"accent_color"`
	PortalName   string `json:"portal_name"`
	LogoURL      string `json:"logo_url,omitempty"`
}

type TenantConfig struct {
	Timezone           string   `json:"timezone"`
	Currency           string   `json:"currency"`
	AutoAssignDelivery bool     `json:"auto_assign_delivery"`
	AllowedBranchTypes []string `json:"allowed_branch_types,omitempty"`
}

type Tenant struct {
	ID                 uuid.UUID       `json:"id"`
	Name               string          `json:"name"`
	Code               string          `json:"code"`
	Slug               string          `json:"slug"`
	Tier               string          `json:"tier"`
	QuotaParcelsPerDay int             `json:"quota_parcels_per_day"`
	QuotaBranches      int             `json:"quota_branches"`
	Branding           TenantBranding  `json:"branding"`
	Config             TenantConfig    `json:"config"`
	IsActive           bool            `json:"is_active"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type TenantCreateRequest struct {
	Name               string          `json:"name"`
	Code               string          `json:"code"`
	Slug               string          `json:"slug"`
	Tier               string          `json:"tier"`
	QuotaParcelsPerDay int             `json:"quota_parcels_per_day"`
	QuotaBranches      int             `json:"quota_branches"`
	Branding           *TenantBranding `json:"branding,omitempty"`
	Config             *TenantConfig   `json:"config,omitempty"`
}

func (r *TenantCreateRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("tenant name is required")
	}
	if strings.TrimSpace(r.Code) == "" {
		return errors.New("tenant code is required")
	}
	if r.Slug == "" {
		r.Slug = strings.ToLower(strings.TrimSpace(r.Code))
	}
	if r.Tier == "" {
		r.Tier = "STANDARD"
	}
	if r.QuotaParcelsPerDay <= 0 {
		r.QuotaParcelsPerDay = 5000
	}
	if r.QuotaBranches <= 0 {
		r.QuotaBranches = 50
	}
	return nil
}

type TenantUpdateRequest struct {
	Name               *string         `json:"name,omitempty"`
	Slug               *string         `json:"slug,omitempty"`
	Tier               *string         `json:"tier,omitempty"`
	QuotaParcelsPerDay *int            `json:"quota_parcels_per_day,omitempty"`
	QuotaBranches      *int            `json:"quota_branches,omitempty"`
	Branding           *TenantBranding `json:"branding,omitempty"`
	Config             *TenantConfig   `json:"config,omitempty"`
	IsActive           *bool           `json:"is_active,omitempty"`
}

type GeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (p GeoPoint) Validate() error {
	if p.Latitude < -90 || p.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %f", p.Latitude)
	}
	if p.Longitude < -180 || p.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %f", p.Longitude)
	}
	return nil
}

type OperatingHours struct {
	Open     string `json:"open"`
	Close    string `json:"close"`
	Timezone string `json:"timezone"`
}

type Branch struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	Code           string          `json:"code"`
	Name           string          `json:"name"`
	BranchType     BranchType      `json:"branch_type"`
	Address        string          `json:"address"`
	City           string          `json:"city"`
	State          string          `json:"state"`
	PostalCode     string          `json:"postal_code"`
	Country        string          `json:"country"`
	ContactPhone   string          `json:"contact_phone"`
	ContactEmail   string          `json:"contact_email"`
	OperatingHours OperatingHours  `json:"operating_hours"`
	DailyCapacity  int             `json:"daily_capacity"`
	Status         BranchStatus    `json:"status"`
	Location       GeoPoint        `json:"location"`
	ServiceArea    json.RawMessage `json:"service_area"` // GeoJSON Polygon
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type BranchSummary struct {
	ID           uuid.UUID    `json:"id"`
	Code         string       `json:"code"`
	Name         string       `json:"name"`
	BranchType   BranchType   `json:"branch_type"`
	Address      string       `json:"address"`
	City         string       `json:"city"`
	ContactPhone string       `json:"contact_phone"`
	Location     GeoPoint     `json:"location"`
	Status       BranchStatus `json:"status"`
}

type BranchCreateRequest struct {
	Code           string          `json:"code"`
	Name           string          `json:"name"`
	BranchType     BranchType      `json:"branch_type"`
	Address        string          `json:"address"`
	City           string          `json:"city"`
	State          string          `json:"state"`
	PostalCode     string          `json:"postal_code"`
	Country        string          `json:"country"`
	ContactPhone   string          `json:"contact_phone"`
	ContactEmail   string          `json:"contact_email"`
	OperatingHours *OperatingHours `json:"operating_hours,omitempty"`
	DailyCapacity  int             `json:"daily_capacity"`
	Status         BranchStatus    `json:"status"`
	Location       GeoPoint        `json:"location"`
	ServiceArea    json.RawMessage `json:"service_area"` // GeoJSON Polygon geometry or coordinates
}

func (r *BranchCreateRequest) Validate() error {
	if strings.TrimSpace(r.Code) == "" {
		return errors.New("branch code is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("branch name is required")
	}
	if !r.BranchType.IsValid() {
		return fmt.Errorf("invalid branch_type: %s", r.BranchType)
	}
	if strings.TrimSpace(r.Address) == "" {
		return errors.New("branch address is required")
	}
	if err := r.Location.Validate(); err != nil {
		return err
	}
	if r.DailyCapacity <= 0 {
		r.DailyCapacity = 1000
	}
	if r.Status == "" {
		r.Status = BranchStatusActive
	} else if !r.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	if len(r.ServiceArea) == 0 {
		return errors.New("service_area GeoJSON is required")
	}
	return nil
}

type BranchUpdateRequest struct {
	Name           *string          `json:"name,omitempty"`
	BranchType     *BranchType      `json:"branch_type,omitempty"`
	Address        *string          `json:"address,omitempty"`
	City           *string          `json:"city,omitempty"`
	State          *string          `json:"state,omitempty"`
	PostalCode     *string          `json:"postal_code,omitempty"`
	Country        *string          `json:"country,omitempty"`
	ContactPhone   *string          `json:"contact_phone,omitempty"`
	ContactEmail   *string          `json:"contact_email,omitempty"`
	OperatingHours *OperatingHours  `json:"operating_hours,omitempty"`
	DailyCapacity  *int             `json:"daily_capacity,omitempty"`
	Status         *BranchStatus    `json:"status,omitempty"`
	Location       *GeoPoint        `json:"location,omitempty"`
	ServiceArea    *json.RawMessage `json:"service_area,omitempty"`
}

type ServiceabilityCheckRequest struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Address   string    `json:"address,omitempty"`
}

func (r *ServiceabilityCheckRequest) Validate() error {
	if r.TenantID == uuid.Nil {
		return errors.New("tenant_id is required")
	}
	p := GeoPoint{Latitude: r.Latitude, Longitude: r.Longitude}
	return p.Validate()
}

type ServiceabilityCheckResponse struct {
	IsServiceable  bool           `json:"is_serviceable"`
	Status         string         `json:"status"` // SERVICEABLE, OUT_OF_COVERAGE, INACTIVE_BRANCH
	Message        string         `json:"message"`
	MatchedBranch  *BranchSummary `json:"matched_branch,omitempty"`
	NearestBranch  *BranchSummary `json:"nearest_branch,omitempty"`
	DistanceMeters float64        `json:"distance_meters"`
	EstimatedETA   string         `json:"estimated_eta,omitempty"`
}

// GeoJSON Types for Map Visualization
type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Geometry   json.RawMessage        `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}
