/**
 * LogiFlows Authoritative Routing & Geocoding Service Client
 * 
 * Supports OpenRouteService API with automatic resilient fallback.
 */

export interface LatLng {
  latitude: number;
  longitude: number;
}

export interface RouteStep {
  instruction: string;
  distance_meters: number;
  duration_seconds: number;
  name?: string;
}

export interface RouteResult {
  distance_meters: number;
  duration_seconds: number;
  distance_km: number;
  duration_minutes: number;
  eta_formatted: string;
  geometry: [number, number][]; // [lat, lon] points for Leaflet polyline
  steps: RouteStep[];
  source: 'LIVE_OPENROUTESERVICE' | 'SIMULATED_GEODESIC';
  provider: string;
}

export interface RoutingStatus {
  status: string;
  provider: string;
  has_api_key: boolean;
  key_preview?: string;
  mode: 'LIVE_CLOUD' | 'SIMULATED_FALLBACK';
}

/**
 * Fetch the operational status of the Routing Service from the backend.
 */
export async function getRoutingStatus(): Promise<RoutingStatus> {
  try {
    const res = await fetch('/api/v1/routing/status');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const json = await res.json();
    return json.data;
  } catch (err) {
    return {
      status: 'FALLBACK',
      provider: 'Local Geospatial Fallback',
      has_api_key: false,
      mode: 'SIMULATED_FALLBACK',
    };
  }
}

/**
 * Calculate the driving route geometry and ETA between origin and destination.
 */
export async function getDirections(
  origin: LatLng,
  destination: LatLng,
  waypoints: LatLng[] = [],
  profile: string = 'driving-car'
): Promise<RouteResult> {
  try {
    const res = await fetch('/api/v1/routing/directions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        origin,
        destination,
        waypoints,
        profile,
      }),
    });

    if (res.ok) {
      const json = await res.json();
      return json.data;
    }
  } catch (err) {
    console.warn('Backend routing unavailable, generating client fallback geometry', err);
  }

  // Resilient Client-Side Fallback Geometry
  return generateClientFallbackRoute(origin, destination, waypoints);
}

/**
 * Generates smooth curved road geometry with realistic city driving metrics
 */
function generateClientFallbackRoute(
  origin: LatLng,
  destination: LatLng,
  waypoints: LatLng[] = []
): RouteResult {
  const points = [origin, ...waypoints, destination];
  const polyline: [number, number][] = [];
  let totalDistanceMeters = 0;

  for (let i = 0; i < points.length - 1; i++) {
    const p1 = points[i];
    const p2 = points[i + 1];

    const dist = haversineDistance(p1.latitude, p1.longitude, p2.latitude, p2.longitude) * 1.25;
    totalDistanceMeters += dist;

    const segments = 8;
    for (let s = 0; s <= segments; s++) {
      const t = s / segments;
      let lat = p1.latitude + t * (p2.latitude - p1.latitude);
      let lon = p1.longitude + t * (p2.longitude - p1.longitude);

      // Add curvature
      const curve = Math.sin(t * Math.PI) * 0.0025;
      lat += curve * 0.7;
      lon += curve * 0.5;

      if (s > 0 || i === 0) {
        polyline.push([Number(lat.toFixed(5)), Number(lon.toFixed(5))]);
      }
    }
  }

  // 36 km/h city average speed = 10 m/s
  const durationSeconds = Math.round(totalDistanceMeters / 10);
  const distanceKm = Math.round((totalDistanceMeters / 1000) * 10) / 10;
  const durationMinutes = Math.round(durationSeconds / 60);

  let etaFormatted = `${durationMinutes} mins`;
  if (durationMinutes >= 60) {
    const hrs = Math.floor(durationMinutes / 60);
    const rem = durationMinutes % 60;
    etaFormatted = rem > 0 ? `${hrs} hr ${rem} mins` : `${hrs} hr`;
  }

  return {
    distance_meters: Math.round(totalDistanceMeters),
    duration_seconds: durationSeconds,
    distance_km: distanceKm,
    duration_minutes: durationMinutes,
    eta_formatted: etaFormatted,
    geometry: polyline,
    steps: [
      {
        instruction: 'Depart facility on primary freight lane',
        distance_meters: totalDistanceMeters * 0.2,
        duration_seconds: durationSeconds * 0.2,
        name: 'Arterial Corridor',
      },
      {
        instruction: 'Continue on urban delivery route',
        distance_meters: totalDistanceMeters * 0.6,
        duration_seconds: durationSeconds * 0.6,
        name: 'Expressway Outer Ring',
      },
      {
        instruction: 'Arrive at destination delivery address',
        distance_meters: totalDistanceMeters * 0.2,
        duration_seconds: durationSeconds * 0.2,
        name: 'Consignee Drop Zone',
      },
    ],
    source: 'SIMULATED_GEODESIC',
    provider: 'LogiFlows Spatial Geometry Engine',
  };
}

function haversineDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 6371000;
  const dLat = ((lat2 - lat1) * Math.PI) / 180;
  const dLon = ((lon2 - lon1) * Math.PI) / 180;
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos((lat1 * Math.PI) / 180) *
      Math.cos((lat2 * Math.PI) / 180) *
      Math.sin(dLon / 2) *
      Math.sin(dLon / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c;
}
