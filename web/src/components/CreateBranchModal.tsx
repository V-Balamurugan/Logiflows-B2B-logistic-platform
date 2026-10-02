import React, { useState } from 'react';
import { X, Building2, MapPin } from 'lucide-react';
import { BranchType, BranchCreatePayload, Branch } from '../types/tenancy';

interface CreateBranchModalProps {
  tenantId: string;
  isOpen: boolean;
  onClose: () => void;
  onCreated: (branch: Branch) => void;
}

export const CreateBranchModal: React.FC<CreateBranchModalProps> = ({ tenantId, isOpen, onClose, onCreated }) => {
  const [code, setCode] = useState('');
  const [name, setName] = useState('');
  const [branchType, setBranchType] = useState<BranchType>('DISTRIBUTION_CENTER');
  const [address, setAddress] = useState('');
  const [city, setCity] = useState('Bengaluru');
  const [state, setState] = useState('Karnataka');
  const [postalCode, setPostalCode] = useState('560001');
  const [contactPhone, setContactPhone] = useState('+91 80 ');
  const [contactEmail, setContactEmail] = useState('');
  const [dailyCapacity, setDailyCapacity] = useState(5000);
  const [latitude, setLatitude] = useState('12.9716');
  const [longitude, setLongitude] = useState('77.5946');
  const [coverageRadiusKm, setCoverageRadiusKm] = useState(8);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  // Helper to generate a rough circular polygon buffer around coordinates
  const generatePolygon = (lat: number, lon: number, radiusKm: number) => {
    const points = 16;
    const coords: [number, number][] = [];
    const earthRadiusKm = 6371;

    for (let i = 0; i < points; i++) {
      const angle = (i * 360) / points;
      const rad = (angle * Math.PI) / 180;
      const dLat = (radiusKm / earthRadiusKm) * (180 / Math.PI);
      const dLon = ((radiusKm / earthRadiusKm) * (180 / Math.PI)) / Math.cos((lat * Math.PI) / 180);

      const pLat = lat + dLat * Math.cos(rad);
      const pLon = lon + dLon * Math.sin(rad);
      coords.push([parseFloat(pLon.toFixed(5)), parseFloat(pLat.toFixed(5))]);
    }
    // Close polygon ring
    coords.push(coords[0]);

    return {
      type: 'Polygon',
      coordinates: [coords],
    };
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    const lat = parseFloat(latitude);
    const lon = parseFloat(longitude);
    if (isNaN(lat) || isNaN(lon)) {
      setError('Please provide valid latitude and longitude numbers');
      setLoading(false);
      return;
    }

    const serviceArea = generatePolygon(lat, lon, coverageRadiusKm);

    const payload: BranchCreatePayload = {
      code: code.trim().toUpperCase(),
      name: name.trim(),
      branch_type: branchType,
      address: address.trim(),
      city: city.trim(),
      state: state.trim(),
      postal_code: postalCode.trim(),
      country: 'India',
      contact_phone: contactPhone.trim(),
      contact_email: contactEmail.trim(),
      daily_capacity: dailyCapacity,
      status: 'ACTIVE',
      location: { latitude: lat, longitude: lon },
      service_area: serviceArea,
    };

    try {
      const token = localStorage.getItem('access_token');
      const res = await fetch(`/api/v1/tenants/${tenantId}/branches`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });

      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.error?.message || 'Failed to create branch');
      }

      onCreated(json.data);
      onClose();
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm overflow-y-auto">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-2xl shadow-2xl overflow-hidden my-8">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-slate-800">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-indigo-500/10 text-indigo-400 rounded-xl border border-indigo-500/20">
              <Building2 className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-slate-100">Deploy New Branch or Hub</h2>
              <p className="text-xs text-slate-400">Configure spatial coordinates, type, and service boundary</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-slate-400 hover:text-white rounded-lg hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          {error && (
            <div className="p-3 bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs rounded-xl">
              {error}
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Branch Code *</label>
              <input
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder="e.g. BLR-HUB-02"
                required
                className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 font-mono uppercase focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Branch Name *</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. East Bengaluru Sorting Hub"
                required
                className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Branch Type *</label>
              <select
                value={branchType}
                onChange={(e) => setBranchType(e.target.value as BranchType)}
                className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
              >
                <option value="SORTING_HUB">SORTING_HUB (Regional Super Hub)</option>
                <option value="DISTRIBUTION_CENTER">DISTRIBUTION_CENTER (Mid-mile DC)</option>
                <option value="LOCAL_OFFICE">LOCAL_OFFICE (Last-mile Delivery Office)</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Daily Capacity (Parcels) *</label>
              <input
                type="number"
                value={dailyCapacity}
                onChange={(e) => setDailyCapacity(parseInt(e.target.value) || 0)}
                required
                className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1">Street Address *</label>
            <input
              type="text"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder="e.g. 45 Outer Ring Road, Mahadevapura"
              required
              className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Contact Phone</label>
              <input
                type="text"
                value={contactPhone}
                onChange={(e) => setContactPhone(e.target.value)}
                placeholder="+91 80 1234 5678"
                className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Contact Email</label>
              <input
                type="email"
                value={contactEmail}
                onChange={(e) => setContactEmail(e.target.value)}
                placeholder="branch@logiflows.io"
                className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">City</label>
              <input
                type="text"
                value={city}
                onChange={(e) => setCity(e.target.value)}
                className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">State</label>
              <input
                type="text"
                value={state}
                onChange={(e) => setState(e.target.value)}
                className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1">Postal Code</label>
              <input
                type="text"
                value={postalCode}
                onChange={(e) => setPostalCode(e.target.value)}
                className="w-full bg-slate-950 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          {/* Spatial Coordinates & Service Boundary */}
          <div className="p-4 bg-slate-950/60 border border-slate-800 rounded-xl space-y-3">
            <div className="flex items-center gap-2 text-xs font-bold text-indigo-400">
              <MapPin className="w-4 h-4" />
              <span>Geospatial PostGIS Coordinates & Coverage Area</span>
            </div>

            <div className="grid grid-cols-3 gap-3">
              <div>
                <label className="block text-[11px] font-semibold text-slate-400 mb-1">Latitude</label>
                <input
                  type="text"
                  value={latitude}
                  onChange={(e) => setLatitude(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:border-indigo-500"
                />
              </div>
              <div>
                <label className="block text-[11px] font-semibold text-slate-400 mb-1">Longitude</label>
                <input
                  type="text"
                  value={longitude}
                  onChange={(e) => setLongitude(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:border-indigo-500"
                />
              </div>
              <div>
                <label className="block text-[11px] font-semibold text-slate-400 mb-1">Coverage Radius</label>
                <div className="flex items-center gap-2">
                  <input
                    type="range"
                    min="2"
                    max="25"
                    value={coverageRadiusKm}
                    onChange={(e) => setCoverageRadiusKm(parseInt(e.target.value))}
                    className="w-full accent-indigo-500"
                  />
                  <span className="text-xs font-semibold text-slate-200 w-12">{coverageRadiusKm} km</span>
                </div>
              </div>
            </div>
            <div className="text-[11px] text-slate-400">
              Generates a 16-point closed PostGIS polygon service area centered on the coordinate.
            </div>
          </div>

          <div className="flex justify-end gap-3 pt-4 border-t border-slate-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs font-medium text-slate-300 hover:text-white rounded-xl hover:bg-slate-800 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-5 py-2 text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 rounded-xl shadow-lg shadow-indigo-600/20 transition-colors"
            >
              {loading ? 'Registering with PostGIS...' : 'Deploy Branch'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
