-- Phase 2: Multi-Tenant Enterprise Hierarchy, Branches, Hubs & Geospatial Serviceability

-- 1. Enhance Tenants table with enterprise metadata, tiers, quotas, and branding
ALTER TABLE tenants 
    ADD COLUMN IF NOT EXISTS slug VARCHAR(100),
    ADD COLUMN IF NOT EXISTS tier VARCHAR(50) NOT NULL DEFAULT 'STANDARD',
    ADD COLUMN IF NOT EXISTS quota_parcels_per_day INT NOT NULL DEFAULT 5000,
    ADD COLUMN IF NOT EXISTS quota_branches INT NOT NULL DEFAULT 50,
    ADD COLUMN IF NOT EXISTS branding JSONB NOT NULL DEFAULT '{"primary_color":"#1e40af","accent_color":"#3b82f6","portal_name":"LogiFlows"}',
    ADD COLUMN IF NOT EXISTS config JSONB NOT NULL DEFAULT '{"timezone":"Asia/Kolkata","currency":"INR","auto_assign_delivery":true}';

-- Backfill slug from code if empty
UPDATE tenants SET slug = LOWER(code) WHERE slug IS NULL OR slug = '';

-- Ensure slug is unique
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);

-- 2. Branches & Logistics Hubs Table
CREATE TABLE IF NOT EXISTS branches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    branch_type VARCHAR(50) NOT NULL, -- SORTING_HUB, DISTRIBUTION_CENTER, LOCAL_OFFICE
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL DEFAULT '',
    state VARCHAR(100) NOT NULL DEFAULT '',
    postal_code VARCHAR(30) NOT NULL DEFAULT '',
    country VARCHAR(100) NOT NULL DEFAULT 'India',
    contact_phone VARCHAR(30) NOT NULL DEFAULT '',
    contact_email VARCHAR(255) NOT NULL DEFAULT '',
    operating_hours JSONB NOT NULL DEFAULT '{"open":"08:00","close":"20:00","timezone":"Asia/Kolkata"}',
    daily_capacity INT NOT NULL DEFAULT 1000,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, INACTIVE, MAINTENANCE
    location GEOMETRY(Point, 4326) NOT NULL,
    service_area GEOMETRY(Polygon, 4326) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_branch_tenant_code UNIQUE (tenant_id, code)
);

-- Spatial and relational indices for performance and multi-tenancy
CREATE INDEX IF NOT EXISTS idx_branches_location ON branches USING GIST (location);
CREATE INDEX IF NOT EXISTS idx_branches_service_area ON branches USING GIST (service_area);
CREATE INDEX IF NOT EXISTS idx_branches_tenant ON branches(tenant_id);
CREATE INDEX IF NOT EXISTS idx_branches_status ON branches(status);
CREATE INDEX IF NOT EXISTS idx_branches_type ON branches(branch_type);
