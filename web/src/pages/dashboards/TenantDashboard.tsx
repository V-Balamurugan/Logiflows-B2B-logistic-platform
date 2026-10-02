import React, { useEffect, useState } from 'react';
import { Building2, Package, Truck, Users, MapPin, LogOut, ShieldCheck, Zap } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { Branch } from '../../types/tenancy';
import { GeospatialMap } from '../../components/GeospatialMap';

interface TenantDashboardProps {
  onNavigate: (path: string) => void;
}

export const TenantDashboard: React.FC<TenantDashboardProps> = ({ onNavigate }) => {
  const { user, logout } = useAuth();
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  const tenantId = user?.tenant_id || '00000000-0000-0000-0000-000000000001';

  useEffect(() => {
    const fetchBranches = async () => {
      try {
        const token = localStorage.getItem('access_token');
        const res = await fetch(`/api/v1/tenants/${tenantId}/branches`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        const json = await res.json();
        if (res.ok) {
          setBranches(json.data || []);
        }
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
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      {/* Header */}
      <header className="glass-panel border-b border-slate-800 px-6 py-4 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400">
              <Building2 className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h1 className="font-extrabold text-lg text-white">Logistics Company Headquarters</h1>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-blue-500/20 text-blue-300 border border-blue-500/30">
                  TENANT
                </span>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 flex items-center gap-1">
                  <Zap className="w-3 h-3" /> ENTERPRISE TIER
                </span>
              </div>
              <p className="text-xs text-slate-400">{user?.email} • Tenant ID: {tenantId}</p>
            </div>
          </div>

          <div className="flex items-center space-x-3">
            <button
              onClick={() => onNavigate('/')}
              className="px-3.5 py-1.5 rounded-xl text-xs font-semibold bg-slate-900 hover:bg-slate-800 text-slate-300 border border-slate-800 transition-colors"
            >
              System Pulse
            </button>
            <button
              onClick={async () => {
                await logout();
                onNavigate('/login');
              }}
              className="flex items-center space-x-1.5 px-3.5 py-1.5 rounded-xl text-xs font-semibold bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/20 transition-colors"
            >
              <LogOut className="w-3.5 h-3.5" />
              <span>Sign Out</span>
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-8 space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Active Hubs & Branches</span>
              <MapPin className="w-4 h-4 text-blue-400" />
            </div>
            <p className="text-2xl font-black text-white">{branches.length} Nodes</p>
            <p className="text-[11px] text-blue-400 mt-1">Geospatial service zones active</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Daily Network Capacity</span>
              <Package className="w-4 h-4 text-purple-400" />
            </div>
            <p className="text-2xl font-black text-white">{loading ? '...' : totalCapacity.toLocaleString()}</p>
            <p className="text-[11px] text-purple-400 mt-1">Parcels/day threshold</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Assigned Fleet</span>
              <Truck className="w-4 h-4 text-indigo-400" />
            </div>
            <p className="text-2xl font-black text-white">24 Vehicles</p>
            <p className="text-[11px] text-slate-400 mt-1">Vans, trucks & two-wheelers</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Branch Staff</span>
              <Users className="w-4 h-4 text-emerald-400" />
            </div>
            <p className="text-2xl font-black text-white">38 Staff</p>
            <p className="text-[11px] text-emerald-400 mt-1">Active mobile heartbeats</p>
          </div>
        </div>

        {/* Live Geospatial Map Section */}
        <div className="glass-card rounded-2xl p-6 border border-slate-800 space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-base font-bold text-white">Enterprise Geospatial Serviceability Network</h2>
              <p className="text-xs text-slate-400">PostGIS polygon boundary zones and dynamic customer point-in-polygon verification</p>
            </div>
            <div className="flex items-center gap-2 text-xs font-semibold text-emerald-400 bg-emerald-500/10 px-3 py-1.5 rounded-xl border border-emerald-500/20">
              <ShieldCheck className="w-4 h-4" />
              <span>Strict Multi-Tenant Isolation</span>
            </div>
          </div>

          <GeospatialMap tenantId={tenantId} branches={branches} />
        </div>
      </main>
    </div>
  );
};
