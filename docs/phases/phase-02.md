# Phase 2 — Multi-Tenant Enterprise Hierarchy, Branches, Hubs & Geospatial Serviceability

## 1. Executive Summary
Phase 2 establishes the multi-tenant enterprise hierarchy and PostGIS-powered geospatial logistics infrastructure for **LogiFlows**. It introduces authoritative branch and hub management (`SORTING_HUB`, `DISTRIBUTION_CENTER`, `LOCAL_OFFICE`), spatial polygon boundaries for coverage zones, point-in-polygon serviceability resolution, and nearest-hub spatial indexing.

---

## 2. Key Deliverables

### 2.1 Database & PostGIS 16 (`0003_tenants_branches_serviceability.sql`)
- **Tenants Table Enhancements**:
  - `slug`: unique URL identifier.
  - `tier`: enterprise tier (`STANDARD`, `ENTERPRISE`).
  - `quota_parcels_per_day`, `quota_branches`.
  - `branding`: custom portal theme, primary/accent colors, logo URL.
  - `config`: timezone, currency, auto-assign policies.
- **Branches Table**:
  - `id`: UUID primary key.
  - `tenant_id`: foreign key with cascade delete.
  - `code`: unique per tenant.
  - `name`: descriptive name.
  - `branch_type`: `SORTING_HUB`, `DISTRIBUTION_CENTER`, `LOCAL_OFFICE`.
  - `address`, `city`, `state`, `postal_code`, `country`, `contact_phone`, `contact_email`.
  - `daily_capacity`: volume threshold.
  - `status`: `ACTIVE`, `INACTIVE`, `MAINTENANCE`.
  - `location`: `GEOMETRY(Point, 4326)` representing spatial coordinates.
  - `service_area`: `GEOMETRY(Polygon, 4326)` representing delivery service boundaries.
- **Spatial Indices**:
  - `idx_branches_location` using `GIST (location)`.
  - `idx_branches_service_area` using `GIST (service_area)`.
  - `idx_branches_tenant` on `branches(tenant_id)`.

---

### 2.2 Authoritative Go Backend (`internal/tenancy`)
- **Spatial Point-in-Polygon Engine**:
  - Evaluates containment using `ST_Contains(service_area, ST_SetSRID(ST_MakePoint(lon, lat), 4326))`.
  - Calculates great-circle ellipsoid distance in meters using `ST_Distance(location::geography, ...::geography)`.
  - Nearest-neighbor fallback via GiST spatial distance operator `<->`.
- **Tenant Isolation**:
  - All branch queries strictly enforce `tenant_id = $1`.
  - Verified by integration tests: Tenant A can never query or service Tenant B's branches.
- **GeoJSON Service Area Export**:
  - `ST_AsGeoJSON(service_area)` generates GeoJSON `FeatureCollection` for interactive web mapping.

---

### 2.3 REST API Endpoints
| Method | Route | Permission / Role | Description |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/serviceability/check` | Public / Authenticated | Evaluates if coordinate is serviceable by tenant |
| `GET` | `/api/v1/serviceability/check` | Public / Authenticated | Query param coordinate serviceability test |
| `POST` | `/api/v1/tenants` | `PLATFORM_ADMIN` | Provisions enterprise tenant |
| `GET` | `/api/v1/tenants` | `PermTenantRead` | Lists tenants (scoped to caller) |
| `GET` | `/api/v1/tenants/{tenant_id}` | `PermTenantRead` | Retrieves tenant metadata |
| `PUT` | `/api/v1/tenants/{tenant_id}` | `PermTenantUpdate` | Updates tenant configuration/branding |
| `GET` | `/api/v1/tenants/{tenant_id}/serviceability/coverage` | Authenticated | Returns GeoJSON FeatureCollection of all branch polygons |
| `GET` | `/api/v1/tenants/{tenant_id}/branches` | Authenticated | Lists tenant's branches |
| `POST` | `/api/v1/tenants/{tenant_id}/branches` | `PermBranchManage` | Deploys new branch or hub |
| `GET` | `/api/v1/tenants/{tenant_id}/branches/{branch_id}` | Authenticated | Retrieves branch details |
| `PUT` | `/api/v1/tenants/{tenant_id}/branches/{branch_id}` | `PermBranchManage` | Updates branch details and polygon |
| `DELETE`| `/api/v1/tenants/{tenant_id}/branches/{branch_id}` | `PermBranchManage` | Removes branch from network |

---

### 2.4 Single Portal Web UI (`web/`)
- **Interactive PostGIS Serviceability Map (`GeospatialMap.tsx`)**:
  - Dark-mode OpenStreetMap tiles (CartoDB Dark Matter).
  - Colored branch markers with pulsation animation (`SORTING_HUB` in Purple, `DISTRIBUTION_CENTER` in Blue, `LOCAL_OFFICE` in Emerald).
  - Polygon overlay rendering with translucent fills and dashed borders for coverage zones.
  - Interactive click-to-test pin drop with real-time HUD evaluation badge showing distance, matched branch, and transit ETA.
  - Preset test coordinate buttons (MG Road, Hebbal, Koramangala, Out of Area).
- **Branch Deployment Modal (`CreateBranchModal.tsx`)**:
  - Modal form for deploying new branches with coordinate inputs and automated 16-point closed polygon generator.
- **Tenant & Tenant Admin Dashboards**:
  - Real-time aggregate statistics: Total Branches, Sorting Hubs, Distribution Centers, and Total Daily Parcel Quota.

---

## 3. Test Verification Results

### 3.1 Live PostGIS 16 Integration Tests (`internal/tenancy/integration_test.go`)
- **Point inside Central Hub polygon**: matched `BLR-HUB-01` (`PASS`).
- **Point inside North DC polygon**: matched `BLR-DC-NORTH` (`PASS`).
- **Point inside South DC polygon**: matched `BLR-DC-SOUTH` (`PASS`).
- **Multi-tenant isolation**: Chennai coordinates tested against Bengaluru tenant returned `OUT_OF_COVERAGE`, but tested against Chennai tenant returned `SERVICEABLE` with matched `CHN-HUB-01` (`PASS`).
- **Coverage GeoJSON FeatureCollection**: successfully returned 3 polygon features (`PASS`).

### 3.2 HTTP API & Unit Tests
- `go test -v ./internal/tenancy/...`: `PASS`
- `go test -v ./internal/httpapi/...`: `PASS`
- `npm run build` in `web/`: `PASS`
- `flutter test` in `mobile/`: `PASS`
