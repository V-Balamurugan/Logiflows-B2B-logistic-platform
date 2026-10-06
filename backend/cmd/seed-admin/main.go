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

	// 5. Seed Vehicles & Telematics
	type SeedVehicle struct {
		ID            uuid.UUID
		TenantID      uuid.UUID
		BranchCode    string
		LicensePlate  string
		VIN           string
		Make          string
		Model         string
		Year          int
		VehicleType   string
		FuelType      string
		CapacityKG    float64
		VolumeM3      float64
		MaxParcels    int
		MileageKM     float64
		BatteryFuel   float64
		Status        string
		Lat           float64
		Lon           float64
		Speed         float64
		Heading       float64
	}

	veh1ID := uuid.MustParse("00000000-0000-0000-0001-000000000001")
	veh2ID := uuid.MustParse("00000000-0000-0000-0001-000000000002")
	veh3ID := uuid.MustParse("00000000-0000-0000-0001-000000000003")
	veh4ID := uuid.MustParse("00000000-0000-0000-0002-000000000001")

	seedVehicles := []SeedVehicle{
		{
			ID:           veh1ID,
			TenantID:     tenant1ID,
			BranchCode:   "BLR-HUB-01",
			LicensePlate: "KA-01-EV-1001",
			VIN:          "MBBLREVITO2024001",
			Make:         "Tata",
			Model:        "Ace EV",
			Year:         2024,
			VehicleType:  "ELECTRIC_VAN",
			FuelType:     "ELECTRIC",
			CapacityKG:   800.0,
			VolumeM3:     4.5,
			MaxParcels:   150,
			MileageKM:    1240.5,
			BatteryFuel:  88.5,
			Status:       "ON_ROUTE",
			Lat:          12.9750,
			Lon:          77.5980,
			Speed:        38.5,
			Heading:      135.0,
		},
		{
			ID:           veh2ID,
			TenantID:     tenant1ID,
			BranchCode:   "BLR-DC-NORTH",
			LicensePlate: "KA-04-MC-2002",
			VIN:          "ATH450XBLR2024002",
			Make:         "Ather",
			Model:        "450X Gen 3",
			Year:         2024,
			VehicleType:  "MOTORCYCLE",
			FuelType:     "ELECTRIC",
			CapacityKG:   65.0,
			VolumeM3:     0.5,
			MaxParcels:   30,
			MileageKM:    3450.0,
			BatteryFuel:  74.0,
			Status:       "AVAILABLE",
			Lat:          13.0360,
			Lon:          77.5975,
			Speed:        0.0,
			Heading:      0.0,
		},
		{
			ID:           veh3ID,
			TenantID:     tenant1ID,
			BranchCode:   "BLR-DC-SOUTH",
			LicensePlate: "KA-51-TK-3003",
			VIN:          "ALDOSTBLR2023003",
			Make:         "Ashok Leyland",
			Model:        "Bada Dost",
			Year:         2023,
			VehicleType:  "TRUCK",
			FuelType:     "DIESEL",
			CapacityKG:   1800.0,
			VolumeM3:     8.5,
			MaxParcels:   350,
			MileageKM:    9600.0,
			BatteryFuel:  65.0,
			Status:       "AVAILABLE",
			Lat:          12.9280,
			Lon:          77.6275,
			Speed:        0.0,
			Heading:      0.0,
		},
		{
			ID:           veh4ID,
			TenantID:     tenant2ID,
			BranchCode:   "CHN-HUB-01",
			LicensePlate: "TN-01-EV-4004",
			VIN:          "MAHTREOCHN2024004",
			Make:         "Mahindra",
			Model:        "Zor Grand EV",
			Year:         2024,
			VehicleType:  "ELECTRIC_VAN",
			FuelType:     "ELECTRIC",
			CapacityKG:   500.0,
			VolumeM3:     3.5,
			MaxParcels:   100,
			MileageKM:    2100.0,
			BatteryFuel:  92.0,
			Status:       "AVAILABLE",
			Lat:          13.0830,
			Lon:          80.2710,
			Speed:        0.0,
			Heading:      0.0,
		},
	}

	for _, sv := range seedVehicles {
		var branchID *uuid.UUID
		err := db.QueryRowContext(ctx, "SELECT id FROM branches WHERE tenant_id = $1 AND code = $2", sv.TenantID, sv.BranchCode).Scan(&branchID)
		if err != nil {
			logger.Warn("Could not resolve branch for vehicle", slog.String("branch", sv.BranchCode), slog.String("plate", sv.LicensePlate))
		}

		_, err = db.ExecContext(ctx, `
			INSERT INTO vehicles (
				id, tenant_id, assigned_branch_id, license_plate, vin, make, model, year,
				vehicle_type, fuel_type, capacity_kg, capacity_volume_m3, max_parcels,
				current_mileage_km, battery_or_fuel_level_percent, status, metadata
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8,
				$9, $10, $11, $12, $13,
				$14, $15, $16, '{}'
			) ON CONFLICT (tenant_id, license_plate) DO UPDATE SET
				assigned_branch_id = EXCLUDED.assigned_branch_id,
				make = EXCLUDED.make,
				model = EXCLUDED.model,
				vehicle_type = EXCLUDED.vehicle_type,
				fuel_type = EXCLUDED.fuel_type,
				capacity_kg = EXCLUDED.capacity_kg,
				capacity_volume_m3 = EXCLUDED.capacity_volume_m3,
				max_parcels = EXCLUDED.max_parcels,
				current_mileage_km = EXCLUDED.current_mileage_km,
				battery_or_fuel_level_percent = EXCLUDED.battery_or_fuel_level_percent,
				status = EXCLUDED.status;
		`, sv.ID, sv.TenantID, branchID, sv.LicensePlate, sv.VIN, sv.Make, sv.Model, sv.Year,
			sv.VehicleType, sv.FuelType, sv.CapacityKG, sv.VolumeM3, sv.MaxParcels,
			sv.MileageKM, sv.BatteryFuel, sv.Status)

		if err != nil {
			logger.Error("Failed to seed vehicle", slog.String("plate", sv.LicensePlate), slog.String("error", err.Error()))
			continue
		}

		// Insert GPS telematics point
		_, err = db.ExecContext(ctx, `
			INSERT INTO vehicle_telematics (
				vehicle_id, tenant_id, location, latitude, longitude,
				speed_kmh, heading_degrees, battery_or_fuel_percent, odometer_km, recorded_at
			) VALUES (
				$1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), $4, $3,
				$5, $6, $7, $8, CURRENT_TIMESTAMP
			);
		`, sv.ID, sv.TenantID, sv.Lon, sv.Lat, sv.Speed, sv.Heading, sv.BatteryFuel, sv.MileageKM)

		if err != nil {
			logger.Warn("Failed to seed telematics ping", slog.String("plate", sv.LicensePlate), slog.String("error", err.Error()))
		}

		logger.Info("Seeded vehicle and telematics", slog.String("plate", sv.LicensePlate), slog.String("type", sv.VehicleType))
	}

	// 6. Seed Maintenance Record for Truck (approaching 10,000 km service)
	nextDueKM := 10000.0
	_, err = db.ExecContext(ctx, `
		INSERT INTO maintenance_records (
			vehicle_id, tenant_id, service_type, description, cost,
			odometer_reading_km, serviced_at, next_service_due_km
		) VALUES (
			$1, $2, 'ENGINE_OIL_CHANGE', 'Full synthetic 15W-40 lube replacement and filter tuneup',
			250.00, 5000.00, CURRENT_TIMESTAMP - INTERVAL '90 days', $3
		);
	`, veh3ID, tenant1ID, nextDueKM)
	if err != nil {
		logger.Warn("Maintenance seed skipped/exists", slog.String("error", err.Error()))
	} else {
		logger.Info("Seeded maintenance record for vehicle", slog.String("plate", "KA-51-TK-3003"))
	}

	logger.Info("Database seeding completed successfully!")
}
