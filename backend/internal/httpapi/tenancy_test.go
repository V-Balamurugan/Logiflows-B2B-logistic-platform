package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/tenancy"
)

type mockTenancyService struct {
	tenants  map[uuid.UUID]*tenancy.Tenant
	branches map[uuid.UUID]*tenancy.Branch
}

func newMockTenancyService() *mockTenancyService {
	return &mockTenancyService{
		tenants:  make(map[uuid.UUID]*tenancy.Tenant),
		branches: make(map[uuid.UUID]*tenancy.Branch),
	}
}

func (m *mockTenancyService) CreateTenant(ctx context.Context, req *tenancy.TenantCreateRequest) (*tenancy.Tenant, error) {
	id := uuid.New()
	t := &tenancy.Tenant{
		ID:                 id,
		Name:               req.Name,
		Code:               req.Code,
		Slug:               req.Slug,
		Tier:               req.Tier,
		QuotaParcelsPerDay: req.QuotaParcelsPerDay,
		QuotaBranches:      req.QuotaBranches,
		IsActive:           true,
	}
	m.tenants[id] = t
	return t, nil
}

func (m *mockTenancyService) GetTenantByID(ctx context.Context, id uuid.UUID) (*tenancy.Tenant, error) {
	t, ok := m.tenants[id]
	if !ok {
		return &tenancy.Tenant{ID: id, Name: "Mock Tenant", Code: "MOCK", Slug: "mock", IsActive: true}, nil
	}
	return t, nil
}

func (m *mockTenancyService) GetTenantByCode(ctx context.Context, code string) (*tenancy.Tenant, error) {
	for _, t := range m.tenants {
		if t.Code == code {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockTenancyService) ListTenants(ctx context.Context) ([]tenancy.Tenant, error) {
	var list []tenancy.Tenant
	for _, t := range m.tenants {
		list = append(list, *t)
	}
	return list, nil
}

func (m *mockTenancyService) UpdateTenant(ctx context.Context, id uuid.UUID, req *tenancy.TenantUpdateRequest) (*tenancy.Tenant, error) {
	return m.GetTenantByID(ctx, id)
}

func (m *mockTenancyService) CreateBranch(ctx context.Context, tenantID uuid.UUID, req *tenancy.BranchCreateRequest) (*tenancy.Branch, error) {
	id := uuid.New()
	b := &tenancy.Branch{
		ID:            id,
		TenantID:      tenantID,
		Code:          req.Code,
		Name:          req.Name,
		BranchType:    req.BranchType,
		Address:       req.Address,
		DailyCapacity: req.DailyCapacity,
		Status:        req.Status,
		Location:      req.Location,
		ServiceArea:   req.ServiceArea,
	}
	m.branches[id] = b
	return b, nil
}

func (m *mockTenancyService) GetBranchByID(ctx context.Context, tenantID, branchID uuid.UUID) (*tenancy.Branch, error) {
	b, ok := m.branches[branchID]
	if !ok {
		return &tenancy.Branch{ID: branchID, TenantID: tenantID, Code: "B1", Name: "Mock Branch"}, nil
	}
	return b, nil
}

func (m *mockTenancyService) ListBranches(ctx context.Context, tenantID uuid.UUID) ([]tenancy.Branch, error) {
	var list []tenancy.Branch
	for _, b := range m.branches {
		if b.TenantID == tenantID {
			list = append(list, *b)
		}
	}
	return list, nil
}

func (m *mockTenancyService) UpdateBranch(ctx context.Context, tenantID, branchID uuid.UUID, req *tenancy.BranchUpdateRequest) (*tenancy.Branch, error) {
	return m.GetBranchByID(ctx, tenantID, branchID)
}

func (m *mockTenancyService) DeleteBranch(ctx context.Context, tenantID, branchID uuid.UUID) error {
	delete(m.branches, branchID)
	return nil
}

func (m *mockTenancyService) CheckServiceability(ctx context.Context, req *tenancy.ServiceabilityCheckRequest) (*tenancy.ServiceabilityCheckResponse, error) {
	// If Bangalore central (lat: 12.9716, lon: 77.5946), return serviceable
	if req.Latitude > 12.90 && req.Latitude < 13.05 && req.Longitude > 77.50 && req.Longitude < 77.70 {
		return &tenancy.ServiceabilityCheckResponse{
			IsServiceable: true,
			Status:        "SERVICEABLE",
			Message:       "Location is fully serviceable by Central Sorting Hub (BLR-HUB-01)",
			MatchedBranch: &tenancy.BranchSummary{
				ID:         uuid.New(),
				Code:       "BLR-HUB-01",
				Name:       "Bengaluru Central Sorting Hub",
				BranchType: tenancy.BranchTypeSortingHub,
				Address:    "Majestic, Bengaluru",
				Status:     tenancy.BranchStatusActive,
			},
			DistanceMeters: 1450.0,
			EstimatedETA:   "Within 2 Hours (Ultra-Fast Hub)",
		}, nil
	}

	return &tenancy.ServiceabilityCheckResponse{
		IsServiceable:  false,
		Status:         "OUT_OF_COVERAGE",
		Message:        "Outside service area. Nearest branch is BLR-HUB-01, approx 45.2 km away.",
		DistanceMeters: 45200.0,
	}, nil
}

func (m *mockTenancyService) GetCoverageGeoJSON(ctx context.Context, tenantID uuid.UUID) (*tenancy.GeoJSONFeatureCollection, error) {
	return &tenancy.GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: []tenancy.GeoJSONFeature{},
	}, nil
}

func TestServiceabilityCheckEndpoint(t *testing.T) {
	mockSvc := newMockTenancyService()
	cfg := &config.Config{JWTSecret: "test_secret_key_at_least_32_bytes_long"}

	router := BuildRouter(ServerDeps{
		Config:         cfg,
		TenancyService: mockSvc,
		Version:        "0.2.0-test",
	})

	t.Run("serviceable location in Bangalore returns 200 and serviceable true", func(t *testing.T) {
		reqBody := tenancy.ServiceabilityCheckRequest{
			TenantID:  uuid.New(),
			Latitude:  12.9716,
			Longitude: 77.5946,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/serviceability/check", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Data tenancy.ServiceabilityCheckResponse `json:"data"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if !resp.Data.IsServiceable {
			t.Fatal("expected location to be serviceable")
		}
		if resp.Data.MatchedBranch == nil || resp.Data.MatchedBranch.Code != "BLR-HUB-01" {
			t.Fatalf("expected matched branch BLR-HUB-01, got %+v", resp.Data.MatchedBranch)
		}
	})

	t.Run("location outside coverage area returns serviceable false", func(t *testing.T) {
		reqBody := tenancy.ServiceabilityCheckRequest{
			TenantID:  uuid.New(),
			Latitude:  15.5000,
			Longitude: 80.0000,
		}
		jsonBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/serviceability/check", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Data tenancy.ServiceabilityCheckResponse `json:"data"`
		}
		_ = json.NewDecoder(rr.Body).Decode(&resp)

		if resp.Data.IsServiceable {
			t.Fatal("expected location outside coverage to be unserviceable")
		}
		if resp.Data.Status != "OUT_OF_COVERAGE" {
			t.Fatalf("expected status OUT_OF_COVERAGE, got %s", resp.Data.Status)
		}
	})
}

func TestBranchTenantIsolationEndpoint(t *testing.T) {
	mockSvc := newMockTenancyService()
	cfg := &config.Config{JWTSecret: "test_secret_key_at_least_32_bytes_long"}

	router := BuildRouter(ServerDeps{
		Config:         cfg,
		TenancyService: mockSvc,
		Version:        "0.2.0-test",
	})

	tenantA := uuid.New()
	tenantB := uuid.New()

	tokenTenantA := createTestToken(t, uuid.New(), auth.RoleTenantAdmin, &tenantA, cfg.JWTSecret, false)

	// Tenant A user attempting to access Tenant B's branches MUST receive 403 Forbidden
	req := httptest.NewRequest("GET", "/api/v1/tenants/"+tenantB.String()+"/branches", nil)
	req.Header.Set("Authorization", "Bearer "+tokenTenantA)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for cross-tenant branch access, got %d", rr.Code)
	}

	// Tenant A user accessing Tenant A's own branches is allowed (200 OK)
	req2 := httptest.NewRequest("GET", "/api/v1/tenants/"+tenantA.String()+"/branches", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenTenantA)
	rr2 := httptest.NewRecorder()

	router.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for accessing own tenant's branches, got %d: %s", rr2.Code, rr2.Body.String())
	}
}
