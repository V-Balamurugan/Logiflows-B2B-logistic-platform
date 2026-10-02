import React, { useEffect, useState } from 'react';
import { Activity, CheckCircle2, AlertTriangle, RefreshCw, Server, Database, Brain, Radio } from 'lucide-react';

interface ReadinessData {
  status: string;
  dependencies: Record<string, string>;
  timestamp: string;
}

interface LivenessData {
  status: string;
  version: string;
  uptime_seconds: number;
}

export const HealthMonitor: React.FC = () => {
  const [liveness, setLiveness] = useState<LivenessData | null>(null);
  const [readiness, setReadiness] = useState<ReadinessData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchHealth = async () => {
    setLoading(true);
    setError(null);
    try {
      const [liveRes, readyRes] = await Promise.all([
        fetch('/api/v1/healthz'),
        fetch('/api/v1/readyz')
      ]);

      if (!liveRes.ok || !readyRes.ok) {
        throw new Error(`API responded with status: Live (${liveRes.status}), Ready (${readyRes.status})`);
      }

      const liveJson = await liveRes.json();
      const readyJson = await readyRes.json();

      setLiveness(liveJson.data);
      setReadiness(readyJson.data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to reach API gateway');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHealth();
    const interval = setInterval(fetchHealth, 15000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div className="glass-card rounded-2xl p-6 border border-slate-800">
      <div className="flex items-center justify-between pb-4 border-b border-slate-800/80 mb-6">
        <div className="flex items-center space-x-3">
          <div className="p-2.5 bg-indigo-500/10 rounded-xl text-indigo-400 border border-indigo-500/20">
            <Activity className="w-5 h-5 animate-pulse" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white tracking-tight">System Infrastructure Pulse</h2>
            <p className="text-xs text-slate-400">Phase 0 Foundation • Real-Time Core Telemetry</p>
          </div>
        </div>
        <button
          onClick={fetchHealth}
          disabled={loading}
          className="flex items-center space-x-2 px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-300 transition-colors border border-slate-700"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          <span>Refresh</span>
        </button>
      </div>

      {error ? (
        <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-300 flex items-start space-x-3">
          <AlertTriangle className="w-5 h-5 flex-shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-semibold">Backend Unreachable (Standby Mode)</p>
            <p className="text-xs text-amber-400/80 mt-1">{error}</p>
            <p className="text-xs text-slate-400 mt-2">Ensure Go backend is running locally on port 8080 (<code>go run ./cmd/api</code>).</p>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {/* Go Backend Authoritative Node */}
          <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex flex-col justify-between">
            <div className="flex items-center justify-between mb-3">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                <Server className="w-3.5 h-3.5 text-indigo-400" /> Go Backend
              </span>
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                <CheckCircle2 className="w-3 h-3" /> {liveness?.status ?? 'ONLINE'}
              </span>
            </div>
            <div>
              <p className="text-lg font-bold text-white">{liveness?.version ?? '0.0.1-foundation'}</p>
              <p className="text-xs text-slate-400 mt-0.5">Authoritative Business Engine</p>
            </div>
          </div>

          {/* PostgreSQL & PostGIS Spatial DB */}
          <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex flex-col justify-between">
            <div className="flex items-center justify-between mb-3">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                <Database className="w-3.5 h-3.5 text-blue-400" /> PostGIS
              </span>
              <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium ${readiness?.dependencies?.postgres?.includes('UP') ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-slate-800 text-slate-400'}`}>
                {readiness?.dependencies?.postgres ?? 'READY'}
              </span>
            </div>
            <div>
              <p className="text-sm font-semibold text-white">PostgreSQL 16</p>
              <p className="text-xs text-slate-400 mt-0.5">Spatial & Transactional Truth</p>
            </div>
          </div>

          {/* Redis Coordination */}
          <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex flex-col justify-between">
            <div className="flex items-center justify-between mb-3">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                <Radio className="w-3.5 h-3.5 text-rose-400" /> Redis 7
              </span>
              <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium ${readiness?.dependencies?.redis?.includes('UP') ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-slate-800 text-slate-400'}`}>
                {readiness?.dependencies?.redis ?? 'READY'}
              </span>
            </div>
            <div>
              <p className="text-sm font-semibold text-white">Pub/Sub & Caching</p>
              <p className="text-xs text-slate-400 mt-0.5">Transient GPS Telemetry</p>
            </div>
          </div>

          {/* Python AI Advisory Engine */}
          <div className="p-4 rounded-xl bg-slate-900/60 border border-slate-800 flex flex-col justify-between">
            <div className="flex items-center justify-between mb-3">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                <Brain className="w-3.5 h-3.5 text-purple-400" /> Python AI
              </span>
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium bg-purple-500/10 text-purple-300 border border-purple-500/20">
                ADVISORY
              </span>
            </div>
            <div>
              <p className="text-sm font-semibold text-white">FastAPI Engine</p>
              <p className="text-xs text-slate-400 mt-0.5">Predictive Delay Intelligence</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
