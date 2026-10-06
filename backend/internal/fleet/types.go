package fleet

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// VehicleType represents the physical class of delivery vehicle
type VehicleType string

const (
	VehicleTypeVan         VehicleType = "VAN"
	VehicleTypeMotorcycle  VehicleType = "MOTORCYCLE"
	VehicleTypeTruck        VehicleType = "TRUCK"
	VehicleTypeElectricVan VehicleType = "ELECTRIC_VAN"
	VehicleTypeBicycle     VehicleType = "BICYCLE"
)

func (v VehicleType) IsValid() bool {
	switch v {
	case VehicleTypeVan, VehicleTypeMotorcycle, VehicleTypeTruck, VehicleTypeElectricVan, VehicleTypeBicycle:
		return true
	default:
		return false
	}
}

// FuelType represents the powertrain energy source
type FuelType string

const (
	FuelTypeDiesel   FuelType = "DIESEL"
	FuelTypePetrol   FuelType = "PETROL"
	FuelTypeElectric FuelType = "ELECTRIC"
	FuelTypeHybrid   FuelType = "HYBRID"
)

func (f FuelType) IsValid() bool {
	switch f {
	case FuelTypeDiesel, FuelTypePetrol, FuelTypeElectric, FuelTypeHybrid:
		return true
	default:
		return false
	}
}

// VehicleStatus represents the operational readiness of the asset
type VehicleStatus string

const (
	VehicleStatusAvailable      VehicleStatus = "AVAILABLE"
	VehicleStatusOnRoute        VehicleStatus = "ON_ROUTE"
	VehicleStatusMaintenance    VehicleStatus = "MAINTENANCE"
	VehicleStatusDecommissioned VehicleStatus = "DECOMMISSIONED"
)

func (s VehicleStatus) IsValid() bool {
	switch s {
	case VehicleStatusAvailable, VehicleStatusOnRoute, VehicleStatusMaintenance, VehicleStatusDecommissioned:
		return true
	default:
		return false
	}
}

// Vehicle represents an authoritative fleet vehicle asset
type Vehicle struct {
	ID                       uuid.UUID       `json:"id"`
	TenantID                 uuid.UUID       `json:"tenant_id"`
	AssignedBranchID         *uuid.UUID      `json:"assigned_branch_id,omitempty"`
	AssignedDriverID         *uuid.UUID      `json:"assigned_driver_id,omitempty"`
	LicensePlate             string          `json:"license_plate"`
	VIN                      string          `json:"vin"`
	Make                     string          `json:"make"`
	Model                    string          `json:"model"`
	Year                     int             `json:"year"`
	VehicleType              VehicleType     `json:"vehicle_type"`
	FuelType                 FuelType        `json:"fuel_type"`
	CapacityKG               float64         `json:"capacity_kg"`
	CapacityVolumeM3         float64         `json:"capacity_volume_m3"`
	MaxParcels               int             `json:"max_parcels"`
	CurrentMileageKM         float64         `json:"current_mileage_km"`
	BatteryOrFuelLevelPercent float64        `json:"battery_or_fuel_level_percent"`
	Status                   VehicleStatus   `json:"status"`
	Metadata                 json.RawMessage `json:"metadata"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
}

func (v *Vehicle) Validate() error {
	v.LicensePlate = strings.ToUpper(strings.TrimSpace(v.LicensePlate))
	if v.LicensePlate == "" {
		return errors.New("license_plate is required")
	}
	if v.TenantID == uuid.Nil {
		return errors.New("tenant_id is required")
	}
	v.Make = strings.TrimSpace(v.Make)
	if v.Make == "" {
		return errors.New("make is required")
	}
	v.Model = strings.TrimSpace(v.Model)
	if v.Model == "" {
		return errors.New("model is required")
	}
	if v.Year < 1990 || v.Year > time.Now().Year()+2 {
		return fmt.Errorf("invalid vehicle year %d", v.Year)
	}
	if !v.VehicleType.IsValid() {
		return fmt.Errorf("invalid vehicle_type: %s", v.VehicleType)
	}
	if !v.FuelType.IsValid() {
		return fmt.Errorf("invalid fuel_type: %s", v.FuelType)
	}
	if !v.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", v.Status)
	}
	if v.CapacityKG <= 0 {
		return errors.New("capacity_kg must be greater than zero")
	}
	if v.CapacityVolumeM3 <= 0 {
		return errors.New("capacity_volume_m3 must be greater than zero")
	}
	if v.MaxParcels <= 0 {
		return errors.New("max_parcels must be greater than zero")
	}
	if v.BatteryOrFuelLevelPercent < 0 || v.BatteryOrFuelLevelPercent > 100 {
		return errors.New("battery_or_fuel_level_percent must be between 0 and 100")
	}
	if len(v.Metadata) == 0 {
		v.Metadata = json.RawMessage("{}")
	}
	return nil
}

// CreateVehicleDTO represents payload to register a new vehicle
type CreateVehicleDTO struct {
	AssignedBranchID          *uuid.UUID      `json:"assigned_branch_id,omitempty"`
	AssignedDriverID          *uuid.UUID      `json:"assigned_driver_id,omitempty"`
	LicensePlate              string          `json:"license_plate"`
	VIN                       string          `json:"vin,omitempty"`
	Make                      string          `json:"make"`
	Model                     string          `json:"model"`
	Year                      int             `json:"year"`
	VehicleType               VehicleType     `json:"vehicle_type"`
	FuelType                  FuelType        `json:"fuel_type"`
	CapacityKG                float64         `json:"capacity_kg"`
	CapacityVolumeM3          float64         `json:"capacity_volume_m3"`
	MaxParcels                int             `json:"max_parcels"`
	CurrentMileageKM          float64         `json:"current_mileage_km,omitempty"`
	BatteryOrFuelLevelPercent float64         `json:"battery_or_fuel_level_percent,omitempty"`
	Status                    VehicleStatus   `json:"status,omitempty"`
	Metadata                  json.RawMessage `json:"metadata,omitempty"`
}

// UpdateVehicleDTO represents payload to update an existing vehicle
type UpdateVehicleDTO struct {
	AssignedBranchID          *uuid.UUID      `json:"assigned_branch_id,omitempty"`
	AssignedDriverID          *uuid.UUID      `json:"assigned_driver_id,omitempty"`
	Make                      *string         `json:"make,omitempty"`
	Model                     *string         `json:"model,omitempty"`
	Year                      *int            `json:"year,omitempty"`
	VehicleType               *VehicleType    `json:"vehicle_type,omitempty"`
	FuelType                  *FuelType       `json:"fuel_type,omitempty"`
	CapacityKG                *float64        `json:"capacity_kg,omitempty"`
	CapacityVolumeM3          *float64        `json:"capacity_volume_m3,omitempty"`
	MaxParcels                *int            `json:"max_parcels,omitempty"`
	CurrentMileageKM          *float64        `json:"current_mileage_km,omitempty"`
	BatteryOrFuelLevelPercent *float64        `json:"battery_or_fuel_level_percent,omitempty"`
	Status                    *VehicleStatus  `json:"status,omitempty"`
	Metadata                  json.RawMessage `json:"metadata,omitempty"`
}

// VehicleFilter specifies listing criteria
type VehicleFilter struct {
	BranchID   *uuid.UUID     `json:"branch_id,omitempty"`
	DriverID   *uuid.UUID     `json:"driver_id,omitempty"`
	Status     *VehicleStatus `json:"status,omitempty"`
	VehicleType *VehicleType  `json:"vehicle_type,omitempty"`
}

// VehicleTelematics represents a spatial telemetry snapshot from vehicle GPS/OBD-II
type VehicleTelematics struct {
	ID                   uuid.UUID `json:"id"`
	VehicleID            uuid.UUID `json:"vehicle_id"`
	TenantID             uuid.UUID `json:"tenant_id"`
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	SpeedKMH             float64   `json:"speed_kmh"`
	HeadingDegrees       float64   `json:"heading_degrees"`
	BatteryOrFuelPercent float64   `json:"battery_or_fuel_percent"`
	OdometerKM           float64   `json:"odometer_km"`
	RecordedAt           time.Time `json:"recorded_at"`
}

func (t *VehicleTelematics) Validate() error {
	if t.VehicleID == uuid.Nil {
		return errors.New("vehicle_id is required")
	}
	if t.TenantID == uuid.Nil {
		return errors.New("tenant_id is required")
	}
	if t.Latitude < -90 || t.Latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if t.Longitude < -180 || t.Longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	if t.BatteryOrFuelPercent < 0 || t.BatteryOrFuelPercent > 100 {
		return errors.New("battery_or_fuel_percent must be between 0 and 100")
	}
	if t.RecordedAt.IsZero() {
		t.RecordedAt = time.Now().UTC()
	}
	return nil
}

// IngestTelematicsDTO represents GPS ping payload from on-board telematics unit
type IngestTelematicsDTO struct {
	Latitude             float64   `json:"latitude"`
	Longitude            float64   `json:"longitude"`
	SpeedKMH             float64   `json:"speed_kmh"`
	HeadingDegrees       float64   `json:"heading_degrees"`
	BatteryOrFuelPercent float64   `json:"battery_or_fuel_percent"`
	OdometerKM           float64   `json:"odometer_km"`
	RecordedAt           time.Time `json:"recorded_at,omitempty"`
}

// LiveVehiclePosition provides combined vehicle & real-time telemetry coordinates
type LiveVehiclePosition struct {
	VehicleID            uuid.UUID     `json:"vehicle_id"`
	TenantID             uuid.UUID     `json:"tenant_id"`
	LicensePlate         string        `json:"license_plate"`
	Make                 string        `json:"make"`
	Model                string        `json:"model"`
	VehicleType          VehicleType   `json:"vehicle_type"`
	Status               VehicleStatus `json:"status"`
	AssignedBranchID     *uuid.UUID    `json:"assigned_branch_id,omitempty"`
	AssignedDriverID     *uuid.UUID    `json:"assigned_driver_id,omitempty"`
	Latitude             float64       `json:"latitude"`
	Longitude            float64       `json:"longitude"`
	SpeedKMH             float64       `json:"speed_kmh"`
	HeadingDegrees       float64       `json:"heading_degrees"`
	BatteryOrFuelPercent float64       `json:"battery_or_fuel_percent"`
	OdometerKM           float64       `json:"odometer_km"`
	LastPingAt           time.Time     `json:"last_ping_at"`
}

// MaintenanceRecord represents a scheduled or performed maintenance intervention
type MaintenanceRecord struct {
	ID                uuid.UUID `json:"id"`
	VehicleID         uuid.UUID `json:"vehicle_id"`
	TenantID          uuid.UUID `json:"tenant_id"`
	ServiceType       string    `json:"service_type"` // OIL_CHANGE, TIRE_ROTATION, BATTERY_CHECK, BRAKE_INSPECTION, ANNUAL_OVERHAUL
	Description       string    `json:"description"`
	Cost              float64   `json:"cost"`
	OdometerReadingKM float64   `json:"odometer_reading_km"`
	ServicedAt        time.Time `json:"serviced_at"`
	NextServiceDueKM  *float64  `json:"next_service_due_km,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

func (m *MaintenanceRecord) Validate() error {
	if m.VehicleID == uuid.Nil {
		return errors.New("vehicle_id is required")
	}
	if m.TenantID == uuid.Nil {
		return errors.New("tenant_id is required")
	}
	m.ServiceType = strings.TrimSpace(m.ServiceType)
	if m.ServiceType == "" {
		return errors.New("service_type is required")
	}
	if m.Cost < 0 {
		return errors.New("cost cannot be negative")
	}
	if m.ServicedAt.IsZero() {
		m.ServicedAt = time.Now().UTC()
	}
	return nil
}

// CreateMaintenanceDTO represents payload to log a maintenance event
type CreateMaintenanceDTO struct {
	ServiceType       string    `json:"service_type"`
	Description       string    `json:"description,omitempty"`
	Cost              float64   `json:"cost"`
	OdometerReadingKM float64   `json:"odometer_reading_km"`
	ServicedAt        time.Time `json:"serviced_at,omitempty"`
	NextServiceDueKM  *float64  `json:"next_service_due_km,omitempty"`
}
