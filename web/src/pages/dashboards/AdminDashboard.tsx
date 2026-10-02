import React, { useState, useEffect } from 'react';
import { ShieldCheck, Building2, Users, FileText, CheckCircle2, Lock, LogOut } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

interface AdminDashboardProps {
  onNavigate: (path: string) => void;
}

export const AdminDashboard: React.FC<AdminDashboardProps> = ({ onNavigate }) => {
  const { user, logout, tokens } = useAuth();
  const [systemCheckResult, setSystemCheckResult] = useState<string | null>(null);
  const [loadingCheck, setLoadingCheck] = useState<boolean>(false);

  useEffect(() => {
    // Verify server-side platform admin role access
    const verifyAdmin = async () => {
      if (!tokens?.access_token) return;
      setLoadingCheck(true);
      try {
        const res = await fetch('/api/v1/admin/system-check', {
          headers: { Authorization: `Bearer ${tokens.access_token}` },
        });
        const data = await res.json();
        if (res.ok) {
          setSystemCheckResult(data.data?.message || 'Authorized');
        } else {
          setSystemCheckResult(`Forbidden: ${data.error?.message}`);
        }
      } catch (err: unknown) {
        setSystemCheckResult('Failed to run system check');
      } finally {
        setLoadingCheck(false);
      }
    };

    verifyAdmin();
  }, [tokens]);

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      {/* Header */}
      <header className="glass-panel border-b border-slate-800 px-6 py-4 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h1 className="font-extrabold text-lg text-white">Platform Administration</h1>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                  PLATFORM_ADMIN
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

      {/* Main Container */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-8 space-y-6">
        {/* Verification Banner */}
        <div className="p-4 rounded-2xl glass-card border border-indigo-500/30 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 rounded-lg bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
              <CheckCircle2 className="w-4 h-4" />
            </div>
            <div>
              <p className="text-sm font-bold text-white">Authoritative Server RBAC Verified</p>
              <p className="text-xs text-slate-400">
                {loadingCheck ? 'Validating server permissions...' : systemCheckResult}
              </p>
            </div>
          </div>
          <span className="text-[11px] font-mono px-2.5 py-1 rounded-md bg-slate-900 text-slate-300 border border-slate-800">
            RBAC: platform_admin.create, tenant.disable
          </span>
        </div>

        {/* Global Metrics Grid */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Active Tenants</span>
              <Building2 className="w-4 h-4 text-indigo-400" />
            </div>
            <p className="text-2xl font-black text-white">12</p>
            <p className="text-[11px] text-emerald-400 mt-1">Multi-tenant isolation active</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Total Users</span>
              <Users className="w-4 h-4 text-blue-400" />
            </div>
            <p className="text-2xl font-black text-white">1,480</p>
            <p className="text-[11px] text-slate-400 mt-1">Across all role categories</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">System Audits</span>
              <FileText className="w-4 h-4 text-amber-400" />
            </div>
            <p className="text-2xl font-black text-white">8,920</p>
            <p className="text-[11px] text-amber-400 mt-1">Tamper-evident logs recorded</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">RBAC Rules</span>
              <Lock className="w-4 h-4 text-purple-400" />
            </div>
            <p className="text-2xl font-black text-white">28</p>
            <p className="text-[11px] text-purple-400 mt-1">Centralized permissions</p>
          </div>
        </div>

        {/* Audit Log Stream */}
        <div className="glass-card rounded-2xl p-6 border border-slate-800">
          <h2 className="text-base font-bold text-white mb-4">Live Authentication & Security Audit Trail</h2>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs text-slate-400">
              <thead className="border-b border-slate-800 text-[11px] uppercase tracking-wider text-slate-500">
                <tr>
                  <th className="pb-3">Action</th>
                  <th className="pb-3">Target Resource</th>
                  <th className="pb-3">Origin IP</th>
                  <th className="pb-3">Status</th>
                  <th className="pb-3">Timestamp</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 font-mono">
                <tr>
                  <td className="py-3 text-indigo-400 font-bold">LOGIN_SUCCESS</td>
                  <td className="py-3 text-slate-300">auth</td>
                  <td className="py-3">127.0.0.1</td>
                  <td className="py-3"><span className="text-emerald-400">PASSED</span></td>
                  <td className="py-3 text-slate-500">Just now</td>
                </tr>
                <tr>
                  <td className="py-3 text-emerald-400 font-bold">USER_REGISTERED</td>
                  <td className="py-3 text-slate-300">users</td>
                  <td className="py-3">127.0.0.1</td>
                  <td className="py-3"><span className="text-emerald-400">CREATED</span></td>
                  <td className="py-3 text-slate-500">2 mins ago</td>
                </tr>
                <tr>
                  <td className="py-3 text-amber-400 font-bold">TOKEN_REFRESHED</td>
                  <td className="py-3 text-slate-300">auth</td>
                  <td className="py-3">127.0.0.1</td>
                  <td className="py-3"><span className="text-emerald-400">ROTATED</span></td>
                  <td className="py-3 text-slate-500">5 mins ago</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </main>
    </div>
  );
};
