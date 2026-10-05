package fleet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("fleet entity not found")
	ErrAlreadyExists = errors.New("fleet entity already exists")
)

type Repository interface {
	CreateVehicle(ctx context.Context, v *Vehicle) error
	GetVehicleByID(ctx context.Context, tenantID, vehicleID uuid.UUID) (*Vehicle, error)
	ListVehicles(ctx context.Context, tenantID uuid.UUID, filter VehicleFilter) ([]Vehicle, error)
	UpdateVehicle(ctx context.Context, v *Vehicle) error
	DeleteVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) error

	IngestTelematics(ctx context.Context, t *VehicleTelematics) error
	GetLatestTelematics(ctx context.Context, tenantID, vehicleID uuid.UUID) (*VehicleTelematics, error)
	GetTelematicsHistory(ctx context.Context, tenantID, vehicleID uuid.UUID, limit int) ([]VehicleTelematics, error)
	GetLiveFleetPositions(ctx context.Context, tenantID uuid.UUID) ([]LiveVehiclePosition, error)

	CreateMaintenanceRecord(ctx context.Context, m *MaintenanceRecord) error
	ListMaintenanceRecords(ctx context.Context, tenantID, vehicleID uuid.UUID) ([]MaintenanceRecord, error)
	ListUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]Vehicle, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateVehicle(ctx context.Context, v *Vehicle) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	v.CreatedAt = time.Now().UTC()
	v.UpdatedAt = v.CreatedAt

	query := `
		INSERT INTO vehicles (
			id, tenant_id, assigned_branch_id, assigned_driver_id,
			license_plate, vin, make, model, year, vehicle_type, fuel_type,
			capacity_kg, capacity_volume_m3, max_parcels, current_mileage_km,
			battery_or_fuel_level_percent, status, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		) RETURNING id, created_at, updated_at;
	`

	err := r.db.QueryRowContext(
		ctx, query,
		v.ID, v.TenantID, v.AssignedBranchID, v.AssignedDriverID,
		v.LicensePlate, v.VIN, v.Make, v.Model, v.Year, string(v.VehicleType), string(v.FuelType),
		v.CapacityKG, v.CapacityVolumeM3, v.MaxParcels, v.CurrentMileageKM,
		v.BatteryOrFuelLevelPercent, string(v.Status), v.Metadata, v.CreatedAt, v.UpdatedAt,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "uq_vehicle_tenant_license") {
			return ErrAlreadyExists
		}
		return fmt.Errorf("failed to insert vehicle: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetVehicleByID(ctx context.Context, tenantID, vehicleID uuid.UUID) (*Vehicle, error) {
	query := `
		SELECT 
			id, tenant_id, assigned_branch_id, assigned_driver_id,
			license_plate, vin, make, model, year, vehicle_type, fuel_type,
			capacity_kg, capacity_volume_m3, max_parcels, current_mileage_km,
			battery_or_fuel_level_percent, status, metadata, created_at, updated_at
		FROM vehicles
		WHERE id = $1 AND tenant_id = $2;
	`

	var v Vehicle
	var vehicleType, fuelType, status string

	err := r.db.QueryRowContext(ctx, query, vehicleID, tenantID).Scan(
		&v.ID, &v.TenantID, &v.AssignedBranchID, &v.AssignedDriverID,
		&v.LicensePlate, &v.VIN, &v.Make, &v.Model, &v.Year, &vehicleType, &fuelType,
		&v.CapacityKG, &v.CapacityVolumeM3, &v.MaxParcels, &v.CurrentMileageKM,
		&v.BatteryOrFuelLevelPercent, &status, &v.Metadata, &v.CreatedAt, &v.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query vehicle: %w", err)
	}

	v.VehicleType = VehicleType(vehicleType)
	v.FuelType = FuelType(fuelType)
	v.Status = VehicleStatus(status)

	return &v, nil
}

func (r *PostgresRepository) ListVehicles(ctx context.Context, tenantID uuid.UUID, filter VehicleFilter) ([]Vehicle, error) {
	query := `
		SELECT 
			id, tenant_id, assigned_branch_id, assigned_driver_id,
			license_plate, vin, make, model, year, vehicle_type, fuel_type,
			capacity_kg, capacity_volume_m3, max_parcels, current_mileage_km,
			battery_or_fuel_level_percent, status, metadata, created_at, updated_at
		FROM vehicles
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argIdx := 2

	if filter.BranchID != nil {
		query += fmt.Sprintf(" AND assigned_branch_id = $%d", argIdx)
		args = append(args, *filter.BranchID)
		argIdx++
	}
	if filter.DriverID != nil {
		query += fmt.Sprintf(" AND assigned_driver_id = $%d", argIdx)
		args = append(args, *filter.DriverID)
		argIdx++
	}
	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, string(*filter.Status))
		argIdx++
	}
	if filter.VehicleType != nil {
		query += fmt.Sprintf(" AND vehicle_type = $%d", argIdx)
		args = append(args, string(*filter.VehicleType))
		argIdx++
	}

	query += " ORDER BY created_at DESC;"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list vehicles: %w", err)
	}
	defer rows.Close()

	var vehicles []Vehicle
	for rows.Next() {
		var v Vehicle
		var vehicleType, fuelType, status string
		err := rows.Scan(
			&v.ID, &v.TenantID, &v.AssignedBranchID, &v.AssignedDriverID,
			&v.LicensePlate, &v.VIN, &v.Make, &v.Model, &v.Year, &vehicleType, &fuelType,
			&v.CapacityKG, &v.CapacityVolumeM3, &v.MaxParcels, &v.CurrentMileageKM,
			&v.BatteryOrFuelLevelPercent, &status, &v.Metadata, &v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vehicle: %w", err)
		}
		v.VehicleType = VehicleType(vehicleType)
		v.FuelType = FuelType(fuelType)
		v.Status = VehicleStatus(status)
		vehicles = append(vehicles, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return vehicles, nil
}

func (r *PostgresRepository) UpdateVehicle(ctx context.Context, v *Vehicle) error {
	v.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE vehicles SET
			assigned_branch_id = $1,
			assigned_driver_id = $2,
			make = $3,
			model = $4,
			year = $5,
			vehicle_type = $6,
			fuel_type = $7,
			capacity_kg = $8,
			capacity_volume_m3 = $9,
			max_parcels = $10,
			current_mileage_km = $11,
			battery_or_fuel_level_percent = $12,
			status = $13,
			metadata = $14,
			updated_at = $15
		WHERE id = $16 AND tenant_id = $17;
	`

	res, err := r.db.ExecContext(
		ctx, query,
		v.AssignedBranchID, v.AssignedDriverID,
		v.Make, v.Model, v.Year, string(v.VehicleType), string(v.FuelType),
		v.CapacityKG, v.CapacityVolumeM3, v.MaxParcels, v.CurrentMileageKM,
		v.BatteryOrFuelLevelPercent, string(v.Status), v.Metadata, v.UpdatedAt,
		v.ID, v.TenantID,
	)

	if err != nil {
		return fmt.Errorf("failed to update vehicle: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *PostgresRepository) DeleteVehicle(ctx context.Context, tenantID, vehicleID uuid.UUID) error {
	query := `DELETE FROM vehicles WHERE id = $1 AND tenant_id = $2;`
	res, err := r.db.ExecContext(ctx, query, vehicleID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) IngestTelematics(ctx context.Context, t *VehicleTelematics) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.RecordedAt.IsZero() {
		t.RecordedAt = time.Now().UTC()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify vehicle exists and belongs to tenant
	var currentMileage float64
	err = tx.QueryRowContext(ctx, "SELECT current_mileage_km FROM vehicles WHERE id = $1 AND tenant_id = $2 FOR UPDATE;", t.VehicleID, t.TenantID).Scan(&currentMileage)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to verify vehicle: %w", err)
	}

	// Insert telematics record with spatial PostGIS Point (SRID 4326)
	insertQuery := `
		INSERT INTO vehicle_telematics (
			id, vehicle_id, tenant_id, location, latitude, longitude,
			speed_kmh, heading_degrees, battery_or_fuel_percent, odometer_km, recorded_at
		) VALUES (
			$1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $5, $4,
			$6, $7, $8, $9, $10
		);
	`
	_, err = tx.ExecContext(
		ctx, insertQuery,
		t.ID, t.VehicleID, t.TenantID, t.Longitude, t.Latitude,
		t.SpeedKMH, t.HeadingDegrees, t.BatteryOrFuelPercent, t.OdometerKM, t.RecordedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert telematics: %w", err)
	}

	// Update vehicle current mileage and battery/fuel status
	newMileage := currentMileage
	if t.OdometerKM > currentMileage {
		newMileage = t.OdometerKM
	}

	updateVehicleQuery := `
		UPDATE vehicles SET
			current_mileage_km = $1,
			battery_or_fuel_level_percent = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND tenant_id = $4;
	`
	_, err = tx.ExecContext(ctx, updateVehicleQuery, newMileage, t.BatteryOrFuelPercent, t.VehicleID, t.TenantID)
	if err != nil {
		return fmt.Errorf("failed to update vehicle stats: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit telematics transaction: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetLatestTelematics(ctx context.Context, tenantID, vehicleID uuid.UUID) (*VehicleTelematics, error) {
	query := `
		SELECT 
			id, vehicle_id, tenant_id, latitude, longitude,
			speed_kmh, heading_degrees, battery_or_fuel_percent, odometer_km, recorded_at
		FROM vehicle_telematics
		WHERE vehicle_id = $1 AND tenant_id = $2
		ORDER BY recorded_at DESC
		LIMIT 1;
	`

	var t VehicleTelematics
	err := r.db.QueryRowContext(ctx, query, vehicleID, tenantID).Scan(
		&t.ID, &t.VehicleID, &t.TenantID, &t.Latitude, &t.Longitude,
		&t.SpeedKMH, &t.HeadingDegrees, &t.BatteryOrFuelPercent, &t.OdometerKM, &t.RecordedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to query latest telematics: %w", err)
	}

	return &t, nil
}

func (r *PostgresRepository) GetTelematicsHistory(ctx context.Context, tenantID, vehicleID uuid.UUID, limit int) ([]VehicleTelematics, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query := `
		SELECT 
			id, vehicle_id, tenant_id, latitude, longitude,
			speed_kmh, heading_degrees, battery_or_fuel_percent, odometer_km, recorded_at
		FROM vehicle_telematics
		WHERE vehicle_id = $1 AND tenant_id = $2
		ORDER BY recorded_at ASC
		LIMIT $3;
	`

	rows, err := r.db.QueryContext(ctx, query, vehicleID, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query telematics history: %w", err)
	}
	defer rows.Close()

	var history []VehicleTelematics
	for rows.Next() {
		var t VehicleTelematics
		err := rows.Scan(
			&t.ID, &t.VehicleID, &t.TenantID, &t.Latitude, &t.Longitude,
			&t.SpeedKMH, &t.HeadingDegrees, &t.BatteryOrFuelPercent, &t.OdometerKM, &t.RecordedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telematics history: %w", err)
		}
		history = append(history, t)
	}

	return history, nil
}

func (r *PostgresRepository) GetLiveFleetPositions(ctx context.Context, tenantID uuid.UUID) ([]LiveVehiclePosition, error) {
	query := `
		SELECT DISTINCT ON (v.id)
			v.id, v.tenant_id, v.license_plate, v.make, v.model, v.vehicle_type, v.status,
			v.assigned_branch_id, v.assigned_driver_id,
			COALESCE(t.latitude, 0.0),
			COALESCE(t.longitude, 0.0),
			COALESCE(t.speed_kmh, 0.0),
			COALESCE(t.heading_degrees, 0.0),
			COALESCE(t.battery_or_fuel_percent, v.battery_or_fuel_level_percent),
			COALESCE(t.odometer_km, v.current_mileage_km),
			COALESCE(t.recorded_at, v.updated_at)
		FROM vehicles v
		LEFT JOIN vehicle_telematics t ON v.id = t.vehicle_id AND v.tenant_id = t.tenant_id
		WHERE v.tenant_id = $1 AND v.status != 'DECOMMISSIONED'
		ORDER BY v.id, t.recorded_at DESC NULLS LAST;
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query live fleet positions: %w", err)
	}
	defer rows.Close()

	var positions []LiveVehiclePosition
	for rows.Next() {
		var p LiveVehiclePosition
		var vType, vStatus string
		err := rows.Scan(
			&p.VehicleID, &p.TenantID, &p.LicensePlate, &p.Make, &p.Model, &vType, &vStatus,
			&p.AssignedBranchID, &p.AssignedDriverID,
			&p.Latitude, &p.Longitude, &p.SpeedKMH, &p.HeadingDegrees,
			&p.BatteryOrFuelPercent, &p.OdometerKM, &p.LastPingAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan live fleet position: %w", err)
		}
		p.VehicleType = VehicleType(vType)
		p.Status = VehicleStatus(vStatus)
		positions = append(positions, p)
	}

	return positions, nil
}

func (r *PostgresRepository) CreateMaintenanceRecord(ctx context.Context, m *MaintenanceRecord) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	if m.ServicedAt.IsZero() {
		m.ServicedAt = m.CreatedAt
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify vehicle belongs to tenant
	var currentMileage float64
	err = tx.QueryRowContext(ctx, "SELECT current_mileage_km FROM vehicles WHERE id = $1 AND tenant_id = $2 FOR UPDATE;", m.VehicleID, m.TenantID).Scan(&currentMileage)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to verify vehicle: %w", err)
	}

	query := `
		INSERT INTO maintenance_records (
			id, vehicle_id, tenant_id, service_type, description, cost,
			odometer_reading_km, serviced_at, next_service_due_km, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, created_at;
	`

	err = tx.QueryRowContext(
		ctx, query,
		m.ID, m.VehicleID, m.TenantID, m.ServiceType, m.Description, m.Cost,
		m.OdometerReadingKM, m.ServicedAt, m.NextServiceDueKM, m.CreatedAt,
	).Scan(&m.ID, &m.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert maintenance record: %w", err)
	}

	// Update vehicle status back to AVAILABLE if it was in MAINTENANCE
	updateStatus := `UPDATE vehicles SET status = 'AVAILABLE', updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND tenant_id = $2 AND status = 'MAINTENANCE';`
	_, _ = tx.ExecContext(ctx, updateStatus, m.VehicleID, m.TenantID)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit maintenance record: %w", err)
	}

	return nil
}

func (r *PostgresRepository) ListMaintenanceRecords(ctx context.Context, tenantID, vehicleID uuid.UUID) ([]MaintenanceRecord, error) {
	query := `
		SELECT 
			id, vehicle_id, tenant_id, service_type, description, cost,
			odometer_reading_km, serviced_at, next_service_due_km, created_at
		FROM maintenance_records
		WHERE vehicle_id = $1 AND tenant_id = $2
		ORDER BY serviced_at DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, vehicleID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query maintenance records: %w", err)
	}
	defer rows.Close()

	var records []MaintenanceRecord
	for rows.Next() {
		var m MaintenanceRecord
		err := rows.Scan(
			&m.ID, &m.VehicleID, &m.TenantID, &m.ServiceType, &m.Description, &m.Cost,
			&m.OdometerReadingKM, &m.ServicedAt, &m.NextServiceDueKM, &m.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance record: %w", err)
		}
		records = append(records, m)
	}

	return records, nil
}

func (r *PostgresRepository) ListUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]Vehicle, error) {
	// Query vehicles whose current mileage is within 500km of next_service_due_km or already overdue
	query := `
		SELECT DISTINCT ON (v.id)
			v.id, v.tenant_id, v.assigned_branch_id, v.assigned_driver_id,
			v.license_plate, v.vin, v.make, v.model, v.year, v.vehicle_type, v.fuel_type,
			v.capacity_kg, v.capacity_volume_m3, v.max_parcels, v.current_mileage_km,
			v.battery_or_fuel_level_percent, v.status, v.metadata, v.created_at, v.updated_at
		FROM vehicles v
		INNER JOIN maintenance_records m ON v.id = m.vehicle_id AND v.tenant_id = m.tenant_id
		WHERE v.tenant_id = $1 
		  AND m.next_service_due_km IS NOT NULL 
		  AND (v.current_mileage_km >= (m.next_service_due_km - 500))
		ORDER BY v.id, m.serviced_at DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming maintenance: %w", err)
	}
	defer rows.Close()

	var vehicles []Vehicle
	for rows.Next() {
		var v Vehicle
		var vehicleType, fuelType, status string
		err := rows.Scan(
			&v.ID, &v.TenantID, &v.AssignedBranchID, &v.AssignedDriverID,
			&v.LicensePlate, &v.VIN, &v.Make, &v.Model, &v.Year, &vehicleType, &fuelType,
			&v.CapacityKG, &v.CapacityVolumeM3, &v.MaxParcels, &v.CurrentMileageKM,
			&v.BatteryOrFuelLevelPercent, &status, &v.Metadata, &v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vehicle for upcoming maintenance: %w", err)
		}
		v.VehicleType = VehicleType(vehicleType)
		v.FuelType = FuelType(fuelType)
		v.Status = VehicleStatus(status)
		vehicles = append(vehicles, v)
	}

	return vehicles, nil
}
