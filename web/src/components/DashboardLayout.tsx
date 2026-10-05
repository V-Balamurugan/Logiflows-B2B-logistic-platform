import React, { useState } from 'react';
import {
  Package, ShieldCheck, Building2, Network, Truck, Users,
  LogOut, Home, ChevronLeft, ChevronRight, Bell, Search,
  Settings, Menu, X
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { Role } from '../types/auth';

interface NavItem {
  icon: React.ReactNode;
  label: string;
  badge?: string | number;
}

interface DashboardLayoutProps {
  children: React.ReactNode;
  onNavigate: (path: string) => void;
  navItems?: NavItem[];
  activeNav?: string;
  onNavClick?: (label: string) => void;
}

const ROLE_META: Record<Role, { label: string; color: string; bgColor: string; borderColor: string; icon: React.ReactNode }> = {
  PLATFORM_ADMIN: {
    label: 'Platform Admin',
    color: 'text-indigo-300',
    bgColor: 'bg-indigo-500/10',
    borderColor: 'border-indigo-500/30',
    icon: <ShieldCheck className="w-4 h-4" />,
  },
  TENANT: {
    label: 'Tenant HQ',
    color: 'text-blue-300',
    bgColor: 'bg-blue-500/10',
    borderColor: 'border-blue-500/30',
    icon: <Building2 className="w-4 h-4" />,
  },
  TENANT_ADMIN: {
    label: 'Hub Operations',
    color: 'text-amber-300',
    bgColor: 'bg-amber-500/10',
    borderColor: 'border-amber-500/30',
    icon: <Network className="w-4 h-4" />,
  },
  EMPLOYEE: {
    label: 'Courier',
    color: 'text-emerald-300',
    bgColor: 'bg-emerald-500/10',
    borderColor: 'border-emerald-500/30',
    icon: <Truck className="w-4 h-4" />,
  },
  CUSTOMER: {
    label: 'Customer',
    color: 'text-purple-300',
    bgColor: 'bg-purple-500/10',
    borderColor: 'border-purple-500/30',
    icon: <Users className="w-4 h-4" />,
  },
};

export const DashboardLayout: React.FC<DashboardLayoutProps> = ({
  children,
  onNavigate,
  navItems = [],
  activeNav,
  onNavClick,
}) => {
  const { user, logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  const roleMeta = user?.role ? ROLE_META[user.role] : null;

  const handleLogout = async () => {
    await logout();
    onNavigate('/login');
  };

  const Sidebar = () => (
    <aside
      className={`
        flex flex-col h-full bg-slate-950/95 border-r border-slate-800/80
        backdrop-blur-xl transition-all duration-300 ease-in-out
        ${collapsed ? 'w-16' : 'w-60'}
      `}
    >
      {/* Logo */}
      <div className="flex items-center justify-between px-3 py-4 border-b border-slate-800/60 min-h-[64px]">
        {!collapsed && (
          <button onClick={() => onNavigate('/')} className="flex items-center gap-2.5 group">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-indigo-600 via-indigo-500 to-cyan-400 flex items-center justify-center shadow-md shadow-indigo-500/30 shrink-0">
              <Package className="w-4 h-4 text-white" />
            </div>
            <span className="font-extrabold text-sm text-white group-hover:text-indigo-300 transition-colors">LogiFlows</span>
          </button>
        )}
        {collapsed && (
          <button onClick={() => onNavigate('/')} className="mx-auto">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-indigo-600 via-indigo-500 to-cyan-400 flex items-center justify-center shadow-md shadow-indigo-500/30">
              <Package className="w-4 h-4 text-white" />
            </div>
          </button>
        )}
        {!collapsed && (
          <button
            onClick={() => setCollapsed(true)}
            className="p-1 rounded-lg text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-all"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Role Badge */}
      {!collapsed && roleMeta && (
        <div className={`mx-3 mt-3 px-3 py-2 rounded-xl ${roleMeta.bgColor} border ${roleMeta.borderColor} flex items-center gap-2`}>
          <span className={roleMeta.color}>{roleMeta.icon}</span>
          <div className="min-w-0">
            <p className={`text-xs font-bold ${roleMeta.color}`}>{roleMeta.label}</p>
            <p className="text-[10px] text-slate-500 truncate">{user?.email}</p>
          </div>
        </div>
      )}

      {/* Nav Items */}
      <nav className="flex-1 overflow-y-auto px-2 py-3 space-y-1">
        {navItems.map((item) => {
          const isActive = activeNav === item.label;
          return (
            <button
              key={item.label}
              onClick={() => onNavClick?.(item.label)}
              title={collapsed ? item.label : undefined}
              className={`
                w-full flex items-center gap-3 px-2.5 py-2.5 rounded-xl text-xs font-semibold transition-all
                ${isActive
                  ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30'
                  : 'text-slate-400 hover:bg-slate-800/60 hover:text-slate-200'
                }
                ${collapsed ? 'justify-center' : ''}
              `}
            >
              <span className={`shrink-0 ${isActive ? 'text-indigo-400' : ''}`}>{item.icon}</span>
              {!collapsed && (
                <>
                  <span className="flex-1 text-left">{item.label}</span>
                  {item.badge !== undefined && (
                    <span className="px-1.5 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 text-[10px] font-bold">
                      {item.badge}
                    </span>
                  )}
                </>
              )}
            </button>
          );
        })}
      </nav>

      {/* Bottom */}
      <div className="border-t border-slate-800/60 px-2 py-3 space-y-1">
        <button
          onClick={() => onNavigate('/')}
          title={collapsed ? 'System Pulse' : undefined}
          className={`w-full flex items-center gap-3 px-2.5 py-2.5 rounded-xl text-xs font-semibold text-slate-400 hover:bg-slate-800/60 hover:text-slate-200 transition-all ${collapsed ? 'justify-center' : ''}`}
        >
          <Home className="w-4 h-4 shrink-0" />
          {!collapsed && <span>System Pulse</span>}
        </button>
        <button
          onClick={handleLogout}
          title={collapsed ? 'Sign Out' : undefined}
          className={`w-full flex items-center gap-3 px-2.5 py-2.5 rounded-xl text-xs font-semibold text-rose-400 hover:bg-rose-500/10 transition-all ${collapsed ? 'justify-center' : ''}`}
        >
          <LogOut className="w-4 h-4 shrink-0" />
          {!collapsed && <span>Sign Out</span>}
        </button>
        {collapsed && (
          <button
            onClick={() => setCollapsed(false)}
            className="w-full flex items-center justify-center px-2.5 py-2.5 rounded-xl text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-all"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        )}
      </div>
    </aside>
  );

  return (
    <div className="flex h-screen bg-slate-950 text-slate-100 overflow-hidden">
      {/* Desktop Sidebar */}
      <div className="hidden md:flex flex-col h-full shrink-0">
        <Sidebar />
      </div>

      {/* Mobile Sidebar Overlay */}
      {mobileOpen && (
        <div className="md:hidden fixed inset-0 z-50 flex">
          <div className="w-60 flex flex-col h-full">
            <Sidebar />
          </div>
          <div
            className="flex-1 bg-black/60 backdrop-blur-sm"
            onClick={() => setMobileOpen(false)}
          />
        </div>
      )}

      {/* Main Area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Top Bar */}
        <header className="shrink-0 h-16 flex items-center justify-between px-4 md:px-6 border-b border-slate-800/80 bg-slate-950/80 backdrop-blur-xl z-40">
          <div className="flex items-center gap-3">
            <button
              className="md:hidden p-2 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-all"
              onClick={() => setMobileOpen(!mobileOpen)}
            >
              {mobileOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </button>

            {/* Search */}
            <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-xl bg-slate-900 border border-slate-800 text-xs text-slate-400 w-52 cursor-text hover:border-slate-700 transition-colors">
              <Search className="w-3.5 h-3.5 shrink-0" />
              <span>Quick search...</span>
              <kbd className="ml-auto text-[10px] px-1.5 py-0.5 rounded bg-slate-800 text-slate-500 font-mono">⌘K</kbd>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {/* Notification Bell */}
            <button className="relative p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition-all">
              <Bell className="w-4 h-4" />
              <span className="absolute top-1.5 right-1.5 w-1.5 h-1.5 rounded-full bg-indigo-500 ring-2 ring-slate-950" />
            </button>

            {/* Settings */}
            <button className="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition-all">
              <Settings className="w-4 h-4" />
            </button>

            {/* Avatar */}
            <div className="flex items-center gap-2 pl-2 ml-1 border-l border-slate-800">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold ${roleMeta?.bgColor} ${roleMeta?.color} border ${roleMeta?.borderColor}`}>
                {user?.first_name?.[0]}{user?.last_name?.[0]}
              </div>
              <div className="hidden sm:block">
                <p className="text-xs font-semibold text-white">{user?.first_name} {user?.last_name}</p>
                <p className={`text-[10px] font-bold ${roleMeta?.color}`}>{user?.role}</p>
              </div>
            </div>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 overflow-y-auto">
          {children}
        </main>
      </div>
    </div>
  );
};
