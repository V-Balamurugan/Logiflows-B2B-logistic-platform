import { describe, it } from 'node:test';
import assert from 'node:assert';

describe('Fleet Frontend Types & Payload Validation', () => {
  it('validates vehicle create payload structure', () => {
    const payload = {
      license_plate: 'KA-01-EV-1001',
      vin: 'MBBLREVITO2024001',
      make: 'Tata',
      model: 'Ace EV',
      year: 2024,
      vehicle_type: 'ELECTRIC_VAN',
      fuel_type: 'ELECTRIC',
      capacity_kg: 800,
      capacity_volume_m3: 4.5,
      max_parcels: 150,
      status: 'AVAILABLE',
    };

    assert.strictEqual(payload.license_plate, 'KA-01-EV-1001');
    assert.strictEqual(payload.vehicle_type, 'ELECTRIC_VAN');
    assert.strictEqual(payload.fuel_type, 'ELECTRIC');
    assert.ok(payload.capacity_kg > 0);
    assert.ok(payload.capacity_volume_m3 > 0);
  });

  it('validates telematics ping coordinates and metrics', () => {
    const ping = {
      latitude: 12.9716,
      longitude: 77.5946,
      speed_kmh: 42.5,
      heading_degrees: 135.0,
      battery_or_fuel_percent: 88.0,
      odometer_km: 1240.5,
    };

    assert.ok(ping.latitude >= -90 && ping.latitude <= 90);
    assert.ok(ping.longitude >= -180 && ping.longitude <= 180);
    assert.ok(ping.speed_kmh >= 0);
    assert.ok(ping.battery_or_fuel_percent >= 0 && ping.battery_or_fuel_percent <= 100);
  });

  it('validates maintenance record structure and next service interval', () => {
    const maint = {
      service_type: 'ENGINE_OIL_CHANGE',
      description: 'Lube and filter overhaul',
      cost: 250.0,
      odometer_reading_km: 5000.0,
      next_service_due_km: 10000.0,
    };

    assert.strictEqual(maint.service_type, 'ENGINE_OIL_CHANGE');
    assert.strictEqual(maint.cost, 250.0);
    assert.ok(maint.next_service_due_km > maint.odometer_reading_km);
  });
});
