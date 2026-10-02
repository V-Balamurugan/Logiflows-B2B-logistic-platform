import React, { useEffect, useState } from 'react';
import { Network, Plus, Building2, MapPin, Phone, Mail, LogOut, CheckCircle2, ShieldCheck } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { Branch } from '../../types/tenancy';
import { GeospatialMap } from '../../components/GeospatialMap';
import { CreateBranchModal } from '../../components/CreateBranchModal';

interface TenantAdminDashboardProps {
  onNavigate: (path: string) => void;
}

export const TenantAdminDashboard: React.FC<TenantAdminDashboardProps> = ({ onNavigate }) => {
  const { user, logout } = useAuth();
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const [activeTab, setActiveTab] = useState<'map' | 'list'>('map');

  const tenantId = user?.tenant_id || '00000000-0000-0000-0000-000000000001';

  const fetchBranches = async () => {
    setLoading(true);
    setError(null);
    try {
      const token = localStorage.getItem('access_token');
      const res = await fetch(`/api/v1/tenants/${tenantId}/branches`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      const json = await res.json();
      if (res.ok) {
        setBranches(json.data || []);
      } else {
        setError(json.error?.message || 'Failed to fetch branches');
      }
    } catch (err: any) {
      setError(err.message || 'Network error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBranches();
  }, [tenantId]);

  const handleBranchCreated = (newBranch: Branch) => {
    setBranches((prev) => [...prev, newBranch]);
  };

  // Metrics
  const totalCapacity = branches.reduce((acc, b) => acc + (b.daily_capacity || 0), 0);
  const hubCount = branches.filter((b) => b.branch_type === 'SORTING_HUB').length;
  const dcCount = branches.filter((b) => b.branch_type === 'DISTRIBUTION_CENTER').length;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col">
      {/* Header */}
      <header className="glass-panel border-b border-slate-800 px-6 py-4 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400">
              <Network className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h1 className="font-extrabold text-lg text-white">Hub & Branch Operations</h1>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30">
                  TENANT_ADMIN
                </span>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 flex items-center gap-1">
                  <ShieldCheck className="w-3 h-3" /> PostGIS 16 Active
                </span>
              </div>
              <p className="text-xs text-slate-400">{user?.email}</p>
            </div>
          </div>

          <div className="flex items-center space-x-3">
            <button
              onClick={() => setIsModalOpen(true)}
              className="flex items-center space-x-1.5 px-3.5 py-1.5 rounded-xl text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white shadow-lg shadow-indigo-600/20 transition-all"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Deploy Branch</span>
            </button>
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

      {/* Main Container */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-8 space-y-6">
        {/* Metric Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Total Active Branches</span>
              <Building2 className="w-4 h-4 text-indigo-400" />
            </div>
            <p className="text-2xl font-black text-white">{branches.length}</p>
            <p className="text-[11px] text-indigo-400 mt-1">Operating with spatial coverage</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Sorting Hubs</span>
              <Network className="w-4 h-4 text-purple-400" />
            </div>
            <p className="text-2xl font-black text-white">{hubCount}</p>
            <p className="text-[11px] text-purple-400 mt-1">Super regional triage nodes</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Distribution Centers</span>
              <Building2 className="w-4 h-4 text-blue-400" />
            </div>
            <p className="text-2xl font-black text-white">{dcCount}</p>
            <p className="text-[11px] text-blue-400 mt-1">Mid-mile dispatch zones</p>
          </div>

          <div className="glass-card rounded-2xl p-5 border border-slate-800">
            <div className="flex items-center justify-between text-slate-400 mb-2">
              <span className="text-xs font-bold uppercase tracking-wider">Daily Processing Quota</span>
              <CheckCircle2 className="w-4 h-4 text-emerald-400" />
            </div>
            <p className="text-2xl font-black text-white">{loading ? '...' : totalCapacity.toLocaleString()}</p>
            <p className="text-[11px] text-emerald-400 mt-1">Parcels/day aggregate</p>
          </div>
        </div>

        {/* View Toggle Tabs */}
        <div className="flex items-center justify-between border-b border-slate-800 pb-3">
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setActiveTab('map')}
              className={`px-4 py-2 rounded-xl text-xs font-bold transition-all ${
                activeTab === 'map'
                  ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/20'
                  : 'bg-slate-900 text-slate-400 hover:text-white'
              }`}
            >
              Interactive PostGIS Coverage Map
            </button>
            <button
              onClick={() => setActiveTab('list')}
              className={`px-4 py-2 rounded-xl text-xs font-bold transition-all ${
                activeTab === 'list'
                  ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-600/20'
                  : 'bg-slate-900 text-slate-400 hover:text-white'
              }`}
            >
              Branch Network Registry ({branches.length})
            </button>
          </div>

          <div className="text-xs text-slate-400">
            SRID 4326 PostGIS Polygon Point-in-Polygon Engine
          </div>
        </div>

        {error && (
          <div className="p-4 bg-rose-500/10 border border-rose-500/20 text-rose-300 rounded-2xl text-xs">
            {error}
          </div>
        )}

        {/* Tab 1: Interactive Geospatial Map */}
        {activeTab === 'map' && (
          <div className="space-y-4">
            <GeospatialMap tenantId={tenantId} branches={branches} />
          </div>
        )}

        {/* Tab 2: Branch Network Registry */}
        {activeTab === 'list' && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {branches.map((b) => (
              <div
                key={b.id}
                className="glass-card rounded-2xl p-5 border border-slate-800 hover:border-slate-700 transition-all flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <span className="font-mono text-[10px] font-bold px-2 py-0.5 rounded bg-slate-800 text-indigo-400 border border-slate-700">
                        {b.code}
                      </span>
                      <h3 className="font-bold text-sm text-slate-100 mt-1.5">{b.name}</h3>
                    </div>
                    <span
                      className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                        b.branch_type === 'SORTING_HUB'
                          ? 'bg-purple-500/20 text-purple-300 border border-purple-500/30'
                          : b.branch_type === 'DISTRIBUTION_CENTER'
                          ? 'bg-blue-500/20 text-blue-300 border border-blue-500/30'
                          : 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
                      }`}
                    >
                      {b.branch_type}
                    </span>
                  </div>

                  <div className="space-y-2 text-xs text-slate-400 mt-3">
                    <div className="flex items-start gap-2">
                      <MapPin className="w-3.5 h-3.5 text-slate-500 shrink-0 mt-0.5" />
                      <span>{b.address}, {b.city}, {b.postal_code}</span>
                    </div>
                    {b.contact_phone && (
                      <div className="flex items-center gap-2">
                        <Phone className="w-3.5 h-3.5 text-slate-500 shrink-0" />
                        <span>{b.contact_phone}</span>
                      </div>
                    )}
                    {b.contact_email && (
                      <div className="flex items-center gap-2">
                        <Mail className="w-3.5 h-3.5 text-slate-500 shrink-0" />
                        <span>{b.contact_email}</span>
                      </div>
                    )}
                  </div>
                </div>

                <div className="mt-4 pt-4 border-t border-slate-800/80 flex items-center justify-between text-xs">
                  <div>
                    <span className="text-slate-500">Capacity:</span>{' '}
                    <span className="text-slate-200 font-bold">{b.daily_capacity.toLocaleString()}</span>
                  </div>
                  <div className="font-mono text-[11px] text-slate-400">
                    {b.location.latitude.toFixed(3)}, {b.location.longitude.toFixed(3)}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      {/* Deploy Branch Modal */}
      <CreateBranchModal
        tenantId={tenantId}
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onCreated={handleBranchCreated}
      />
    </div>
  );
};
