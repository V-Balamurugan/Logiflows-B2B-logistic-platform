import React, { useState, useEffect, useCallback } from 'react';
import {
  Truck,
  Plus,
  Wrench,
  Search,
  Filter,
  Clock,
  RefreshCw,
  AlertTriangle,
  Zap,
} from 'lucide-react';
import { Vehicle, MaintenanceRecord } from '../types/fleet';
import { Branch } from '../types/tenancy';
import { fleetService } from '../services/fleetService';
import { FleetTrackingMap } from './FleetTrackingMap';
import { CreateVehicleModal } from './CreateVehicleModal';

interface FleetManagementViewProps {
  token: string;
  tenantId?: string;
  branches: Branch[];
}

export const FleetManagementView: React.FC<FleetManagementViewProps> = ({
  token,
  tenantId,
  branches,
}) => {
  const [vehicles, setVehicles] = useState<Vehicle[]>([]);
  const [upcomingMaintenance, setUpcomingMaintenance] = useState<Vehicle[]>([]);
  const [activeTab, setActiveTab] = useState<'MAP' | 'INVENTORY' | 'MAINTENANCE'>('MAP');
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState<boolean>(false);
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('ALL');

  // Maintenance Logging state
  const [selectedVehicleForMaint, setSelectedVehicleForMaint] = useState<Vehicle | null>(null);
  const [maintServiceType, setMaintServiceType] = useState<string>('OIL_CHANGE');
  const [maintDescription, setMaintDescription] = useState<string>('');
  const [maintCost, setMaintCost] = useState<number>(120);
  const [maintNextDueKm, setMaintNextDueKm] = useState<number>(15000);
  const [isLoggingMaint, setIsLoggingMaint] = useState<boolean>(false);
  const [maintHistory, setMaintHistory] = useState<MaintenanceRecord[]>([]);

  const loadFleetData = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [vehList, upcomingList] = await Promise.all([
        fleetService.getVehicles(token, tenantId),
        fleetService.getUpcomingMaintenance(token, tenantId).catch(() => []),
      ]);
      setVehicles(vehList);
      setUpcomingMaintenance(upcomingList);
    } catch (err: any) {
      setError(err?.message || 'Failed to load fleet data');
    } finally {
      setIsLoading(false);
    }
  }, [token, tenantId]);

  useEffect(() => {
    loadFleetData();
  }, [loadFleetData]);

  const loadVehicleMaintenance = async (v: Vehicle) => {
    setSelectedVehicleForMaint(v);
    try {
      const records = await fleetService.getMaintenanceRecords(token, v.id, tenantId);
      setMaintHistory(records);
    } catch (err) {
      console.error('Failed to load maintenance records:', err);
    }
  };

  const handleRecordMaintenance = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedVehicleForMaint) return;
    setIsLoggingMaint(true);
    try {
      await fleetService.createMaintenanceRecord(
        token,
        selectedVehicleForMaint.id,
        {
          service_type: maintServiceType,
          description: maintDescription,
          cost: Number(maintCost),
          odometer_reading_km: selectedVehicleForMaint.current_mileage_km,
          next_service_due_km: Number(maintNextDueKm),
        },
        tenantId
      );
      await loadVehicleMaintenance(selectedVehicleForMaint);
      await loadFleetData();
      setMaintDescription('');
    } catch (err: any) {
      alert(`Error logging maintenance: ${err?.message}`);
    } finally {
      setIsLoggingMaint(false);
    }
  };

  // KPI calculations
  const totalVehicles = vehicles.length;
  const onRouteCount = vehicles.filter((v) => v.status === 'ON_ROUTE').length;
  const evCount = vehicles.filter((v) => v.vehicle_type === 'ELECTRIC_VAN' || v.vehicle_type === 'BICYCLE').length;
  const evPercentage = totalVehicles > 0 ? ((evCount / totalVehicles) * 100).toFixed(0) : '0';
  const maintenanceCount = vehicles.filter((v) => v.status === 'MAINTENANCE').length + upcomingMaintenance.length;

  const filteredVehicles = vehicles.filter((v) => {
    const matchesSearch =
      v.license_plate.toLowerCase().includes(searchTerm.toLowerCase()) ||
      v.make.toLowerCase().includes(searchTerm.toLowerCase()) ||
      v.model.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesType = typeFilter === 'ALL' || v.vehicle_type === typeFilter;
    return matchesSearch && matchesType;
  });

  return (
    <div className="space-y-6">
      {error && (
        <div className="p-3 bg-red-950/60 border border-red-800 text-red-300 text-xs rounded-xl flex items-center gap-2">
          <AlertTriangle className="w-4 h-4 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Fleet KPI Metric Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="p-5 bg-slate-900 border border-slate-800 rounded-2xl shadow-lg relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">Total Fleet Assets</span>
            <div className="p-2 rounded-xl bg-blue-500/10 text-blue-400 border border-blue-500/20">
              <Truck className="w-5 h-5" />
            </div>
          </div>
          <p className="text-2xl font-bold text-white mt-3">{totalVehicles}</p>
          <p className="text-xs text-slate-400 mt-1">Authorized delivery vehicles</p>
        </div>

        <div className="p-5 bg-slate-900 border border-slate-800 rounded-2xl shadow-lg relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">Active On Route</span>
            <div className="p-2 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <Clock className="w-5 h-5" />
            </div>
          </div>
          <p className="text-2xl font-bold text-emerald-400 mt-3">{onRouteCount}</p>
          <p className="text-xs text-slate-400 mt-1">Live dispatch telematics active</p>
        </div>

        <div className="p-5 bg-slate-900 border border-slate-800 rounded-2xl shadow-lg relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">Green Fleet (EVs)</span>
            <div className="p-2 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              <Zap className="w-5 h-5" />
            </div>
          </div>
          <p className="text-2xl font-bold text-indigo-400 mt-3">{evPercentage}%</p>
          <p className="text-xs text-slate-400 mt-1">{evCount} zero-emission vehicles</p>
        </div>

        <div className="p-5 bg-slate-900 border border-slate-800 rounded-2xl shadow-lg relative overflow-hidden">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold uppercase tracking-wider text-slate-400">Service Attention</span>
            <div className="p-2 rounded-xl bg-amber-500/10 text-amber-400 border border-amber-500/20">
              <Wrench className="w-5 h-5" />
            </div>
          </div>
          <p className="text-2xl font-bold text-amber-400 mt-3">{maintenanceCount}</p>
          <p className="text-xs text-slate-400 mt-1">Due or in maintenance bay</p>
        </div>
      </div>

      {/* Main Container with Navigation & Header */}
      <div className="bg-slate-900 border border-slate-800 rounded-2xl shadow-xl overflow-hidden">
        <div className="p-5 border-b border-slate-800 flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-2 bg-slate-800/60 p-1 rounded-xl border border-slate-700/60">
            <button
              onClick={() => setActiveTab('MAP')}
              className={`px-4 py-2 rounded-lg text-xs font-bold transition-all ${
                activeTab === 'MAP'
                  ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              Live Telematics Map
            </button>
            <button
              onClick={() => setActiveTab('INVENTORY')}
              className={`px-4 py-2 rounded-lg text-xs font-bold transition-all ${
                activeTab === 'INVENTORY'
                  ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              Vehicle Inventory ({vehicles.length})
            </button>
            <button
              onClick={() => setActiveTab('MAINTENANCE')}
              className={`px-4 py-2 rounded-lg text-xs font-bold transition-all ${
                activeTab === 'MAINTENANCE'
                  ? 'bg-indigo-600 text-white shadow-md shadow-indigo-600/20'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              Maintenance Schedules {upcomingMaintenance.length > 0 && `(${upcomingMaintenance.length})`}
            </button>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={loadFleetData}
              disabled={isLoading}
              className="p-2 rounded-xl bg-slate-800 text-slate-300 hover:text-white border border-slate-700 transition-colors"
              title="Refresh Fleet Data"
            >
              <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
            <button
              onClick={() => setIsCreateModalOpen(true)}
              className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold rounded-xl shadow-lg shadow-indigo-600/20 transition-all flex items-center gap-2"
            >
              <Plus className="w-4 h-4" />
              Register Vehicle
            </button>
          </div>
        </div>

        {/* Tab 1: Live Telematics Map */}
        {activeTab === 'MAP' && (
          <div className="p-5 h-[620px]">
            <FleetTrackingMap
              token={token}
              tenantId={tenantId}
              onSelectVehicle={(v) => {
                const found = vehicles.find((x) => x.id === v.vehicle_id);
                if (found) loadVehicleMaintenance(found);
              }}
            />
          </div>
        )}

        {/* Tab 2: Vehicle Inventory Table */}
        {activeTab === 'INVENTORY' && (
          <div className="p-5 space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="relative flex-1 min-w-[240px] max-w-md">
                <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
                <input
                  type="text"
                  placeholder="Search by license plate, make, or model..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 bg-slate-800/80 border border-slate-700 rounded-xl text-white text-xs placeholder-slate-400 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="flex items-center gap-2">
                <Filter className="w-4 h-4 text-slate-400" />
                <select
                  value={typeFilter}
                  onChange={(e) => setTypeFilter(e.target.value)}
                  className="px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-xs outline-none cursor-pointer"
                >
                  <option value="ALL">All Vehicle Types</option>
                  <option value="ELECTRIC_VAN">Electric Van</option>
                  <option value="VAN">Delivery Van</option>
                  <option value="MOTORCYCLE">Motorcycle</option>
                  <option value="TRUCK">Truck</option>
                </select>
              </div>
            </div>

            <div className="overflow-x-auto rounded-xl border border-slate-800">
              <table className="w-full text-left text-xs text-slate-300">
                <thead className="bg-slate-800/80 text-slate-400 font-semibold uppercase tracking-wider">
                  <tr>
                    <th className="py-3 px-4">Vehicle</th>
                    <th className="py-3 px-4">Class & Fuel</th>
                    <th className="py-3 px-4">Payload Capacity</th>
                    <th className="py-3 px-4">Odometer</th>
                    <th className="py-3 px-4">Battery / Fuel</th>
                    <th className="py-3 px-4">Status</th>
                    <th className="py-3 px-4 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/50">
                  {filteredVehicles.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="text-center py-8 text-slate-500">
                        No fleet vehicles match search filter.
                      </td>
                    </tr>
                  ) : (
                    filteredVehicles.map((v) => (
                      <tr key={v.id} className="hover:bg-slate-800/30 transition-colors">
                        <td className="py-3.5 px-4">
                          <div className="font-mono font-bold text-white text-sm">
                            {v.license_plate}
                          </div>
                          <div className="text-[11px] text-slate-400">
                            {v.year} {v.make} {v.model}
                          </div>
                        </td>
                        <td className="py-3.5 px-4">
                          <div className="font-medium text-slate-200">{v.vehicle_type}</div>
                          <div className="text-[10px] text-indigo-400 uppercase">{v.fuel_type}</div>
                        </td>
                        <td className="py-3.5 px-4">
                          <div>{v.capacity_kg} kg • {v.capacity_volume_m3} m³</div>
                          <div className="text-[10px] text-slate-400">Max {v.max_parcels} parcels</div>
                        </td>
                        <td className="py-3.5 px-4 font-mono font-medium text-slate-200">
                          {v.current_mileage_km.toFixed(1)} km
                        </td>
                        <td className="py-3.5 px-4">
                          <div className="flex items-center gap-2">
                            <div className="w-16 h-2 rounded-full bg-slate-800 overflow-hidden">
                              <div
                                style={{ width: `${v.battery_or_fuel_level_percent}%` }}
                                className={`h-full rounded-full ${
                                  v.battery_or_fuel_level_percent > 40
                                    ? 'bg-emerald-500'
                                    : v.battery_or_fuel_level_percent > 20
                                    ? 'bg-amber-500'
                                    : 'bg-red-500'
                                }`}
                              />
                            </div>
                            <span className="font-mono font-semibold text-xs">
                              {v.battery_or_fuel_level_percent.toFixed(0)}%
                            </span>
                          </div>
                        </td>
                        <td className="py-3.5 px-4">
                          <span
                            className={`px-2 py-0.5 text-[10px] font-bold rounded-md uppercase ${
                              v.status === 'ON_ROUTE'
                                ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                                : v.status === 'MAINTENANCE'
                                ? 'bg-amber-500/20 text-amber-400 border border-amber-500/30'
                                : 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
                            }`}
                          >
                            {v.status}
                          </span>
                        </td>
                        <td className="py-3.5 px-4 text-right">
                          <button
                            onClick={() => {
                              loadVehicleMaintenance(v);
                              setActiveTab('MAINTENANCE');
                            }}
                            className="px-2.5 py-1 text-xs font-semibold rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition-colors"
                          >
                            Service Logs
                          </button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Tab 3: Maintenance & Upcoming Servicing */}
        {activeTab === 'MAINTENANCE' && (
          <div className="p-5 space-y-6">
            {upcomingMaintenance.length > 0 && (
              <div className="p-4 bg-amber-500/10 border border-amber-500/30 rounded-xl">
                <div className="flex items-center gap-2 text-amber-400 font-bold text-sm mb-2">
                  <AlertTriangle className="w-4 h-4" />
                  <span>Upcoming Maintenance Due Alerts ({upcomingMaintenance.length})</span>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                  {upcomingMaintenance.map((uv) => (
                    <div
                      key={uv.id}
                      className="p-3 bg-slate-900 border border-amber-500/20 rounded-lg flex items-center justify-between"
                    >
                      <div>
                        <span className="font-mono font-bold text-white text-xs">{uv.license_plate}</span>
                        <span className="text-slate-400 text-xs ml-2">
                          {uv.make} {uv.model}
                        </span>
                        <p className="text-[11px] text-amber-300/80 mt-0.5">
                          Odometer: {uv.current_mileage_km.toFixed(0)} km (Overdue/Approaching Service)
                        </p>
                      </div>
                      <button
                        onClick={() => loadVehicleMaintenance(uv)}
                        className="px-2.5 py-1 text-xs font-bold bg-amber-500 hover:bg-amber-400 text-slate-950 rounded-lg"
                      >
                        Inspect
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Service Logging Form & History */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Log Service Form */}
              <div className="p-5 bg-slate-800/40 border border-slate-800 rounded-xl">
                <h3 className="text-sm font-bold text-white mb-1 flex items-center gap-2">
                  <Wrench className="w-4 h-4 text-indigo-400" />
                  Log Maintenance Intervention
                </h3>
                <p className="text-xs text-slate-400 mb-4">
                  {selectedVehicleForMaint
                    ? `Logging service for ${selectedVehicleForMaint.license_plate}`
                    : 'Select a vehicle from the list below'}
                </p>

                <form onSubmit={handleRecordMaintenance} className="space-y-3">
                  <div>
                    <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                      Vehicle *
                    </label>
                    <select
                      value={selectedVehicleForMaint?.id || ''}
                      onChange={(e) => {
                        const found = vehicles.find((v) => v.id === e.target.value);
                        if (found) loadVehicleMaintenance(found);
                      }}
                      className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-xs focus:outline-none focus:border-indigo-500"
                      required
                    >
                      <option value="">Select vehicle...</option>
                      {vehicles.map((v) => (
                        <option key={v.id} value={v.id}>
                          {v.license_plate} ({v.make} {v.model})
                        </option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                      Service Type *
                    </label>
                    <select
                      value={maintServiceType}
                      onChange={(e) => setMaintServiceType(e.target.value)}
                      className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-xs focus:outline-none focus:border-indigo-500"
                    >
                      <option value="OIL_CHANGE">Oil & Filter Change</option>
                      <option value="TIRE_ROTATION">Tire Rotation & Alignment</option>
                      <option value="BATTERY_CHECK">EV Battery Diagnostic</option>
                      <option value="BRAKE_INSPECTION">Brake Pads & Fluid</option>
                      <option value="ANNUAL_OVERHAUL">Full Annual Inspection</option>
                    </select>
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        Cost ($)
                      </label>
                      <input
                        type="number"
                        step="0.01"
                        value={maintCost}
                        onChange={(e) => setMaintCost(Number(e.target.value))}
                        className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-xs focus:outline-none focus:border-indigo-500"
                      />
                    </div>
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        Next Due (km)
                      </label>
                      <input
                        type="number"
                        value={maintNextDueKm}
                        onChange={(e) => setMaintNextDueKm(Number(e.target.value))}
                        className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-xs focus:outline-none focus:border-indigo-500"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                      Service Notes
                    </label>
                    <textarea
                      rows={2}
                      placeholder="Technician details, parts swapped..."
                      value={maintDescription}
                      onChange={(e) => setMaintDescription(e.target.value)}
                      className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white text-xs focus:outline-none focus:border-indigo-500"
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={isLoggingMaint || !selectedVehicleForMaint}
                    className="w-full py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-semibold text-xs rounded-lg transition-colors flex items-center justify-center gap-2"
                  >
                    {isLoggingMaint ? (
                      <span className="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin" />
                    ) : (
                      <Wrench className="w-3.5 h-3.5" />
                    )}
                    Commit Service Record
                  </button>
                </form>
              </div>

              {/* Maintenance History List */}
              <div className="lg:col-span-2 p-5 bg-slate-800/40 border border-slate-800 rounded-xl space-y-3">
                <h3 className="text-sm font-bold text-white flex items-center justify-between">
                  <span>
                    Past Maintenance Logs {selectedVehicleForMaint && `for ${selectedVehicleForMaint.license_plate}`}
                  </span>
                  <span className="text-xs text-slate-400 font-normal">
                    {maintHistory.length} records recorded
                  </span>
                </h3>

                {maintHistory.length === 0 ? (
                  <div className="text-center py-12 text-slate-500 text-xs">
                    No maintenance records found for the selected vehicle.
                  </div>
                ) : (
                  <div className="space-y-2 max-h-[360px] overflow-y-auto pr-1">
                    {maintHistory.map((rec) => (
                      <div
                        key={rec.id}
                        className="p-3 bg-slate-900 border border-slate-800 rounded-lg flex items-center justify-between text-xs"
                      >
                        <div>
                          <div className="font-bold text-white flex items-center gap-2">
                            <span>{rec.service_type}</span>
                            <span className="text-[10px] font-mono text-emerald-400 font-normal">
                              ${rec.cost.toFixed(2)}
                            </span>
                          </div>
                          <p className="text-slate-400 text-[11px] mt-0.5">{rec.description || 'Standard service'}</p>
                          <p className="text-[10px] text-slate-500 mt-1 font-mono">
                            Odometer: {rec.odometer_reading_km.toFixed(0)} km • Serviced:{' '}
                            {new Date(rec.serviced_at).toLocaleDateString()}
                          </p>
                        </div>
                        {rec.next_service_due_km && (
                          <div className="text-right">
                            <span className="text-[10px] text-slate-400">Next Service</span>
                            <p className="font-mono font-bold text-slate-200">{rec.next_service_due_km.toFixed(0)} km</p>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Modal for Creating Vehicle */}
      <CreateVehicleModal
        token={token}
        tenantId={tenantId}
        branches={branches}
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onVehicleCreated={(newVeh) => {
          setVehicles((prev) => [newVeh, ...prev]);
        }}
      />
    </div>
  );
};
