package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"logiflows/backend/internal/auth"
	"logiflows/backend/internal/config"
	"logiflows/backend/internal/fleet"
)

type mockFleetRepoForAPI struct {
	vehicles    map[string]*fleet.Vehicle
	telematics  map[string][]fleet.VehicleTelematics
	maintenance map[string][]fleet.MaintenanceRecord
}

func newMockFleetRepoForAPI() *mockFleetRepoForAPI {
	return &mockFleetRepoForAPI{
		vehicles:    make(map[string]*fleet.Vehicle),
		telematics:  make(map[string][]fleet.VehicleTelematics),
		maintenance: make(map[string][]fleet.MaintenanceRecord),
	}
}

func (m *mockFleetRepoForAPI) CreateVehicle(ctx context.Context, v *fleet.Vehicle) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	for _, existing := range m.vehicles {
		if existing.TenantID == v.TenantID && existing.LicensePlate == v.LicensePlate {
			return fleet.ErrAlreadyExists
		}
	}
	v.CreatedAt = time.Now().UTC()
	v.UpdatedAt = v.CreatedAt
	m.vehicles[v.ID.String()] = v
	return nil
}

func (m *mockFleetRepoForAPI) GetVehicleByID(ctx context.Context, tenantID, vehicleID uuid.UUID) (*fleet.Vehicle, error) {
	v, exists := m.vehicles[vehicleID.String()]
	if !exists || v.TenantID != tenantID {
		return nil, fleet.ErrNotFound
	}
	return v, nil
}

func (m *mockFleetRepoForAPI) ListVehicles(ctx context.Context, tenantID uuid.UUID, filter fleet.VehicleFilter) ([]fleet.Vehicle, error) {
	var list []fleet.Vehicle
	for _, v := range m.vehicles {
		if v.TenantID != tenantID {
			continue
		}
		if filter.Status != nil && v.Status != *filter.Status {
			continue
		}
		if filter.VehicleType != nil && v.VehicleType != *filter.VehicleType {
			continue
		}
		list = append(list, *v)
	}
	return list, nil
}

func (m *mockFleetRepoForAPI) UpdateVehicle(ctx context.Context, v *fleet.Vehicle) error {
	existing, exists := m.vehicles[v.ID.String()]
	if !exists || existing.TenantID != v.TenantID {
		return fleet.ErrNotFound
	}
	v.UpdatedAt = time.Now().UTC()
	m.vehicles[v.ID.String()] = v
	return nil
}

func (m *mockFleetRepoForAPI) DeleteVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) error {
	existing, exists := m.vehicles[vehicleID.String()]
	if !exists || existing.TenantID != tenantID {
		return fleet.ErrNotFound
	}
	delete(m.vehicles, vehicleID.String())
	return nil
}

func (m *mockFleetRepoForAPI) IngestTelematics(ctx context.Context, t *fleet.VehicleTelematics) error {
	v, exists := m.vehicles[t.VehicleID.String()]
	if !exists || v.TenantID != t.TenantID {
		return fleet.ErrNotFound
	}
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	m.telematics[t.VehicleID.String()] = append(m.telematics[t.VehicleID.String()], *t)
	if t.OdometerKM > v.CurrentMileageKM {
		v.CurrentMileageKM = t.OdometerKM
	}
	v.BatteryOrFuelLevelPercent = t.BatteryOrFuelPercent
	v.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *mockFleetRepoForAPI) GetLatestTelematics(ctx context.Context, tenantID, vehicleID uuid.UUID) (*fleet.VehicleTelematics, error) {
	list, exists := m.telematics[vehicleID.String()]
	if !exists || len(list) == 0 {
		return nil, fleet.ErrNotFound
	}
	return &list[len(list)-1], nil
}

func (m *mockFleetRepoForAPI) GetTelematicsHistory(ctx context.Context, tenantID, vehicleID uuid.UUID, limit int) ([]fleet.VehicleTelematics, error) {
	list := m.telematics[vehicleID.String()]
	if len(list) > limit {
		return list[len(list)-limit:], nil
	}
	return list, nil
}

func (m *mockFleetRepoForAPI) GetLiveFleetPositions(ctx context.Context, tenantID uuid.UUID) ([]fleet.LiveVehiclePosition, error) {
	var positions []fleet.LiveVehiclePosition
	for _, v := range m.vehicles {
		if v.TenantID != tenantID || v.Status == fleet.VehicleStatusDecommissioned {
			continue
		}
		p := fleet.LiveVehiclePosition{
			VehicleID:            v.ID,
			TenantID:             v.TenantID,
			LicensePlate:         v.LicensePlate,
			Make:                 v.Make,
			Model:                v.Model,
			VehicleType:          v.VehicleType,
			Status:               v.Status,
			AssignedBranchID:     v.AssignedBranchID,
			AssignedDriverID:     v.AssignedDriverID,
			BatteryOrFuelPercent: v.BatteryOrFuelLevelPercent,
			OdometerKM:           v.CurrentMileageKM,
			LastPingAt:           v.UpdatedAt,
		}
		if hist := m.telematics[v.ID.String()]; len(hist) > 0 {
			latest := hist[len(hist)-1]
			p.Latitude = latest.Latitude
			p.Longitude = latest.Longitude
			p.SpeedKMH = latest.SpeedKMH
			p.HeadingDegrees = latest.HeadingDegrees
			p.LastPingAt = latest.RecordedAt
		}
		positions = append(positions, p)
	}
	return positions, nil
}

func (m *mockFleetRepoForAPI) CreateMaintenanceRecord(ctx context.Context, rec *fleet.MaintenanceRecord) error {
	v, exists := m.vehicles[rec.VehicleID.String()]
	if !exists || v.TenantID != rec.TenantID {
		return fleet.ErrNotFound
	}
	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}
	m.maintenance[rec.VehicleID.String()] = append(m.maintenance[rec.VehicleID.String()], *rec)
	v.Status = fleet.VehicleStatusAvailable
	return nil
}

func (m *mockFleetRepoForAPI) ListMaintenanceRecords(ctx context.Context, tenantID, vehicleID uuid.UUID) ([]fleet.MaintenanceRecord, error) {
	return m.maintenance[vehicleID.String()], nil
}

func (m *mockFleetRepoForAPI) ListUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]fleet.Vehicle, error) {
	var list []fleet.Vehicle
	for _, v := range m.vehicles {
		if v.TenantID != tenantID {
			continue
		}
		for _, rec := range m.maintenance[v.ID.String()] {
			if rec.NextServiceDueKM != nil && v.CurrentMileageKM >= (*rec.NextServiceDueKM-500) {
				list = append(list, *v)
				break
			}
		}
	}
	return list, nil
}

func TestFleetHTTPAPI_FullLifecycleAndRBAC(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test_secret_key_at_least_32_bytes_long",
	}
	repo := newMockFleetRepoForAPI()
	fleetService := fleet.NewService(repo, nil)

	router := BuildRouter(ServerDeps{
		Config:       cfg,
		FleetService: fleetService,
		Version:      "test-0.3.0",
	})

	tenantA := uuid.New()
	tenantB := uuid.New()
	userAdminA := uuid.New()
	userCustomerA := uuid.New()

	tokenAdminA := createTestToken(t, userAdminA, auth.RoleTenantAdmin, &tenantA, cfg.JWTSecret, false)
	tokenCustomerA := createTestToken(t, userCustomerA, auth.RoleCustomer, &tenantA, cfg.JWTSecret, false)
	tokenAdminB := createTestToken(t, uuid.New(), auth.RoleTenantAdmin, &tenantB, cfg.JWTSecret, false)

	// 1. RoleCustomer should be forbidden from creating a vehicle (403 Forbidden)
	vehicleBody := []byte(`{
		"license_plate": "KA-01-EQ-1000",
		"make": "Mercedes-Benz",
		"model": "eVito",
		"year": 2024,
		"vehicle_type": "ELECTRIC_VAN",
		"fuel_type": "ELECTRIC",
		"capacity_kg": 1000.0,
		"capacity_volume_m3": 6.6,
		"max_parcels": 200,
		"battery_or_fuel_level_percent": 100.0
	}`)

	reqCust := httptest.NewRequest("POST", "/api/v1/vehicles", bytes.NewReader(vehicleBody))
	reqCust.Header.Set("Authorization", "Bearer "+tokenCustomerA)
	reqCust.Header.Set("Content-Type", "application/json")
	rrCust := httptest.NewRecorder()
	router.ServeHTTP(rrCust, reqCust)

	if rrCust.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for customer creating vehicle, got %d: %s", rrCust.Code, rrCust.Body.String())
	}

	// 2. RoleTenantAdmin should successfully create vehicle (201 Created)
	reqAdmin := httptest.NewRequest("POST", "/api/v1/vehicles", bytes.NewReader(vehicleBody))
	reqAdmin.Header.Set("Authorization", "Bearer "+tokenAdminA)
	reqAdmin.Header.Set("Content-Type", "application/json")
	rrAdmin := httptest.NewRecorder()
	router.ServeHTTP(rrAdmin, reqAdmin)

	if rrAdmin.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for admin, got %d: %s", rrAdmin.Code, rrAdmin.Body.String())
	}

	type testEnv[T any] struct {
		Data T `json:"data"`
	}

	var createdEnv testEnv[fleet.Vehicle]
	if err := json.Unmarshal(rrAdmin.Body.Bytes(), &createdEnv); err != nil {
		t.Fatalf("failed to decode created vehicle: %v", err)
	}
	createdVehicle := createdEnv.Data
	if createdVehicle.LicensePlate != "KA-01-EQ-1000" {
		t.Errorf("expected plate KA-01-EQ-1000, got %s", createdVehicle.LicensePlate)
	}

	// 3. List vehicles for Tenant A
	reqList := httptest.NewRequest("GET", "/api/v1/vehicles", nil)
	reqList.Header.Set("Authorization", "Bearer "+tokenAdminA)
	rrList := httptest.NewRecorder()
	router.ServeHTTP(rrList, reqList)

	if rrList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing vehicles, got %d", rrList.Code)
	}
	var listEnv testEnv[[]fleet.Vehicle]
	if err := json.Unmarshal(rrList.Body.Bytes(), &listEnv); err != nil {
		t.Fatalf("failed to decode vehicles list: %v", err)
	}
	vehiclesList := listEnv.Data
	if len(vehiclesList) != 1 {
		t.Fatalf("expected 1 vehicle for Tenant A, got %d", len(vehiclesList))
	}

	// 4. Multi-tenant isolation: Tenant B should NOT see Tenant A's vehicle
	reqListB := httptest.NewRequest("GET", "/api/v1/vehicles", nil)
	reqListB.Header.Set("Authorization", "Bearer "+tokenAdminB)
	rrListB := httptest.NewRecorder()
	router.ServeHTTP(rrListB, reqListB)

	if rrListB.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing vehicles for Tenant B, got %d", rrListB.Code)
	}
	var listEnvB testEnv[[]fleet.Vehicle]
	if err := json.Unmarshal(rrListB.Body.Bytes(), &listEnvB); err != nil {
		t.Fatalf("failed to decode vehicles list B: %v", err)
	}
	vehiclesListB := listEnvB.Data
	if len(vehiclesListB) != 0 {
		t.Fatalf("expected 0 vehicles for Tenant B, got %d", len(vehiclesListB))
	}

	// 5. Ingest GPS telematics ping
	telemBody := []byte(`{
		"latitude": 13.0827,
		"longitude": 80.2707,
		"speed_kmh": 48.0,
		"heading_degrees": 90.0,
		"battery_or_fuel_percent": 92.5,
		"odometer_km": 340.2
	}`)
	reqTelem := httptest.NewRequest("POST", "/api/v1/vehicles/"+createdVehicle.ID.String()+"/telematics", bytes.NewReader(telemBody))
	reqTelem.Header.Set("Authorization", "Bearer "+tokenAdminA)
	reqTelem.Header.Set("Content-Type", "application/json")
	rrTelem := httptest.NewRecorder()
	router.ServeHTTP(rrTelem, reqTelem)

	if rrTelem.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for telematics ping, got %d: %s", rrTelem.Code, rrTelem.Body.String())
	}

	// 6. Fetch latest telematics
	reqLatest := httptest.NewRequest("GET", "/api/v1/vehicles/"+createdVehicle.ID.String()+"/telematics/latest", nil)
	reqLatest.Header.Set("Authorization", "Bearer "+tokenAdminA)
	rrLatest := httptest.NewRecorder()
	router.ServeHTTP(rrLatest, reqLatest)

	if rrLatest.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for latest telematics, got %d: %s", rrLatest.Code, rrLatest.Body.String())
	}

	var latestEnv testEnv[fleet.VehicleTelematics]
	if err := json.Unmarshal(rrLatest.Body.Bytes(), &latestEnv); err != nil {
		t.Fatalf("failed to decode latest telematics: %v", err)
	}
	latestTelem := latestEnv.Data
	if latestTelem.SpeedKMH != 48.0 {
		t.Errorf("expected speed 48.0, got %f", latestTelem.SpeedKMH)
	}

	// 7. Get live fleet positions
	reqLive := httptest.NewRequest("GET", "/api/v1/fleet/live", nil)
	reqLive.Header.Set("Authorization", "Bearer "+tokenAdminA)
	rrLive := httptest.NewRecorder()
	router.ServeHTTP(rrLive, reqLive)

	if rrLive.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for live fleet positions, got %d: %s", rrLive.Code, rrLive.Body.String())
	}
	var liveEnv testEnv[[]fleet.LiveVehiclePosition]
	if err := json.Unmarshal(rrLive.Body.Bytes(), &liveEnv); err != nil {
		t.Fatalf("failed to decode live positions: %v", err)
	}
	livePositions := liveEnv.Data
	if len(livePositions) != 1 {
		t.Fatalf("expected 1 live vehicle, got %d", len(livePositions))
	}
	if livePositions[0].Latitude != 13.0827 || livePositions[0].Longitude != 80.2707 {
		t.Errorf("expected live position (13.0827, 80.2707), got (%f, %f)", livePositions[0].Latitude, livePositions[0].Longitude)
	}

	// 8. Log maintenance record
	maintBody := []byte(`{
		"service_type": "TIRE_ROTATION",
		"description": "Front axle tire rotation and pressure calibration",
		"cost": 85.00,
		"odometer_reading_km": 340.2,
		"next_service_due_km": 10000.0
	}`)
	reqMaint := httptest.NewRequest("POST", "/api/v1/vehicles/"+createdVehicle.ID.String()+"/maintenance", bytes.NewReader(maintBody))
	reqMaint.Header.Set("Authorization", "Bearer "+tokenAdminA)
	reqMaint.Header.Set("Content-Type", "application/json")
	rrMaint := httptest.NewRecorder()
	router.ServeHTTP(rrMaint, reqMaint)

	if rrMaint.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for maintenance, got %d: %s", rrMaint.Code, rrMaint.Body.String())
	}

	// 9. List maintenance records
	reqListMaint := httptest.NewRequest("GET", "/api/v1/vehicles/"+createdVehicle.ID.String()+"/maintenance", nil)
	reqListMaint.Header.Set("Authorization", "Bearer "+tokenAdminA)
	rrListMaint := httptest.NewRecorder()
	router.ServeHTTP(rrListMaint, reqListMaint)

	if rrListMaint.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for list maintenance, got %d", rrListMaint.Code)
	}
	var maintEnv testEnv[[]fleet.MaintenanceRecord]
	if err := json.Unmarshal(rrListMaint.Body.Bytes(), &maintEnv); err != nil {
		t.Fatalf("failed to decode maintenance records: %v", err)
	}
	maintRecords := maintEnv.Data
	if len(maintRecords) != 1 {
		t.Fatalf("expected 1 maintenance record, got %d", len(maintRecords))
	}
}
