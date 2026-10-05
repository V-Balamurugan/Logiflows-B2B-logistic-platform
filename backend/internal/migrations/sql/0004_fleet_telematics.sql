-- Phase 3: Fleet Management, Vehicle Telematics & Asset Tracking

-- 1. Vehicles Table
CREATE TABLE IF NOT EXISTS vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    assigned_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    assigned_driver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    license_plate VARCHAR(32) NOT NULL,
    vin VARCHAR(64) NOT NULL DEFAULT '',
    make VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INT NOT NULL DEFAULT 2024,
    vehicle_type VARCHAR(50) NOT NULL, -- VAN, MOTORCYCLE, TRUCK, ELECTRIC_VAN, BICYCLE
    fuel_type VARCHAR(50) NOT NULL DEFAULT 'DIESEL', -- DIESEL, PETROL, ELECTRIC, HYBRID
    capacity_kg NUMERIC(10, 2) NOT NULL DEFAULT 500.00,
    capacity_volume_m3 NUMERIC(10, 2) NOT NULL DEFAULT 3.50,
    max_parcels INT NOT NULL DEFAULT 150,
    current_mileage_km NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    battery_or_fuel_level_percent NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    status VARCHAR(50) NOT NULL DEFAULT 'AVAILABLE', -- AVAILABLE, ON_ROUTE, MAINTENANCE, DECOMMISSIONED
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_vehicle_tenant_license UNIQUE (tenant_id, license_plate)
);

CREATE INDEX IF NOT EXISTS idx_vehicles_tenant ON vehicles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_branch ON vehicles(assigned_branch_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_driver ON vehicles(assigned_driver_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles(status);
CREATE INDEX IF NOT EXISTS idx_vehicles_type ON vehicles(vehicle_type);

-- 2. Vehicle Telematics Table (GPS breadcrumbs and spatial telemetry points)
CREATE TABLE IF NOT EXISTS vehicle_telematics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    location GEOMETRY(Point, 4326) NOT NULL,
    latitude NUMERIC(10, 6) NOT NULL,
    longitude NUMERIC(10, 6) NOT NULL,
    speed_kmh NUMERIC(6, 2) NOT NULL DEFAULT 0.00,
    heading_degrees NUMERIC(6, 2) NOT NULL DEFAULT 0.00,
    battery_or_fuel_percent NUMERIC(5, 2) NOT NULL DEFAULT 100.00,
    odometer_km NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vehicle_telematics_vehicle ON vehicle_telematics(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_telematics_tenant ON vehicle_telematics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_telematics_recorded_at ON vehicle_telematics(recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_vehicle_telematics_location ON vehicle_telematics USING GIST (location);

-- 3. Maintenance Records Table
CREATE TABLE IF NOT EXISTS maintenance_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    service_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    cost NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    odometer_reading_km NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    serviced_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    next_service_due_km NUMERIC(10, 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_maintenance_vehicle ON maintenance_records(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_tenant ON maintenance_records(tenant_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_serviced_at ON maintenance_records(serviced_at DESC);
