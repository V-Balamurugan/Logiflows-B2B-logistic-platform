import React, { useState, useRef } from 'react';
import {
  QrCode, CheckCircle, Navigation, Clock, Package,
  MapPin, Phone, AlertCircle, Radio, BarChart3, Zap,
  ChevronRight, Camera, PenLine, Route as RouteIcon, UploadCloud, X
} from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import { DashboardLayout } from '../../components/DashboardLayout';
import { getDirections, RouteResult } from '../../services/routingService';
import { uploadProofOfDelivery, ProofOfDeliveryUploadResult } from '../../services/firebaseService';

interface EmployeeDashboardProps {
  onNavigate: (path: string) => void;
}

const NAV = [
  { icon: <BarChart3 className="w-4 h-4" />, label: 'Overview' },
  { icon: <Navigation className="w-4 h-4" />, label: 'Run Sheet', badge: 6 },
  { icon: <Radio className="w-4 h-4" />, label: 'GPS Telemetry' },
  { icon: <QrCode className="w-4 h-4" />, label: 'Scanner' },
  { icon: <PenLine className="w-4 h-4" />, label: 'Proof of Delivery' },
];

type StopStatus = 'pending' | 'in-transit' | 'delivered' | 'failed';

interface Stop {
  id: string;
  trackingId: string;
  recipient: string;
  address: string;
  phone: string;
  priority: 'high' | 'normal' | 'express';
  status: StopStatus;
  eta: string;
  weight: string;
  instructions?: string;
  lat: number;
  lon: number;
}

const STOPS: Stop[] = [
  { id: '1', trackingId: 'LF-8923-4412', recipient: 'Ranjith Kumar', address: '42 Elm Street, Koramangala', phone: '+91 98801 12345', priority: 'express', status: 'in-transit', eta: '10 min', weight: '2.4 kg', lat: 12.9345, lon: 77.6265 },
  { id: '2', trackingId: 'LF-8924-9901', recipient: 'Priya Sharma', address: '8/B, 3rd Cross, Indiranagar', phone: '+91 97722 98765', priority: 'high', status: 'pending', eta: '28 min', weight: '0.8 kg', instructions: 'Call before delivery', lat: 12.9784, lon: 77.6408 },
  { id: '3', trackingId: 'LF-8925-2211', recipient: 'Nataraj P.', address: '77 MG Road, Richmond Circle', phone: '+91 91103 44111', priority: 'normal', status: 'delivered', eta: '—', weight: '5.1 kg', lat: 12.9756, lon: 77.6066 },
  { id: '4', trackingId: 'LF-8926-6650', recipient: 'Ananya Reddy', address: '22 Residency Rd, Shivaji Nagar', phone: '+91 98450 67890', priority: 'normal', status: 'pending', eta: '41 min', weight: '1.2 kg', lat: 12.9719, lon: 77.6012 },
  { id: '5', trackingId: 'LF-8927-3381', recipient: 'Mohammed Iqbal', address: '113A Frazer Town', phone: '+91 96001 23456', priority: 'high', status: 'delivered', eta: '—', weight: '3.0 kg', lat: 12.9982, lon: 77.6124 },
  { id: '6', trackingId: 'LF-8928-7750', recipient: 'Kavitha L.', address: '5 Bannerghatta Rd', phone: '+91 77820 11002', priority: 'express', status: 'failed', eta: 'Retry', weight: '0.5 kg', instructions: 'Gate locked, contact society', lat: 12.8988, lon: 77.5997 },
];

const STATUS_STYLES: Record<StopStatus, { label: string; cls: string; dot: string }> = {
  'pending': { label: 'Pending', cls: 'bg-amber-500/15 text-amber-300 border-amber-500/30', dot: 'bg-amber-400' },
  'in-transit': { label: 'In Transit', cls: 'bg-blue-500/15 text-blue-300 border-blue-500/30', dot: 'bg-blue-400' },
  'delivered': { label: 'Delivered', cls: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30', dot: 'bg-emerald-400' },
  'failed': { label: 'Attempt Failed', cls: 'bg-rose-500/15 text-rose-300 border-rose-500/30', dot: 'bg-rose-400' },
};

const PRIORITY_STYLES: Record<string, string> = {
  express: 'bg-rose-500/20 text-rose-300 border-rose-500/30',
  high: 'bg-amber-500/20 text-amber-300 border-amber-500/30',
  normal: 'bg-slate-800 text-slate-400 border-slate-700',
};

export const EmployeeDashboard: React.FC<EmployeeDashboardProps> = ({ onNavigate }) => {
  const { user } = useAuth();
  const [activeNav, setActiveNav] = useState('Overview');
  const [stops, setStops] = useState<Stop[]>(STOPS);
  const [selected, setSelected] = useState<string | null>(null);

  // Live Navigation State
  const [activeNavRoute, setActiveNavRoute] = useState<RouteResult | null>(null);
  const [navigatingStop, setNavigatingStop] = useState<Stop | null>(null);
  const [isRouting, setIsRouting] = useState(false);

  // Firebase Proof of Delivery State
  const [podUploads, setPodUploads] = useState<Record<string, ProofOfDeliveryUploadResult>>({});
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [uploadingStopId, setUploadingStopId] = useState<string | null>(null);

  const delivered = stops.filter((s) => s.status === 'delivered').length;
  const inTransit = stops.filter((s) => s.status === 'in-transit').length;
  const pending = stops.filter((s) => s.status === 'pending').length;
  const failed = stops.filter((s) => s.status === 'failed').length;
  const progress = Math.round((delivered / stops.length) * 100);

  const markDelivered = (id: string) => {
    setStops((prev) => prev.map((s) => s.id === id ? { ...s, status: 'delivered' as StopStatus, eta: '—' } : s));
    setSelected(null);
  };

  const handleStartNavigation = async (stop: Stop) => {
    setIsRouting(true);
    setNavigatingStop(stop);
    try {
      // Driver origin at Hub (12.9716, 77.5946) to stop location
      const route = await getDirections(
        { latitude: 12.9716, longitude: 77.5946 },
        { latitude: stop.lat, longitude: stop.lon }
      );
      setActiveNavRoute(route);
    } catch (e) {
      console.error(e);
    } finally {
      setIsRouting(false);
    }
  };

  const handleTriggerUpload = (stopId: string) => {
    setUploadingStopId(stopId);
    if (fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file && uploadingStopId) {
      const stop = stops.find((s) => s.id === uploadingStopId);
      const trackingId = stop?.trackingId || uploadingStopId;
      const res = await uploadProofOfDelivery(file, trackingId);
      setPodUploads((prev) => ({ ...prev, [uploadingStopId]: res }));
    }
    setUploadingStopId(null);
  };

  return (
    <DashboardLayout onNavigate={onNavigate} navItems={NAV} activeNav={activeNav} onNavClick={setActiveNav}>
      <div className="p-6 space-y-6 fade-in-up">

        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-extrabold text-white">Courier Operations</h1>
            <p className="text-xs text-slate-400 mt-0.5">
              {user?.first_name} {user?.last_name} • Route optimized by PostGIS
            </p>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold">
              <Radio className="w-3.5 h-3.5 pulse-dot" />
              GPS Active
            </div>
            <a
              href="http://localhost:5174"
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold shadow-md shadow-indigo-600/20 transition-all"
            >
              <Zap className="w-3.5 h-3.5" /> Mobile App
            </a>
          </div>
        </div>

        {/* Progress Banner */}
        <div className="glass-card rounded-2xl p-5 border border-slate-800">
          <div className="flex items-center justify-between mb-3">
            <div>
              <p className="text-sm font-bold text-white">Today's Delivery Progress</p>
              <p className="text-xs text-slate-400">Route: Koramangala → MG Road → Hebbal circuit</p>
            </div>
            <span className="text-2xl font-black text-emerald-400">{progress}%</span>
          </div>
          <div className="h-2.5 bg-slate-800 rounded-full overflow-hidden">
            <div
              className="h-full rounded-full bg-gradient-to-r from-emerald-600 to-emerald-400 progress-fill"
              style={{ width: `${progress}%` }}
            />
          </div>
          <div className="flex items-center gap-4 mt-3 text-xs">
            <div className="flex items-center gap-1.5"><span className="w-2 h-2 rounded-full bg-emerald-400" /><span className="text-slate-400">{delivered} Delivered</span></div>
            <div className="flex items-center gap-1.5"><span className="w-2 h-2 rounded-full bg-blue-400" /><span className="text-slate-400">{inTransit} In Transit</span></div>
            <div className="flex items-center gap-1.5"><span className="w-2 h-2 rounded-full bg-amber-400" /><span className="text-slate-400">{pending} Pending</span></div>
            {failed > 0 && <div className="flex items-center gap-1.5"><span className="w-2 h-2 rounded-full bg-rose-400" /><span className="text-slate-400">{failed} Failed</span></div>}
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: "Assigned Stops", value: stops.length, icon: <Navigation className="w-4 h-4" />, color: 'indigo', sub: 'Total today' },
            { label: 'Delivered', value: delivered, icon: <CheckCircle className="w-4 h-4" />, color: 'emerald', sub: 'POD captured' },
            { label: 'Remaining', value: pending + inTransit, icon: <Clock className="w-4 h-4" />, color: 'amber', sub: 'In queue' },
            { label: 'Est. Earnings', value: '₹480', icon: <Package className="w-4 h-4" />, color: 'purple', sub: 'Performance bonus eligible' },
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

        {/* Run Sheet */}
        <div className="glass-card rounded-2xl border border-slate-800 overflow-hidden">
          <div className="px-6 py-4 border-b border-slate-800/80 flex items-center justify-between">
            <div>
              <h2 className="text-sm font-bold text-white">Delivery Run Sheet</h2>
              <p className="text-[11px] text-slate-400 mt-0.5">PostGIS optimized sequence</p>
            </div>
            <div className="flex items-center gap-1.5 text-[10px] font-bold text-indigo-300 bg-indigo-500/10 px-3 py-1.5 rounded-xl border border-indigo-500/20">
              <Navigation className="w-3.5 h-3.5" /> Route Active
            </div>
          </div>

          <div className="divide-y divide-slate-800/60">
            {stops.map((stop, idx) => {
              const st = STATUS_STYLES[stop.status];
              const isSelected = selected === stop.id;
              return (
                <div key={stop.id} className={`transition-all ${isSelected ? 'bg-indigo-500/5' : 'hover:bg-slate-800/20'}`}>
                  <button
                    className="w-full px-5 py-4 flex items-start gap-4 text-left"
                    onClick={() => setSelected(isSelected ? null : stop.id)}
                  >
                    {/* Sequence Number */}
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-black shrink-0 mt-0.5 ${
                      stop.status === 'delivered' ? 'bg-emerald-500/20 text-emerald-400' :
                      stop.status === 'in-transit' ? 'bg-blue-500/20 text-blue-400' :
                      stop.status === 'failed' ? 'bg-rose-500/20 text-rose-400' :
                      'bg-slate-800 text-slate-400'
                    }`}>{idx + 1}</div>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-mono text-[10px] font-bold text-indigo-400">{stop.trackingId}</span>
                          <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full border ${PRIORITY_STYLES[stop.priority]}`}>
                            {stop.priority.toUpperCase()}
                          </span>
                          <span className={`text-[10px] font-bold px-2 py-0.5 rounded-full border flex items-center gap-1 ${st.cls}`}>
                            <span className={`w-1.5 h-1.5 rounded-full ${st.dot}`} />
                            {st.label}
                          </span>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          {stop.eta !== '—' && (
                            <div className="flex items-center gap-1 text-amber-400 text-xs font-semibold">
                              <Clock className="w-3.5 h-3.5" /> {stop.eta}
                            </div>
                          )}
                          <ChevronRight className={`w-4 h-4 text-slate-500 transition-transform ${isSelected ? 'rotate-90' : ''}`} />
                        </div>
                      </div>
                      <p className="text-sm font-semibold text-white mt-1">{stop.recipient}</p>
                      <div className="flex items-center gap-1.5 text-xs text-slate-400 mt-0.5">
                        <MapPin className="w-3 h-3 text-slate-500 shrink-0" />
                        <span>{stop.address}</span>
                        <span className="text-slate-600">•</span>
                        <span>{stop.weight}</span>
                      </div>
                      {stop.instructions && (
                        <div className="flex items-center gap-1.5 mt-1 text-[11px] text-amber-400">
                          <AlertCircle className="w-3 h-3" />
                          {stop.instructions}
                        </div>
                      )}

                      {/* POD Upload Badge */}
                      {podUploads[stop.id] && (
                        <div className="flex items-center gap-2 mt-2 p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs">
                          <UploadCloud className="w-4 h-4 text-emerald-400 shrink-0" />
                          <div>
                            <span className="font-bold">Proof of Delivery Captured:</span>{' '}
                            <span className="font-mono text-[11px]">{podUploads[stop.id].storage_path.split('/').pop()}</span>{' '}
                            <span className="text-[10px] text-emerald-400/80">({podUploads[stop.id].mode === 'FIREBASE_STORAGE' ? 'Firebase Cloud' : 'Device Storage'})</span>
                          </div>
                        </div>
                      )}
                    </div>
                  </button>

                  {/* Expanded Actions */}
                  {isSelected && stop.status !== 'delivered' && (
                    <div className="px-5 pb-4 flex items-center gap-2 flex-wrap">
                      <button
                        onClick={() => handleStartNavigation(stop)}
                        disabled={isRouting}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold bg-sky-600 hover:bg-sky-500 text-white shadow-md shadow-sky-600/20 transition-all"
                      >
                        <RouteIcon className="w-3.5 h-3.5" />
                        {isRouting && navigatingStop?.id === stop.id ? 'Routing...' : 'Navigate Route'}
                      </button>
                      <button
                        onClick={() => handleTriggerUpload(stop.id)}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-slate-900 text-slate-300 border border-slate-800 hover:border-slate-700 transition-all"
                      >
                        <Camera className="w-3.5 h-3.5" />
                        {podUploads[stop.id] ? 'Re-take POD' : 'Capture POD Photo'}
                      </button>
                      <a
                        href={`tel:${stop.phone}`}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-slate-900 text-slate-300 border border-slate-800 hover:border-slate-700 transition-all"
                      >
                        <Phone className="w-3.5 h-3.5" /> Call Recipient
                      </a>
                      <button
                        onClick={() => markDelivered(stop.id)}
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-bold bg-emerald-600 hover:bg-emerald-500 text-white shadow-md shadow-emerald-600/20 transition-all"
                      >
                        <CheckCircle className="w-3.5 h-3.5" /> Mark Delivered
                      </button>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Hidden File Input for POD Photo */}
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          capture="environment"
          onChange={handleFileChange}
          className="hidden"
        />

        {/* Live Turn-by-Turn Navigation Modal/Drawer */}
        {activeNavRoute && navigatingStop && (
          <div className="glass-card rounded-2xl p-5 border border-sky-500/30 bg-slate-950/90 shadow-2xl relative">
            <div className="flex items-start justify-between gap-4 mb-4">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-sky-500/20 text-sky-400 flex items-center justify-center border border-sky-500/30">
                  <Navigation className="w-5 h-5 animate-pulse" />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-bold text-white">Active Route to {navigatingStop.recipient}</h3>
                    <span className="text-[10px] px-2 py-0.5 rounded-full font-mono bg-sky-950 text-sky-300 border border-sky-800">
                      {activeNavRoute.source === 'LIVE_OPENROUTESERVICE' ? 'Live ORS Engine' : 'LogiFlows Spatial'}
                    </span>
                  </div>
                  <p className="text-xs text-slate-400 mt-0.5">{navigatingStop.address}</p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <div className="text-right">
                  <div className="text-sm font-bold text-emerald-400">{activeNavRoute.eta_formatted}</div>
                  <div className="text-[11px] text-slate-400">{activeNavRoute.distance_km} km total</div>
                </div>
                <button
                  onClick={() => setActiveNavRoute(null)}
                  className="p-1.5 rounded-xl bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 transition-all"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Step-by-Step Directions */}
            <div className="space-y-2 max-h-48 overflow-y-auto pr-1">
              {activeNavRoute.steps.map((step, idx) => (
                <div key={idx} className="flex items-center justify-between p-2.5 rounded-xl bg-slate-900/60 border border-slate-800 text-xs">
                  <div className="flex items-center gap-2.5">
                    <span className="w-5 h-5 rounded-full bg-slate-800 text-slate-300 font-bold flex items-center justify-center text-[10px]">
                      {idx + 1}
                    </span>
                    <span className="text-slate-200">{step.instruction}</span>
                  </div>
                  <span className="text-[11px] font-mono text-slate-400 shrink-0">
                    {(step.distance_meters / 1000).toFixed(1)} km
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}

      </div>
    </DashboardLayout>
  );
};
