# Phase 3 — Fleet Management, Vehicle Telematics & Asset Tracking

## 1. Executive Summary
Phase 3 implements the authoritative fleet management, vehicle telematics, and preventative asset tracking subsystem for **LogiFlows**. It introduces physical vehicle tracking across 5 vehicle classes (`VAN`, `MOTORCYCLE`, `TRUCK`, `ELECTRIC_VAN`, `BICYCLE`), sub-second GPS/OBD-II spatial telemetry ingestion, live fleet position querying, historical breadcrumbs playback, preventative maintenance logging, and resilient routing via OpenRouteService with automated geodesic fallback when API keys are unconfigured or invalid.

---

## 2. Key Deliverables

### 2.1 Resilient OpenRouteService (ORS) Routing Engine (`internal/routing`)
- **Configurable API Key**:
  - Automatically loads `ORS_API_KEY` from environment / `.env`.
  - Distinguishes between valid configured keys, missing keys, and invalid/unauthorized keys (`401`/`403`).
- **Graceful Simulated Geodesic Fallback**:
  - Automatically intercepts invalid key errors or rate limits, falling back to multi-point great-circle geodesic route geometry with realistic road waypoint perturbations.
  - Zero fatal errors or hard outages when external routing providers experience key revocation or quota depletion.
- **Key Status Diagnostics**:
  - Added `KeyStatus()` reporting key configured status, validity, and error state.

---

### 2.2 Database & PostGIS 16 Telematics Schema (`0004_fleet_telematics.sql`)
- **Vehicles Table (`vehicles`)**:
  - `id`: UUID primary key.
  - `tenant_id`: Foreign key enforcing multi-tenant isolation.
  - `assigned_branch_id`: Hub/office assignment (`branches.id`).
  - `assigned_driver_id`: Dedicated driver identity (`users.id`).
  - `license_plate`: Normalized uppercase registration plate.
  - `vin`: Vehicle Identification Number.
  - `make`, `model`, `year`.
  - `vehicle_type`: `VAN`, `MOTORCYCLE`, `TRUCK`, `ELECTRIC_VAN`, `BICYCLE`.
  - `fuel_type`: `DIESEL`, `PETROL`, `ELECTRIC`, `HYBRID`.
  - `capacity_kg`, `capacity_volume_m3`, `max_parcels`.
  - `current_mileage_km`, `battery_or_fuel_level_percent`.
  - `status`: `AVAILABLE`, `ON_ROUTE`, `MAINTENANCE`, `DECOMMISSIONED`.
  - `metadata`: Flexible JSONB properties.
- **Vehicle Telematics Table (`vehicle_telematics`)**:
  - `id`: UUID primary key.
  - `vehicle_id`, `tenant_id`.
  - `location`: `GEOMETRY(Point, 4326)` spatial coordinate.
  - `speed_kmh`, `heading_degrees`, `battery_or_fuel_percent`, `odometer_km`.
  - `recorded_at`: Timestamp of telemetry ping.
  - Spatial index: `GIST (location)`.
  - Compound time-series index: `(vehicle_id, recorded_at DESC)`.
- **Maintenance Records Table (`maintenance_records`)**:
  - `id`: UUID primary key.
  - `vehicle_id`, `tenant_id`.
  - `service_type`: `OIL_CHANGE`, `TIRE_ROTATION`, `BATTERY_CHECK`, `BRAKE_INSPECTION`, `ANNUAL_OVERHAUL`.
  - `description`, `cost`, `odometer_reading_km`, `serviced_at`, `next_service_due_km`.

---

### 2.3 Authoritative Go Fleet Domain & REST API (`internal/fleet`, `internal/httpapi`)
- **Fleet Repository & Service**:
  - Full CRUD operations with strict tenant scoping.
  - High-throughput spatial telematics ingestion updating both `vehicle_telematics` and vehicle's live odometer/battery level in a transaction.
  - Live fleet snapshot combining vehicle attributes with their latest telemetry coordinates (`ST_Y(location)`, `ST_X(location)`).
  - Breadcrumbs history with configurable limit for trajectory route rendering.
  - Preventative maintenance logging and upcoming service alert engine (`odometer >= next_service_due_km - 1000`).
- **REST Endpoints**:
  | Method | Route | Permission / Role | Description |
  | :--- | :--- | :--- | :--- |
  | `GET` | `/api/v1/vehicles` | Authenticated | List tenant fleet vehicles with filters |
  | `POST` | `/api/v1/vehicles` | `PermVehicleManage` | Register a new fleet vehicle |
  | `GET` | `/api/v1/vehicles/{id}` | Authenticated | Get vehicle specifications and stats |
  | `PUT` | `/api/v1/vehicles/{id}` | `PermVehicleManage` | Update vehicle parameters |
  | `DELETE`| `/api/v1/vehicles/{id}` | `PermVehicleManage` | Decommission / remove vehicle |
  | `POST` | `/api/v1/vehicles/{id}/telematics` | Authenticated | Ingest GPS/OBD-II telematics ping |
  | `GET` | `/api/v1/vehicles/{id}/telematics/latest` | Authenticated | Get latest vehicle GPS location and telemetry |
  | `GET` | `/api/v1/vehicles/{id}/telematics/history` | Authenticated | Get GPS breadcrumbs trajectory history |
  | `GET` | `/api/v1/fleet/live` | Authenticated | Get real-time positions for all active vehicles |
  | `GET` | `/api/v1/vehicles/{id}/maintenance` | Authenticated | List vehicle maintenance service logs |
  | `POST` | `/api/v1/vehicles/{id}/maintenance` | `PermVehicleManage` | Log maintenance intervention |
  | `GET` | `/api/v1/fleet/maintenance/upcoming` | Authenticated | List vehicles due for maintenance |

---

### 2.4 Mobile Courier Telematics Integration (`mobile/`)
- **API Client Extensions**:
  - `getVehicles()`: Retrieves list of assigned branch vehicles.
  - `sendTelematicsPing(vehicleId, payload)`: Transmits courier GPS coordinates, heading, speed, and battery status.
  - `getLiveFleet()`: Retrieves real-time coordinates of nearby courier units.
- **Courier UI & Driver Workflow**:
  - Vehicle selection dropdown allowing drivers to pick their assigned vehicle.
  - Real-time telemetry dashboard card displaying live odometer, battery %, and GPS status.
  - Interactive "Simulate GPS Ping" action for immediate spatial coordinate transmission.

---

### 2.5 Single Portal Web UI & Telematics Map (`web/`)
- **Live Fleet Tracking Map (`FleetTrackingMap.tsx`)**:
  - Dark-mode CartoDB map rendering live vehicle markers.
  - Directional heading rotation using vehicle azimuth angle (`heading_degrees`).
  - Vehicle type icons and color-coded status badges (`AVAILABLE` in Emerald, `ON_ROUTE` in Sky, `MAINTENANCE` in Amber).
  - Selected vehicle breadcrumb path overlay (polyline) showing trajectory history.
  - Real-time HUD summary bar displaying total vehicles, active on route, available, and vehicles needing maintenance.
- **Vehicle Provisioning (`CreateVehicleModal.tsx`)**:
  - Enterprise modal dialog to register vehicles with physical capacities (weight, volume, parcels), powertrain, and branch assignment.
- **Fleet Management Hub (`FleetManagementView.tsx`)**:
  - Tabbed interface switching between Live Map view, Inventory Data Table, and Upcoming Maintenance alerts.
  - Embedded into `TenantAdminDashboard`, `TenantDashboard`, and `AdminDashboard`.

---

## 3. Test Verification Results

### 3.1 Backend Test Coverage
- **Routing Engine (`internal/routing/service_test.go`)**:
  - Tests valid key simulation, invalid key handling (`401`/`403`), and fallback geodesic geometry (`PASS`).
- **Migrations (`internal/migrations/migrate_test.go`)**:
  - Verifies migration files 0001 through 0004 (`PASS`).
- **Fleet Domain (`internal/fleet/fleet_test.go`)**:
  - Unit tests for vehicle validation, telematics validation, and maintenance validation (`PASS`).
- **HTTP API & PostGIS Integration (`internal/httpapi/fleet_test.go`)**:
  - Complete integration lifecycle test against live PostgreSQL 16 / PostGIS database:
    - Vehicle registration and retrieval.
    - PostGIS GPS telematics ingestion.
    - Latest telematics coordinate resolution (`ST_Y`, `ST_X`).
    - Breadcrumbs history ordering.
    - Live fleet spatial query.
    - Maintenance logging and upcoming maintenance detection.
    - Multi-tenant boundary isolation (Tenant B blocked from accessing Tenant A's vehicles).
    - Status: `PASS`.
- **OpenAPI & Swagger Documentation (`internal/swagger`)**:
  - Specification validation and embedded UI handler test (`PASS`).

### 3.2 Mobile App Test Coverage (`mobile/test/api_client_test.dart`)
- Telematics ping transmission test (`PASS`).
- Vehicles list fetching test (`PASS`).
- Live fleet coordinates retrieval test (`PASS`).
- Overall: 5/5 tests passed.

### 3.3 Web App Test Coverage (`web/tests/fleet.test.js`)
- Fleet service unit tests: list vehicles, fetch live fleet, fetch breadcrumbs history, and create vehicle (`PASS`).
- Production bundle compilation: `npm run build` completed with zero TypeScript/JSX errors.
