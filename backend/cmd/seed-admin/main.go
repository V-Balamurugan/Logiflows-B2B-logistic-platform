package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/database"
	"logiflows/backend/internal/logging"
	"logiflows/backend/internal/migrations"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	logger := logging.NewLogger("info", false)
	logger.Info("Starting database seed tool")

	db, err := database.NewPostgresDB(cfg, logger)
	if err != nil || db == nil {
		logger.Error("Database connection failed", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ensure migrations are up to date
	if err := migrations.RunMigrations(ctx, db, logger); err != nil {
		logger.Error("Migrations failed during seed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 1. Seed Demo Tenant 1: Speedy Express Logistics (Bengaluru)
	tenant1ID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tenant1Name := "Speedy Express Logistics"
	tenant1Code := "speedy-express"
	tenant1Slug := "speedy-express"

	_, err = db.ExecContext(ctx, `
	INSERT INTO tenants (
		id, name, code, slug, tier, quota_parcels_per_day, quota_branches,
		branding, config, is_active
	)
	VALUES (
		$1, $2, $3, $4, 'ENTERPRISE', 50000, 100,
		'{"primary_color":"#2563eb","accent_color":"#60a5fa","portal_name":"Speedy Express Portal"}',
		'{"timezone":"Asia/Kolkata","currency":"INR","auto_assign_delivery":true}',
		true
	)
	ON CONFLICT (code) DO UPDATE SET
		name = EXCLUDED.name,
		slug = EXCLUDED.slug,
		tier = EXCLUDED.tier,
		quota_parcels_per_day = EXCLUDED.quota_parcels_per_day,
		quota_branches = EXCLUDED.quota_branches,
		branding = EXCLUDED.branding,
		config = EXCLUDED.config`,
		tenant1ID, tenant1Name, tenant1Code, tenant1Slug,
	)
	if err != nil {
		logger.Error("Failed to seed tenant 1", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Seeded demo tenant 1", slog.String("code", tenant1Code))

	// Seed Demo Tenant 2: Metro Courier Services (Chennai)
	tenant2ID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	tenant2Name := "Metro Courier Services"
	tenant2Code := "metro-couriers"
	tenant2Slug := "metro-couriers"

	_, err = db.ExecContext(ctx, `
	INSERT INTO tenants (
		id, name, code, slug, tier, quota_parcels_per_day, quota_branches,
		branding, config, is_active
	)
	VALUES (
		$1, $2, $3, $4, 'STANDARD', 10000, 20,
		'{"primary_color":"#059669","accent_color":"#34d399","portal_name":"Metro Couriers Hub"}',
		'{"timezone":"Asia/Kolkata","currency":"INR","auto_assign_delivery":false}',
		true
	)
	ON CONFLICT (code) DO UPDATE SET
		name = EXCLUDED.name,
		slug = EXCLUDED.slug,
		tier = EXCLUDED.tier,
		quota_parcels_per_day = EXCLUDED.quota_parcels_per_day,
		quota_branches = EXCLUDED.quota_branches,
		branding = EXCLUDED.branding,
		config = EXCLUDED.config`,
		tenant2ID, tenant2Name, tenant2Code, tenant2Slug,
	)
	if err != nil {
		logger.Error("Failed to seed tenant 2", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Seeded demo tenant 2", slog.String("code", tenant2Code))

	// 2. Seed Users
	users := []struct {
		Email     string
		Password  string
		FirstName string
		LastName  string
		Role      auth.Role
		TenantID  *uuid.UUID
	}{
		{
			Email:     "admin@logiflows.io",
			Password:  "Admin@LogiFlows2026!",
			FirstName: "Platform",
			LastName:  "SuperAdmin",
			Role:      auth.RolePlatformAdmin,
			TenantID:  nil,
		},
		{
			Email:     "tenant@speedycourier.com",
			Password:  "Tenant@Speedy2026!",
			FirstName: "Alice",
			LastName:  "Owner",
			Role:      auth.RoleTenant,
			TenantID:  &tenant1ID,
		},
		{
			Email:     "manager@speedycourier.com",
			Password:  "Manager@Speedy2026!",
			FirstName: "Bob",
			LastName:  "HubManager",
			Role:      auth.RoleTenantAdmin,
			TenantID:  &tenant1ID,
		},
		{
			Email:     "courier@speedycourier.com",
			Password:  "Courier@Speedy2026!",
			FirstName: "Charlie",
			LastName:  "Rider",
			Role:      auth.RoleEmployee,
			TenantID:  &tenant1ID,
		},
		{
			Email:     "customer@gmail.com",
			Password:  "Customer@LogiFlows2026!",
			FirstName: "Diana",
			LastName:  "Shopper",
			Role:      auth.RoleCustomer,
			TenantID:  nil,
		},
		{
			Email:     "metro_admin@metrocourier.com",
			Password:  "Metro@2026!",
			FirstName: "Murugan",
			LastName:  "MetroAdmin",
			Role:      auth.RoleTenantAdmin,
			TenantID:  &tenant2ID,
		},
	}

	for _, u := range users {
		pwdHash, _ := auth.HashPassword(u.Password)
		var uid uuid.UUID
		err := db.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role
		RETURNING id`,
			u.Email, pwdHash, u.FirstName, u.LastName, string(u.Role),
		).Scan(&uid)
		if err != nil {
			logger.Error("Failed to seed user", slog.String("email", u.Email), slog.String("error", err.Error()))
			continue
		}

		if u.TenantID != nil {
			_, _ = db.ExecContext(ctx, `
			INSERT INTO memberships (tenant_id, user_id, role, is_default)
			VALUES ($1, $2, $3, true)
			ON CONFLICT (tenant_id, user_id) DO NOTHING`,
				*u.TenantID, uid, string(u.Role),
			)
		}

		logger.Info("Seeded user successfully", slog.String("email", u.Email), slog.String("role", string(u.Role)))
	}

	// 3. Seed Realistic Branches with PostGIS Points and Polygons
	branches := []struct {
		TenantID      uuid.UUID
		Code          string
		Name          string
		BranchType    string
		Address       string
		City          string
		State         string
		PostalCode    string
		Phone         string
		Email         string
		Capacity      int
		Status        string
		Lat           float64
		Lon           float64
		ServiceAreaWKT string // WKT Polygon
	}{
		// Branch 1: BLR Central Sorting Hub
		{
			TenantID:      tenant1ID,
			Code:          "BLR-HUB-01",
			Name:          "Bengaluru Central Sorting Hub",
			BranchType:    "SORTING_HUB",
			Address:       "Plot 12, Majestic Logistics Yard, Gubbi Thotadappa Rd",
			City:          "Bengaluru",
			State:         "Karnataka",
			PostalCode:    "560009",
			Phone:         "+91 80 2234 5678",
			Email:         "blr.hub@speedycourier.com",
			Capacity:      25000,
			Status:        "ACTIVE",
			Lat:           12.9716,
			Lon:           77.5946,
			// Central Bengaluru Polygon (Majestic, MG Road, Rajajinagar, Malleshwaram, Shantinagar)
			ServiceAreaWKT: "POLYGON((77.5500 12.9400, 77.6300 12.9400, 77.6300 13.0100, 77.5500 13.0100, 77.5500 12.9400))",
		},
		// Branch 2: BLR North Distribution Center
		{
			TenantID:      tenant1ID,
			Code:          "BLR-DC-NORTH",
			Name:          "Bengaluru North Distribution Center",
			BranchType:    "DISTRIBUTION_CENTER",
			Address:       "Hub 4, Hebbal Flyover Junction, Outer Ring Rd",
			City:          "Bengaluru",
			State:         "Karnataka",
			PostalCode:    "560024",
			Phone:         "+91 80 2345 6789",
			Email:         "blr.north@speedycourier.com",
			Capacity:      10000,
			Status:        "ACTIVE",
			Lat:           13.0358,
			Lon:           77.5970,
			// North Bengaluru Polygon (Hebbal, Yelahanka, Manyata Tech Park, Sahakarnagar)
			ServiceAreaWKT: "POLYGON((77.5500 13.0100, 77.6500 13.0100, 77.6500 13.1200, 77.5500 13.1200, 77.5500 13.0100))",
		},
		// Branch 3: BLR South Distribution Center
		{
			TenantID:      tenant1ID,
			Code:          "BLR-DC-SOUTH",
			Name:          "Bengaluru South Distribution Center",
			BranchType:    "DISTRIBUTION_CENTER",
			Address:       "Tech Depot 8, Koramangala 4th Block, 80 Feet Rd",
			City:          "Bengaluru",
			State:         "Karnataka",
			PostalCode:    "560034",
			Phone:         "+91 80 2456 7890",
			Email:         "blr.south@speedycourier.com",
			Capacity:      12000,
			Status:        "ACTIVE",
			Lat:           12.9279,
			Lon:           77.6271,
			// South Bengaluru Polygon (Koramangala, HSR Layout, BTM, Electronic City)
			ServiceAreaWKT: "POLYGON((77.5800 12.8300, 77.7000 12.8300, 77.7000 12.9400, 77.5800 12.9400, 77.5800 12.8300))",
		},
		// Branch 4: CHN Central Hub (Tenant 2)
		{
			TenantID:      tenant2ID,
			Code:          "CHN-HUB-01",
			Name:          "Chennai Central Sorting Hub",
			BranchType:    "SORTING_HUB",
			Address:       "150 Mount Road, Anna Salai",
			City:          "Chennai",
			State:         "Tamil Nadu",
			PostalCode:    "600002",
			Phone:         "+91 44 2850 1234",
			Email:         "chn.hub@metrocourier.com",
			Capacity:      20000,
			Status:        "ACTIVE",
			Lat:           13.0827,
			Lon:           80.2707,
			// Chennai Central Polygon
			ServiceAreaWKT: "POLYGON((80.2000 12.9800, 80.3200 12.9800, 80.3200 13.1500, 80.2000 13.1500, 80.2000 12.9800))",
		},
	}

	for _, b := range branches {
		_, err := db.ExecContext(ctx, `
		INSERT INTO branches (
			tenant_id, code, name, branch_type, address, city, state, postal_code, country,
			contact_phone, contact_email, daily_capacity, status,
			location, service_area
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, 'India',
			$9, $10, $11, $12,
			ST_SetSRID(ST_MakePoint($13, $14), 4326),
			ST_SetSRID(ST_GeomFromText($15), 4326)
		)
		ON CONFLICT (tenant_id, code) DO UPDATE SET
			name = EXCLUDED.name,
			branch_type = EXCLUDED.branch_type,
			address = EXCLUDED.address,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			postal_code = EXCLUDED.postal_code,
			contact_phone = EXCLUDED.contact_phone,
			contact_email = EXCLUDED.contact_email,
			daily_capacity = EXCLUDED.daily_capacity,
			status = EXCLUDED.status,
			location = EXCLUDED.location,
			service_area = EXCLUDED.service_area`,
			b.TenantID, b.Code, b.Name, b.BranchType, b.Address, b.City, b.State, b.PostalCode,
			b.Phone, b.Email, b.Capacity, b.Status,
			b.Lon, b.Lat, b.ServiceAreaWKT,
		)
		if err != nil {
			logger.Error("Failed to seed branch", slog.String("code", b.Code), slog.String("error", err.Error()))
			continue
		}
		logger.Info("Seeded branch successfully", slog.String("code", b.Code), slog.String("name", b.Name))
	}

	logger.Info("Database seeding completed successfully!")
}
