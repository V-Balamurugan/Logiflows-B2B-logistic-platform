package tenancy

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Repository interface {
	// Tenant operations
	CreateTenant(ctx context.Context, req *TenantCreateRequest) (*Tenant, error)
	GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetTenantByCode(ctx context.Context, code string) (*Tenant, error)
	GetTenantBySlug(ctx context.Context, slug string) (*Tenant, error)
	ListTenants(ctx context.Context) ([]Tenant, error)
	UpdateTenant(ctx context.Context, id uuid.UUID, req *TenantUpdateRequest) (*Tenant, error)

	// Branch operations
	CreateBranch(ctx context.Context, tenantID uuid.UUID, req *BranchCreateRequest) (*Branch, error)
	GetBranchByID(ctx context.Context, tenantID, branchID uuid.UUID) (*Branch, error)
	GetBranchByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Branch, error)
	ListBranches(ctx context.Context, tenantID uuid.UUID) ([]Branch, error)
	UpdateBranch(ctx context.Context, tenantID, branchID uuid.UUID, req *BranchUpdateRequest) (*Branch, error)
	DeleteBranch(ctx context.Context, tenantID, branchID uuid.UUID) error

	// Geospatial serviceability operations
	FindServiceableBranch(ctx context.Context, tenantID uuid.UUID, lat, lon float64) (*BranchSummary, float64, error)
	FindNearestBranch(ctx context.Context, tenantID uuid.UUID, lat, lon float64) (*BranchSummary, float64, error)
	GetCoverageGeoJSON(ctx context.Context, tenantID uuid.UUID) (*GeoJSONFeatureCollection, error)
}

type pgRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &pgRepository{db: db}
}

func (r *pgRepository) CreateTenant(ctx context.Context, req *TenantCreateRequest) (*Tenant, error) {
	brandingJSON, err := json.Marshal(req.Branding)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal branding: %w", err)
	}
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO tenants (name, code, slug, tier, quota_parcels_per_day, quota_branches, branding, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, name, code, slug, tier, quota_parcels_per_day, quota_branches, branding, config, is_active, created_at, updated_at
	`
	var t Tenant
	var brandingBytes, configBytes []byte

	err = r.db.QueryRowContext(
		ctx, query,
		req.Name, req.Code, req.Slug, req.Tier, req.QuotaParcelsPerDay, req.QuotaBranches, brandingJSON, configJSON,
	).Scan(
		&t.ID, &t.Name, &t.Code, &t.Slug, &t.Tier, &t.QuotaParcelsPerDay, &t.QuotaBranches,
		&brandingBytes, &configBytes, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(brandingBytes, &t.Branding)
	_ = json.Unmarshal(configBytes, &t.Config)
	return &t, nil
}

func (r *pgRepository) GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	query := `
		SELECT id, name, code, COALESCE(slug, code), tier, quota_parcels_per_day, quota_branches,
		       branding, config, is_active, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`
	var t Tenant
	var brandingBytes, configBytes []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Name, &t.Code, &t.Slug, &t.Tier, &t.QuotaParcelsPerDay, &t.QuotaBranches,
		&brandingBytes, &configBytes, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

	_ = json.Unmarshal(brandingBytes, &t.Branding)
	_ = json.Unmarshal(configBytes, &t.Config)
	return &t, nil
}

func (r *pgRepository) GetTenantByCode(ctx context.Context, code string) (*Tenant, error) {
	query := `
		SELECT id, name, code, COALESCE(slug, code), tier, quota_parcels_per_day, quota_branches,
		       branding, config, is_active, created_at, updated_at
		FROM tenants
		WHERE code = $1
	`
	var t Tenant
	var brandingBytes, configBytes []byte

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&t.ID, &t.Name, &t.Code, &t.Slug, &t.Tier, &t.QuotaParcelsPerDay, &t.QuotaBranches,
		&brandingBytes, &configBytes, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

	_ = json.Unmarshal(brandingBytes, &t.Branding)
	_ = json.Unmarshal(configBytes, &t.Config)
	return &t, nil
}

func (r *pgRepository) GetTenantBySlug(ctx context.Context, slug string) (*Tenant, error) {
	query := `
		SELECT id, name, code, COALESCE(slug, code), tier, quota_parcels_per_day, quota_branches,
		       branding, config, is_active, created_at, updated_at
		FROM tenants
		WHERE slug = $1 OR code = $1
	`
	var t Tenant
	var brandingBytes, configBytes []byte

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&t.ID, &t.Name, &t.Code, &t.Slug, &t.Tier, &t.QuotaParcelsPerDay, &t.QuotaBranches,
		&brandingBytes, &configBytes, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

	_ = json.Unmarshal(brandingBytes, &t.Branding)
	_ = json.Unmarshal(configBytes, &t.Config)
	return &t, nil
}

func (r *pgRepository) ListTenants(ctx context.Context) ([]Tenant, error) {
	query := `
		SELECT id, name, code, COALESCE(slug, code), tier, quota_parcels_per_day, quota_branches,
		       branding, config, is_active, created_at, updated_at
		FROM tenants
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []Tenant
	for rows.Next() {
		var t Tenant
		var brandingBytes, configBytes []byte
		if err := rows.Scan(
			&t.ID, &t.Name, &t.Code, &t.Slug, &t.Tier, &t.QuotaParcelsPerDay, &t.QuotaBranches,
			&brandingBytes, &configBytes, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(brandingBytes, &t.Branding)
		_ = json.Unmarshal(configBytes, &t.Config)
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

func (r *pgRepository) UpdateTenant(ctx context.Context, id uuid.UUID, req *TenantUpdateRequest) (*Tenant, error) {
	existing, err := r.GetTenantByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Slug != nil {
		existing.Slug = *req.Slug
	}
	if req.Tier != nil {
		existing.Tier = *req.Tier
	}
	if req.QuotaParcelsPerDay != nil {
		existing.QuotaParcelsPerDay = *req.QuotaParcelsPerDay
	}
	if req.QuotaBranches != nil {
		existing.QuotaBranches = *req.QuotaBranches
	}
	if req.Branding != nil {
		existing.Branding = *req.Branding
	}
	if req.Config != nil {
		existing.Config = *req.Config
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	brandingJSON, _ := json.Marshal(existing.Branding)
	configJSON, _ := json.Marshal(existing.Config)

	query := `
		UPDATE tenants
		SET name = $1, slug = $2, tier = $3, quota_parcels_per_day = $4, quota_branches = $5,
		    branding = $6, config = $7, is_active = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9
		RETURNING updated_at
	`
	err = r.db.QueryRowContext(
		ctx, query,
		existing.Name, existing.Slug, existing.Tier, existing.QuotaParcelsPerDay, existing.QuotaBranches,
		brandingJSON, configJSON, existing.IsActive, id,
	).Scan(&existing.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

// Branch operations
func (r *pgRepository) CreateBranch(ctx context.Context, tenantID uuid.UUID, req *BranchCreateRequest) (*Branch, error) {
	opHoursJSON, err := json.Marshal(req.OperatingHours)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operating hours: %w", err)
	}

	query := `
		INSERT INTO branches (
			tenant_id, code, name, branch_type, address, city, state, postal_code, country,
			contact_phone, contact_email, operating_hours, daily_capacity, status,
			location, service_area
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14,
			ST_SetSRID(ST_MakePoint($15, $16), 4326),
			ST_SetSRID(ST_GeomFromGeoJSON($17), 4326)
		)
		RETURNING id, created_at, updated_at
	`
	var branch Branch
	branch.TenantID = tenantID
	branch.Code = req.Code
	branch.Name = req.Name
	branch.BranchType = req.BranchType
	branch.Address = req.Address
	branch.City = req.City
	branch.State = req.State
	branch.PostalCode = req.PostalCode
	branch.Country = req.Country
	branch.ContactPhone = req.ContactPhone
	branch.ContactEmail = req.ContactEmail
	if req.OperatingHours != nil {
		branch.OperatingHours = *req.OperatingHours
	}
	branch.DailyCapacity = req.DailyCapacity
	branch.Status = req.Status
	branch.Location = req.Location
	branch.ServiceArea = req.ServiceArea

	err = r.db.QueryRowContext(
		ctx, query,
		tenantID, req.Code, req.Name, req.BranchType, req.Address, req.City, req.State, req.PostalCode, req.Country,
		req.ContactPhone, req.ContactEmail, opHoursJSON, req.DailyCapacity, req.Status,
		req.Location.Longitude, req.Location.Latitude, string(req.ServiceArea),
	).Scan(&branch.ID, &branch.CreatedAt, &branch.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert branch into postgis: %w", err)
	}

	return &branch, nil
}

func (r *pgRepository) GetBranchByID(ctx context.Context, tenantID, branchID uuid.UUID) (*Branch, error) {
	query := `
		SELECT id, tenant_id, code, name, branch_type, address, city, state, postal_code, country,
		       contact_phone, contact_email, operating_hours, daily_capacity, status,
		       ST_Y(location) as lat, ST_X(location) as lon,
		       ST_AsGeoJSON(service_area) as service_area_geojson,
		       created_at, updated_at
		FROM branches
		WHERE tenant_id = $1 AND id = $2
	`
	var b Branch
	var opHoursBytes []byte
	var serviceAreaStr string

	err := r.db.QueryRowContext(ctx, query, tenantID, branchID).Scan(
		&b.ID, &b.TenantID, &b.Code, &b.Name, &b.BranchType, &b.Address, &b.City, &b.State, &b.PostalCode, &b.Country,
		&b.ContactPhone, &b.ContactEmail, &opHoursBytes, &b.DailyCapacity, &b.Status,
		&b.Location.Latitude, &b.Location.Longitude,
		&serviceAreaStr,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("branch not found")
		}
		return nil, err
	}

	_ = json.Unmarshal(opHoursBytes, &b.OperatingHours)
	b.ServiceArea = json.RawMessage(serviceAreaStr)
	return &b, nil
}

func (r *pgRepository) GetBranchByCode(ctx context.Context, tenantID uuid.UUID, code string) (*Branch, error) {
	query := `
		SELECT id, tenant_id, code, name, branch_type, address, city, state, postal_code, country,
		       contact_phone, contact_email, operating_hours, daily_capacity, status,
		       ST_Y(location) as lat, ST_X(location) as lon,
		       ST_AsGeoJSON(service_area) as service_area_geojson,
		       created_at, updated_at
		FROM branches
		WHERE tenant_id = $1 AND code = $2
	`
	var b Branch
	var opHoursBytes []byte
	var serviceAreaStr string

	err := r.db.QueryRowContext(ctx, query, tenantID, code).Scan(
		&b.ID, &b.TenantID, &b.Code, &b.Name, &b.BranchType, &b.Address, &b.City, &b.State, &b.PostalCode, &b.Country,
		&b.ContactPhone, &b.ContactEmail, &opHoursBytes, &b.DailyCapacity, &b.Status,
		&b.Location.Latitude, &b.Location.Longitude,
		&serviceAreaStr,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("branch not found")
		}
		return nil, err
	}

	_ = json.Unmarshal(opHoursBytes, &b.OperatingHours)
	b.ServiceArea = json.RawMessage(serviceAreaStr)
	return &b, nil
}

func (r *pgRepository) ListBranches(ctx context.Context, tenantID uuid.UUID) ([]Branch, error) {
	query := `
		SELECT id, tenant_id, code, name, branch_type, address, city, state, postal_code, country,
		       contact_phone, contact_email, operating_hours, daily_capacity, status,
		       ST_Y(location) as lat, ST_X(location) as lon,
		       ST_AsGeoJSON(service_area) as service_area_geojson,
		       created_at, updated_at
		FROM branches
		WHERE tenant_id = $1
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []Branch
	for rows.Next() {
		var b Branch
		var opHoursBytes []byte
		var serviceAreaStr string

		if err := rows.Scan(
			&b.ID, &b.TenantID, &b.Code, &b.Name, &b.BranchType, &b.Address, &b.City, &b.State, &b.PostalCode, &b.Country,
			&b.ContactPhone, &b.ContactEmail, &opHoursBytes, &b.DailyCapacity, &b.Status,
			&b.Location.Latitude, &b.Location.Longitude,
			&serviceAreaStr,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(opHoursBytes, &b.OperatingHours)
		b.ServiceArea = json.RawMessage(serviceAreaStr)
		branches = append(branches, b)
	}
	return branches, rows.Err()
}

func (r *pgRepository) UpdateBranch(ctx context.Context, tenantID, branchID uuid.UUID, req *BranchUpdateRequest) (*Branch, error) {
	existing, err := r.GetBranchByID(ctx, tenantID, branchID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.BranchType != nil {
		existing.BranchType = *req.BranchType
	}
	if req.Address != nil {
		existing.Address = *req.Address
	}
	if req.City != nil {
		existing.City = *req.City
	}
	if req.State != nil {
		existing.State = *req.State
	}
	if req.PostalCode != nil {
		existing.PostalCode = *req.PostalCode
	}
	if req.Country != nil {
		existing.Country = *req.Country
	}
	if req.ContactPhone != nil {
		existing.ContactPhone = *req.ContactPhone
	}
	if req.ContactEmail != nil {
		existing.ContactEmail = *req.ContactEmail
	}
	if req.OperatingHours != nil {
		existing.OperatingHours = *req.OperatingHours
	}
	if req.DailyCapacity != nil {
		existing.DailyCapacity = *req.DailyCapacity
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Location != nil {
		existing.Location = *req.Location
	}
	if req.ServiceArea != nil {
		existing.ServiceArea = *req.ServiceArea
	}

	opHoursJSON, _ := json.Marshal(existing.OperatingHours)

	query := `
		UPDATE branches
		SET name = $1, branch_type = $2, address = $3, city = $4, state = $5, postal_code = $6, country = $7,
		    contact_phone = $8, contact_email = $9, operating_hours = $10, daily_capacity = $11, status = $12,
		    location = ST_SetSRID(ST_MakePoint($13, $14), 4326),
		    service_area = ST_SetSRID(ST_GeomFromGeoJSON($15), 4326),
		    updated_at = CURRENT_TIMESTAMP
		WHERE tenant_id = $16 AND id = $17
		RETURNING updated_at
	`
	err = r.db.QueryRowContext(
		ctx, query,
		existing.Name, existing.BranchType, existing.Address, existing.City, existing.State, existing.PostalCode, existing.Country,
		existing.ContactPhone, existing.ContactEmail, opHoursJSON, existing.DailyCapacity, existing.Status,
		existing.Location.Longitude, existing.Location.Latitude, string(existing.ServiceArea),
		tenantID, branchID,
	).Scan(&existing.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update branch: %w", err)
	}

	return existing, nil
}

func (r *pgRepository) DeleteBranch(ctx context.Context, tenantID, branchID uuid.UUID) error {
	query := `DELETE FROM branches WHERE tenant_id = $1 AND id = $2`
	result, err := r.db.ExecContext(ctx, query, tenantID, branchID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("branch not found")
	}
	return nil
}

// PostGIS Geospatial Serviceability Operations
func (r *pgRepository) FindServiceableBranch(ctx context.Context, tenantID uuid.UUID, lat, lon float64) (*BranchSummary, float64, error) {
	// Point-in-polygon query using PostGIS ST_Contains and spatial index
	query := `
		SELECT id, code, name, branch_type, address, city, contact_phone, status,
		       ST_Y(location) as lat, ST_X(location) as lon,
		       ST_Distance(location::geography, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography) as distance_meters
		FROM branches
		WHERE tenant_id = $1
		  AND status = 'ACTIVE'
		  AND ST_Contains(service_area, ST_SetSRID(ST_MakePoint($2, $3), 4326))
		ORDER BY distance_meters ASC
		LIMIT 1
	`
	var b BranchSummary
	var distanceMeters float64

	err := r.db.QueryRowContext(ctx, query, tenantID, lon, lat).Scan(
		&b.ID, &b.Code, &b.Name, &b.BranchType, &b.Address, &b.City, &b.ContactPhone, &b.Status,
		&b.Location.Latitude, &b.Location.Longitude,
		&distanceMeters,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, nil // Not found inside any service area
		}
		return nil, 0, fmt.Errorf("geospatial serviceability query failed: %w", err)
	}

	return &b, distanceMeters, nil
}

func (r *pgRepository) FindNearestBranch(ctx context.Context, tenantID uuid.UUID, lat, lon float64) (*BranchSummary, float64, error) {
	// KNN query using PostGIS <-> geometry distance operator
	query := `
		SELECT id, code, name, branch_type, address, city, contact_phone, status,
		       ST_Y(location) as lat, ST_X(location) as lon,
		       ST_Distance(location::geography, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography) as distance_meters
		FROM branches
		WHERE tenant_id = $1
		  AND status = 'ACTIVE'
		ORDER BY location <-> ST_SetSRID(ST_MakePoint($2, $3), 4326) ASC
		LIMIT 1
	`
	var b BranchSummary
	var distanceMeters float64

	err := r.db.QueryRowContext(ctx, query, tenantID, lon, lat).Scan(
		&b.ID, &b.Code, &b.Name, &b.BranchType, &b.Address, &b.City, &b.ContactPhone, &b.Status,
		&b.Location.Latitude, &b.Location.Longitude,
		&distanceMeters,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("nearest branch query failed: %w", err)
	}

	return &b, distanceMeters, nil
}

func (r *pgRepository) GetCoverageGeoJSON(ctx context.Context, tenantID uuid.UUID) (*GeoJSONFeatureCollection, error) {
	query := `
		SELECT id, code, name, branch_type, status,
		       ST_AsGeoJSON(service_area) as service_area_geojson
		FROM branches
		WHERE tenant_id = $1
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fc := &GeoJSONFeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]GeoJSONFeature, 0),
	}

	for rows.Next() {
		var id uuid.UUID
		var code, name string
		var bType BranchType
		var status BranchStatus
		var geomStr string

		if err := rows.Scan(&id, &code, &name, &bType, &status, &geomStr); err != nil {
			return nil, err
		}

		feature := GeoJSONFeature{
			Type:     "Feature",
			Geometry: json.RawMessage(geomStr),
			Properties: map[string]interface{}{
				"branch_id":   id.String(),
				"code":        code,
				"name":        name,
				"branch_type": bType,
				"status":      status,
			},
		}
		fc.Features = append(fc.Features, feature)
	}

	return fc, rows.Err()
}
