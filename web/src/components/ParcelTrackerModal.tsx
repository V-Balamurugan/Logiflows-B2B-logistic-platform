import React, { useState, useEffect } from 'react';
import {
  X, Search, Package, MapPin, CheckCircle2,
  AlertCircle, ArrowRight, Loader2, Calendar
} from 'lucide-react';

interface ParcelTrackerModalProps {
  isOpen: boolean;
  onClose: () => void;
  initialTrackingNumber?: string;
}

export const ParcelTrackerModal: React.FC<ParcelTrackerModalProps> = ({
  isOpen,
  onClose,
  initialTrackingNumber = '',
}) => {
  const [trackingNumber, setTrackingNumber] = useState(initialTrackingNumber);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [trackingData, setTrackingData] = useState<any | null>(null);

  useEffect(() => {
    if (initialTrackingNumber) {
      setTrackingNumber(initialTrackingNumber);
      fetchTracking(initialTrackingNumber);
    }
  }, [initialTrackingNumber]);

  if (!isOpen) return null;

  const fetchTracking = async (numberToTrack: string) => {
    if (!numberToTrack.trim()) return;
    setLoading(true);
    setError(null);

    try {
      const res = await fetch(`/api/v1/parcels/track/${encodeURIComponent(numberToTrack.trim())}`);
      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error?.message || 'Shipment not found');
      }
      setTrackingData(data.data);
    } catch (err: any) {
      setError(err.message || 'Unable to retrieve tracking information');
      setTrackingData(null);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    fetchTracking(trackingNumber);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'DELIVERED':
        return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20';
      case 'OUT_FOR_DELIVERY':
      case 'IN_TRANSIT':
        return 'text-blue-400 bg-blue-500/10 border-blue-500/20';
      case 'CANCELLED':
      case 'FAILED':
        return 'text-rose-400 bg-rose-500/10 border-rose-500/20';
      default:
        return 'text-amber-400 bg-amber-500/10 border-amber-500/20';
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-fade-in">
      <div className="relative w-full max-w-2xl max-h-[90vh] overflow-y-auto glass-card rounded-3xl border border-slate-700/60 shadow-2xl bg-slate-900/95 text-slate-100 p-6 md:p-8">
        {/* Header */}
        <div className="flex items-center justify-between pb-4 border-b border-slate-800">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-2xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400">
              <Search className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">Live Consignment Tracker</h2>
              <p className="text-xs text-slate-400">Real-time custody tracking across hubs &amp; couriers</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="w-9 h-9 rounded-xl bg-slate-800/60 hover:bg-slate-800 flex items-center justify-center text-slate-400 hover:text-white transition-all"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Search Bar */}
        <form onSubmit={handleSearch} className="flex gap-2.5 mt-5">
          <div className="relative flex-1">
            <div className="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-500">
              <Package className="w-4 h-4" />
            </div>
            <input
              type="text"
              value={trackingNumber}
              onChange={(e) => setTrackingNumber(e.target.value)}
              placeholder="Enter tracking number (e.g. LF-20261006-XXXXXX)"
              className="w-full pl-10 pr-4 py-2.5 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-purple-500/50"
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="px-5 py-2.5 rounded-xl bg-purple-600 hover:bg-purple-500 text-xs font-bold text-white flex items-center gap-1.5 transition-all disabled:opacity-50"
          >
            {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Search className="w-4 h-4" />}
            Track
          </button>
        </form>

        {error && (
          <div className="mt-4 p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs flex items-center gap-2.5">
            <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
            <span>{error}</span>
          </div>
        )}

        {/* Tracking Details & Timeline */}
        {trackingData && (
          <div className="mt-6 space-y-6 animate-fade-in">
            {/* Status Card */}
            <div className="p-4 rounded-2xl bg-slate-950/70 border border-slate-800 flex items-center justify-between">
              <div>
                <span className="text-[10px] text-slate-500 uppercase font-bold tracking-wider">Current Status</span>
                <p className="font-mono text-sm font-bold text-white mt-0.5">{trackingData.tracking_number}</p>
                <p className="text-xs text-slate-400 mt-1 flex items-center gap-1">
                  <span>{trackingData.origin_city || 'Origin'}</span>
                  <ArrowRight className="w-3 h-3 text-slate-600" />
                  <span>{trackingData.destination_city || 'Destination'}</span>
                </p>
              </div>
              <div className={`px-3 py-1.5 rounded-xl text-xs font-bold border ${getStatusColor(trackingData.status)}`}>
                {trackingData.status}
              </div>
            </div>

            {/* Estimated Delivery */}
            {trackingData.estimated_delivery_at && (
              <div className="p-3.5 rounded-xl bg-slate-900 border border-slate-800 text-xs flex items-center justify-between">
                <span className="text-slate-400 flex items-center gap-1.5">
                  <Calendar className="w-3.5 h-3.5 text-indigo-400" /> Estimated Delivery:
                </span>
                <span className="font-semibold text-emerald-400">
                  {new Date(trackingData.estimated_delivery_at).toLocaleDateString('en-US', {
                    month: 'short',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </span>
              </div>
            )}

            {/* Custody Timeline */}
            <div>
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-3">Custody Event Timeline</h3>
              <div className="space-y-3">
                {trackingData.timeline && trackingData.timeline.length > 0 ? (
                  trackingData.timeline.map((evt: any, i: number) => (
                    <div key={i} className="flex items-start gap-3 p-3 rounded-xl bg-slate-950/40 border border-slate-800/80">
                      <div className="w-7 h-7 rounded-lg bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 flex items-center justify-center shrink-0 mt-0.5">
                        <CheckCircle2 className="w-4 h-4" />
                      </div>
                      <div className="flex-1 text-xs">
                        <div className="flex items-center justify-between">
                          <span className="font-bold text-white">{evt.event_type}</span>
                          <span className="text-[10px] text-slate-500">
                            {new Date(evt.recorded_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </span>
                        </div>
                        <p className="text-slate-400 text-[11px] mt-0.5">{evt.notes || 'Status confirmed'}</p>
                        {evt.location_name && (
                          <p className="text-[10px] text-slate-500 mt-1 flex items-center gap-1">
                            <MapPin className="w-3 h-3 text-slate-600" /> {evt.location_name}
                          </p>
                        )}
                      </div>
                    </div>
                  ))
                ) : (
                  <p className="text-xs text-slate-500 italic">No custody handover events logged yet.</p>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
