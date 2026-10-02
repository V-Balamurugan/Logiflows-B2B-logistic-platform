import React from 'react';
import { Package, ShieldCheck, MapPin, Truck, Brain, Layers, GitBranch, ArrowUpRight } from 'lucide-react';
import { HealthMonitor } from './components/HealthMonitor';

export const App: React.FC = () => {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col antialiased">
      {/* Top Navigation */}
      <header className="sticky top-0 z-50 glass-panel border-b border-slate-800/80 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 via-indigo-500 to-cyan-400 flex items-center justify-center shadow-lg shadow-indigo-500/20">
              <Package className="w-5 h-5 text-white" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="font-extrabold text-xl tracking-tight text-white">LogiFlows</span>
                <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                  Phase 0 Foundation
                </span>
              </div>
              <p className="text-xs text-slate-400">Smart Postal & Courier Delivery Intelligence</p>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <div className="hidden sm:flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-xs text-slate-300">
              <GitBranch className="w-3.5 h-3.5 text-indigo-400" />
              <span className="font-mono">phase/00-foundation</span>
            </div>
            <a
              href="/api/v1/healthz"
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-md shadow-indigo-600/20"
            >
              API Gateway <ArrowUpRight className="w-3.5 h-3.5" />
            </a>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-6 py-8 space-y-8">
        {/* Hero Banner */}
        <div className="relative overflow-hidden rounded-3xl glass-card p-8 md:p-10 border border-slate-800 glow-indigo">
          <div className="relative z-10 max-w-3xl space-y-4">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">
              <ShieldCheck className="w-3.5 h-3.5 text-indigo-400" />
              <span>Multi-Tenant Architectural Core Established</span>
            </div>
            <h1 className="text-3xl md:text-4xl lg:text-5xl font-extrabold tracking-tight text-white leading-tight">
              AI-Powered Intelligent Logistics & Real-Time Postal Platform
            </h1>
            <p className="text-base text-slate-300 leading-relaxed">
              LogiFlows combines an authoritative <strong>Go backend</strong>, <strong>PostgreSQL + PostGIS</strong> geospatial engine, <strong>Redis</strong> transient coordination, <strong>Python FastAPI</strong> delay risk prediction, and a cross-platform <strong>Flutter</strong> mobile suite.
            </p>
          </div>
        </div>

        {/* Live Infrastructure Health Component */}
        <HealthMonitor />

        {/* Vertical Architectural Slices Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div className="glass-card rounded-2xl p-6 border border-slate-800 hover:border-indigo-500/40 transition-all">
            <div className="w-10 h-10 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400 mb-4">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Authoritative Backend</h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              Written in Go. Authoritative owner of business rules, RBAC, tenant isolation, QR custody state machine, and tracking APIs.
            </p>
          </div>

          <div className="glass-card rounded-2xl p-6 border border-slate-800 hover:border-blue-500/40 transition-all">
            <div className="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-blue-400 mb-4">
              <MapPin className="w-5 h-5" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">PostGIS Geospatial Engine</h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              PostgreSQL 16 with PostGIS extension for polygon serviceability zones, hub distance matrix calculation, and spatial routing.
            </p>
          </div>

          <div className="glass-card rounded-2xl p-6 border border-slate-800 hover:border-purple-500/40 transition-all">
            <div className="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400 mb-4">
              <Brain className="w-5 h-5" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Advisory AI Intelligence</h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              Python FastAPI service providing predictive delay intelligence (LOW, MEDIUM, HIGH) without ever bypassing Go authority.
            </p>
          </div>

          <div className="glass-card rounded-2xl p-6 border border-slate-800 hover:border-rose-500/40 transition-all">
            <div className="w-10 h-10 rounded-xl bg-rose-500/10 border border-rose-500/20 flex items-center justify-center text-rose-400 mb-4">
              <Layers className="w-5 h-5" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Redis Real-Time Pub/Sub</h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              Decoupled transient state handling for courier live GPS points and WebSocket broadcasting with automatic safe degradation.
            </p>
          </div>

          <div className="glass-card rounded-2xl p-6 border border-slate-800 hover:border-cyan-500/40 transition-all">
            <div className="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/20 flex items-center justify-center text-cyan-400 mb-4">
              <Truck className="w-5 h-5" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Flutter Mobile Suite</h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              Dedicated mobile operations client with offline queueing, barcode/QR scanner, real-time GPS telemetry, and Proof of Delivery.
            </p>
          </div>

          <div className="glass-card rounded-2xl p-6 border border-slate-800 hover:border-amber-500/40 transition-all">
            <div className="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400 mb-4">
              <Package className="w-5 h-5" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Single Portal RBAC</h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              Unified web application dynamically projecting role-specific dashboards for Platform Admins, Tenants, Employees, and Customers.
            </p>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-800/80 px-6 py-6 mt-12 bg-slate-950/80">
        <div className="max-w-7xl mx-auto flex flex-col sm:flex-row items-center justify-between text-xs text-slate-500 gap-4">
          <p>© 2026 LogiFlows Logistics Platform. All rights reserved.</p>
          <p className="font-mono">Commit: feat(phase-00): establish LogiFlows foundation</p>
        </div>
      </footer>
    </div>
  );
};

export default App;
