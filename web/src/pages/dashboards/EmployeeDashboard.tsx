import React from 'react';
import { Truck, QrCode, CheckCircle, Navigation, LogOut } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

interface EmployeeDashboardProps {
  onNavigate: (path: string) => void;
}

export const EmployeeDashboard: React.FC<EmployeeDashboardProps> = ({ onNavigate }) => {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      <header className="glass-panel border-b border-slate-800 px-6 py-4 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
              <Truck className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h1 className="font-extrabold text-lg text-white">Courier Operations</h1>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                  EMPLOYEE
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
              <span className="text-xs font-bold uppercase tracking-wider">Today's Assigned</span>
              <Navigation className="w-4 h-4 text-emerald-400" />
            </div>
            <p className="text-2xl font-black text-white">14 Stops</p>
            <p className="text-[11px] text-emerald-400 mt-1">Route optimized by PostGIS</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Completed</span>
              <CheckCircle className="w-4 h-4 text-blue-400" />
            </div>
            <p className="text-2xl font-black text-white">8 Delivered</p>
            <p className="text-[11px] text-slate-400 mt-1">Proof of Delivery captured</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Remaining</span>
              <QrCode className="w-4 h-4 text-amber-400" />
            </div>
            <p className="text-2xl font-black text-white">6 In Transit</p>
            <p className="text-[11px] text-amber-400 mt-1">Next: 42 Elm Street</p>
          </div>
        </div>

        <div className="p-6 rounded-2xl glass-card border border-slate-800">
          <div className="flex items-center space-x-3 mb-3">
            <QrCode className="w-5 h-5 text-indigo-400" />
            <h2 className="text-base font-bold text-white">Mobile Companion Client</h2>
          </div>
          <p className="text-sm text-slate-400 leading-relaxed mb-4">
            Couriers and field staff typically perform on-the-road scanning, live GPS telemetry broadcasting, and recipient signature capture via the <strong>Flutter Employee Mobile App</strong> (configured in <code>mobile/</code> on port 5174).
          </p>
          <a
            href="http://localhost:5174"
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center space-x-2 px-4 py-2 rounded-xl text-xs font-bold bg-indigo-600 hover:bg-indigo-500 text-white shadow-md shadow-indigo-600/30 transition-all"
          >
            <span>Launch Flutter Mobile Client</span>
          </a>
        </div>
      </main>
    </div>
  );
};
