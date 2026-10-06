package tenancy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
)

type Service interface {
	// Tenant
	CreateTenant(ctx context.Context, req *TenantCreateRequest) (*Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetTenantByCode(ctx context.Context, code string) (*Tenant, error)
	ListTenants(ctx context.Context) ([]Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, req *TenantUpdateRequest) (*Tenant, error)

	// Branch
	CreateBranch(ctx context.Context, tenantID uuid.UUID, req *BranchCreateRequest) (*Branch, error)
	GetBranchByID(ctx context.Context, tenantID, branchID uuid.UUID) (*Branch, error)
	ListBranches(ctx context.Context, tenantID uuid.UUID) ([]Branch, error)
	UpdateBranch(ctx context.Context, tenantID, branchID uuid.UUID, req *BranchUpdateRequest) (*Branch, error)
	DeleteBranch(ctx context.Context, tenantID, branchID uuid.UUID) error

	// Serviceability
	CheckServiceability(ctx context.Context, req *ServiceabilityCheckRequest) (*ServiceabilityCheckResponse, error)
	GetCoverageGeoJSON(ctx context.Context, tenantID uuid.UUID) (*GeoJSONFeatureCollection, error)
}

type service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) Service {
	return &service{
		repo:   repo,
		logger: logger,
	}
}

// Tenant methods
func (s *service) CreateTenant(ctx context.Context, req *TenantCreateRequest) (*Tenant, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify code is not already in use
	existing, _ := s.repo.GetTenantByCode(ctx, req.Code)
	if existing != nil {
		return nil, errors.New("tenant with this code already exists")
	}

	if req.Branding == nil {
		req.Branding = &TenantBranding{
			PrimaryColor: "#1e40af",
			AccentColor:  "#3b82f6",
			PortalName:   req.Name,
		}
	}
	if req.Config == nil {
		req.Config = &TenantConfig{
			Timezone:           "Asia/Kolkata",
			Currency:           "INR",
			AutoAssignDelivery: true,
			AllowedBranchTypes: []string{
				string(BranchTypeSortingHub),
				string(BranchTypeDistributionCenter),
				string(BranchTypeLocalOffice),
			},
		}
	}

	tenant, err := s.repo.CreateTenant(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create tenant", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.InfoContext(ctx, "Tenant created successfully", slog.String("tenant_id", tenant.ID.String()), slog.String("code", tenant.Code))
	return tenant, nil
}

func (s *service) GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	return s.repo.GetTenantByID(ctx, id)
}

func (s *service) GetTenantByCode(ctx context.Context, code string) (*Tenant, error) {
	return s.repo.GetTenantByCode(ctx, code)
}

func (s *service) ListTenants(ctx context.Context) ([]Tenant, error) {
	return s.repo.ListTenants(ctx)
}

func (s *service) UpdateTenant(ctx context.Context, id uuid.UUID, req *TenantUpdateRequest) (*Tenant, error) {
	return s.repo.UpdateTenant(ctx, id, req)
}

// Branch methods
func (s *service) CreateBranch(ctx context.Context, tenantID uuid.UUID, req *BranchCreateRequest) (*Branch, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required")
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if branch code exists for this tenant
	existing, _ := s.repo.GetBranchByCode(ctx, tenantID, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("branch code '%s' already exists for this tenant", req.Code)
	}

	// Validate GeoJSON format
	validatedGeoJSON, err := validateAndFormatPolygonGeoJSON(req.ServiceArea)
	if err != nil {
		return nil, fmt.Errorf("invalid service_area polygon: %w", err)
	}
	req.ServiceArea = validatedGeoJSON

	if req.OperatingHours == nil {
		req.OperatingHours = &OperatingHours{
			Open:     "08:00",
			Close:    "20:00",
			Timezone: "Asia/Kolkata",
		}
	}

	branch, err := s.repo.CreateBranch(ctx, tenantID, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to create branch", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.InfoContext(ctx, "Branch created successfully",
		slog.String("tenant_id", tenantID.String()),
		slog.String("branch_id", branch.ID.String()),
		slog.String("branch_code", branch.Code),
	)
	return branch, nil
}

func (s *service) GetBranchByID(ctx context.Context, tenantID, branchID uuid.UUID) (*Branch, error) {
	return s.repo.GetBranchByID(ctx, tenantID, branchID)
}

func (s *service) ListBranches(ctx context.Context, tenantID uuid.UUID) ([]Branch, error) {
	return s.repo.ListBranches(ctx, tenantID)
}

func (s *service) UpdateBranch(ctx context.Context, tenantID, branchID uuid.UUID, req *BranchUpdateRequest) (*Branch, error) {
	if req.Location != nil {
		if err := req.Location.Validate(); err != nil {
			return nil, err
		}
	}
	if req.ServiceArea != nil {
		validatedGeoJSON, err := validateAndFormatPolygonGeoJSON(*req.ServiceArea)
		if err != nil {
			return nil, fmt.Errorf("invalid service_area polygon: %w", err)
		}
		req.ServiceArea = &validatedGeoJSON
	}

	return s.repo.UpdateBranch(ctx, tenantID, branchID, req)
}

func (s *service) DeleteBranch(ctx context.Context, tenantID, branchID uuid.UUID) error {
	return s.repo.DeleteBranch(ctx, tenantID, branchID)
}

// Geospatial Serviceability Engine
func (s *service) CheckServiceability(ctx context.Context, req *ServiceabilityCheckRequest) (*ServiceabilityCheckResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 1. Check if coordinate falls inside any active branch's polygon service area
	matched, distMeters, err := s.repo.FindServiceableBranch(ctx, req.TenantID, req.Latitude, req.Longitude)
	if err != nil {
		return nil, fmt.Errorf("serviceability check failed: %w", err)
	}

	if matched != nil {
		eta := "Standard Next-Day"
		if distMeters < 5000 {
			eta = "Within 2 Hours (Ultra-Fast Hub)"
		} else if distMeters < 15000 {
			eta = "Same-Day Evening Delivery"
		}

		return &ServiceabilityCheckResponse{
			IsServiceable:  true,
			Status:         "SERVICEABLE",
			Message:        fmt.Sprintf("Location is fully serviceable by %s (%s)", matched.Name, matched.Code),
			MatchedBranch:  matched,
			DistanceMeters: distMeters,
			EstimatedETA:   eta,
		}, nil
	}

	// 2. Coordinate is not covered by any active polygon. Query nearest hub for user feedback
	nearest, nearestDist, err := s.repo.FindNearestBranch(ctx, req.TenantID, req.Latitude, req.Longitude)
	if err != nil {
		s.logger.WarnContext(ctx, "Failed to query nearest branch", slog.String("error", err.Error()))
	}

	msg := "Coordinates fall outside current serviceable delivery zones for this tenant."
	if nearest != nil {
		distKm := nearestDist / 1000.0
		msg = fmt.Sprintf("Outside service area. Nearest branch is %s (%s), approx %.1f km away.", nearest.Name, nearest.Code, distKm)
	}

	return &ServiceabilityCheckResponse{
		IsServiceable:  false,
		Status:         "OUT_OF_COVERAGE",
		Message:        msg,
		NearestBranch:  nearest,
		DistanceMeters: nearestDist,
	}, nil
}

func (s *service) GetCoverageGeoJSON(ctx context.Context, tenantID uuid.UUID) (*GeoJSONFeatureCollection, error) {
	return s.repo.GetCoverageGeoJSON(ctx, tenantID)
}

// GeoJSON polygon validation helper
func validateAndFormatPolygonGeoJSON(raw json.RawMessage) (json.RawMessage, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, errors.New("service_area must be a valid JSON object")
	}

	geomType, ok := parsed["type"].(string)
	if !ok {
		return nil, errors.New("service_area missing 'type' field (expected 'Polygon')")
	}

	if strings.ToUpper(geomType) != "POLYGON" {
		return nil, fmt.Errorf("service_area geometry type must be 'Polygon', got '%s'", geomType)
	}

	coordsRaw, ok := parsed["coordinates"]
	if !ok {
		return nil, errors.New("service_area polygon missing 'coordinates'")
	}

	// Ensure coordinates is a 3D array: [][][2]float64
	coordsBytes, err := json.Marshal(coordsRaw)
	if err != nil {
		return nil, err
	}

	var rings [][][2]float64
	if err := json.Unmarshal(coordsBytes, &rings); err != nil {
		return nil, fmt.Errorf("coordinates must be an array of rings [[[lon, lat], ...]]: %w", err)
	}

	if len(rings) == 0 || len(rings[0]) < 4 {
		return nil, errors.New("polygon must have at least one outer ring with at least 4 coordinate pairs")
	}

	// In GeoJSON, first and last coordinate pair must match to close the polygon
	first := rings[0][0]
	last := rings[0][len(rings[0])-1]
	if first[0] != last[0] || first[1] != last[1] {
		// Auto-close polygon ring
		rings[0] = append(rings[0], first)
		parsed["coordinates"] = rings
		return json.Marshal(parsed)
	}

	return raw, nil
}
