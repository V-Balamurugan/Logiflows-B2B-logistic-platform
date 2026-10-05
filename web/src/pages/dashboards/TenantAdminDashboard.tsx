import React, { useEffect, useState } from 'react';
import {
  Network, Plus, Building2, MapPin, Phone, Mail, CheckCircle2,
  ShieldCheck, Package, BarChart3, Layers, Filter, ArrowUpRight
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { Branch } from '../../types/tenancy';
import { GeospatialMap } from '../../components/GeospatialMap';
import { CreateBranchModal } from '../../components/CreateBranchModal';
import { DashboardLayout } from '../../components/DashboardLayout';

interface TenantAdminDashboardProps {
  onNavigate: (path: string) => void;
}

const NAV = [
  { icon: <BarChart3 className="w-4 h-4" />, label: 'Overview' },
  { icon: <MapPin className="w-4 h-4" />, label: 'PostGIS Map' },
  { icon: <Building2 className="w-4 h-4" />, label: 'Branch Registry' },
  { icon: <Layers className="w-4 h-4" />, label: 'Zone Coverage' },
  { icon: <Package className="w-4 h-4" />, label: 'Capacity Planner' },
];

const BRANCH_TYPE_COLOR: Record<string, string> = {
  SORTING_HUB: 'bg-purple-500/20 text-purple-300 border-purple-500/30',
  DISTRIBUTION_CENTER: 'bg-blue-500/20 text-blue-300 border-blue-500/30',
  DELIVERY_POINT: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
};

export const TenantAdminDashboard: React.FC<TenantAdminDashboardProps> = ({ onNavigate }) => {
  const { user } = useAuth();
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [activeNav, setActiveNav] = useState('Overview');
  const [filterType, setFilterType] = useState<string>('ALL');

  const tenantId = user?.tenant_id || '00000000-0000-0000-0000-000000000001';

  const fetchBranches = async () => {
    setLoading(true);
    setError(null);
    try {
      const token = localStorage.getItem('access_token');
      const res = await fetch(`/api/v1/tenants/${tenantId}/branches`, {
        headers: { Authorization: `Bearer ${token}` },
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

  useEffect(() => { fetchBranches(); }, [tenantId]);

  const handleBranchCreated = (newBranch: Branch) => {
    setBranches((prev) => [...prev, newBranch]);
  };

  const totalCapacity = branches.reduce((acc, b) => acc + (b.daily_capacity || 0), 0);
  const hubCount = branches.filter((b) => b.branch_type === 'SORTING_HUB').length;
  const dcCount = branches.filter((b) => b.branch_type === 'DISTRIBUTION_CENTER').length;
  const dpCount = branches.filter((b) => b.branch_type === 'LOCAL_OFFICE').length;

  const filteredBranches = filterType === 'ALL'
    ? branches
    : branches.filter((b) => b.branch_type === (filterType as Branch['branch_type']));


  // Capacity bar widths per branch type
  const maxCap = Math.max(...branches.map((b) => b.daily_capacity || 1), 1);

  return (
    <DashboardLayout onNavigate={onNavigate} navItems={NAV} activeNav={activeNav} onNavClick={setActiveNav}>
      <div className="p-6 space-y-6 fade-in-up">

        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-extrabold text-white">Hub & Branch Operations</h1>
              <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
                <ShieldCheck className="w-3 h-3" /> PostGIS 16
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-0.5">Manage branch network & SRID 4326 serviceability zones</p>
          </div>
          <button
            onClick={() => setIsModalOpen(true)}
            className="flex items-center gap-1.5 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-lg shadow-indigo-600/20 transition-all"
          >
            <Plus className="w-3.5 h-3.5" /> Deploy Branch
          </button>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'Total Branches', value: branches.length, sub: 'Spatial coverage active', color: 'indigo', icon: <Building2 className="w-4 h-4" /> },
            { label: 'Sorting Hubs', value: hubCount, sub: 'Regional triage nodes', color: 'purple', icon: <Network className="w-4 h-4" /> },
            { label: 'Distribution Centers', value: dcCount, sub: 'Mid-mile dispatch', color: 'blue', icon: <Building2 className="w-4 h-4" /> },
            { label: 'Daily Quota', value: loading ? '…' : totalCapacity.toLocaleString(), sub: 'Parcels/day aggregate', color: 'emerald', icon: <CheckCircle2 className="w-4 h-4" /> },
          ].map((s) => (
            <div key={s.label} className="stat-card glass-card rounded-2xl p-5 border border-slate-800">
              <div className="flex items-center justify-between mb-3">
                <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">{s.label}</span>
                <div className={`w-7 h-7 rounded-lg bg-${s.color}-500/10 border border-${s.color}-500/20 flex items-center justify-center text-${s.color}-400`}>
                  {s.icon}
                </div>
              </div>
              <p className="text-2xl font-black text-white">{s.value}</p>
              <p className="text-[11px] text-slate-400 mt-1">{s.sub}</p>
            </div>
          ))}
        </div>

        {/* Capacity Distribution */}
        <div className="glass-card rounded-2xl p-6 border border-slate-800">
          <div className="flex items-center justify-between mb-5">
            <div>
              <h2 className="text-sm font-bold text-white">Capacity Distribution by Branch Type</h2>
              <p className="text-[11px] text-slate-400 mt-0.5">Daily parcel processing quota per facility</p>
            </div>
          </div>
          <div className="space-y-3">
            {[
              { type: 'SORTING_HUB', label: 'Sorting Hubs', count: hubCount, color: 'bg-purple-500', cap: branches.filter(b => b.branch_type === 'SORTING_HUB').reduce((a,b) => a + (b.daily_capacity||0), 0) },
              { type: 'DISTRIBUTION_CENTER', label: 'Distribution Centers', count: dcCount, color: 'bg-blue-500', cap: branches.filter(b => b.branch_type === 'DISTRIBUTION_CENTER').reduce((a,b) => a + (b.daily_capacity||0), 0) },
              { type: 'LOCAL_OFFICE', label: 'Local Offices', count: dpCount, color: 'bg-emerald-500', cap: branches.filter(b => b.branch_type === 'LOCAL_OFFICE').reduce((a,b) => a + (b.daily_capacity||0), 0) },
            ].map((row) => (
              <div key={row.type} className="flex items-center gap-4">
                <div className="w-36 shrink-0">
                  <div className="text-xs font-semibold text-slate-300">{row.label}</div>
                  <div className="text-[10px] text-slate-500">{row.count} facilities • {row.cap.toLocaleString()} cap</div>
                </div>
                <div className="flex-1 h-2 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full progress-fill ${row.color}`}
                    style={{ width: totalCapacity > 0 ? `${(row.cap / totalCapacity) * 100}%` : '0%' }}
                  />
                </div>
                <span className="text-xs font-bold text-slate-300 w-8 text-right">
                  {totalCapacity > 0 ? `${Math.round((row.cap / totalCapacity) * 100)}%` : '0%'}
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* PostGIS Map */}
        {(activeNav === 'Overview' || activeNav === 'PostGIS Map' || activeNav === 'Zone Coverage') && (
          <div className="glass-card rounded-2xl p-6 border border-slate-800 space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-sm font-bold text-white">Interactive PostGIS Coverage Map</h2>
                <p className="text-xs text-slate-400">SRID 4326 polygon point-in-polygon engine</p>
              </div>
              <button
                onClick={() => setIsModalOpen(true)}
                className="flex items-center gap-1.5 text-xs font-semibold text-indigo-400 hover:text-indigo-300 transition-colors"
              >
                <Plus className="w-3.5 h-3.5" /> Add Zone <ArrowUpRight className="w-3.5 h-3.5" />
              </button>
            </div>
            {error && (
              <div className="p-3 bg-rose-500/10 border border-rose-500/20 text-rose-300 rounded-xl text-xs">{error}</div>
            )}
            <GeospatialMap tenantId={tenantId} branches={branches} />
          </div>
        )}

        {/* Branch Registry */}
        {(activeNav === 'Overview' || activeNav === 'Branch Registry') && (
          <div className="glass-card rounded-2xl border border-slate-800 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-800/80 flex items-center justify-between">
              <h2 className="text-sm font-bold text-white">Branch Network Registry ({branches.length})</h2>
              <div className="flex items-center gap-2">
                <Filter className="w-3.5 h-3.5 text-slate-400" />
                {['ALL', 'SORTING_HUB', 'DISTRIBUTION_CENTER', 'LOCAL_OFFICE'].map((t) => (
                  <button
                    key={t}
                    onClick={() => setFilterType(t)}
                    className={`px-2.5 py-1 rounded-lg text-[10px] font-bold transition-all ${
                      filterType === t
                        ? 'bg-indigo-600 text-white'
                        : 'bg-slate-900 text-slate-400 hover:text-white border border-slate-800'
                    }`}
                  >
                    {t === 'ALL' ? 'All' : t.replace('_', ' ')}
                  </button>
                ))}
              </div>
            </div>
            {loading ? (
              <div className="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {[...Array(3)].map((_, i) => <div key={i} className="skeleton h-36" />)}
              </div>
            ) : (
              <div className="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {filteredBranches.map((b) => (
                  <div
                    key={b.id}
                    className="stat-card glass-card rounded-2xl p-5 border border-slate-800 hover:border-slate-700 flex flex-col justify-between"
                  >
                    <div>
                      <div className="flex items-start justify-between mb-3">
                        <div>
                          <span className="font-mono text-[10px] font-bold px-2 py-0.5 rounded bg-slate-800 text-indigo-400 border border-slate-700">
                            {b.code}
                          </span>
                          <h3 className="font-bold text-sm text-slate-100 mt-1.5">{b.name}</h3>
                        </div>
                        <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full border ${BRANCH_TYPE_COLOR[b.branch_type] || 'bg-slate-800 text-slate-400 border-slate-700'}`}>
                          {b.branch_type.replace('_', ' ')}
                        </span>
                      </div>
                      <div className="space-y-1.5 text-xs text-slate-400">
                        <div className="flex items-start gap-2">
                          <MapPin className="w-3.5 h-3.5 text-slate-500 shrink-0 mt-0.5" />
                          <span>{b.address}, {b.city}, {b.postal_code}</span>
                        </div>
                        {b.contact_phone && (
                          <div className="flex items-center gap-2">
                            <Phone className="w-3.5 h-3.5 text-slate-500" />
                            <span>{b.contact_phone}</span>
                          </div>
                        )}
                        {b.contact_email && (
                          <div className="flex items-center gap-2">
                            <Mail className="w-3.5 h-3.5 text-slate-500" />
                            <span className="truncate">{b.contact_email}</span>
                          </div>
                        )}
                      </div>
                    </div>
                    <div className="mt-4 pt-3 border-t border-slate-800/80">
                      <div className="flex items-center justify-between text-xs mb-2">
                        <span className="text-slate-500">Daily Capacity</span>
                        <span className="font-bold text-slate-200">{b.daily_capacity.toLocaleString()}</span>
                      </div>
                      <div className="h-1.5 bg-slate-800 rounded-full overflow-hidden">
                        <div
                          className="h-full rounded-full bg-indigo-500 progress-fill"
                          style={{ width: `${(b.daily_capacity / maxCap) * 100}%` }}
                        />
                      </div>
                      <div className="mt-2 font-mono text-[10px] text-slate-500 text-right">
                        {b.location.latitude.toFixed(4)}, {b.location.longitude.toFixed(4)}
                      </div>
                    </div>
                  </div>
                ))}
                {filteredBranches.length === 0 && (
                  <div className="col-span-3 py-12 text-center text-slate-500 text-sm">
                    No branches found for this filter.
                  </div>
                )}
              </div>
            )}
          </div>
        )}

      </div>

      <CreateBranchModal
        tenantId={tenantId}
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onCreated={handleBranchCreated}
      />
    </DashboardLayout>
  );
};
