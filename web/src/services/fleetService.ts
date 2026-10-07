import {
  Vehicle,
  VehicleCreatePayload,
  LiveVehiclePosition,
  VehicleTelematics,
  MaintenanceRecord,
  MaintenanceCreatePayload,
} from '../types/fleet';

const API_BASE = '/api/v1';

function getAuthHeaders(token: string) {
  return {
    'Content-Type': 'application/json',
    Accept: 'application/json',
    Authorization: `Bearer ${token}`,
  };
}

export const fleetService = {
  async getVehicles(token: string, tenantId?: string): Promise<Vehicle[]> {
    const url = tenantId ? `${API_BASE}/tenants/${tenantId}/vehicles` : `${API_BASE}/vehicles`;
    const res = await fetch(url, { headers: getAuthHeaders(token) });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to fetch vehicles (${res.status})`);
    }
    const data = await res.json();
    return data.data || [];
  },

  async createVehicle(token: string, payload: VehicleCreatePayload, tenantId?: string): Promise<Vehicle> {
    const url = tenantId ? `${API_BASE}/tenants/${tenantId}/vehicles` : `${API_BASE}/vehicles`;
    const res = await fetch(url, {
      method: 'POST',
      headers: getAuthHeaders(token),
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to register vehicle (${res.status})`);
    }
    const data = await res.json();
    return data.data;
  },

  async getVehicle(token: string, vehicleId: string, tenantId?: string): Promise<Vehicle> {
    const url = tenantId ? `${API_BASE}/tenants/${tenantId}/vehicles/${vehicleId}` : `${API_BASE}/vehicles/${vehicleId}`;
    const res = await fetch(url, { headers: getAuthHeaders(token) });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to fetch vehicle (${res.status})`);
    }
    const data = await res.json();
    return data.data;
  },

  async updateVehicle(
    token: string,
    vehicleId: string,
    payload: Partial<VehicleCreatePayload>,
    tenantId?: string
  ): Promise<Vehicle> {
    const url = tenantId ? `${API_BASE}/tenants/${tenantId}/vehicles/${vehicleId}` : `${API_BASE}/vehicles/${vehicleId}`;
    const res = await fetch(url, {
      method: 'PUT',
      headers: getAuthHeaders(token),
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to update vehicle (${res.status})`);
    }
    const data = await res.json();
    return data.data;
  },

  async deleteVehicle(token: string, vehicleId: string, tenantId?: string): Promise<void> {
    const url = tenantId ? `${API_BASE}/tenants/${tenantId}/vehicles/${vehicleId}` : `${API_BASE}/vehicles/${vehicleId}`;
    const res = await fetch(url, {
      method: 'DELETE',
      headers: getAuthHeaders(token),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to delete vehicle (${res.status})`);
    }
  },

  async getLiveFleet(token: string, tenantId?: string): Promise<LiveVehiclePosition[]> {
    const url = tenantId ? `${API_BASE}/tenants/${tenantId}/fleet/live` : `${API_BASE}/fleet/live`;
    const res = await fetch(url, { headers: getAuthHeaders(token) });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to fetch live fleet (${res.status})`);
    }
    const data = await res.json();
    return data.data || [];
  },

  async sendTelematicsPing(
    token: string,
    vehicleId: string,
    telematics: {
      latitude: number;
      longitude: number;
      speed_kmh: number;
      heading_degrees: number;
      battery_or_fuel_percent: number;
      odometer_km: number;
    },
    tenantId?: string
  ): Promise<VehicleTelematics> {
    const url = tenantId
      ? `${API_BASE}/tenants/${tenantId}/vehicles/${vehicleId}/telematics`
      : `${API_BASE}/vehicles/${vehicleId}/telematics`;
    const res = await fetch(url, {
      method: 'POST',
      headers: getAuthHeaders(token),
      body: JSON.stringify(telematics),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to send telematics (${res.status})`);
    }
    const data = await res.json();
    return data.data;
  },

  async getTelematicsHistory(
    token: string,
    vehicleId: string,
    limit: number = 100,
    tenantId?: string
  ): Promise<VehicleTelematics[]> {
    const url = tenantId
      ? `${API_BASE}/tenants/${tenantId}/vehicles/${vehicleId}/telematics/history?limit=${limit}`
      : `${API_BASE}/vehicles/${vehicleId}/telematics/history?limit=${limit}`;
    const res = await fetch(url, { headers: getAuthHeaders(token) });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to fetch telematics history (${res.status})`);
    }
    const data = await res.json();
    return data.data || [];
  },

  async createMaintenanceRecord(
    token: string,
    vehicleId: string,
    payload: MaintenanceCreatePayload,
    tenantId?: string
  ): Promise<MaintenanceRecord> {
    const url = tenantId
      ? `${API_BASE}/tenants/${tenantId}/vehicles/${vehicleId}/maintenance`
      : `${API_BASE}/vehicles/${vehicleId}/maintenance`;
    const res = await fetch(url, {
      method: 'POST',
      headers: getAuthHeaders(token),
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to log maintenance (${res.status})`);
    }
    const data = await res.json();
    return data.data;
  },

  async getMaintenanceRecords(token: string, vehicleId: string, tenantId?: string): Promise<MaintenanceRecord[]> {
    const url = tenantId
      ? `${API_BASE}/tenants/${tenantId}/vehicles/${vehicleId}/maintenance`
      : `${API_BASE}/vehicles/${vehicleId}/maintenance`;
    const res = await fetch(url, { headers: getAuthHeaders(token) });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to fetch maintenance records (${res.status})`);
    }
    const data = await res.json();
    return data.data || [];
  },

  async getUpcomingMaintenance(token: string, tenantId?: string): Promise<Vehicle[]> {
    const url = tenantId
      ? `${API_BASE}/tenants/${tenantId}/fleet/maintenance/upcoming`
      : `${API_BASE}/fleet/maintenance/upcoming`;
    const res = await fetch(url, { headers: getAuthHeaders(token) });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err?.error?.message || `Failed to fetch upcoming maintenance (${res.status})`);
    }
    const data = await res.json();
    return data.data || [];
  },
};
