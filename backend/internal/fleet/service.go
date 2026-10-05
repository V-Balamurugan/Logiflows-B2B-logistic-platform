package fleet

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service interface {
	RegisterVehicle(ctx context.Context, tenantID uuid.UUID, dto CreateVehicleDTO) (*Vehicle, error)
	GetVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) (*Vehicle, error)
	ListVehicles(ctx context.Context, tenantID uuid.UUID, filter VehicleFilter) ([]Vehicle, error)
	UpdateVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID, dto UpdateVehicleDTO) (*Vehicle, error)
	DeleteVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) error

	RecordTelematics(ctx context.Context, tenantID, vehicleID uuid.UUID, dto IngestTelematicsDTO) (*VehicleTelematics, error)
	GetLatestTelematics(ctx context.Context, tenantID, vehicleID uuid.UUID) (*VehicleTelematics, error)
	GetTelematicsHistory(ctx context.Context, tenantID, vehicleID uuid.UUID, limit int) ([]VehicleTelematics, error)
	GetLiveFleetPositions(ctx context.Context, tenantID uuid.UUID) ([]LiveVehiclePosition, error)

	RecordMaintenance(ctx context.Context, tenantID, vehicleID uuid.UUID, dto CreateMaintenanceDTO) (*MaintenanceRecord, error)
	ListMaintenanceRecords(ctx context.Context, tenantID, vehicleID uuid.UUID) ([]MaintenanceRecord, error)
	ListUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]Vehicle, error)
}

type fleetService struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &fleetService{
		repo:   repo,
		logger: logger,
	}
}

func (s *fleetService) RegisterVehicle(ctx context.Context, tenantID uuid.UUID, dto CreateVehicleDTO) (*Vehicle, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required")
	}

	status := dto.Status
	if status == "" {
		status = VehicleStatusAvailable
	}

	fuelType := dto.FuelType
	if fuelType == "" {
		fuelType = FuelTypeDiesel
	}

	batteryOrFuel := dto.BatteryOrFuelLevelPercent
	if batteryOrFuel <= 0 {
		batteryOrFuel = 100.0
	}

	v := &Vehicle{
		TenantID:                  tenantID,
		AssignedBranchID:         dto.AssignedBranchID,
		AssignedDriverID:         dto.AssignedDriverID,
		LicensePlate:              strings.ToUpper(strings.TrimSpace(dto.LicensePlate)),
		VIN:                       strings.TrimSpace(dto.VIN),
		Make:                      strings.TrimSpace(dto.Make),
		Model:                     strings.TrimSpace(dto.Model),
		Year:                      dto.Year,
		VehicleType:               dto.VehicleType,
		FuelType:                  fuelType,
		CapacityKG:                dto.CapacityKG,
		CapacityVolumeM3:          dto.CapacityVolumeM3,
		MaxParcels:                dto.MaxParcels,
		CurrentMileageKM:          dto.CurrentMileageKM,
		BatteryOrFuelLevelPercent: batteryOrFuel,
		Status:                    status,
		Metadata:                  dto.Metadata,
	}

	if err := v.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.CreateVehicle(ctx, v); err != nil {
		s.logger.ErrorContext(ctx, "failed to register vehicle", slog.String("error", err.Error()), slog.String("plate", v.LicensePlate))
		return nil, err
	}

	s.logger.InfoContext(ctx, "vehicle registered successfully", slog.String("id", v.ID.String()), slog.String("plate", v.LicensePlate))
	return v, nil
}

func (s *fleetService) GetVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) (*Vehicle, error) {
	return s.repo.GetVehicleByID(ctx, tenantID, vehicleID)
}

func (s *fleetService) ListVehicles(ctx context.Context, tenantID uuid.UUID, filter VehicleFilter) ([]Vehicle, error) {
	return s.repo.ListVehicles(ctx, tenantID, filter)
}

func (s *fleetService) UpdateVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID, dto UpdateVehicleDTO) (*Vehicle, error) {
	v, err := s.repo.GetVehicleByID(ctx, tenantID, vehicleID)
	if err != nil {
		return nil, err
	}

	if dto.AssignedBranchID != nil {
		v.AssignedBranchID = dto.AssignedBranchID
	}
	if dto.AssignedDriverID != nil {
		v.AssignedDriverID = dto.AssignedDriverID
	}
	if dto.Make != nil {
		v.Make = strings.TrimSpace(*dto.Make)
	}
	if dto.Model != nil {
		v.Model = strings.TrimSpace(*dto.Model)
	}
	if dto.Year != nil {
		v.Year = *dto.Year
	}
	if dto.VehicleType != nil {
		v.VehicleType = *dto.VehicleType
	}
	if dto.FuelType != nil {
		v.FuelType = *dto.FuelType
	}
	if dto.CapacityKG != nil {
		v.CapacityKG = *dto.CapacityKG
	}
	if dto.CapacityVolumeM3 != nil {
		v.CapacityVolumeM3 = *dto.CapacityVolumeM3
	}
	if dto.MaxParcels != nil {
		v.MaxParcels = *dto.MaxParcels
	}
	if dto.CurrentMileageKM != nil {
		v.CurrentMileageKM = *dto.CurrentMileageKM
	}
	if dto.BatteryOrFuelLevelPercent != nil {
		v.BatteryOrFuelLevelPercent = *dto.BatteryOrFuelLevelPercent
	}
	if dto.Status != nil {
		v.Status = *dto.Status
	}
	if len(dto.Metadata) > 0 {
		v.Metadata = dto.Metadata
	}

	if err := v.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateVehicle(ctx, v); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "vehicle updated successfully", slog.String("id", v.ID.String()))
	return v, nil
}

func (s *fleetService) DeleteVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) error {
	return s.repo.DeleteVehicle(ctx, tenantID, vehicleID)
}

func (s *fleetService) RecordTelematics(ctx context.Context, tenantID, vehicleID uuid.UUID, dto IngestTelematicsDTO) (*VehicleTelematics, error) {
	recordedAt := dto.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = time.Now().UTC()
	}

	t := &VehicleTelematics{
		VehicleID:            vehicleID,
		TenantID:             tenantID,
		Latitude:             dto.Latitude,
		Longitude:            dto.Longitude,
		SpeedKMH:             dto.SpeedKMH,
		HeadingDegrees:       dto.HeadingDegrees,
		BatteryOrFuelPercent: dto.BatteryOrFuelPercent,
		OdometerKM:           dto.OdometerKM,
		RecordedAt:           recordedAt,
	}

	if err := t.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.IngestTelematics(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *fleetService) GetLatestTelematics(ctx context.Context, tenantID, vehicleID uuid.UUID) (*VehicleTelematics, error) {
	return s.repo.GetLatestTelematics(ctx, tenantID, vehicleID)
}

func (s *fleetService) GetTelematicsHistory(ctx context.Context, tenantID, vehicleID uuid.UUID, limit int) ([]VehicleTelematics, error) {
	return s.repo.GetTelematicsHistory(ctx, tenantID, vehicleID, limit)
}

func (s *fleetService) GetLiveFleetPositions(ctx context.Context, tenantID uuid.UUID) ([]LiveVehiclePosition, error) {
	return s.repo.GetLiveFleetPositions(ctx, tenantID)
}

func (s *fleetService) RecordMaintenance(ctx context.Context, tenantID, vehicleID uuid.UUID, dto CreateMaintenanceDTO) (*MaintenanceRecord, error) {
	servicedAt := dto.ServicedAt
	if servicedAt.IsZero() {
		servicedAt = time.Now().UTC()
	}

	m := &MaintenanceRecord{
		VehicleID:         vehicleID,
		TenantID:          tenantID,
		ServiceType:       strings.ToUpper(strings.TrimSpace(dto.ServiceType)),
		Description:       strings.TrimSpace(dto.Description),
		Cost:              dto.Cost,
		OdometerReadingKM: dto.OdometerReadingKM,
		ServicedAt:        servicedAt,
		NextServiceDueKM:  dto.NextServiceDueKM,
	}

	if err := m.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.CreateMaintenanceRecord(ctx, m); err != nil {
		return nil, err
	}

	s.logger.InfoContext(ctx, "maintenance record logged", slog.String("id", m.ID.String()), slog.String("vehicle_id", vehicleID.String()))
	return m, nil
}

func (s *fleetService) ListMaintenanceRecords(ctx context.Context, tenantID, vehicleID uuid.UUID) ([]MaintenanceRecord, error) {
	return s.repo.ListMaintenanceRecords(ctx, tenantID, vehicleID)
}

func (s *fleetService) ListUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]Vehicle, error) {
	return s.repo.ListUpcomingMaintenance(ctx, tenantID)
}
