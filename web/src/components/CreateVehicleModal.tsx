import React, { useState } from 'react';
import { X, Truck, Zap, AlertCircle } from 'lucide-react';
import { VehicleType, FuelType, VehicleCreatePayload, Vehicle } from '../types/fleet';
import { Branch } from '../types/tenancy';
import { fleetService } from '../services/fleetService';

interface CreateVehicleModalProps {
  token: string;
  tenantId?: string;
  branches: Branch[];
  isOpen: boolean;
  onClose: () => void;
  onVehicleCreated: (vehicle: Vehicle) => void;
}

export const CreateVehicleModal: React.FC<CreateVehicleModalProps> = ({
  token,
  tenantId,
  branches,
  isOpen,
  onClose,
  onVehicleCreated,
}) => {
  const [licensePlate, setLicensePlate] = useState('');
  const [vin, setVin] = useState('');
  const [make, setMake] = useState('');
  const [model, setModel] = useState('');
  const [year, setYear] = useState<number>(2024);
  const [vehicleType, setVehicleType] = useState<VehicleType>('ELECTRIC_VAN');
  const [fuelType, setFuelType] = useState<FuelType>('ELECTRIC');
  const [capacityKg, setCapacityKg] = useState<number>(800);
  const [capacityVolumeM3, setCapacityVolumeM3] = useState<number>(4.5);
  const [maxParcels, setMaxParcels] = useState<number>(150);
  const [assignedBranchId, setAssignedBranchId] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!licensePlate.trim()) {
      setError('License plate is required');
      return;
    }
    if (!make.trim() || !model.trim()) {
      setError('Vehicle make and model are required');
      return;
    }

    setIsSubmitting(true);
    try {
      const payload: VehicleCreatePayload = {
        license_plate: licensePlate.trim().toUpperCase(),
        vin: vin.trim(),
        make: make.trim(),
        model: model.trim(),
        year: Number(year),
        vehicle_type: vehicleType,
        fuel_type: fuelType,
        capacity_kg: Number(capacityKg),
        capacity_volume_m3: Number(capacityVolumeM3),
        max_parcels: Number(maxParcels),
        assigned_branch_id: assignedBranchId ? assignedBranchId : undefined,
        current_mileage_km: 0,
        battery_or_fuel_level_percent: 100,
        status: 'AVAILABLE',
      };

      const created = await fleetService.createVehicle(token, payload, tenantId);
      onVehicleCreated(created);
      onClose();
    } catch (err: any) {
      setError(err?.message || 'Failed to register vehicle');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-[2000] flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-fade-in">
      <div className="relative w-full max-w-lg bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl overflow-hidden">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800 bg-slate-900/50">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              <Truck className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">Register Fleet Vehicle</h2>
              <p className="text-xs text-slate-400">Add a delivery asset to tenant fleet</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4 max-h-[80vh] overflow-y-auto">
          {error && (
            <div className="flex items-center gap-2 p-3 bg-red-950/50 border border-red-800 text-red-300 text-xs rounded-xl">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">
                License Plate *
              </label>
              <input
                type="text"
                placeholder="e.g. KA-01-EV-1001"
                value={licensePlate}
                onChange={(e) => setLicensePlate(e.target.value.toUpperCase())}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">
                VIN (Vehicle ID Number)
              </label>
              <input
                type="text"
                placeholder="e.g. MBBLREVITO2024001"
                value={vin}
                onChange={(e) => setVin(e.target.value)}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Make *</label>
              <input
                type="text"
                placeholder="Tata, Ather..."
                value={make}
                onChange={(e) => setMake(e.target.value)}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Model *</label>
              <input
                type="text"
                placeholder="Ace EV, 450X..."
                value={model}
                onChange={(e) => setModel(e.target.value)}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Year</label>
              <input
                type="number"
                min="1990"
                max={new Date().getFullYear() + 2}
                value={year}
                onChange={(e) => setYear(Number(e.target.value))}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Vehicle Type</label>
              <select
                value={vehicleType}
                onChange={(e) => {
                  const vt = e.target.value as VehicleType;
                  setVehicleType(vt);
                  if (vt === 'ELECTRIC_VAN' || vt === 'BICYCLE') {
                    setFuelType('ELECTRIC');
                  }
                }}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
              >
                <option value="ELECTRIC_VAN">Electric Van (EV)</option>
                <option value="VAN">Standard Delivery Van</option>
                <option value="MOTORCYCLE">Motorcycle / Scooter</option>
                <option value="TRUCK">Heavy / Medium Truck</option>
                <option value="BICYCLE">Cargo Bicycle</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Energy / Fuel</label>
              <select
                value={fuelType}
                onChange={(e) => setFuelType(e.target.value as FuelType)}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
              >
                <option value="ELECTRIC">Electric (EV Battery)</option>
                <option value="DIESEL">Diesel</option>
                <option value="PETROL">Petrol</option>
                <option value="HYBRID">Hybrid</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Max Weight (kg)</label>
              <input
                type="number"
                step="0.5"
                min="1"
                value={capacityKg}
                onChange={(e) => setCapacityKg(Number(e.target.value))}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Volume (m³)</label>
              <input
                type="number"
                step="0.1"
                min="0.1"
                value={capacityVolumeM3}
                onChange={(e) => setCapacityVolumeM3(Number(e.target.value))}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Max Parcels</label>
              <input
                type="number"
                min="1"
                value={maxParcels}
                onChange={(e) => setMaxParcels(Number(e.target.value))}
                className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1">
              Assigned Base Branch
            </label>
            <select
              value={assignedBranchId}
              onChange={(e) => setAssignedBranchId(e.target.value)}
              className="w-full px-3 py-2 bg-slate-800 border border-slate-700 rounded-xl text-white text-sm focus:outline-none focus:border-indigo-500"
            >
              <option value="">No Assigned Branch (Floating)</option>
              {branches.map((b) => (
                <option key={b.id} value={b.id}>
                  {b.name} ({b.code}) - {b.city}
                </option>
              ))}
            </select>
          </div>

          <div className="pt-4 border-t border-slate-800 flex justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm font-semibold rounded-xl transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className="px-5 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-sm font-semibold rounded-xl shadow-lg shadow-indigo-600/20 transition-all flex items-center gap-2"
            >
              {isSubmitting ? (
                <span className="inline-block w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin" />
              ) : (
                <Zap className="w-4 h-4" />
              )}
              Register Vehicle
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
