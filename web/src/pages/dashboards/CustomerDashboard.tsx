import React, { useState, useEffect } from 'react';
import {
  Package, Search, PlusCircle, Clock, CheckCircle2, ShieldCheck,
  MapPin, Truck, Box, FileText, ArrowRight,
  Star, Bell, BarChart3, Route as RouteIcon
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { DashboardLayout } from '../../components/DashboardLayout';
import { getDirections, RouteResult } from '../../services/routingService';

interface CustomerDashboardProps {
  onNavigate: (path: string) => void;
}

const NAV = [
  { icon: <BarChart3 className="w-4 h-4" />, label: 'Overview' },
  { icon: <Search className="w-4 h-4" />, label: 'Track Shipment' },
  { icon: <Package className="w-4 h-4" />, label: 'Book Delivery' },
  { icon: <Clock className="w-4 h-4" />, label: 'Order History', badge: 5 },
  { icon: <MapPin className="w-4 h-4" />, label: 'Serviceability' },
  { icon: <Bell className="w-4 h-4" />, label: 'Notifications', badge: 2 },
];

interface TimelineEvent {
  label: string;
  location: string;
  time: string;
  done: boolean;
  active?: boolean;
}

const TIMELINE: TimelineEvent[] = [
  { label: 'Order Placed', location: 'Online', time: 'Oct 4, 09:12 AM', done: true },
  { label: 'Picked Up', location: 'Koramangala Pickup Hub', time: 'Oct 4, 11:40 AM', done: true },
  { label: 'In Transit', location: 'MG Road Sorting Hub', time: 'Oct 4, 02:15 PM', done: true, active: true },
  { label: 'Out for Delivery', location: 'Hebbal Distribution Center', time: 'Estimated Oct 5, 10:00 AM', done: false },
  { label: 'Delivered', location: 'Your Doorstep', time: 'Estimated Oct 5, 12:30 PM', done: false },
];

const RECENT_ORDERS = [
  { id: 'LF-8923-4412', destination: 'Mumbai, MH', status: 'In Transit', date: 'Oct 4', items: 1, amount: '₹349' },
  { id: 'LF-8801-1123', destination: 'Chennai, TN', status: 'Delivered', date: 'Sep 29', items: 3, amount: '₹890' },
  { id: 'LF-8750-6612', destination: 'Hyderabad, TS', status: 'Delivered', date: 'Sep 22', items: 1, amount: '₹210' },
  { id: 'LF-8701-9911', destination: 'Pune, MH', status: 'Cancelled', date: 'Sep 15', items: 2, amount: '₹540' },
];

const STATUS_STYLES: Record<string, string> = {
  'In Transit': 'bg-blue-500/15 text-blue-300 border-blue-500/30',
  'Delivered': 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30',
  'Cancelled': 'bg-rose-500/15 text-rose-300 border-rose-500/30',
};

const SERVICEABLE_ZONES = [
  { name: 'MG Road / Central Hub', lat: 12.9756, lng: 77.6066 },
  { name: 'Hebbal / North DC', lat: 13.0358, lng: 77.5970 },
  { name: 'Koramangala / South DC', lat: 12.9345, lng: 77.6265 },
];

export const CustomerDashboard: React.FC<CustomerDashboardProps> = ({ onNavigate }) => {
  const { user, tokens } = useAuth();
  const [activeNav, setActiveNav] = useState('Overview');
  const [trackingInput, setTrackingInput] = useState('');
  const [permissionCheck, setPermissionCheck] = useState<string | null>(null);
  const [serviceResult, setServiceResult] = useState<Record<string, string>>({});
  const [checkingZone, setCheckingZone] = useState<string | null>(null);
  const [liveRoute, setLiveRoute] = useState<RouteResult | null>(null);

  useEffect(() => {
    // Fetch live routing calculation for active shipment
    const fetchActiveShipmentRoute = async () => {
      try {
        const route = await getDirections(
          { latitude: 12.9756, longitude: 77.6066 }, // MG Road Hub
          { latitude: 13.0358, longitude: 77.5970 }  // Hebbal DC
        );
        setLiveRoute(route);
      } catch (e) {
        console.error(e);
      }
    };
    fetchActiveShipmentRoute();
  }, []);

  useEffect(() => {
    const verifyCustomerPerms = async () => {
      if (!tokens?.access_token) return;
      try {
        const res = await fetch('/api/v1/parcels/permission-check', {
          headers: { Authorization: `Bearer ${tokens.access_token}` },
        });
        const data = await res.json();
        setPermissionCheck(res.ok ? (data.data?.message || 'Authorized') : `Denied: ${data.error?.message}`);
      } catch {
        setPermissionCheck('Failed to verify');
      }
    };
    verifyCustomerPerms();
  }, [tokens]);

  const checkServiceability = async (zone: typeof SERVICEABLE_ZONES[0]) => {
    setCheckingZone(zone.name);
    try {
      const res = await fetch('/api/v1/serviceability/check', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          tenant_id: '00000000-0000-0000-0000-000000000001',
          latitude: zone.lat,
          longitude: zone.lng,
        }),
      });
      const data = await res.json();
      setServiceResult((prev) => ({ ...prev, [zone.name]: data.data?.message || 'Check complete' }));
    } catch {
      setServiceResult((prev) => ({ ...prev, [zone.name]: 'Error checking serviceability' }));
    } finally {
      setCheckingZone(null);
    }
  };

  return (
    <DashboardLayout onNavigate={onNavigate} navItems={NAV} activeNav={activeNav} onNavClick={setActiveNav}>
      <div className="p-6 space-y-6 fade-in-up">

        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-extrabold text-white">
              Hello, {user?.first_name}! 👋
            </h1>
            <p className="text-xs text-slate-400 mt-0.5">
              Track, book, and manage your shipments
            </p>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-[11px] font-mono px-3 py-1.5 rounded-xl bg-slate-900 text-slate-400 border border-slate-800 flex items-center gap-1.5">
              <ShieldCheck className="w-3.5 h-3.5 text-purple-400" />
              {permissionCheck ?? 'Verifying permissions…'}
            </span>
          </div>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'Active Shipments', value: '1', icon: <Truck className="w-4 h-4" />, color: 'blue', sub: 'In transit now' },
            { label: 'Delivered (30d)', value: '3', icon: <CheckCircle2 className="w-4 h-4" />, color: 'emerald', sub: 'Successfully received' },
            { label: 'Total Spent', value: '₹1,449', icon: <FileText className="w-4 h-4" />, color: 'purple', sub: 'All time' },
            { label: 'Saved Addresses', value: '4', icon: <MapPin className="w-4 h-4" />, color: 'amber', sub: 'Quick booking' },
          ].map((s) => (
            <div key={s.label} className="stat-card glass-card rounded-2xl p-5 border border-slate-800">
              <div className="flex items-center justify-between mb-3">
                <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">{s.label}</span>
                <div className={`w-7 h-7 rounded-lg bg-${s.color}-500/10 border border-${s.color}-500/20 flex items-center justify-center text-${s.color}-400`}>
                  {s.icon}
                </div>
              </div>
              <p className="text-2xl font-black text-white">{s.value}</p>
              <p className="text-[11px] text-slate-400 mt-1">{s.sub}</p>
            </div>
          ))}
        </div>

        {/* Main Grid: Tracking + Timeline */}
        <div className="grid grid-cols-1 xl:grid-cols-5 gap-6">

          {/* Tracking & Live Timeline */}
          <div className="xl:col-span-3 space-y-4">
            {/* Search */}
            <div className="glass-card rounded-2xl p-5 border border-slate-800">
              <h2 className="text-sm font-bold text-white mb-3">Track a Shipment</h2>
              <div className="flex gap-3">
                <div className="relative flex-1">
                  <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                    <Search className="h-4 w-4 text-slate-500" />
                  </div>
                  <input
                    type="text"
                    value={trackingInput}
                    onChange={(e) => setTrackingInput(e.target.value)}
                    placeholder="Enter tracking number (e.g. LF-8923-4412)"
                    className="w-full pl-10 pr-4 py-2.5 bg-slate-900 border border-slate-800 rounded-xl text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500/50 transition-all"
                  />
                </div>
                <button className="px-5 py-2.5 rounded-xl font-bold text-xs bg-purple-600 hover:bg-purple-500 text-white shadow-md shadow-purple-600/30 transition-all">
                  Track
                </button>
              </div>
            </div>

            {/* Live Shipment Timeline */}
            <div className="glass-card rounded-2xl p-6 border border-slate-800">
              <div className="flex items-center justify-between mb-5">
                <div>
                  <h2 className="text-sm font-bold text-white">Live Shipment</h2>
                  <p className="font-mono text-[11px] text-indigo-400 mt-0.5">LF-8923-4412</p>
                </div>
                <div className="flex items-center gap-1.5 text-[10px] font-bold text-blue-400 bg-blue-500/10 px-3 py-1.5 rounded-xl border border-blue-500/20">
                  <span className="w-1.5 h-1.5 rounded-full bg-blue-400 pulse-dot" />
                  In Transit
                </div>
              </div>

              {/* Authoritative Route & Real-Time Tracking Card */}
              {liveRoute && (
                <div className="mb-5 p-3.5 rounded-xl bg-slate-900/70 border border-sky-500/20 text-xs">
                  <div className="flex items-center justify-between mb-2">
                    <span className="flex items-center gap-1.5 font-bold text-sky-300">
                      <RouteIcon className="w-3.5 h-3.5" />
                      Live Transit Route Calculation
                    </span>
                    <span className="text-[10px] px-2 py-0.5 rounded font-mono bg-sky-950 text-sky-300 border border-sky-800">
                      {liveRoute.source === 'LIVE_OPENROUTESERVICE' ? 'OpenRouteService API' : 'Spatial Geometry Fallback'}
                    </span>
                  </div>
                  <div className="grid grid-cols-3 gap-2 py-2 border-y border-slate-800 text-[11px]">
                    <div>
                      <span className="text-slate-400">Road Distance:</span>{' '}
                      <b className="text-white">{liveRoute.distance_km} km</b>
                    </div>
                    <div>
                      <span className="text-slate-400">Driving Time:</span>{' '}
                      <b className="text-emerald-400">{liveRoute.eta_formatted}</b>
                    </div>
                    <div>
                      <span className="text-slate-400">Waypoints:</span>{' '}
                      <b className="text-slate-200">{liveRoute.geometry.length} GPS nodes</b>
                    </div>
                  </div>
                  <div className="flex items-center justify-between mt-2 text-[10px] text-slate-400">
                    <span>Carrier: SpeedEx Express</span>
                    <span className="text-emerald-400 font-semibold flex items-center gap-1">
                      <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping" />
                      Active GPS Tracking
                    </span>
                  </div>
                </div>
              )}

              {/* Timeline */}
              <div className="space-y-0">
                {TIMELINE.map((evt, i) => (
                  <div key={i} className="flex items-start gap-4">
                    {/* Timeline Node */}
                    <div className="flex flex-col items-center">
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 ${
                        evt.done && !evt.active ? 'bg-emerald-500/20 border border-emerald-500/40' :
                        evt.active ? 'bg-blue-500/20 border border-blue-500/40 ring-2 ring-blue-500/20' :
                        'bg-slate-800 border border-slate-700'
                      }`}>
                        {evt.done && !evt.active
                          ? <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                          : evt.active
                          ? <Truck className="w-4 h-4 text-blue-400" />
                          : <div className="w-2.5 h-2.5 rounded-full bg-slate-600" />
                        }
                      </div>
                      {i < TIMELINE.length - 1 && (
                        <div className={`w-0.5 h-8 ${evt.done ? 'bg-emerald-500/40' : 'bg-slate-800'}`} />
                      )}
                    </div>

                    {/* Event Info */}
                    <div className="flex-1 pb-4">
                      <div className="flex items-center justify-between">
                        <p className={`text-sm font-bold ${evt.active ? 'text-blue-300' : evt.done ? 'text-white' : 'text-slate-500'}`}>
                          {evt.label}
                        </p>
                        <span className="text-[10px] text-slate-500">{evt.time}</span>
                      </div>
                      <div className="flex items-center gap-1.5 text-xs text-slate-400 mt-0.5">
                        <MapPin className="w-3 h-3 text-slate-500" />
                        {evt.location}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Right Column: Book + Serviceability */}
          <div className="xl:col-span-2 space-y-4">
            {/* Book New Delivery */}
            <div className="glass-card rounded-2xl p-5 border border-slate-800">
              <div className="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400 mb-4">
                <PlusCircle className="w-5 h-5" />
              </div>
              <h3 className="text-base font-bold text-white mb-1">Book New Delivery</h3>
              <p className="text-xs text-slate-400 leading-relaxed mb-4">
                Generate secure QR identity, calculate dynamic pricing, and dispatch a courier to your pickup location.
              </p>
              <div className="p-3 rounded-xl bg-slate-900/60 border border-slate-800 text-[11px] text-slate-500 font-mono flex items-center gap-2">
                <Box className="w-3.5 h-3.5" />
                Coming in Phase 4: Customer &amp; Parcel Booking
              </div>
              <button className="mt-3 w-full py-2 rounded-xl text-xs font-bold bg-slate-900 text-slate-400 border border-slate-800 cursor-not-allowed opacity-70">
                Coming Soon
              </button>
            </div>

            {/* Serviceability Checker */}
            <div className="glass-card rounded-2xl p-5 border border-slate-800">
              <div className="flex items-center justify-between mb-4">
                <div>
                  <h3 className="text-sm font-bold text-white">Serviceability Checker</h3>
                  <p className="text-[11px] text-slate-400 mt-0.5">Live PostGIS 16 polygon query</p>
                </div>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                  Live API
                </span>
              </div>
              <div className="space-y-3">
                {SERVICEABLE_ZONES.map((zone) => (
                  <button
                    key={zone.name}
                    onClick={() => checkServiceability(zone)}
                    disabled={checkingZone === zone.name}
                    className="w-full p-3.5 bg-slate-900/60 hover:bg-slate-800/60 border border-slate-800 hover:border-slate-700 rounded-xl text-left transition-all disabled:opacity-60 disabled:cursor-wait"
                  >
                    <div className="flex items-center justify-between">
                      <div className="font-semibold text-xs text-slate-200">{zone.name}</div>
                      <ArrowRight className="w-3.5 h-3.5 text-slate-500" />
                    </div>
                    <div className="text-[10px] text-slate-500 mt-0.5 font-mono">
                      {zone.lat}°N, {zone.lng}°E
                    </div>
                    {serviceResult[zone.name] && (
                      <div className="mt-2 text-[11px] text-emerald-400 font-semibold">
                        ✓ {serviceResult[zone.name]}
                      </div>
                    )}
                    {checkingZone === zone.name && (
                      <div className="mt-2 text-[11px] text-indigo-400">Querying PostGIS…</div>
                    )}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Order History */}
        <div className="glass-card rounded-2xl border border-slate-800 overflow-hidden">
          <div className="px-6 py-4 border-b border-slate-800/80 flex items-center justify-between">
            <h2 className="text-sm font-bold text-white">Order History</h2>
            <button className="text-xs text-purple-400 hover:text-purple-300 transition-colors font-semibold">View All →</button>
          </div>
          <div className="divide-y divide-slate-800/60">
            {RECENT_ORDERS.map((order) => (
              <div key={order.id} className="px-6 py-4 flex items-center justify-between hover:bg-slate-800/20 transition-colors">
                <div className="flex items-center gap-4">
                  <div className="w-9 h-9 rounded-xl bg-slate-900 border border-slate-800 flex items-center justify-center text-slate-500">
                    <Package className="w-4 h-4" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-[11px] font-bold text-indigo-400">{order.id}</span>
                      <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full border ${STATUS_STYLES[order.status] || 'bg-slate-800 text-slate-400 border-slate-700'}`}>
                        {order.status}
                      </span>
                    </div>
                    <p className="text-xs text-slate-400 mt-0.5">{order.destination} • {order.items} item{order.items > 1 ? 's' : ''}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-bold text-white">{order.amount}</p>
                  <p className="text-[11px] text-slate-500">{order.date}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Rating Banner */}
        <div className="glass-card rounded-2xl p-5 border border-amber-500/20 bg-amber-500/5 flex items-center justify-between">
          <div>
            <p className="text-sm font-bold text-white">Rate your last delivery</p>
            <p className="text-xs text-slate-400 mt-0.5">LF-8801-1123 delivered to Chennai</p>
          </div>
          <div className="flex items-center gap-1">
            {[1, 2, 3, 4, 5].map((star) => (
              <Star key={star} className="w-5 h-5 text-amber-400 hover:scale-125 transition-transform cursor-pointer" fill="#f59e0b" />
            ))}
          </div>
        </div>

      </div>
    </DashboardLayout>
  );
};
