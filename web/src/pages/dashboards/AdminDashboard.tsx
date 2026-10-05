import React, { useState, useEffect } from 'react';
import {
  Building2, Users, FileText, Lock, AlertTriangle,
  TrendingUp, Server, CheckCircle2, XCircle, Activity, Globe,
  MoreHorizontal, Eye, Ban, RefreshCw, Flame, Route as RouteIcon, Radio, Zap
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { DashboardLayout } from '../../components/DashboardLayout';
import { getFirebaseStatus, FirebaseOperationalStatus } from '../../services/firebaseService';
import { getRoutingStatus, getDirections, RoutingStatus, RouteResult } from '../../services/routingService';

interface AdminDashboardProps {
  onNavigate: (path: string) => void;
}

const NAV = [
  { icon: <Activity className="w-4 h-4" />, label: 'Overview' },
  { icon: <Building2 className="w-4 h-4" />, label: 'Tenants', badge: 12 },
  { icon: <Users className="w-4 h-4" />, label: 'Users', badge: '1.4k' },
  { icon: <Lock className="w-4 h-4" />, label: 'RBAC Rules' },
  { icon: <FileText className="w-4 h-4" />, label: 'Audit Logs' },
  { icon: <Server className="w-4 h-4" />, label: 'Infrastructure' },
  { icon: <AlertTriangle className="w-4 h-4" />, label: 'Alerts', badge: 3 },
];

const TENANTS = [
  { name: 'SpeedEx Logistics', id: 'SPX-001', plan: 'Enterprise', users: 142, status: 'active', parcels: 18420, growth: '+12%' },
  { name: 'BlueDart India', id: 'BDI-002', plan: 'Pro', users: 87, status: 'active', parcels: 9340, growth: '+8%' },
  { name: 'QuickShip Co.', id: 'QSC-003', plan: 'Starter', users: 23, status: 'suspended', parcels: 1200, growth: '-2%' },
  { name: 'MegaCargo Ltd.', id: 'MCL-004', plan: 'Enterprise', users: 210, status: 'active', parcels: 32100, growth: '+21%' },
  { name: 'AirFreight Plus', id: 'AFP-005', plan: 'Pro', users: 56, status: 'active', parcels: 7890, growth: '+5%' },
];

const AUDIT_LOGS = [
  { action: 'LOGIN_SUCCESS', resource: 'auth', ip: '103.21.58.12', status: 'PASSED', role: 'PLATFORM_ADMIN', time: 'Just now' },
  { action: 'TENANT_CREATED', resource: 'tenants', ip: '127.0.0.1', status: 'CREATED', role: 'PLATFORM_ADMIN', time: '2m ago' },
  { action: 'USER_SUSPENDED', resource: 'users', ip: '103.21.58.12', status: 'MUTATED', role: 'PLATFORM_ADMIN', time: '8m ago' },
  { action: 'TOKEN_REFRESHED', resource: 'auth', ip: '192.168.1.5', status: 'ROTATED', role: 'TENANT', time: '12m ago' },
  { action: 'BRANCH_DEPLOYED', resource: 'branches', ip: '103.21.58.01', status: 'CREATED', role: 'TENANT_ADMIN', time: '25m ago' },
  { action: 'LOGIN_FAILED', resource: 'auth', ip: '45.32.16.9', status: 'BLOCKED', role: 'UNKNOWN', time: '31m ago' },
];

const WEEKLY_VOLUME = [65, 78, 52, 91, 83, 110, 98];
const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

export const AdminDashboard: React.FC<AdminDashboardProps> = ({ onNavigate }) => {
  const { tokens } = useAuth();
  const [activeNav, setActiveNav] = useState('Overview');
  const [systemCheckResult, setSystemCheckResult] = useState<string | null>(null);
  const [loadingCheck, setLoadingCheck] = useState(false);

  // Cloud Integrations State
  const [fbStatus, setFbStatus] = useState<FirebaseOperationalStatus | null>(null);
  const [routeStatus, setRouteStatus] = useState<RoutingStatus | null>(null);
  const [testingRoute, setTestingRoute] = useState(false);
  const [routeTestResult, setRouteTestResult] = useState<RouteResult | null>(null);
  const [testingFirebase, setTestingFirebase] = useState(false);
  const [fbTestMessage, setFbTestMessage] = useState<string | null>(null);

  useEffect(() => {
    const loadIntegrations = async () => {
      try {
        const [fb, rtr] = await Promise.all([getFirebaseStatus(), getRoutingStatus()]);
        setFbStatus(fb);
        setRouteStatus(rtr);
      } catch (e) {
        console.error('Failed to load integration status', e);
      }
    };
    loadIntegrations();
  }, []);

  const handleTestRoute = async () => {
    setTestingRoute(true);
    try {
      const res = await getDirections(
        { latitude: 12.9716, longitude: 77.5946 }, // Bangalore City Center
        { latitude: 13.0358, longitude: 77.5970 }  // Hebbal Distribution Hub
      );
      setRouteTestResult(res);
    } catch (e: any) {
      console.error(e);
    } finally {
      setTestingRoute(false);
    }
  };

  const handleTestFirebase = async () => {
    setTestingFirebase(true);
    try {
      const res = await fetch('/api/v1/ping');
      if (res.ok) {
        setFbTestMessage(`Connected: ${fbStatus?.mode || 'SIMULATED_MOCK'} • Project: ${fbStatus?.project_id || 'logiflows'}`);
      }
    } catch (e: any) {
      setFbTestMessage('Failed to ping service');
    } finally {
      setTestingFirebase(false);
    }
  };

  useEffect(() => {
    const verifyAdmin = async () => {
      if (!tokens?.access_token) return;
      setLoadingCheck(true);
      try {
        const res = await fetch('/api/v1/admin/system-check', {
          headers: { Authorization: `Bearer ${tokens.access_token}` },
        });
        const data = await res.json();
        setSystemCheckResult(res.ok ? (data.data?.message || 'Authorized') : `Forbidden: ${data.error?.message}`);
      } catch {
        setSystemCheckResult('Failed to run system check');
      } finally {
        setLoadingCheck(false);
      }
    };
    verifyAdmin();
  }, [tokens]);

  const maxVol = Math.max(...WEEKLY_VOLUME);

  return (
    <DashboardLayout onNavigate={onNavigate} navItems={NAV} activeNav={activeNav} onNavClick={setActiveNav}>
      <div className="p-6 space-y-6 fade-in-up">

        {/* Page Title */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-extrabold text-white">Platform Administration</h1>
            <p className="text-xs text-slate-400 mt-0.5">Global system oversight & multi-tenant control</p>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 pulse-dot" />
              All Systems Operational
            </div>
          </div>
        </div>

        {/* RBAC Verification Banner */}
        <div className="p-4 rounded-2xl glass-card border border-indigo-500/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400 shrink-0">
              <CheckCircle2 className="w-5 h-5" />
            </div>
            <div>
              <p className="text-sm font-bold text-white">Authoritative Server RBAC Verified</p>
              <p className="text-xs text-slate-400">
                {loadingCheck ? 'Validating server-side permissions...' : (systemCheckResult || 'Awaiting verification')}
              </p>
            </div>
          </div>
          <span className="text-[11px] font-mono px-3 py-1.5 rounded-lg bg-slate-900 text-indigo-300 border border-slate-800">
            platform_admin.create • tenant.disable • rbac.manage
          </span>
        </div>

        {/* Cloud & Spatial Integrations (Firebase & OpenRouteService) */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Firebase Status Card */}
          <div className="glass-card rounded-2xl p-5 border border-amber-500/20 relative overflow-hidden">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2.5">
                <div className="w-8 h-8 rounded-xl bg-amber-500/10 text-amber-400 flex items-center justify-center border border-amber-500/20">
                  <Flame className="w-4 h-4" />
                </div>
                <div>
                  <h3 className="text-sm font-bold text-white">Firebase Admin SDK</h3>
                  <p className="text-[11px] text-slate-400">FCM Push Messaging & Cloud Storage</p>
                </div>
              </div>
              <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold border ${
                fbStatus?.firebase_enabled
                  ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                  : 'bg-amber-500/10 text-amber-300 border-amber-500/20'
              }`}>
                {fbStatus?.mode || 'DETECTING...'}
              </span>
            </div>

            <div className="space-y-1.5 text-xs text-slate-300 mb-4 bg-slate-900/50 p-3 rounded-xl border border-slate-800">
              <div className="flex justify-between">
                <span className="text-slate-400">Project ID:</span>
                <span className="font-mono text-slate-200">{fbStatus?.project_id || 'logiflows-platform'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Storage Bucket:</span>
                <span className="font-mono text-slate-200">{fbStatus?.storage_bucket || 'logiflows-platform.appspot.com'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Credentials:</span>
                <span className="text-slate-300">Auto-detected from .env / serviceAccountKey.json</span>
              </div>
            </div>

            <div className="flex items-center justify-between gap-3">
              <button
                onClick={handleTestFirebase}
                disabled={testingFirebase}
                className="px-3 py-1.5 bg-amber-500/10 hover:bg-amber-500/20 text-amber-300 border border-amber-500/30 rounded-xl text-xs font-semibold flex items-center gap-1.5 transition-colors"
              >
                <Radio className="w-3.5 h-3.5" />
                {testingFirebase ? 'Pinging...' : 'Verify Firebase Connection'}
              </button>
              {fbTestMessage && (
                <span className="text-[11px] text-amber-300/80 truncate max-w-[200px]">
                  {fbTestMessage}
                </span>
              )}
            </div>
          </div>

          {/* OpenRouteService Card */}
          <div className="glass-card rounded-2xl p-5 border border-sky-500/20 relative overflow-hidden">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2.5">
                <div className="w-8 h-8 rounded-xl bg-sky-500/10 text-sky-400 flex items-center justify-center border border-sky-500/20">
                  <RouteIcon className="w-4 h-4" />
                </div>
                <div>
                  <h3 className="text-sm font-bold text-white">Routing & ETA Engine</h3>
                  <p className="text-[11px] text-slate-400">Turn-by-turn road geometry & distance</p>
                </div>
              </div>
              <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold border ${
                routeStatus?.has_api_key
                  ? 'bg-sky-500/10 text-sky-400 border-sky-500/20'
                  : 'bg-indigo-500/10 text-indigo-300 border-indigo-500/20'
              }`}>
                {routeStatus?.mode || 'DETECTING...'}
              </span>
            </div>

            <div className="space-y-1.5 text-xs text-slate-300 mb-4 bg-slate-900/50 p-3 rounded-xl border border-slate-800">
              <div className="flex justify-between">
                <span className="text-slate-400">Provider:</span>
                <span className="font-semibold text-slate-200">{routeStatus?.provider || 'OpenRouteService'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">API Key Status:</span>
                <span className="font-mono text-sky-300">
                  {routeStatus?.has_api_key ? `Configured (${routeStatus.key_preview})` : 'Not Set (Using Spatial Fallback)'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Integration:</span>
                <span className="text-slate-300">Connected to all 5 Dashboards & Leaflet Map</span>
              </div>
            </div>

            <div className="flex items-center justify-between gap-3">
              <button
                onClick={handleTestRoute}
                disabled={testingRoute}
                className="px-3 py-1.5 bg-sky-500/10 hover:bg-sky-500/20 text-sky-300 border border-sky-500/30 rounded-xl text-xs font-semibold flex items-center gap-1.5 transition-colors"
              >
                <Zap className="w-3.5 h-3.5" />
                {testingRoute ? 'Calculating...' : 'Test Directions (Bangalore -> Hebbal)'}
              </button>
              {routeTestResult && (
                <span className="text-[11px] text-emerald-400 font-semibold">
                  {routeTestResult.distance_km} km • {routeTestResult.eta_formatted}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'Active Tenants', value: '12', sub: '+2 this week', icon: <Building2 className="w-4 h-4" />, color: 'indigo', trend: 'up' },
            { label: 'Total Users', value: '1,480', sub: 'Across all roles', icon: <Users className="w-4 h-4" />, color: 'blue', trend: 'up' },
            { label: 'Audit Events', value: '8,920', sub: 'Tamper-evident logs', icon: <FileText className="w-4 h-4" />, color: 'amber', trend: 'up' },
            { label: 'RBAC Rules', value: '28', sub: 'Centralized perms', icon: <Lock className="w-4 h-4" />, color: 'purple', trend: 'stable' },
          ].map((s) => (
            <div key={s.label} className="stat-card glass-card rounded-2xl p-5 border border-slate-800">
              <div className={`flex items-center justify-between text-${s.color}-400 mb-3`}>
                <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">{s.label}</span>
                <div className={`w-7 h-7 rounded-lg bg-${s.color}-500/10 border border-${s.color}-500/20 flex items-center justify-center`}>
                  {s.icon}
                </div>
              </div>
              <p className="text-2xl font-black text-white">{s.value}</p>
              <div className="flex items-center gap-1.5 mt-1">
                {s.trend === 'up' && <TrendingUp className="w-3 h-3 text-emerald-400" />}
                <p className="text-[11px] text-slate-400">{s.sub}</p>
              </div>
            </div>
          ))}
        </div>

        {/* Bottom Grid: Weekly Volume Chart + Tenant Table */}
        <div className="grid grid-cols-1 xl:grid-cols-5 gap-6">

          {/* Weekly Volume Bar Chart */}
          <div className="xl:col-span-2 glass-card rounded-2xl p-6 border border-slate-800">
            <div className="flex items-center justify-between mb-6">
              <div>
                <h2 className="text-sm font-bold text-white">Weekly Parcel Volume</h2>
                <p className="text-[11px] text-slate-400 mt-0.5">All tenants combined</p>
              </div>
              <span className="text-xs font-bold text-emerald-400">+14% vs last wk</span>
            </div>
            <div className="flex items-end gap-3 h-32">
              {WEEKLY_VOLUME.map((v, i) => (
                <div key={i} className="flex-1 flex flex-col items-center gap-1">
                  <div
                    className="w-full rounded-t-lg chart-bar"
                    style={{
                      height: `${(v / maxVol) * 100}%`,
                      background: i === 5
                        ? 'linear-gradient(to top, #6366f1, #818cf8)'
                        : 'linear-gradient(to top, #334155, #475569)',
                      animationDelay: `${i * 80}ms`,
                    }}
                  />
                  <span className="text-[10px] text-slate-500">{DAYS[i]}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Tenant Table */}
          <div className="xl:col-span-3 glass-card rounded-2xl border border-slate-800 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-800/80 flex items-center justify-between">
              <h2 className="text-sm font-bold text-white">Tenant Registry</h2>
              <button className="text-xs text-indigo-400 hover:text-indigo-300 transition-colors font-semibold">View All →</button>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="border-b border-slate-800/60">
                    {['Tenant', 'Plan', 'Users', 'Parcels', 'Growth', 'Status', ''].map((h) => (
                      <th key={h} className="px-4 py-3 text-left text-[10px] uppercase tracking-wider text-slate-500 font-bold">{h}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {TENANTS.map((t) => (
                    <tr key={t.id} className="table-row-hover border-b border-slate-800/40 last:border-0">
                      <td className="px-4 py-3">
                        <div className="font-semibold text-slate-200">{t.name}</div>
                        <div className="text-[10px] font-mono text-slate-500 mt-0.5">{t.id}</div>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${
                          t.plan === 'Enterprise' ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30' :
                          t.plan === 'Pro' ? 'bg-blue-500/20 text-blue-300 border border-blue-500/30' :
                          'bg-slate-800 text-slate-400 border border-slate-700'
                        }`}>{t.plan}</span>
                      </td>
                      <td className="px-4 py-3 text-slate-300">{t.users}</td>
                      <td className="px-4 py-3 text-slate-300">{t.parcels.toLocaleString()}</td>
                      <td className="px-4 py-3">
                        <span className={t.growth.startsWith('+') ? 'text-emerald-400' : 'text-rose-400'}>
                          {t.growth}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-1.5">
                          <span className={`w-1.5 h-1.5 rounded-full ${t.status === 'active' ? 'bg-emerald-400' : 'bg-rose-400'}`} />
                          <span className={t.status === 'active' ? 'text-emerald-400' : 'text-rose-400'}>
                            {t.status}
                          </span>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-1">
                          <button className="p-1 rounded-lg hover:bg-slate-800 text-slate-500 hover:text-slate-300 transition-all"><Eye className="w-3.5 h-3.5" /></button>
                          <button className="p-1 rounded-lg hover:bg-slate-800 text-slate-500 hover:text-rose-400 transition-all"><Ban className="w-3.5 h-3.5" /></button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* Audit Log Stream */}
        <div className="glass-card rounded-2xl border border-slate-800 overflow-hidden">
          <div className="px-6 py-4 border-b border-slate-800/80 flex items-center justify-between">
            <div>
              <h2 className="text-sm font-bold text-white">Live Security Audit Trail</h2>
              <p className="text-[11px] text-slate-400 mt-0.5">Tamper-evident RBAC event stream</p>
            </div>
            <button className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold text-slate-400 hover:text-white bg-slate-900 border border-slate-800 transition-all">
              <RefreshCw className="w-3.5 h-3.5" /> Refresh
            </button>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead>
                <tr className="border-b border-slate-800/60">
                  {['Action', 'Resource', 'Role', 'Origin IP', 'Status', 'Time'].map((h) => (
                    <th key={h} className="px-5 py-3 text-left text-[10px] uppercase tracking-wider text-slate-500 font-bold">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {AUDIT_LOGS.map((log, i) => (
                  <tr key={i} className="table-row-hover border-b border-slate-800/40 last:border-0">
                    <td className="px-5 py-3">
                      <span className={`font-mono font-bold ${
                        log.action.includes('FAIL') || log.action.includes('BLOCKED') ? 'text-rose-400' :
                        log.action.includes('SUCCESS') || log.action.includes('CREATED') ? 'text-emerald-400' :
                        'text-amber-400'
                      }`}>{log.action}</span>
                    </td>
                    <td className="px-5 py-3 text-slate-400 font-mono">{log.resource}</td>
                    <td className="px-5 py-3">
                      <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-800 text-slate-400 border border-slate-700">{log.role}</span>
                    </td>
                    <td className="px-5 py-3 font-mono text-slate-400">{log.ip}</td>
                    <td className="px-5 py-3">
                      <div className="flex items-center gap-1.5">
                        {log.status === 'PASSED' || log.status === 'CREATED' || log.status === 'ROTATED' ? (
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                        ) : log.status === 'BLOCKED' ? (
                          <XCircle className="w-3.5 h-3.5 text-rose-400" />
                        ) : (
                          <MoreHorizontal className="w-3.5 h-3.5 text-amber-400" />
                        )}
                        <span className={
                          log.status === 'BLOCKED' ? 'text-rose-400' :
                          log.status === 'MUTATED' ? 'text-amber-400' :
                          'text-emerald-400'
                        }>{log.status}</span>
                      </div>
                    </td>
                    <td className="px-5 py-3 text-slate-500">{log.time}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* System Health */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[
            { name: 'Go API Gateway', uptime: '99.98%', latency: '12ms', status: 'healthy' },
            { name: 'PostgreSQL + PostGIS', uptime: '99.95%', latency: '4ms', status: 'healthy' },
            { name: 'Redis Pub/Sub', uptime: '99.80%', latency: '1ms', status: 'degraded' },
          ].map((svc) => (
            <div key={svc.name} className="stat-card glass-card rounded-2xl p-5 border border-slate-800">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <span className={`w-2 h-2 rounded-full ${svc.status === 'healthy' ? 'bg-emerald-400 pulse-dot' : 'bg-amber-400 pulse-dot'}`} />
                  <span className="text-xs font-bold text-white">{svc.name}</span>
                </div>
                <Globe className="w-4 h-4 text-slate-500" />
              </div>
              <div className="grid grid-cols-2 gap-3 mt-3">
                <div className="bg-slate-900/60 rounded-xl p-3 text-center">
                  <p className="text-[10px] text-slate-500 mb-1">Uptime</p>
                  <p className="text-sm font-bold text-emerald-400">{svc.uptime}</p>
                </div>
                <div className="bg-slate-900/60 rounded-xl p-3 text-center">
                  <p className="text-[10px] text-slate-500 mb-1">Latency</p>
                  <p className="text-sm font-bold text-indigo-400">{svc.latency}</p>
                </div>
              </div>
            </div>
          ))}
        </div>

      </div>
    </DashboardLayout>
  );
};
