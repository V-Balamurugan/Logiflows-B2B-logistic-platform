import React, { useState, useEffect } from 'react';
import { Package, Search, PlusCircle, Clock, CheckCircle2, ShieldCheck, LogOut } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

interface CustomerDashboardProps {
  onNavigate: (path: string) => void;
}

export const CustomerDashboard: React.FC<CustomerDashboardProps> = ({ onNavigate }) => {
  const { user, logout, tokens } = useAuth();
  const [permissionCheck, setPermissionCheck] = useState<string | null>(null);

  useEffect(() => {
    const verifyCustomerPerms = async () => {
      if (!tokens?.access_token) return;
      try {
        const res = await fetch('/api/v1/parcels/permission-check', {
          headers: { Authorization: `Bearer ${tokens.access_token}` },
        });
        const data = await res.json();
        if (res.ok) {
          setPermissionCheck(data.data?.message || 'Authorized');
        } else {
          setPermissionCheck(`Denied: ${data.error?.message}`);
        }
      } catch {
        setPermissionCheck('Failed to verify permissions');
      }
    };

    verifyCustomerPerms();
  }, [tokens]);

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      <header className="glass-panel border-b border-slate-800 px-6 py-4 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-purple-500/10 border border-purple-500/20 text-purple-400">
              <Package className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h1 className="font-extrabold text-lg text-white">Customer Portal</h1>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-purple-500/20 text-purple-300 border border-purple-500/30">
                  CUSTOMER
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
        {/* Permission Confirmation */}
        <div className="p-4 rounded-2xl glass-card border border-purple-500/30 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
              <CheckCircle2 className="w-4 h-4" />
            </div>
            <div>
              <p className="text-sm font-bold text-white">Customer Permissions Active</p>
              <p className="text-xs text-slate-400">
                Server confirmation: {permissionCheck ?? 'Validating parcel.create permission...'}
              </p>
            </div>
          </div>
          <span className="text-[11px] font-mono px-2.5 py-1 rounded-md bg-slate-900 text-slate-300 border border-slate-800 flex items-center gap-1.5">
            <ShieldCheck className="w-3.5 h-3.5 text-purple-400" />
            <span>Perms: parcel.create, parcel.track</span>
          </span>
        </div>

        {/* Tracking Bar */}
        <div className="glass-card rounded-2xl p-6 border border-slate-800">
          <h2 className="text-sm font-bold text-white uppercase tracking-wider mb-3">Quick Parcel Tracking</h2>
          <div className="flex gap-3">
            <div className="relative flex-1">
              <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-500">
                <Search className="h-4 w-4" />
              </div>
              <input
                type="text"
                placeholder="Enter 12-digit tracking number (e.g. LF-8923-4412-901)"
                className="w-full pl-10 pr-4 py-2.5 bg-slate-900/90 border border-slate-800 rounded-xl text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
              />
            </div>
            <button className="px-5 py-2.5 rounded-xl font-bold text-xs bg-purple-600 hover:bg-purple-500 text-white shadow-md shadow-purple-600/30 transition-all">
              Track Shipment
            </button>
          </div>
        </div>

        {/* Action & Status Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="glass-card rounded-2xl p-6 border border-slate-800 flex flex-col justify-between">
            <div>
              <div className="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400 mb-4">
                <PlusCircle className="w-5 h-5" />
              </div>
              <h3 className="text-lg font-bold text-white mb-1">Book New Delivery</h3>
              <p className="text-sm text-slate-400 leading-relaxed mb-6">
                Generate secure QR identity, calculate dynamic pricing, and dispatch a courier to your pickup location.
              </p>
            </div>
            <div className="text-xs text-slate-500 font-mono">
              Coming in Phase 4: Customer & Parcel Booking
            </div>
          </div>

          <div className="glass-card rounded-2xl p-6 border border-slate-800 flex flex-col justify-between">
            <div>
              <div className="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-blue-400 mb-4">
                <Clock className="w-5 h-5" />
              </div>
              <h3 className="text-lg font-bold text-white mb-1">Shipment History</h3>
              <p className="text-sm text-slate-400 leading-relaxed mb-6">
                View verified proof-of-delivery receipts, digital signatures, and timeline progression.
              </p>
            </div>
            <div className="text-xs text-slate-500 font-mono">
              Coming in Phase 5: QR Custody & Timeline
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};
