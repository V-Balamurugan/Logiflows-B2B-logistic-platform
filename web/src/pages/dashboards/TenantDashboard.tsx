import React from 'react';
import { Building2, Package, Truck, Users, MapPin, LogOut } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

interface TenantDashboardProps {
  onNavigate: (path: string) => void;
}

export const TenantDashboard: React.FC<TenantDashboardProps> = ({ onNavigate }) => {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
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
              </div>
              <p className="text-xs text-slate-400">{user?.email} • Tenant ID: {user?.tenant_id ?? 'Speedy Express'}</p>
            </div>
          </div>

          <div className="flex items-center space-x-3">
            <button
              onClick={() => onNavigate('/')}
              className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-slate-900 hover:bg-slate-800 text-slate-300 border border-slate-800 transition-colors"
            >
              System Pulse
            </button>
            <button
              onClick={async () => {
                await logout();
                onNavigate('/login');
              }}
              className="flex items-center space-x-1.5 px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/20 transition-colors"
            >
              <LogOut className="w-3.5 h-3.5" />
              <span>Sign Out</span>
            </button>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-8 space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Active Branches</span>
              <MapPin className="w-4 h-4 text-blue-400" />
            </div>
            <p className="text-2xl font-black text-white">6 Hubs</p>
            <p className="text-[11px] text-slate-400 mt-1">Geospatial service zones active</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Assigned Fleet</span>
              <Truck className="w-4 h-4 text-indigo-400" />
            </div>
            <p className="text-2xl font-black text-white">24 Vehicles</p>
            <p className="text-[11px] text-slate-400 mt-1">Vans, trucks, and two-wheelers</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Active Couriers</span>
              <Users className="w-4 h-4 text-emerald-400" />
            </div>
            <p className="text-2xl font-black text-white">38 Staff</p>
            <p className="text-[11px] text-emerald-400 mt-1">Mobile GPS tracking ready</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Live Parcels</span>
              <Package className="w-4 h-4 text-purple-400" />
            </div>
            <p className="text-2xl font-black text-white">342</p>
            <p className="text-[11px] text-purple-400 mt-1">Under custody state progression</p>
          </div>
        </div>

        <div className="glass-card rounded-2xl p-6 border border-slate-800">
          <h2 className="text-base font-bold text-white mb-2">Tenant Scope Isolation</h2>
          <p className="text-sm text-slate-400 leading-relaxed mb-4">
            All branches, vehicles, personnel, parcels, and customers in this portal are automatically constrained to Tenant ID: <code className="text-indigo-300 font-mono">{user?.tenant_id ?? 'speedy-express'}</code>. Server-side middleware prevents cross-tenant data leaks.
          </p>
          <div className="flex items-center space-x-2 text-xs text-slate-500 font-mono bg-slate-900/80 p-3 rounded-xl border border-slate-800">
            <span>Enforced by:</span>
            <span className="text-emerald-400">backend/internal/middleware/tenant.go</span>
          </div>
        </div>
      </main>
    </div>
  );
};
