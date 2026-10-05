import React, { useState, useEffect } from 'react';
import { Package, ShieldCheck, MapPin, Truck, Brain, Layers, GitBranch, ArrowUpRight, LogIn, UserCheck, Zap, Building2, Users, Eye } from 'lucide-react';
import { AuthProvider, useAuth } from './context/AuthContext';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { AdminDashboard } from './pages/dashboards/AdminDashboard';
import { TenantDashboard } from './pages/dashboards/TenantDashboard';
import { TenantAdminDashboard } from './pages/dashboards/TenantAdminDashboard';
import { EmployeeDashboard } from './pages/dashboards/EmployeeDashboard';
import { CustomerDashboard } from './pages/dashboards/CustomerDashboard';
import { HealthMonitor } from './components/HealthMonitor';
import { Role } from './types/auth';

// ── Role Demo Switcher ──────────────────────────────────────────────────────
const DEMO_ROLES: { role: Role; label: string; path: string; icon: React.ReactNode; color: string }[] = [
  { role: 'PLATFORM_ADMIN', label: 'Platform Admin', path: '/admin/dashboard', icon: <ShieldCheck className="w-3.5 h-3.5" />, color: 'indigo' },
  { role: 'TENANT', label: 'Tenant HQ', path: '/tenant/dashboard', icon: <Building2 className="w-3.5 h-3.5" />, color: 'blue' },
  { role: 'TENANT_ADMIN', label: 'Hub Ops', path: '/tenant-admin/dashboard', icon: <Zap className="w-3.5 h-3.5" />, color: 'amber' },
  { role: 'EMPLOYEE', label: 'Courier', path: '/employee/dashboard', icon: <Truck className="w-3.5 h-3.5" />, color: 'emerald' },
  { role: 'CUSTOMER', label: 'Customer', path: '/customer/dashboard', icon: <Users className="w-3.5 h-3.5" />, color: 'purple' },
];

// ── Auth Guard ───────────────────────────────────────────────────────────────
const PROTECTED_PATHS = [
  '/admin/dashboard',
  '/tenant/dashboard',
  '/tenant-admin/dashboard',
  '/employee/dashboard',
  '/customer/dashboard',
];

const MainPortal: React.FC = () => {
  const [currentPath, setCurrentPath] = useState<string>(window.location.pathname);
  const { user, isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    const handlePopState = () => {
      setCurrentPath(window.location.pathname);
    };
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  const navigate = (path: string) => {
    window.history.pushState({}, '', path);
    setCurrentPath(path);
  };

  const getDashboardPath = () => {
    if (!user) return '/login';
    switch (user.role) {
      case 'PLATFORM_ADMIN': return '/admin/dashboard';
      case 'TENANT': return '/tenant/dashboard';
      case 'TENANT_ADMIN': return '/tenant-admin/dashboard';
      case 'EMPLOYEE': return '/employee/dashboard';
      case 'CUSTOMER': return '/customer/dashboard';
      default: return '/';
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-slate-950 flex items-center justify-center">
        <div className="flex flex-col items-center gap-4">
          <div className="w-12 h-12 rounded-2xl bg-gradient-to-tr from-indigo-600 to-cyan-400 flex items-center justify-center animate-pulse">
            <Package className="w-6 h-6 text-white" />
          </div>
          <p className="text-sm text-slate-400">Restoring session…</p>
        </div>
      </div>
    );
  }

  // Protected Route Guard — redirect to /login if not authenticated
  if (PROTECTED_PATHS.includes(currentPath) && !isAuthenticated) {
    navigate('/login');
    return null;
  }

  // Route Dispatcher
  if (currentPath === '/login') return <LoginPage onNavigate={navigate} />;
  if (currentPath === '/register') return <RegisterPage onNavigate={navigate} />;
  if (currentPath === '/admin/dashboard') return <AdminDashboard onNavigate={navigate} />;
  if (currentPath === '/tenant/dashboard') return <TenantDashboard onNavigate={navigate} />;
  if (currentPath === '/tenant-admin/dashboard') return <TenantAdminDashboard onNavigate={navigate} />;
  if (currentPath === '/employee/dashboard') return <EmployeeDashboard onNavigate={navigate} />;
  if (currentPath === '/customer/dashboard') return <CustomerDashboard onNavigate={navigate} />;

  // ── Home / Overview Route ────────────────────────────────────────────────
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col antialiased">
      {/* Top Navigation */}
      <header className="sticky top-0 z-50 glass-panel border-b border-slate-800/80 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3 cursor-pointer" onClick={() => navigate('/')}>
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 via-indigo-500 to-cyan-400 flex items-center justify-center shadow-lg shadow-indigo-500/20">
              <Package className="w-5 h-5 text-white" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="font-extrabold text-xl tracking-tight text-white">LogiFlows</span>
                <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                  Phase 1 Auth Active
                </span>
              </div>
              <p className="text-xs text-slate-400">Smart Postal & Courier Delivery Intelligence</p>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <div className="hidden sm:flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-xs text-slate-300">
              <GitBranch className="w-3.5 h-3.5 text-indigo-400" />
              <span className="font-mono">phase/01-auth-rbac</span>
            </div>

            {isAuthenticated && user ? (
              <button
                onClick={() => navigate(getDashboardPath())}
                className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 text-white transition-all shadow-md shadow-emerald-600/20"
              >
                <UserCheck className="w-3.5 h-3.5" />
                <span>My Dashboard ({user.role})</span>
              </button>
            ) : (
              <div className="flex items-center space-x-2">
                <button
                  onClick={() => navigate('/login')}
                  className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-md shadow-indigo-600/20"
                >
                  <LogIn className="w-3.5 h-3.5" />
                  <span>Portal Login</span>
                </button>
                <button
                  onClick={() => navigate('/register')}
                  className="hidden sm:inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-slate-900 hover:bg-slate-800 text-slate-300 border border-slate-700 transition-all"
                >
                  <span>Register</span>
                </button>
              </div>
            )}
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-8 space-y-8">
        {/* Hero Banner */}
        <div className="relative overflow-hidden rounded-3xl glass-card p-8 md:p-10 border border-slate-800 glow-indigo">
          <div className="relative z-10 max-w-3xl space-y-4">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">
              <ShieldCheck className="w-3.5 h-3.5 text-indigo-400" />
              <span>Phase 1 Authentication & Server-Side RBAC Complete</span>
            </div>
            <h1 className="text-3xl md:text-4xl lg:text-5xl font-extrabold tracking-tight text-white leading-tight">
              Single Portal Gateway with Authoritative Multi-Tenant RBAC
            </h1>
            <p className="text-base text-slate-300 leading-relaxed">
              LogiFlows enforces centralized, permission-based authorization across <strong>Platform Admins</strong>, <strong>Tenants</strong>, <strong>Tenant Admins</strong>, <strong>Couriers</strong>, and <strong>Customers</strong>. Access is authoritatively determined by the Go backend—never trusted from frontend claims.
            </p>
            <div className="pt-2 flex flex-wrap gap-3">
              <button
                onClick={() => navigate('/login')}
                className="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl font-bold text-xs bg-indigo-600 hover:bg-indigo-500 text-white shadow-lg shadow-indigo-600/30 transition-all"
              >
                <span>Open Single Portal Login</span>
                <ArrowUpRight className="w-4 h-4" />
              </button>
              <button
                onClick={() => navigate('/register')}
                className="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl font-bold text-xs bg-slate-900 hover:bg-slate-800 text-slate-300 border border-slate-800 transition-all"
              >
                <span>Register Account</span>
              </button>
            </div>
          </div>
        </div>

        {/* ── Role Dashboard Preview Switcher ───────────────────────────── */}
        <div className="glass-card rounded-2xl p-6 border border-slate-800">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-base font-bold text-white flex items-center gap-2">
                <Eye className="w-4 h-4 text-indigo-400" />
                Role Dashboard Previewer
              </h2>
              <p className="text-xs text-slate-400 mt-0.5">
                Each role sees a completely different, permission-scoped dashboard. Click to preview.
              </p>
            </div>
            <span className="text-[10px] font-bold px-2.5 py-1 rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30">
              Demo Mode
            </span>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            {DEMO_ROLES.map((d) => (
              <button
                key={d.role}
                onClick={() => navigate(d.path)}
                className={`flex flex-col items-center gap-2.5 p-4 rounded-2xl border transition-all hover:scale-105 active:scale-95 bg-slate-900/60 hover:bg-slate-800/60 border-slate-800 hover:border-${d.color}-500/40 group`}
              >
                <div className={`w-10 h-10 rounded-xl bg-${d.color}-500/10 border border-${d.color}-500/20 flex items-center justify-center text-${d.color}-400 group-hover:bg-${d.color}-500/20 transition-colors`}>
                  {d.icon}
                </div>
                <div className="text-center">
                  <p className="text-xs font-bold text-white">{d.label}</p>
                  <p className={`text-[10px] font-bold mt-0.5 text-${d.color}-400`}>{d.role}</p>
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Live Infrastructure Health */}
        <HealthMonitor />

        {/* Feature Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[
            { icon: <ShieldCheck />, title: 'Single Portal & RBAC', desc: 'Unified authentication at /login. The backend verifies credentials and automatically redirects to the user\'s role-specific dashboard.', color: 'indigo' },
            { icon: <MapPin />, title: 'PostGIS Geospatial Engine', desc: 'PostgreSQL 16 with PostGIS extension for polygon serviceability zones, hub distance matrix calculation, and spatial routing.', color: 'blue' },
            { icon: <Brain />, title: 'Advisory AI Intelligence', desc: 'Python FastAPI service providing predictive delay intelligence (LOW, MEDIUM, HIGH) without ever bypassing Go authority.', color: 'purple' },
            { icon: <Layers />, title: 'Redis Real-Time Pub/Sub', desc: 'Decoupled transient state handling for courier live GPS points and WebSocket broadcasting with automatic safe degradation.', color: 'rose' },
            { icon: <Truck />, title: 'Flutter Mobile Suite', desc: 'Dedicated mobile operations client with offline queueing, barcode/QR scanner, real-time GPS telemetry, and Proof of Delivery.', color: 'cyan' },
            { icon: <Package />, title: 'Multi-Tenancy Isolation', desc: 'Independent company namespaces ensuring Tenant A can never query or mutate Tenant B records, enforced by backend middleware.', color: 'amber' },
          ].map((feat) => (
            <div key={feat.title} className={`glass-card rounded-2xl p-6 border border-slate-800 hover:border-${feat.color}-500/40 transition-all group`}>
              <div className={`w-10 h-10 rounded-xl bg-${feat.color}-500/10 border border-${feat.color}-500/20 flex items-center justify-center text-${feat.color}-400 mb-4 group-hover:bg-${feat.color}-500/20 transition-colors`}>
                {React.cloneElement(feat.icon as React.ReactElement, { className: 'w-5 h-5' })}
              </div>
              <h3 className={`text-lg font-bold text-white mb-2 group-hover:text-${feat.color}-300 transition-colors`}>
                {feat.title}
              </h3>
              <p className="text-sm text-slate-400 leading-relaxed">{feat.desc}</p>
            </div>
          ))}
        </div>
      </main>

      <footer className="border-t border-slate-800/80 px-6 py-6 mt-12 bg-slate-950/80">
        <div className="max-w-7xl mx-auto flex flex-col sm:flex-row items-center justify-between text-xs text-slate-500 gap-4">
          <p>© 2026 LogiFlows Logistics Platform. All rights reserved.</p>
          <p className="font-mono">Branch: phase/01-auth-rbac</p>
        </div>
      </footer>
    </div>
  );
};

export const App: React.FC = () => {
  return (
    <AuthProvider>
      <MainPortal />
    </AuthProvider>
  );
};

export default App;
