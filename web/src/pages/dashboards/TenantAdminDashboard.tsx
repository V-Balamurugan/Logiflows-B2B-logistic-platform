import React from 'react';
import { Network, Package, Users, Truck, ArrowUpRight, LogOut } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

interface TenantAdminDashboardProps {
  onNavigate: (path: string) => void;
}

export const TenantAdminDashboard: React.FC<TenantAdminDashboardProps> = ({ onNavigate }) => {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      <header className="glass-panel border-b border-slate-800 px-6 py-4 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400">
              <Network className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h1 className="font-extrabold text-lg text-white">Branch & Hub Operations</h1>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30">
                  TENANT_ADMIN
                </span>
              </div>
              <p className="text-xs text-slate-400">{user?.email}</p>
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
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Awaiting Dispatch</span>
              <Package className="w-4 h-4 text-amber-400" />
            </div>
            <p className="text-2xl font-black text-white">28</p>
            <p className="text-[11px] text-amber-400 mt-1">Ready for courier route assignment</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Couriers on Duty</span>
              <Users className="w-4 h-4 text-emerald-400" />
            </div>
            <p className="text-2xl font-black text-white">9</p>
            <p className="text-[11px] text-emerald-400 mt-1">Active GPS heartbeats</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Hub Capacity</span>
              <Truck className="w-4 h-4 text-blue-400" />
            </div>
            <p className="text-2xl font-black text-white">68%</p>
            <p className="text-[11px] text-blue-400 mt-1">Optimal volume threshold</p>
          </div>
        </div>

        <div className="glass-card rounded-2xl p-6 border border-slate-800">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-base font-bold text-white">Hub Dispatch Queue</h2>
            <button className="flex items-center space-x-1 text-xs font-semibold text-indigo-400 hover:text-indigo-300">
              <span>View All Runs</span>
              <ArrowUpRight className="w-3.5 h-3.5" />
            </button>
          </div>
          <div className="text-sm text-slate-400">
            Branch management and courier assignment capabilities will activate in Phase 2 & Phase 6.
          </div>
        </div>
      </main>
    </div>
  );
};
