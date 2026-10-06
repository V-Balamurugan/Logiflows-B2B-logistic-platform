package tenancy

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/database"
	"logiflows/backend/internal/logging"
)

func TestPostGISGeospatialIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" && os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping live database integration test (set INTEGRATION_TESTS=true)")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	logger := logging.NewLogger("debug", false)
	db, err := database.NewPostgresDB(cfg, logger)
	if err != nil || db == nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo := NewRepository(db)
	svc := NewService(repo, logger)

	tenant1ID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tenant2ID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	t.Run("Point inside BLR Central Hub polygon is serviceable", func(t *testing.T) {
		// MG Road, Bengaluru (12.9756, 77.6066)
		res, err := svc.CheckServiceability(ctx, &ServiceabilityCheckRequest{
			TenantID:  tenant1ID,
			Latitude:  12.9756,
			Longitude: 77.6066,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.IsServiceable {
			t.Fatalf("expected location inside Central polygon to be serviceable, got: %+v", res)
		}
		if res.MatchedBranch == nil || res.MatchedBranch.Code != "BLR-HUB-01" {
			t.Fatalf("expected matched branch BLR-HUB-01, got: %+v", res.MatchedBranch)
		}
	})

	t.Run("Point inside BLR North DC polygon is serviceable", func(t *testing.T) {
		// Hebbal, Bengaluru (13.0358, 77.5970)
		res, err := svc.CheckServiceability(ctx, &ServiceabilityCheckRequest{
			TenantID:  tenant1ID,
			Latitude:  13.0400,
			Longitude: 77.6000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.IsServiceable {
			t.Fatalf("expected location inside North polygon to be serviceable, got: %+v", res)
		}
		if res.MatchedBranch == nil || res.MatchedBranch.Code != "BLR-DC-NORTH" {
			t.Fatalf("expected matched branch BLR-DC-NORTH, got: %+v", res.MatchedBranch)
		}
	})

	t.Run("Point inside BLR South DC polygon is serviceable", func(t *testing.T) {
		// Koramangala 4th Block (12.9345, 77.6265)
		res, err := svc.CheckServiceability(ctx, &ServiceabilityCheckRequest{
			TenantID:  tenant1ID,
			Latitude:  12.9345,
			Longitude: 77.6265,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.IsServiceable {
			t.Fatalf("expected location inside South polygon to be serviceable, got: %+v", res)
		}
		if res.MatchedBranch == nil || res.MatchedBranch.Code != "BLR-DC-SOUTH" {
			t.Fatalf("expected matched branch BLR-DC-SOUTH, got: %+v", res.MatchedBranch)
		}
	})

	t.Run("Multi-tenant isolation: Chennai location is NOT serviceable by Bengaluru tenant", func(t *testing.T) {
		// Anna Salai, Chennai (13.0827, 80.2707)
		res1, err := svc.CheckServiceability(ctx, &ServiceabilityCheckRequest{
			TenantID:  tenant1ID,
			Latitude:  13.0827,
			Longitude: 80.2707,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res1.IsServiceable {
			t.Fatalf("Bengaluru tenant should NOT service Chennai coordinate, got: %+v", res1)
		}
		if res1.Status != "OUT_OF_COVERAGE" {
			t.Fatalf("expected OUT_OF_COVERAGE status, got %s", res1.Status)
		}

		// BUT for Tenant 2 (Metro Courier Services in Chennai), it IS serviceable!
		res2, err := svc.CheckServiceability(ctx, &ServiceabilityCheckRequest{
			TenantID:  tenant2ID,
			Latitude:  13.0827,
			Longitude: 80.2707,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res2.IsServiceable {
			t.Fatalf("Chennai tenant MUST service Chennai coordinate, got: %+v", res2)
		}
		if res2.MatchedBranch == nil || res2.MatchedBranch.Code != "CHN-HUB-01" {
			t.Fatalf("expected matched branch CHN-HUB-01, got: %+v", res2.MatchedBranch)
		}
	})

	t.Run("Coverage GeoJSON returns FeatureCollection for tenant", func(t *testing.T) {
		fc, err := svc.GetCoverageGeoJSON(ctx, tenant1ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fc.Type != "FeatureCollection" {
			t.Fatalf("expected FeatureCollection, got %s", fc.Type)
		}
		if len(fc.Features) != 3 {
			t.Fatalf("expected 3 features for tenant 1, got %d", len(fc.Features))
		}
	})
}
