export type VehicleType = 'VAN' | 'MOTORCYCLE' | 'TRUCK' | 'ELECTRIC_VAN' | 'BICYCLE';
export type FuelType = 'DIESEL' | 'PETROL' | 'ELECTRIC' | 'HYBRID';
export type VehicleStatus = 'AVAILABLE' | 'ON_ROUTE' | 'MAINTENANCE' | 'DECOMMISSIONED';

export interface Vehicle {
  id: string;
  tenant_id: string;
  assigned_branch_id?: string;
  assigned_driver_id?: string;
  license_plate: string;
  vin: string;
  make: string;
  model: string;
  year: number;
  vehicle_type: VehicleType;
  fuel_type: FuelType;
  capacity_kg: number;
  capacity_volume_m3: number;
  max_parcels: number;
  current_mileage_km: number;
  battery_or_fuel_level_percent: number;
  status: VehicleStatus;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface VehicleCreatePayload {
  assigned_branch_id?: string;
  assigned_driver_id?: string;
  license_plate: string;
  vin?: string;
  make: string;
  model: string;
  year: number;
  vehicle_type: VehicleType;
  fuel_type: FuelType;
  capacity_kg: number;
  capacity_volume_m3: number;
  max_parcels: number;
  current_mileage_km?: number;
  battery_or_fuel_level_percent?: number;
  status?: VehicleStatus;
}

export interface VehicleTelematics {
  id: string;
  vehicle_id: string;
  tenant_id: string;
  latitude: number;
  longitude: number;
  speed_kmh: number;
  heading_degrees: number;
  battery_or_fuel_percent: number;
  odometer_km: number;
  recorded_at: string;
}

export interface LiveVehiclePosition {
  vehicle_id: string;
  tenant_id: string;
  license_plate: string;
  make: string;
  model: string;
  vehicle_type: VehicleType;
  status: VehicleStatus;
  assigned_branch_id?: string;
  assigned_driver_id?: string;
  latitude: number;
  longitude: number;
  speed_kmh: number;
  heading_degrees: number;
  battery_or_fuel_percent: number;
  odometer_km: number;
  last_ping_at: string;
}

export interface MaintenanceRecord {
  id: string;
  vehicle_id: string;
  tenant_id: string;
  service_type: string;
  description: string;
  cost: number;
  odometer_reading_km: number;
  serviced_at: string;
  next_service_due_km?: number;
  created_at: string;
}

export interface MaintenanceCreatePayload {
  service_type: string;
  description?: string;
  cost: number;
  odometer_reading_km: number;
  serviced_at?: string;
  next_service_due_km?: number;
}
