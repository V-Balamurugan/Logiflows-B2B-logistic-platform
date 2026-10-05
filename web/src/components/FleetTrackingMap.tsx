import React, { useEffect, useRef, useState, useCallback } from 'react';
import L from 'leaflet';
import {
  Truck,
  Battery,
  RefreshCw,
  Compass,
  Gauge,
  MapPin,
  Activity,
  Filter,
  AlertTriangle,
} from 'lucide-react';
import { LiveVehiclePosition, VehicleTelematics } from '../types/fleet';
import { fleetService } from '../services/fleetService';

interface FleetTrackingMapProps {
  token: string;
  tenantId?: string;
  onSelectVehicle?: (vehicle: LiveVehiclePosition) => void;
}

export const FleetTrackingMap: React.FC<FleetTrackingMapProps> = ({
  token,
  tenantId,
  onSelectVehicle,
}) => {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapInstanceRef = useRef<L.Map | null>(null);
  const markersLayerRef = useRef<L.LayerGroup>(L.layerGroup());
  const breadcrumbLayerRef = useRef<L.LayerGroup>(L.layerGroup());

  const [positions, setPositions] = useState<LiveVehiclePosition[]>([]);
  const [selectedPosition, setSelectedPosition] = useState<LiveVehiclePosition | null>(null);
  const [breadcrumbs, setBreadcrumbs] = useState<VehicleTelematics[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [typeFilter, setTypeFilter] = useState<string>('ALL');
  const [statusFilter, setStatusFilter] = useState<string>('ALL');
  const [autoRefresh, setAutoRefresh] = useState<boolean>(true);
  const [lastRefreshed, setLastRefreshed] = useState<Date>(new Date());
  const [error, setError] = useState<string | null>(null);

  // Fetch live fleet positions
  const fetchLiveFleet = useCallback(async () => {
    try {
      setError(null);
      const data = await fleetService.getLiveFleet(token, tenantId);
      setPositions(data);
      setLastRefreshed(new Date());
    } catch (err: any) {
      setError(err?.message || 'Failed to fetch live fleet telematics');
    } finally {
      setIsLoading(false);
    }
  }, [token, tenantId]);

  // Load breadcrumbs for selected vehicle
  const loadBreadcrumbs = useCallback(
    async (vehicleId: string) => {
      try {
        const hist = await fleetService.getTelematicsHistory(token, vehicleId, 50, tenantId);
        setBreadcrumbs(hist);
      } catch (err) {
        console.error('Failed to load breadcrumbs:', err);
      }
    },
    [token, tenantId]
  );

  // Initialize Leaflet Map
  useEffect(() => {
    if (!mapContainerRef.current || mapInstanceRef.current) return;

    const map = L.map(mapContainerRef.current, {
      center: [12.9716, 77.5946],
      zoom: 12,
      zoomControl: false,
    });

    L.control.zoom({ position: 'bottomright' }).addTo(map);

    // Carto Dark Matter tiles for ultra-crisp modern dashboard look
    L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
      subdomains: 'abcd',
      maxZoom: 19,
    }).addTo(map);

    markersLayerRef.current.addTo(map);
    breadcrumbLayerRef.current.addTo(map);
    mapInstanceRef.current = map;

    return () => {
      map.remove();
      mapInstanceRef.current = null;
    };
  }, []);

  // Auto-refresh interval
  useEffect(() => {
    fetchLiveFleet();
    if (!autoRefresh) return;
    const interval = setInterval(fetchLiveFleet, 10000);
    return () => clearInterval(interval);
  }, [fetchLiveFleet, autoRefresh]);

  // Create custom HTML marker for vehicle with heading and status styling
  const createVehicleIcon = (pos: LiveVehiclePosition, isSelected: boolean) => {
    const statusColor =
      pos.status === 'ON_ROUTE'
        ? '#10b981' // emerald-500
        : pos.status === 'MAINTENANCE'
        ? '#f59e0b' // amber-500
        : '#3b82f6'; // blue-500

    const pulseEffect =
      pos.status === 'ON_ROUTE'
        ? `<div class="absolute -inset-1 rounded-full bg-emerald-500/30 animate-ping"></div>`
        : '';

    const borderStyle = isSelected ? 'border-2 border-white scale-125 shadow-lg' : 'border border-white/20';

    return L.divIcon({
      className: 'vehicle-marker-wrapper',
      html: `
        <div class="relative flex items-center justify-center cursor-pointer transition-transform duration-300">
          ${pulseEffect}
          <div style="background-color: ${statusColor}; transform: rotate(${pos.heading_degrees}deg);" 
               class="relative flex items-center justify-center w-8 h-8 rounded-full text-white shadow-md ${borderStyle}">
            <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <polygon points="12 2 19 21 12 17 5 21 12 2"></polygon>
            </svg>
          </div>
          <div class="absolute -bottom-4 px-1.5 py-0.5 rounded bg-slate-900/90 text-[10px] font-mono font-semibold text-white whitespace-nowrap shadow-sm border border-slate-700">
            ${pos.license_plate}
          </div>
        </div>
      `,
      iconSize: [32, 32],
      iconAnchor: [16, 16],
    });
  };

  // Update map markers when positions or filters change
  useEffect(() => {
    if (!mapInstanceRef.current) return;

    markersLayerRef.current.clearLayers();

    const filtered = positions.filter((p) => {
      if (typeFilter !== 'ALL' && p.vehicle_type !== typeFilter) return false;
      if (statusFilter !== 'ALL' && p.status !== statusFilter) return false;
      return true;
    });

    const bounds = L.latLngBounds([]);

    filtered.forEach((pos) => {
      if (pos.latitude === 0 && pos.longitude === 0) return;

      const isSelected = selectedPosition?.vehicle_id === pos.vehicle_id;
      const marker = L.marker([pos.latitude, pos.longitude], {
        icon: createVehicleIcon(pos, isSelected),
      });

      marker.on('click', () => {
        setSelectedPosition(pos);
        if (onSelectVehicle) onSelectVehicle(pos);
        loadBreadcrumbs(pos.vehicle_id);
      });

      markersLayerRef.current.addLayer(marker);
      bounds.extend([pos.latitude, pos.longitude]);
    });

    // Auto-fit bounds if we have positions and haven't centered yet
    if (bounds.isValid() && filtered.length > 0 && !selectedPosition) {
      mapInstanceRef.current.fitBounds(bounds, { padding: [50, 50], maxZoom: 14 });
    }
  }, [positions, typeFilter, statusFilter, selectedPosition, onSelectVehicle, loadBreadcrumbs]);

  // Draw breadcrumbs route polyline when selected vehicle changes
  useEffect(() => {
    breadcrumbLayerRef.current.clearLayers();
    if (!mapInstanceRef.current || breadcrumbs.length === 0) return;

    const latlngs: [number, number][] = breadcrumbs.map((b) => [b.latitude, b.longitude]);

    const polyline = L.polyline(latlngs, {
      color: '#38bdf8',
      weight: 4,
      opacity: 0.8,
      dashArray: '8, 8',
    });

    breadcrumbLayerRef.current.addLayer(polyline);

    breadcrumbs.forEach((b, idx) => {
      const circle = L.circleMarker([b.latitude, b.longitude], {
        radius: idx === breadcrumbs.length - 1 ? 6 : 3,
        color: '#0284c7',
        fillColor: '#38bdf8',
        fillOpacity: 0.9,
      });
      circle.bindTooltip(
        `Recorded: ${new Date(b.recorded_at).toLocaleTimeString()}<br/>Speed: ${b.speed_kmh} km/h`,
        { className: 'leaflet-dark-tooltip' }
      );
      breadcrumbLayerRef.current.addLayer(circle);
    });
  }, [breadcrumbs]);

  return (
    <div className="relative w-full h-full flex flex-col bg-slate-900 rounded-2xl overflow-hidden border border-slate-800 shadow-xl">
      {/* Top Map Control Bar */}
      <div className="absolute top-4 left-4 right-4 z-[1000] flex flex-wrap items-center justify-between gap-3 p-3 bg-slate-900/90 backdrop-blur-md rounded-xl border border-slate-700/80 shadow-lg">
        <div className="flex items-center gap-2">
          <div className="p-2 rounded-lg bg-indigo-500/10 text-indigo-400">
            <Truck className="w-5 h-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-sm font-semibold text-white">Live Fleet Telematics</span>
              <span className="px-2 py-0.5 text-xs font-mono rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                {positions.length} Active
              </span>
            </div>
            <p className="text-[11px] text-slate-400">
              Synced: {lastRefreshed.toLocaleTimeString()}
            </p>
          </div>
        </div>

        {/* Filters & Actions */}
        <div className="flex flex-wrap items-center gap-2">
          {/* Vehicle Type Filter */}
          <div className="flex items-center gap-1.5 bg-slate-800/80 px-2.5 py-1.5 rounded-lg border border-slate-700">
            <Filter className="w-3.5 h-3.5 text-slate-400" />
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              className="bg-transparent text-xs text-slate-200 outline-none cursor-pointer"
            >
              <option value="ALL" className="bg-slate-900">All Vehicle Types</option>
              <option value="ELECTRIC_VAN" className="bg-slate-900">Electric Van (EV)</option>
              <option value="VAN" className="bg-slate-900">Delivery Van</option>
              <option value="MOTORCYCLE" className="bg-slate-900">Motorcycle</option>
              <option value="TRUCK" className="bg-slate-900">Truck</option>
              <option value="BICYCLE" className="bg-slate-900">Bicycle</option>
            </select>
          </div>

          {/* Status Filter */}
          <div className="flex items-center gap-1.5 bg-slate-800/80 px-2.5 py-1.5 rounded-lg border border-slate-700">
            <Activity className="w-3.5 h-3.5 text-slate-400" />
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="bg-transparent text-xs text-slate-200 outline-none cursor-pointer"
            >
              <option value="ALL" className="bg-slate-900">All Statuses</option>
              <option value="ON_ROUTE" className="bg-slate-900">On Route (Active)</option>
              <option value="AVAILABLE" className="bg-slate-900">Available / Parked</option>
              <option value="MAINTENANCE" className="bg-slate-900">Under Maintenance</option>
            </select>
          </div>

          {/* Auto Refresh Toggle */}
          <button
            onClick={() => setAutoRefresh(!autoRefresh)}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors border ${
              autoRefresh
                ? 'bg-indigo-600/20 text-indigo-300 border-indigo-500/30'
                : 'bg-slate-800 text-slate-400 border-slate-700 hover:text-white'
            }`}
          >
            <RefreshCw className={`w-3.5 h-3.5 ${autoRefresh ? 'animate-spin' : ''}`} />
            Auto (10s)
          </button>

          <button
            onClick={fetchLiveFleet}
            disabled={isLoading}
            className="p-1.5 rounded-lg bg-slate-800 text-slate-300 hover:text-white border border-slate-700"
            title="Manual Refresh"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} />
          </button>
        </div>
      </div>

      {/* Main Map Canvas */}
      <div ref={mapContainerRef} className="w-full h-full min-h-[480px] z-0" />

      {/* Selected Vehicle Floating HUD */}
      {selectedPosition && (
        <div className="absolute bottom-6 left-6 z-[1000] w-80 p-4 bg-slate-900/95 backdrop-blur-md rounded-xl border border-slate-700/80 shadow-2xl animate-fade-in">
          <div className="flex items-start justify-between">
            <div>
              <div className="flex items-center gap-2">
                <span className="text-base font-bold text-white tracking-wide">
                  {selectedPosition.license_plate}
                </span>
                <span
                  className={`px-2 py-0.5 text-[10px] font-semibold rounded-md uppercase ${
                    selectedPosition.status === 'ON_ROUTE'
                      ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
                      : selectedPosition.status === 'MAINTENANCE'
                      ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30'
                      : 'bg-blue-500/20 text-blue-300 border border-blue-500/30'
                  }`}
                >
                  {selectedPosition.status.replace('_', ' ')}
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                {selectedPosition.make} {selectedPosition.model} • {selectedPosition.vehicle_type}
              </p>
            </div>
            <button
              onClick={() => {
                setSelectedPosition(null);
                setBreadcrumbs([]);
              }}
              className="text-slate-400 hover:text-white text-xs p-1"
            >
              ✕
            </button>
          </div>

          <div className="grid grid-cols-2 gap-2 mt-4 pt-3 border-t border-slate-800">
            <div className="p-2 rounded-lg bg-slate-800/60 border border-slate-700/40">
              <div className="flex items-center gap-1.5 text-slate-400 text-[11px]">
                <Gauge className="w-3.5 h-3.5 text-sky-400" />
                <span>Speed</span>
              </div>
              <p className="text-sm font-semibold text-white mt-1">
                {selectedPosition.speed_kmh.toFixed(1)} <span className="text-[10px] text-slate-400 font-normal">km/h</span>
              </p>
            </div>

            <div className="p-2 rounded-lg bg-slate-800/60 border border-slate-700/40">
              <div className="flex items-center gap-1.5 text-slate-400 text-[11px]">
                <Battery className="w-3.5 h-3.5 text-emerald-400" />
                <span>Energy / Fuel</span>
              </div>
              <p className="text-sm font-semibold text-emerald-400 mt-1">
                {selectedPosition.battery_or_fuel_percent.toFixed(1)}%
              </p>
            </div>

            <div className="p-2 rounded-lg bg-slate-800/60 border border-slate-700/40">
              <div className="flex items-center gap-1.5 text-slate-400 text-[11px]">
                <Compass className="w-3.5 h-3.5 text-purple-400" />
                <span>Heading</span>
              </div>
              <p className="text-sm font-semibold text-white mt-1">
                {selectedPosition.heading_degrees.toFixed(0)}°
              </p>
            </div>

            <div className="p-2 rounded-lg bg-slate-800/60 border border-slate-700/40">
              <div className="flex items-center gap-1.5 text-slate-400 text-[11px]">
                <MapPin className="w-3.5 h-3.5 text-amber-400" />
                <span>Odometer</span>
              </div>
              <p className="text-sm font-semibold text-white mt-1">
                {selectedPosition.odometer_km.toFixed(1)} <span className="text-[10px] text-slate-400 font-normal">km</span>
              </p>
            </div>
          </div>

          <div className="mt-3 pt-2 text-[11px] text-slate-400 flex items-center justify-between border-t border-slate-800">
            <span>Coordinates:</span>
            <span className="font-mono text-slate-300">
              {selectedPosition.latitude.toFixed(4)}, {selectedPosition.longitude.toFixed(4)}
            </span>
          </div>

          {breadcrumbs.length > 0 && (
            <div className="mt-2 text-[11px] text-sky-400 flex items-center gap-1">
              <Activity className="w-3 h-3" />
              <span>Displaying {breadcrumbs.length} recent GPS breadcrumbs</span>
            </div>
          )}
        </div>
      )}

      {/* Error Toast */}
      {error && (
        <div className="absolute bottom-6 right-6 z-[1000] flex items-center gap-2 p-3 bg-red-900/90 border border-red-700 text-red-200 text-xs rounded-xl shadow-lg">
          <AlertTriangle className="w-4 h-4 text-red-400 flex-shrink-0" />
          <span>{error}</span>
        </div>
      )}
    </div>
  );
};
