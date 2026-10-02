import React, { useEffect, useRef, useState } from 'react';
import L from 'leaflet';
import { Navigation, CheckCircle2, XCircle, Clock, Building2, Search, Crosshair } from 'lucide-react';
import { Branch, ServiceabilityCheckResponse } from '../types/tenancy';

interface GeospatialMapProps {
  tenantId: string;
  branches: Branch[];
  onSelectBranch?: (branch: Branch) => void;
}

export const GeospatialMap: React.FC<GeospatialMapProps> = ({ tenantId, branches, onSelectBranch }) => {
  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapInstanceRef = useRef<L.Map | null>(null);
  const layersRef = useRef<{
    markers: L.LayerGroup;
    polygons: L.LayerGroup;
    targetMarker: L.Marker | null;
  }>({
    markers: L.layerGroup(),
    polygons: L.layerGroup(),
    targetMarker: null,
  });

  const [testLat, setTestLat] = useState<string>('12.9756');
  const [testLon, setTestLon] = useState<string>('77.6066');
  const [isChecking, setIsChecking] = useState<boolean>(false);
  const [result, setResult] = useState<ServiceabilityCheckResponse | null>(null);
  const [selectedBranch, setSelectedBranch] = useState<Branch | null>(null);

  // Initialize Map
  useEffect(() => {
    if (!mapContainerRef.current || mapInstanceRef.current) return;

    const map = L.map(mapContainerRef.current, {
      center: [12.9716, 77.5946],
      zoom: 11,
      zoomControl: false,
    });

    L.control.zoom({ position: 'bottomright' }).addTo(map);

    // Dark-themed OpenStreetMap tiles (CartoDB Dark Matter)
    L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
      subdomains: 'abcd',
      maxZoom: 19,
    }).addTo(map);

    layersRef.current.polygons.addTo(map);
    layersRef.current.markers.addTo(map);
    mapInstanceRef.current = map;

    // Click handler to drop test pin and check serviceability
    map.on('click', (e: L.LeafletMouseEvent) => {
      const { lat, lng } = e.latlng;
      setTestLat(lat.toFixed(5));
      setTestLon(lng.toFixed(5));
      triggerCheck(lat, lng);
    });

    return () => {
      map.remove();
      mapInstanceRef.current = null;
    };
  }, []);

  // Update Branch Markers and Service Area Polygons
  useEffect(() => {
    const map = mapInstanceRef.current;
    if (!map) return;

    layersRef.current.markers.clearLayers();
    layersRef.current.polygons.clearLayers();

    const bounds = L.latLngBounds([]);

    branches.forEach((b) => {
      bounds.extend([b.location.latitude, b.location.longitude]);

      // Marker Icon Color based on Branch Type
      const isHub = b.branch_type === 'SORTING_HUB';
      const isDC = b.branch_type === 'DISTRIBUTION_CENTER';
      const color = isHub ? '#8b5cf6' : isDC ? '#3b82f6' : '#10b981';
      const bgClass = isHub ? 'bg-purple-500' : isDC ? 'bg-blue-500' : 'bg-emerald-500';

      const customIcon = L.divIcon({
        className: 'custom-map-marker',
        html: `
          <div class="relative flex items-center justify-center">
            <div class="absolute -top-1 w-7 h-7 rounded-full ${bgClass} opacity-30 animate-ping"></div>
            <div class="w-8 h-8 rounded-xl ${bgClass} text-white flex items-center justify-center shadow-lg border-2 border-slate-900 cursor-pointer hover:scale-110 transition-transform">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 0 1 16 0Z"/><circle cx="12" cy="10" r="3"/></svg>
            </div>
          </div>
        `,
        iconSize: [32, 32],
        iconAnchor: [16, 16],
      });

      const marker = L.marker([b.location.latitude, b.location.longitude], { icon: customIcon });
      marker.bindPopup(`
        <div style="color: #0f172a; font-family: sans-serif; padding: 4px;">
          <div style="font-weight: 700; font-size: 13px;">${b.name}</div>
          <div style="font-size: 11px; color: #64748b; margin-top: 2px;">Code: <b>${b.code}</b> | Type: ${b.branch_type}</div>
          <div style="font-size: 11px; color: #475569; margin-top: 4px;">${b.address}</div>
          <div style="font-size: 11px; color: #0284c7; margin-top: 4px; font-weight: 600;">Capacity: ${b.daily_capacity.toLocaleString()} parcels/day</div>
        </div>
      `);
      marker.on('click', () => {
        setSelectedBranch(b);
        if (onSelectBranch) onSelectBranch(b);
      });
      layersRef.current.markers.addLayer(marker);

      // Service Area Polygon
      if (b.service_area && b.service_area.coordinates) {
        try {
          // GeoJSON Polygon coordinates are [lon, lat], Leaflet polygon needs [lat, lon]
          const ring = b.service_area.coordinates[0];
          const latLngs: L.LatLngTuple[] = ring.map((pt: [number, number]) => [pt[1], pt[0]]);

          const polygon = L.polygon(latLngs, {
            color: color,
            weight: 2,
            dashArray: '4, 6',
            fillColor: color,
            fillOpacity: 0.15,
          });

          polygon.bindTooltip(`<b>${b.code} Coverage Area</b><br/>${b.name}`, { sticky: true });
          layersRef.current.polygons.addLayer(polygon);

          latLngs.forEach((coord) => bounds.extend(coord));
        } catch (e) {
          console.error('Failed to parse polygon for branch', b.code, e);
        }
      }
    });

    if (bounds.isValid()) {
      map.fitBounds(bounds, { padding: [40, 40] });
    }
  }, [branches, onSelectBranch]);

  // Execute Serviceability Check
  const triggerCheck = async (lat: number, lon: number) => {
    setIsChecking(true);
    const map = mapInstanceRef.current;

    // Drop or Move Target Pin
    if (map) {
      if (layersRef.current.targetMarker) {
        layersRef.current.targetMarker.setLatLng([lat, lon]);
      } else {
        const targetIcon = L.divIcon({
          className: 'target-pin',
          html: `
            <div class="relative flex items-center justify-center">
              <div class="w-8 h-8 rounded-full bg-amber-400 text-slate-950 flex items-center justify-center shadow-2xl border-2 border-white font-bold animate-bounce">
                🎯
              </div>
            </div>
          `,
          iconSize: [32, 32],
          iconAnchor: [16, 16],
        });
        layersRef.current.targetMarker = L.marker([lat, lon], { icon: targetIcon }).addTo(map);
      }
    }

    try {
      const res = await fetch(`/api/v1/serviceability/check`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          tenant_id: tenantId,
          latitude: lat,
          longitude: lon,
        }),
      });
      const json = await res.json();
      if (res.ok) {
        setResult(json.data);
      } else {
        setResult({
          is_serviceable: false,
          status: 'ERROR',
          message: json.error?.message || 'Check failed',
          distance_meters: 0,
        });
      }
    } catch (err: any) {
      setResult({
        is_serviceable: false,
        status: 'ERROR',
        message: err.message || 'Network error',
        distance_meters: 0,
      });
    } finally {
      setIsChecking(false);
    }
  };

  const handleManualSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const lat = parseFloat(testLat);
    const lon = parseFloat(testLon);
    if (!isNaN(lat) && !isNaN(lon)) {
      triggerCheck(lat, lon);
      if (mapInstanceRef.current) {
        mapInstanceRef.current.panTo([lat, lon]);
      }
    }
  };

  const setPreset = (_name: string, lat: number, lon: number) => {
    setTestLat(lat.toString());
    setTestLon(lon.toString());
    triggerCheck(lat, lon);
    if (mapInstanceRef.current) {
      mapInstanceRef.current.flyTo([lat, lon], 12);
    }
  };

  return (
    <div className="flex flex-col xl:flex-row gap-6 w-full">
      {/* Map Section */}
      <div className="flex-1 bg-slate-900 border border-slate-800 rounded-2xl overflow-hidden shadow-2xl relative min-h-[550px] flex flex-col">
        {/* Top Control Bar */}
        <div className="absolute top-4 left-4 right-4 z-[1000] flex flex-wrap items-center justify-between gap-3 pointer-events-none">
          <div className="bg-slate-900/90 backdrop-blur-md border border-slate-700/80 px-4 py-2.5 rounded-xl shadow-xl flex items-center gap-3 pointer-events-auto">
            <Crosshair className="w-5 h-5 text-indigo-400 animate-spin-slow" />
            <div>
              <div className="text-xs font-semibold uppercase tracking-wider text-slate-400">Live PostGIS Inspector</div>
              <div className="text-xs text-slate-200">Click anywhere on the map to test serviceability</div>
            </div>
          </div>

          <div className="bg-slate-900/90 backdrop-blur-md border border-slate-700/80 px-3 py-2 rounded-xl shadow-xl flex items-center gap-4 text-xs font-medium pointer-events-auto">
            <span className="flex items-center gap-1.5 text-purple-300">
              <span className="w-2.5 h-2.5 rounded-full bg-purple-500"></span> Sorting Hub
            </span>
            <span className="flex items-center gap-1.5 text-blue-300">
              <span className="w-2.5 h-2.5 rounded-full bg-blue-500"></span> Distribution Center
            </span>
            <span className="flex items-center gap-1.5 text-emerald-300">
              <span className="w-2.5 h-2.5 rounded-full bg-emerald-500"></span> Local Office
            </span>
          </div>
        </div>

        {/* Leaflet Container */}
        <div ref={mapContainerRef} className="w-full h-full min-h-[550px] z-10" />

        {/* Selected Branch Detail Drawer (Bottom of Map) */}
        {selectedBranch && (
          <div className="absolute bottom-4 left-4 right-4 z-[1000] bg-slate-950/95 backdrop-blur-md border border-slate-800 p-4 rounded-xl shadow-2xl flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="p-2.5 bg-indigo-500/10 text-indigo-400 rounded-lg border border-indigo-500/20">
                <Building2 className="w-5 h-5" />
              </div>
              <div>
                <div className="font-bold text-slate-100 flex items-center gap-2">
                  {selectedBranch.name}
                  <span className="text-[10px] px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 border border-slate-700">
                    {selectedBranch.code}
                  </span>
                </div>
                <div className="text-xs text-slate-400 mt-0.5">{selectedBranch.address} • {selectedBranch.city}</div>
              </div>
            </div>
            <div className="flex items-center gap-6 text-xs">
              <div>
                <span className="text-slate-500">Coordinates:</span>{' '}
                <span className="text-slate-300 font-mono">{selectedBranch.location.latitude.toFixed(4)}, {selectedBranch.location.longitude.toFixed(4)}</span>
              </div>
              <div>
                <span className="text-slate-500">Daily Quota:</span>{' '}
                <span className="text-indigo-400 font-semibold">{selectedBranch.daily_capacity.toLocaleString()} pkgs</span>
              </div>
              <button
                onClick={() => setSelectedBranch(null)}
                className="text-slate-400 hover:text-white text-xs underline"
              >
                Close
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Side Inspector & Coordinate Tester */}
      <div className="w-full xl:w-96 flex flex-col gap-5">
        {/* Coordinate Tester Card */}
        <div className="bg-slate-900 border border-slate-800 rounded-2xl p-5 shadow-xl">
          <div className="flex items-center gap-2 font-bold text-slate-100 mb-4">
            <Navigation className="w-5 h-5 text-indigo-400" />
            <span>Serviceability Engine</span>
          </div>

          <form onSubmit={handleManualSubmit} className="space-y-4">
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-semibold text-slate-400 mb-1">Latitude</label>
                <input
                  type="text"
                  value={testLat}
                  onChange={(e) => setTestLat(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-xs font-mono text-slate-100 focus:outline-none focus:border-indigo-500"
                  placeholder="e.g. 12.9716"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-semibold text-slate-400 mb-1">Longitude</label>
                <input
                  type="text"
                  value={testLon}
                  onChange={(e) => setTestLon(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-xs font-mono text-slate-100 focus:outline-none focus:border-indigo-500"
                  placeholder="e.g. 77.5946"
                  required
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={isChecking}
              className="w-full bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-medium py-2.5 px-4 rounded-xl text-xs flex items-center justify-center gap-2 shadow-lg shadow-indigo-600/20 transition-colors"
            >
              {isChecking ? (
                <>
                  <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Checking PostGIS Boundary...
                </>
              ) : (
                <>
                  <Search className="w-4 h-4" />
                  Evaluate Serviceability
                </>
              )}
            </button>
          </form>

          {/* Quick Presets */}
          <div className="mt-4 pt-4 border-t border-slate-800">
            <div className="text-[11px] font-semibold text-slate-400 mb-2">Preset Test Locations:</div>
            <div className="flex flex-wrap gap-1.5">
              <button
                type="button"
                onClick={() => setPreset('MG Road', 12.9756, 77.6066)}
                className="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 rounded-lg text-[11px] text-slate-300 transition-colors"
              >
                MG Road (Central)
              </button>
              <button
                type="button"
                onClick={() => setPreset('Hebbal', 13.0358, 77.5970)}
                className="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 rounded-lg text-[11px] text-slate-300 transition-colors"
              >
                Hebbal (North)
              </button>
              <button
                type="button"
                onClick={() => setPreset('Koramangala', 12.9345, 77.6265)}
                className="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 rounded-lg text-[11px] text-slate-300 transition-colors"
              >
                Koramangala (South)
              </button>
              <button
                type="button"
                onClick={() => setPreset('Hoskote (Out)', 13.0710, 77.7980)}
                className="px-2.5 py-1 bg-rose-950/40 text-rose-300 border border-rose-800/40 hover:bg-rose-900/50 rounded-lg text-[11px] transition-colors"
              >
                Hoskote (Out of Area)
              </button>
            </div>
          </div>
        </div>

        {/* Real-time Verification Result Card */}
        {result && (
          <div
            className={`border rounded-2xl p-5 shadow-xl transition-all ${
              result.is_serviceable
                ? 'bg-emerald-950/20 border-emerald-500/40 text-emerald-200'
                : 'bg-rose-950/20 border-rose-500/40 text-rose-200'
            }`}
          >
            <div className="flex items-start gap-3">
              {result.is_serviceable ? (
                <CheckCircle2 className="w-6 h-6 text-emerald-400 shrink-0 mt-0.5" />
              ) : (
                <XCircle className="w-6 h-6 text-rose-400 shrink-0 mt-0.5" />
              )}
              <div className="flex-1">
                <div className="font-bold text-sm text-slate-100 flex items-center justify-between">
                  <span>{result.is_serviceable ? 'Serviceable Location' : 'Outside Coverage'}</span>
                  <span
                    className={`text-[10px] px-2 py-0.5 rounded-full font-semibold ${
                      result.is_serviceable ? 'bg-emerald-500/20 text-emerald-300' : 'bg-rose-500/20 text-rose-300'
                    }`}
                  >
                    {result.status}
                  </span>
                </div>
                <div className="text-xs text-slate-300 mt-1 leading-relaxed">{result.message}</div>

                {result.matched_branch && (
                  <div className="mt-3.5 pt-3 border-t border-emerald-500/20 space-y-1.5 text-xs">
                    <div className="text-slate-300">
                      <span className="text-slate-400">Assigned Branch:</span>{' '}
                      <b className="text-white">{result.matched_branch.name}</b> ({result.matched_branch.code})
                    </div>
                    <div className="text-slate-300">
                      <span className="text-slate-400">Hub Distance:</span>{' '}
                      <b className="text-white">{(result.distance_meters / 1000).toFixed(2)} km</b>
                    </div>
                    {result.estimated_eta && (
                      <div className="flex items-center gap-1.5 text-emerald-300 font-semibold mt-2">
                        <Clock className="w-3.5 h-3.5" />
                        <span>Transit ETA: {result.estimated_eta}</span>
                      </div>
                    )}
                  </div>
                )}

                {result.nearest_branch && (
                  <div className="mt-3.5 pt-3 border-t border-rose-500/20 space-y-1.5 text-xs">
                    <div className="text-slate-300">
                      <span className="text-slate-400">Nearest Hub:</span>{' '}
                      <b className="text-white">{result.nearest_branch.name}</b> ({result.nearest_branch.code})
                    </div>
                    <div className="text-slate-300">
                      <span className="text-slate-400">Proximity:</span>{' '}
                      <b className="text-white">{(result.distance_meters / 1000).toFixed(2)} km away</b>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
