-- Phase 4: Customers, Saved Address Book, Parcel Booking & Strict State Machine
-- Phase 5: QR Identification & Append-Only Custody Handover Chain

-- 1. Customer Profiles Table
CREATE TABLE IF NOT EXISTS customer_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    company_name VARCHAR(255),
    tax_id VARCHAR(100),
    billing_address TEXT,
    preferences JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_customer_user UNIQUE (user_id)
);

CREATE INDEX IF NOT EXISTS idx_customer_profiles_tenant ON customer_profiles(tenant_id);

-- 2. Customer Address Book Table
CREATE TABLE IF NOT EXISTS addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    label VARCHAR(100) NOT NULL,
    contact_name VARCHAR(150) NOT NULL,
    contact_phone VARCHAR(30) NOT NULL,
    contact_email VARCHAR(255),
    street_line1 TEXT NOT NULL,
    street_line2 TEXT,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(30) NOT NULL,
    country VARCHAR(100) NOT NULL DEFAULT 'India',
    location GEOMETRY(Point, 4326),
    is_default_pickup BOOLEAN NOT NULL DEFAULT false,
    is_default_delivery BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_addresses_customer ON addresses(customer_id);
CREATE INDEX IF NOT EXISTS idx_addresses_tenant ON addresses(tenant_id);
CREATE INDEX IF NOT EXISTS idx_addresses_location ON addresses USING GIST(location);

-- 3. Parcels & Consignment Bookings Table
CREATE TABLE IF NOT EXISTS parcels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    tracking_number VARCHAR(64) NOT NULL UNIQUE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    origin_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    destination_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    current_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    service_type VARCHAR(50) NOT NULL DEFAULT 'STANDARD',
    status VARCHAR(50) NOT NULL DEFAULT 'CREATED',
    
    -- Sender Details
    sender_name VARCHAR(150) NOT NULL,
    sender_phone VARCHAR(30) NOT NULL,
    sender_email VARCHAR(255),
    sender_address TEXT NOT NULL,
    sender_city VARCHAR(100) NOT NULL,
    sender_state VARCHAR(100) NOT NULL,
    sender_postal_code VARCHAR(30) NOT NULL,
    sender_location GEOMETRY(Point, 4326),
    
    -- Recipient Details
    recipient_name VARCHAR(150) NOT NULL,
    recipient_phone VARCHAR(30) NOT NULL,
    recipient_email VARCHAR(255),
    recipient_address TEXT NOT NULL,
    recipient_city VARCHAR(100) NOT NULL,
    recipient_state VARCHAR(100) NOT NULL,
    recipient_postal_code VARCHAR(30) NOT NULL,
    recipient_location GEOMETRY(Point, 4326),
    
    -- Physical Specifications
    weight_kg NUMERIC(8, 2) NOT NULL DEFAULT 1.0,
    length_cm NUMERIC(8, 2) NOT NULL DEFAULT 10.0,
    width_cm NUMERIC(8, 2) NOT NULL DEFAULT 10.0,
    height_cm NUMERIC(8, 2) NOT NULL DEFAULT 10.0,
    declared_value NUMERIC(12, 2) NOT NULL DEFAULT 0.0,
    currency VARCHAR(10) NOT NULL DEFAULT 'INR',
    is_fragile BOOLEAN NOT NULL DEFAULT false,
    special_instructions TEXT,
    
    -- Commercial & Idempotency
    shipping_cost NUMERIC(10, 2) NOT NULL DEFAULT 0.0,
    payment_status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    qr_token VARCHAR(255) NOT NULL UNIQUE,
    idempotency_key VARCHAR(255) UNIQUE,
    
    -- Status Lifecycle Timestamps
    estimated_delivery_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_parcels_tracking_number ON parcels(tracking_number);
CREATE INDEX IF NOT EXISTS idx_parcels_tenant_id ON parcels(tenant_id);
CREATE INDEX IF NOT EXISTS idx_parcels_sender_id ON parcels(sender_id);
CREATE INDEX IF NOT EXISTS idx_parcels_status ON parcels(status);
CREATE INDEX IF NOT EXISTS idx_parcels_service_type ON parcels(service_type);
CREATE INDEX IF NOT EXISTS idx_parcels_qr_token ON parcels(qr_token);
CREATE INDEX IF NOT EXISTS idx_parcels_idempotency_key ON parcels(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_parcels_created_at ON parcels(created_at DESC);

-- 4. Custody Events Table (Append-only immutable chain of custody)
CREATE TABLE IF NOT EXISTS custody_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parcel_id UUID NOT NULL REFERENCES parcels(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,
    previous_status VARCHAR(50),
    new_status VARCHAR(50),
    location GEOMETRY(Point, 4326),
    location_name VARCHAR(255),
    device_info VARCHAR(255),
    notes TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_custody_parcel_id ON custody_events(parcel_id);
CREATE INDEX IF NOT EXISTS idx_custody_tenant_id ON custody_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_custody_recorded_at ON custody_events(parcel_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_custody_event_type ON custody_events(event_type);
