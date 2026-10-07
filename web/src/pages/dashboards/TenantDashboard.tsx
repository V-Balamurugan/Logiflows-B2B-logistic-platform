import React, { useEffect, useState } from 'react';
import {
  Package, Truck, Users, MapPin, Zap, TrendingUp,
  BarChart3, ShieldCheck, ArrowUpRight, Clock, Navigation
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { Branch } from '../../types/tenancy';
import { GeospatialMap } from '../../components/GeospatialMap';
import { FleetManagementView } from '../../components/FleetManagementView';
import { DashboardLayout } from '../../components/DashboardLayout';

interface TenantDashboardProps {
  onNavigate: (path: string) => void;
}

const NAV = [
  { icon: <BarChart3 className="w-4 h-4" />, label: 'Overview' },
  { icon: <MapPin className="w-4 h-4" />, label: 'Network Map' },
  { icon: <Package className="w-4 h-4" />, label: 'Parcels', badge: '18.4k' },
  { icon: <Truck className="w-4 h-4" />, label: 'Fleet', badge: 24 },
  { icon: <Users className="w-4 h-4" />, label: 'Staff' },
  { icon: <TrendingUp className="w-4 h-4" />, label: 'Analytics' },
];

const MONTHLY = [42, 58, 51, 72, 68, 91, 88, 103, 97, 118, 112, 130];
const MONTHS = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];

const LIVE_DELIVERIES = [
  { id: 'LF-8921-4412', courier: 'Ravi K.', status: 'In Transit', location: 'MG Road', eta: '14 min' },
  { id: 'LF-8922-1190', courier: 'Suresh P.', status: 'Out for Delivery', location: 'Koramangala', eta: '28 min' },
  { id: 'LF-8923-6654', courier: 'Anita M.', status: 'Delivered', location: 'Whitefield', eta: '—' },
  { id: 'LF-8924-3312', courier: 'Kiran D.', status: 'In Transit', location: 'Hebbal', eta: '42 min' },
];

export const TenantDashboard: React.FC<TenantDashboardProps> = ({ onNavigate }) => {
  const { user } = useAuth();
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeNav, setActiveNav] = useState('Overview');

  const tenantId = user?.tenant_id || '00000000-0000-0000-0000-000000000001';
  const maxVol = Math.max(...MONTHLY);

  useEffect(() => {
    const fetchBranches = async () => {
      try {
        const token = localStorage.getItem('access_token');
        const res = await fetch(`/api/v1/tenants/${tenantId}/branches`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        const json = await res.json();
        if (res.ok) setBranches(json.data || []);
      } catch (err) {
        console.error('Failed to load tenant branches', err);
      } finally {
        setLoading(false);
      }
    };
    fetchBranches();
  }, [tenantId]);

  const totalCapacity = branches.reduce((acc, b) => acc + (b.daily_capacity || 0), 0);

  return (
    <DashboardLayout onNavigate={onNavigate} navItems={NAV} activeNav={activeNav} onNavClick={setActiveNav}>
      <div className="p-6 space-y-6 fade-in-up">

        {/* Page Title */}
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-extrabold text-white">Logistics Company HQ</h1>
              <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                <Zap className="w-3 h-3" /> ENTERPRISE
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-0.5">Tenant ID: <span className="font-mono text-slate-300">{tenantId}</span></p>
          </div>
          <button
            onClick={() => setActiveNav('Network Map')}
            className="flex items-center gap-1.5 px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold shadow-lg shadow-blue-600/20 transition-all"
          >
            <MapPin className="w-3.5 h-3.5" /> Live Network Map
            <ArrowUpRight className="w-3.5 h-3.5" />
          </button>
        </div>

        {/* Stats Row */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'Active Hubs', value: loading ? '…' : `${branches.length}`, sub: 'Geospatial zones', icon: <MapPin className="w-4 h-4" />, color: 'blue' },
            { label: 'Daily Capacity', value: loading ? '…' : totalCapacity.toLocaleString(), sub: 'Parcels/day', icon: <Package className="w-4 h-4" />, color: 'purple' },
            { label: 'Fleet Size', value: '24', sub: 'Vans, trucks & 2W', icon: <Truck className="w-4 h-4" />, color: 'indigo' },
            { label: 'Active Staff', value: '38', sub: 'Mobile heartbeats', icon: <Users className="w-4 h-4" />, color: 'emerald' },
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

        {/* Main Content: Chart + Live Deliveries */}
        <div className="grid grid-cols-1 xl:grid-cols-5 gap-6">

          {/* Monthly Volume Chart */}
          <div className="xl:col-span-3 glass-card rounded-2xl p-6 border border-slate-800">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h2 className="text-sm font-bold text-white">Monthly Parcel Volume</h2>
                <p className="text-[11px] text-slate-400 mt-0.5">Across all branches</p>
              </div>
              <div className="flex items-center gap-1.5 text-xs font-bold text-emerald-400">
                <TrendingUp className="w-3.5 h-3.5" /> +16% YTD
              </div>
            </div>
            <div className="flex items-end gap-2 h-32">
              {MONTHLY.map((v, i) => (
                <div key={i} className="flex-1 flex flex-col items-center gap-1">
                  <div
                    className="w-full rounded-t-lg chart-bar"
                    style={{
                      height: `${(v / maxVol) * 100}%`,
                      background: i === 9
                        ? 'linear-gradient(to top, #3b82f6, #60a5fa)'
                        : 'linear-gradient(to top, #1e3a5f, #2563eb40)',
                      animationDelay: `${i * 60}ms`,
                    }}
                  />
                  <span className="text-[9px] text-slate-500">{MONTHS[i]}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Live Deliveries */}
          <div className="xl:col-span-2 glass-card rounded-2xl border border-slate-800 overflow-hidden">
            <div className="px-5 py-4 border-b border-slate-800/80 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className="w-2 h-2 rounded-full bg-emerald-400 pulse-dot" />
                <h2 className="text-sm font-bold text-white">Live Deliveries</h2>
              </div>
              <span className="text-[10px] text-slate-500">Today</span>
            </div>
            <div className="divide-y divide-slate-800/60">
              {LIVE_DELIVERIES.map((d) => (
                <div key={d.id} className="px-5 py-3.5 hover:bg-slate-800/20 transition-colors">
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-mono text-[10px] text-indigo-400 font-bold">{d.id}</span>
                    <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                      d.status === 'Delivered' ? 'bg-emerald-500/20 text-emerald-400' :
                      d.status === 'Out for Delivery' ? 'bg-blue-500/20 text-blue-400' :
                      'bg-amber-500/20 text-amber-400'
                    }`}>{d.status}</span>
                  </div>
                  <div className="flex items-center justify-between text-xs text-slate-400">
                    <div className="flex items-center gap-1.5">
                      <Navigation className="w-3 h-3 text-slate-500" />
                      <span>{d.courier} • {d.location}</span>
                    </div>
                    {d.eta !== '—' && (
                      <div className="flex items-center gap-1 text-amber-400">
                        <Clock className="w-3 h-3" />
                        <span className="font-semibold">{d.eta}</span>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Geospatial Map */}
        {(activeNav === 'Overview' || activeNav === 'Network Map') && (
          <div className="glass-card rounded-2xl p-6 border border-slate-800 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-sm font-bold text-white">Enterprise Geospatial Serviceability Network</h2>
                <p className="text-xs text-slate-400">PostGIS polygon boundary zones & dynamic customer point-in-polygon verification</p>
              </div>
              <div className="flex items-center gap-1.5 text-xs font-semibold text-emerald-400 bg-emerald-500/10 px-3 py-1.5 rounded-xl border border-emerald-500/20">
                <ShieldCheck className="w-4 h-4" />
                <span>Multi-Tenant Isolated</span>
              </div>
            </div>
            <GeospatialMap tenantId={tenantId} branches={branches} />
          </div>
        )}

        {/* Fleet & Telematics Management */}
        {(activeNav === 'Fleet' || activeNav === 'Overview') && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-base font-bold text-white">Fleet & Vehicle Telematics Operations</h2>
                <p className="text-xs text-slate-400">Live GPS tracking, road speed, energy levels, and maintenance telemetry</p>
              </div>
            </div>
            <FleetManagementView
              token={localStorage.getItem('access_token') || ''}
              tenantId={tenantId}
              branches={branches}
            />
          </div>
        )}

      </div>
    </DashboardLayout>
  );
};
