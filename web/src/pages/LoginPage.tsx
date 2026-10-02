import React, { useState } from 'react';
import { Package, Lock, Mail, ArrowRight, AlertCircle, ShieldAlert, KeyRound } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { Role } from '../types/auth';

interface LoginPageProps {
  onNavigate: (path: string) => void;
}

export const LoginPage: React.FC<LoginPageProps> = ({ onNavigate }) => {
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const redirectByRole = (role: Role) => {
    switch (role) {
      case 'PLATFORM_ADMIN':
        onNavigate('/admin/dashboard');
        break;
      case 'TENANT':
        onNavigate('/tenant/dashboard');
        break;
      case 'TENANT_ADMIN':
        onNavigate('/tenant-admin/dashboard');
        break;
      case 'EMPLOYEE':
        onNavigate('/employee/dashboard');
        break;
      case 'CUSTOMER':
        onNavigate('/customer/dashboard');
        break;
      default:
        onNavigate('/');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      const authoritativeRole = await login({ email, password });
      redirectByRole(authoritativeRole);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Authentication failed.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleQuickFill = (demoEmail: string, demoPass: string) => {
    setEmail(demoEmail);
    setPassword(demoPass);
    setError(null);
  };

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col justify-center py-12 sm:px-6 lg:px-8 selection:bg-indigo-500 selection:text-white">
      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center">
        <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-gradient-to-tr from-indigo-600 via-indigo-500 to-cyan-400 shadow-xl shadow-indigo-500/25 mb-4">
          <Package className="w-7 h-7 text-white" />
        </div>
        <h2 className="text-3xl font-extrabold tracking-tight text-white">LogiFlows Portal</h2>
        <p className="mt-2 text-sm text-slate-400">
          Intelligent Single Portal Gateway • Authoritative RBAC Access
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md px-4 sm:px-0">
        <div className="glass-card py-8 px-6 shadow-2xl rounded-3xl sm:px-10 border border-slate-800 glow-indigo">
          {error && (
            <div className="mb-6 p-4 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-300 flex items-start space-x-3 text-sm">
              <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          <form className="space-y-5" onSubmit={handleSubmit}>
            <div>
              <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">
                Work / Account Email
              </label>
              <div className="relative rounded-xl shadow-sm">
                <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-500">
                  <Mail className="h-4 w-4" />
                </div>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="name@company.com"
                  className="block w-full pl-10 pr-4 py-2.5 bg-slate-900/90 border border-slate-700/80 rounded-xl text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
                />
              </div>
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">
                Password
              </label>
              <div className="relative rounded-xl shadow-sm">
                <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-500">
                  <Lock className="h-4 w-4" />
                </div>
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="••••••••••••"
                  className="block w-full pl-10 pr-4 py-2.5 bg-slate-900/90 border border-slate-700/80 rounded-xl text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full flex items-center justify-center py-3 px-4 border border-transparent rounded-xl shadow-lg shadow-indigo-600/30 text-sm font-bold text-white bg-indigo-600 hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-all disabled:opacity-50"
            >
              {isSubmitting ? (
                <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <>
                  <span>Sign In</span>
                  <ArrowRight className="ml-2 w-4 h-4" />
                </>
              )}
            </button>
          </form>

          <div className="mt-6 pt-6 border-t border-slate-800 text-center">
            <p className="text-xs text-slate-400">
              New business client or sender?{' '}
              <button
                onClick={() => onNavigate('/register')}
                className="font-semibold text-indigo-400 hover:text-indigo-300 transition-colors"
              >
                Create Account
              </button>
            </p>
          </div>

          {/* Quick Demo Fill Buttons for Examiners */}
          <div className="mt-6 pt-4 border-t border-slate-800/80">
            <div className="flex items-center space-x-1.5 text-xs text-slate-400 mb-3 font-semibold">
              <KeyRound className="w-3.5 h-3.5 text-indigo-400" />
              <span>Demo Role Quick-Fill (Authoritative Credentials)</span>
            </div>
            <div className="grid grid-cols-2 gap-2 text-[11px]">
              <button
                type="button"
                onClick={() => handleQuickFill('admin@logiflows.io', 'Admin@LogiFlows2026!')}
                className="p-2 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-left text-slate-300 transition-all"
              >
                <span className="font-bold text-indigo-400 block">Platform Admin</span>
                <span className="text-[10px] text-slate-500">Super Admin</span>
              </button>

              <button
                type="button"
                onClick={() => handleQuickFill('tenant@speedycourier.com', 'Tenant@Speedy2026!')}
                className="p-2 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-left text-slate-300 transition-all"
              >
                <span className="font-bold text-blue-400 block">Tenant Owner</span>
                <span className="text-[10px] text-slate-500">Speedy Express</span>
              </button>

              <button
                type="button"
                onClick={() => handleQuickFill('manager@speedycourier.com', 'Manager@Speedy2026!')}
                className="p-2 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-left text-slate-300 transition-all"
              >
                <span className="font-bold text-amber-400 block">Tenant Admin</span>
                <span className="text-[10px] text-slate-500">Branch Dispatcher</span>
              </button>

              <button
                type="button"
                onClick={() => handleQuickFill('courier@speedycourier.com', 'Courier@Speedy2026!')}
                className="p-2 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-left text-slate-300 transition-all"
              >
                <span className="font-bold text-emerald-400 block">Employee</span>
                <span className="text-[10px] text-slate-500">Courier / Driver</span>
              </button>

              <button
                type="button"
                onClick={() => handleQuickFill('customer@gmail.com', 'Customer@LogiFlows2026!')}
                className="col-span-2 p-2 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-left text-slate-300 transition-all"
              >
                <span className="font-bold text-purple-400 block">Customer</span>
                <span className="text-[10px] text-slate-500">Sender / Receiver Portal</span>
              </button>
            </div>
          </div>
        </div>

        <div className="mt-6 flex items-center justify-center space-x-2 text-xs text-slate-500">
          <ShieldAlert className="w-3.5 h-3.5" />
          <span>Server-side RBAC: Identity and permissions are authoritative in Go</span>
        </div>
      </div>
    </div>
  );
};
