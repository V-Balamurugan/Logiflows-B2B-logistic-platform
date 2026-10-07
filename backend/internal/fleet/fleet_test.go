package fleet

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockFleetRepo struct {
	vehicles      map[string]*Vehicle
	telematics    map[string][]VehicleTelematics
	maintenance   map[string][]MaintenanceRecord
}

func newMockFleetRepo() *mockFleetRepo {
	return &mockFleetRepo{
		vehicles:    make(map[string]*Vehicle),
		telematics:  make(map[string][]VehicleTelematics),
		maintenance: make(map[string][]MaintenanceRecord),
	}
}

func (m *mockFleetRepo) CreateVehicle(ctx context.Context, v *Vehicle) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	for _, existing := range m.vehicles {
		if existing.TenantID == v.TenantID && existing.LicensePlate == v.LicensePlate {
			return ErrAlreadyExists
		}
	}
	v.CreatedAt = time.Now().UTC()
	v.UpdatedAt = v.CreatedAt
	m.vehicles[v.ID.String()] = v
	return nil
}

func (m *mockFleetRepo) GetVehicleByID(ctx context.Context, tenantID, vehicleID uuid.UUID) (*Vehicle, error) {
	v, exists := m.vehicles[vehicleID.String()]
	if !exists || v.TenantID != tenantID {
		return nil, ErrNotFound
	}
	return v, nil
}

func (m *mockFleetRepo) ListVehicles(ctx context.Context, tenantID uuid.UUID, filter VehicleFilter) ([]Vehicle, error) {
	var list []Vehicle
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

func (m *mockFleetRepo) UpdateVehicle(ctx context.Context, v *Vehicle) error {
	existing, exists := m.vehicles[v.ID.String()]
	if !exists || existing.TenantID != v.TenantID {
		return ErrNotFound
	}
	v.UpdatedAt = time.Now().UTC()
	m.vehicles[v.ID.String()] = v
	return nil
}

func (m *mockFleetRepo) DeleteVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) error {
	existing, exists := m.vehicles[vehicleID.String()]
	if !exists || existing.TenantID != tenantID {
		return ErrNotFound
	}
	delete(m.vehicles, vehicleID.String())
	return nil
}

func (m *mockFleetRepo) IngestTelematics(ctx context.Context, t *VehicleTelematics) error {
	v, exists := m.vehicles[t.VehicleID.String()]
	if !exists || v.TenantID != t.TenantID {
		return ErrNotFound
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

func (m *mockFleetRepo) GetLatestTelematics(ctx context.Context, tenantID, vehicleID uuid.UUID) (*VehicleTelematics, error) {
	list, exists := m.telematics[vehicleID.String()]
	if !exists || len(list) == 0 {
		return nil, ErrNotFound
	}
	return &list[len(list)-1], nil
}

func (m *mockFleetRepo) GetTelematicsHistory(ctx context.Context, tenantID, vehicleID uuid.UUID, limit int) ([]VehicleTelematics, error) {
	list := m.telematics[vehicleID.String()]
	if len(list) > limit {
		return list[len(list)-limit:], nil
	}
	return list, nil
}

func (m *mockFleetRepo) GetLiveFleetPositions(ctx context.Context, tenantID uuid.UUID) ([]LiveVehiclePosition, error) {
	var positions []LiveVehiclePosition
	for _, v := range m.vehicles {
		if v.TenantID != tenantID || v.Status == VehicleStatusDecommissioned {
			continue
		}
		p := LiveVehiclePosition{
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

func (m *mockFleetRepo) CreateMaintenanceRecord(ctx context.Context, rec *MaintenanceRecord) error {
	v, exists := m.vehicles[rec.VehicleID.String()]
	if !exists || v.TenantID != rec.TenantID {
		return ErrNotFound
	}
	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}
	m.maintenance[rec.VehicleID.String()] = append(m.maintenance[rec.VehicleID.String()], *rec)
	v.Status = VehicleStatusAvailable
	return nil
}

func (m *mockFleetRepo) ListMaintenanceRecords(ctx context.Context, tenantID, vehicleID uuid.UUID) ([]MaintenanceRecord, error) {
	return m.maintenance[vehicleID.String()], nil
}

func (m *mockFleetRepo) ListUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]Vehicle, error) {
	var list []Vehicle
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

func TestVehicleValidation(t *testing.T) {
	tenantID := uuid.New()

	valid := &Vehicle{
		TenantID:                  tenantID,
		LicensePlate:              "EV-VAN-101",
		Make:                      "Mercedes-Benz",
		Model:                     "eVito",
		Year:                      2024,
		VehicleType:               VehicleTypeElectricVan,
		FuelType:                  FuelTypeElectric,
		CapacityKG:                1000.0,
		CapacityVolumeM3:          6.6,
		MaxParcels:                200,
		BatteryOrFuelLevelPercent: 95.0,
		Status:                    VehicleStatusAvailable,
		Metadata:                  json.RawMessage("{}"),
	}

	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid vehicle, got %v", err)
	}

	// Missing license plate
	invalidPlate := *valid
	invalidPlate.LicensePlate = ""
	if err := invalidPlate.Validate(); err == nil {
		t.Errorf("expected error for empty license plate")
	}

	// Invalid vehicle type
	invalidType := *valid
	invalidType.VehicleType = "SUBMARINE"
	if err := invalidType.Validate(); err == nil {
		t.Errorf("expected error for invalid vehicle type")
	}

	// Negative capacity
	invalidCapacity := *valid
	invalidCapacity.CapacityKG = -10
	if err := invalidCapacity.Validate(); err == nil {
		t.Errorf("expected error for negative capacity")
	}
}

func TestFleetService_Lifecycle(t *testing.T) {
	repo := newMockFleetRepo()
	svc := NewService(repo, nil)
	ctx := context.Background()
	tenantID := uuid.New()
	otherTenantID := uuid.New()

	// 1. Register vehicle
	v, err := svc.RegisterVehicle(ctx, tenantID, CreateVehicleDTO{
		LicensePlate:              "KA-01-EQ-9999",
		Make:                      "Tata",
		Model:                     "Ace EV",
		Year:                      2024,
		VehicleType:               VehicleTypeElectricVan,
		FuelType:                  FuelTypeElectric,
		CapacityKG:                600.0,
		CapacityVolumeM3:          4.5,
		MaxParcels:                120,
		BatteryOrFuelLevelPercent: 100.0,
	})
	if err != nil {
		t.Fatalf("failed to register vehicle: %v", err)
	}
	if v.Status != VehicleStatusAvailable {
		t.Errorf("expected default status AVAILABLE, got %s", v.Status)
	}

	// 2. Fetch vehicle
	fetched, err := svc.GetVehicle(ctx, tenantID, v.ID)
	if err != nil {
		t.Fatalf("failed to fetch vehicle: %v", err)
	}
	if fetched.LicensePlate != "KA-01-EQ-9999" {
		t.Errorf("expected plate KA-01-EQ-9999, got %s", fetched.LicensePlate)
	}

	// 3. Multi-tenant isolation check: other tenant cannot fetch
	_, err = svc.GetVehicle(ctx, otherTenantID, v.ID)
	if err == nil {
		t.Errorf("expected error fetching vehicle from another tenant")
	}

	// 4. Ingest telematics GPS ping
	telem, err := svc.RecordTelematics(ctx, tenantID, v.ID, IngestTelematicsDTO{
		Latitude:             12.9716,
		Longitude:            77.5946,
		SpeedKMH:             42.5,
		HeadingDegrees:       180.0,
		BatteryOrFuelPercent: 88.0,
		OdometerKM:           125.4,
	})
	if err != nil {
		t.Fatalf("failed to ingest telematics: %v", err)
	}
	if telem.SpeedKMH != 42.5 {
		t.Errorf("expected speed 42.5, got %f", telem.SpeedKMH)
	}

	// Verify vehicle current mileage updated
	updatedV, err := svc.GetVehicle(ctx, tenantID, v.ID)
	if err != nil {
		t.Fatalf("failed to get updated vehicle: %v", err)
	}
	if updatedV.CurrentMileageKM != 125.4 {
		t.Errorf("expected updated mileage 125.4, got %f", updatedV.CurrentMileageKM)
	}
	if updatedV.BatteryOrFuelLevelPercent != 88.0 {
		t.Errorf("expected battery 88.0, got %f", updatedV.BatteryOrFuelLevelPercent)
	}

	// 5. Test Live Fleet Positions
	livePositions, err := svc.GetLiveFleetPositions(ctx, tenantID)
	if err != nil {
		t.Fatalf("failed to get live positions: %v", err)
	}
	if len(livePositions) != 1 {
		t.Fatalf("expected 1 live vehicle, got %d", len(livePositions))
	}
	if livePositions[0].Latitude != 12.9716 || livePositions[0].Longitude != 77.5946 {
		t.Errorf("expected lat 12.9716 lon 77.5946, got %f, %f", livePositions[0].Latitude, livePositions[0].Longitude)
	}

	// 6. Record maintenance event
	nextDue := 5000.0
	maint, err := svc.RecordMaintenance(ctx, tenantID, v.ID, CreateMaintenanceDTO{
		ServiceType:       "BRAKE_INSPECTION",
		Description:       "Routine brake pad and fluid replacement",
		Cost:              150.00,
		OdometerReadingKM: 125.4,
		NextServiceDueKM:  &nextDue,
	})
	if err != nil {
		t.Fatalf("failed to log maintenance: %v", err)
	}
	if maint.Cost != 150.00 {
		t.Errorf("expected cost 150.00, got %f", maint.Cost)
	}

	// 7. List maintenance records
	records, err := svc.ListMaintenanceRecords(ctx, tenantID, v.ID)
	if err != nil {
		t.Fatalf("failed to list maintenance records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}
