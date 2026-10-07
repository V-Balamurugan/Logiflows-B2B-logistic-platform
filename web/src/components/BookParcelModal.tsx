import React, { useState } from 'react';
import {
  X, Package, MapPin, Scale, CheckCircle2,
  AlertCircle, ArrowRight, Loader2, Copy, Check
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';

interface BookParcelModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: (newParcel: any) => void;
}

export const BookParcelModal: React.FC<BookParcelModalProps> = ({ isOpen, onClose, onSuccess }) => {
  const { tokens } = useAuth();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdParcel, setCreatedParcel] = useState<any | null>(null);
  const [copied, setCopied] = useState(false);

  // Form State
  const [serviceType, setServiceType] = useState('STANDARD');
  const [weightKg, setWeightKg] = useState('1.5');
  const [lengthCm, setLengthCm] = useState('15');
  const [widthCm, setWidthCm] = useState('12');
  const heightCm = '10';
  const [declaredValue, setDeclaredValue] = useState('1500');
  const [isFragile, setIsFragile] = useState(false);
  const specialInstructions = '';

  // Sender Details
  const [senderName, setSenderName] = useState('Logistics Partner Hub');
  const senderPhone = '+91 98765 43210';
  const senderEmail = '';
  const [senderAddress, setSenderAddress] = useState('100 Industrial Area, Phase 2');
  const [senderCity, setSenderCity] = useState('Bengaluru');
  const senderState = 'Karnataka';
  const [senderPostalCode, setSenderPostalCode] = useState('560001');

  // Recipient Details
  const [recipientName, setRecipientName] = useState('');
  const recipientPhone = '+91 98765 43210';
  const recipientEmail = '';
  const [recipientAddress, setRecipientAddress] = useState('');
  const [recipientCity, setRecipientCity] = useState('');
  const recipientState = 'Karnataka';
  const [recipientPostalCode, setRecipientPostalCode] = useState('');

  if (!isOpen) return null;

  // Real-time estimated shipping cost
  const calculateEstimatedCost = () => {
    const w = parseFloat(weightKg) || 1.0;
    let base = 80.0;
    if (serviceType === 'SAME_DAY') base = 250.0;
    if (serviceType === 'EXPRESS') base = 150.0;
    if (serviceType === 'COLD_CHAIN') base = 300.0;
    if (serviceType === 'FREIGHT') base = 500.0;

    if (w > 1.0) {
      base += (w - 1.0) * 40.0;
    }
    return Math.round(base);
  };

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!recipientName.trim() || !recipientAddress.trim() || !recipientCity.trim()) {
      setError('Recipient Name, Address, and City are required.');
      return;
    }

    setLoading(true);
    try {
      // Client-generated UUID for deduplication
      const idempotencyKey = `idemp-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;

      const payload = {
        service_type: serviceType,
        sender_name: senderName,
        sender_phone: senderPhone,
        sender_email: senderEmail,
        sender_address: senderAddress,
        sender_city: senderCity,
        sender_state: senderState,
        sender_postal_code: senderPostalCode,
        recipient_name: recipientName,
        recipient_phone: recipientPhone,
        recipient_email: recipientEmail,
        recipient_address: recipientAddress,
        recipient_city: recipientCity,
        recipient_state: recipientState,
        recipient_postal_code: recipientPostalCode,
        weight_kg: parseFloat(weightKg) || 1.0,
        length_cm: parseFloat(lengthCm) || 10.0,
        width_cm: parseFloat(widthCm) || 10.0,
        height_cm: parseFloat(heightCm) || 10.0,
        declared_value: parseFloat(declaredValue) || 0.0,
        currency: 'INR',
        is_fragile: isFragile,
        special_instructions: specialInstructions,
        idempotency_key: idempotencyKey,
      };

      const res = await fetch('/api/v1/parcels/book', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${tokens?.access_token}`,
        },
        body: JSON.stringify(payload),
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error?.message || 'Failed to book parcel');
      }

      setCreatedParcel(data.data);
      onSuccess(data.data);
    } catch (err: any) {
      setError(err.message || 'An unexpected error occurred during booking');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-fade-in">
      <div className="relative w-full max-w-3xl max-h-[90vh] overflow-y-auto glass-card rounded-3xl border border-slate-700/60 shadow-2xl bg-slate-900/95 text-slate-100 p-6 md:p-8">
        {/* Header */}
        <div className="flex items-center justify-between pb-5 border-b border-slate-800">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400">
              <Package className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-lg font-bold text-white">Book Consignment</h2>
              <p className="text-xs text-slate-400">Create authoritative parcel booking with signed QR identity</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="w-9 h-9 rounded-xl bg-slate-800/60 hover:bg-slate-800 flex items-center justify-center text-slate-400 hover:text-white transition-all"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {createdParcel ? (
          /* Confirmation View */
          <div className="py-6 space-y-6 text-center animate-fade-in">
            <div className="w-16 h-16 rounded-3xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center mx-auto">
              <CheckCircle2 className="w-8 h-8" />
            </div>
            <div>
              <h3 className="text-xl font-bold text-white">Consignment Successfully Booked!</h3>
              <p className="text-xs text-slate-400 mt-1">Status: Initial state set to CREATED</p>
            </div>

            {/* Tracking Badge */}
            <div className="p-4 rounded-2xl bg-slate-950/70 border border-slate-800 flex items-center justify-between max-w-md mx-auto">
              <div className="text-left">
                <span className="text-[10px] font-bold text-slate-500 uppercase">Authoritative Tracking ID</span>
                <p className="text-lg font-mono font-bold text-indigo-400">{createdParcel.tracking_number}</p>
              </div>
              <button
                onClick={() => handleCopy(createdParcel.tracking_number)}
                className="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-xs font-semibold text-slate-200 flex items-center gap-1.5 transition-all"
              >
                {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                {copied ? 'Copied' : 'Copy'}
              </button>
            </div>

            {/* Details Summary */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 max-w-lg mx-auto text-left text-xs">
              <div className="p-3 rounded-xl bg-slate-900 border border-slate-800">
                <span className="text-slate-500 block text-[10px]">Service</span>
                <span className="font-semibold text-white">{createdParcel.service_type}</span>
              </div>
              <div className="p-3 rounded-xl bg-slate-900 border border-slate-800">
                <span className="text-slate-500 block text-[10px]">Weight</span>
                <span className="font-semibold text-white">{createdParcel.weight_kg} kg</span>
              </div>
              <div className="p-3 rounded-xl bg-slate-900 border border-slate-800">
                <span className="text-slate-500 block text-[10px]">Estimated Cost</span>
                <span className="font-semibold text-emerald-400">₹{createdParcel.shipping_cost}</span>
              </div>
              <div className="p-3 rounded-xl bg-slate-900 border border-slate-800">
                <span className="text-slate-500 block text-[10px]">QR Token</span>
                <span className="font-semibold text-slate-300 font-mono text-[10px] truncate block" title={createdParcel.qr_token}>
                  {createdParcel.qr_token?.slice(0, 10)}...
                </span>
              </div>
            </div>

            <div className="pt-4 flex justify-center gap-3">
              <button
                onClick={() => {
                  setCreatedParcel(null);
                  setRecipientName('');
                  setRecipientAddress('');
                }}
                className="px-5 py-2.5 rounded-xl text-xs font-bold bg-slate-800 hover:bg-slate-700 text-white transition-all"
              >
                Book Another Parcel
              </button>
              <button
                onClick={onClose}
                className="px-6 py-2.5 rounded-xl text-xs font-bold bg-indigo-600 hover:bg-indigo-500 text-white shadow-lg shadow-indigo-600/30 transition-all"
              >
                Done
              </button>
            </div>
          </div>
        ) : (
          /* Booking Form */
          <form onSubmit={handleSubmit} className="space-y-6 pt-5">
            {error && (
              <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs flex items-center gap-2.5">
                <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
                <span>{error}</span>
              </div>
            )}

            {/* Service Selection */}
            <div>
              <label className="block text-xs font-bold uppercase tracking-wider text-slate-400 mb-2">Service Tier</label>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
                {[
                  { id: 'STANDARD', label: 'Standard', sub: '3-4 Business Days', color: 'slate' },
                  { id: 'EXPRESS', label: 'Express', sub: 'Next Day Delivery', color: 'indigo' },
                  { id: 'SAME_DAY', label: 'Same Day', sub: 'Within 12 Hours', color: 'emerald' },
                  { id: 'COLD_CHAIN', label: 'Cold Chain', sub: 'Temp Controlled', color: 'sky' },
                ].map((s) => (
                  <button
                    key={s.id}
                    type="button"
                    onClick={() => setServiceType(s.id)}
                    className={`p-3 rounded-2xl text-left border transition-all ${
                      serviceType === s.id
                        ? 'bg-indigo-600/20 border-indigo-500 ring-1 ring-indigo-500 text-white'
                        : 'bg-slate-900/60 border-slate-800 hover:border-slate-700 text-slate-300'
                    }`}
                  >
                    <p className="font-bold text-xs">{s.label}</p>
                    <p className="text-[10px] text-slate-400 mt-0.5">{s.sub}</p>
                  </button>
                ))}
              </div>
            </div>

            {/* Sender and Recipient Address Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
              {/* Sender Details */}
              <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800 space-y-3">
                <span className="text-[11px] font-bold text-indigo-400 uppercase tracking-wider flex items-center gap-1.5">
                  <MapPin className="w-3.5 h-3.5" /> Pickup / Sender Origin
                </span>
                <div>
                  <label className="text-[10px] text-slate-400">Sender Name / Entity</label>
                  <input
                    type="text"
                    value={senderName}
                    onChange={(e) => setSenderName(e.target.value)}
                    required
                    className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-[10px] text-slate-400">Pickup Address</label>
                  <input
                    type="text"
                    value={senderAddress}
                    onChange={(e) => setSenderAddress(e.target.value)}
                    required
                    className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="text-[10px] text-slate-400">City</label>
                    <input
                      type="text"
                      value={senderCity}
                      onChange={(e) => setSenderCity(e.target.value)}
                      required
                      className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                    />
                  </div>
                  <div>
                    <label className="text-[10px] text-slate-400">Postal Code</label>
                    <input
                      type="text"
                      value={senderPostalCode}
                      onChange={(e) => setSenderPostalCode(e.target.value)}
                      required
                      className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                    />
                  </div>
                </div>
              </div>

              {/* Recipient Details */}
              <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800 space-y-3">
                <span className="text-[11px] font-bold text-emerald-400 uppercase tracking-wider flex items-center gap-1.5">
                  <MapPin className="w-3.5 h-3.5" /> Destination / Recipient
                </span>
                <div>
                  <label className="text-[10px] text-slate-400">Recipient Full Name *</label>
                  <input
                    type="text"
                    placeholder="e.g. Priya Sharma"
                    value={recipientName}
                    onChange={(e) => setRecipientName(e.target.value)}
                    required
                    className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-[10px] text-slate-400">Delivery Address *</label>
                  <input
                    type="text"
                    placeholder="Street line, Apartment / Suite"
                    value={recipientAddress}
                    onChange={(e) => setRecipientAddress(e.target.value)}
                    required
                    className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="text-[10px] text-slate-400">City *</label>
                    <input
                      type="text"
                      placeholder="e.g. Chennai"
                      value={recipientCity}
                      onChange={(e) => setRecipientCity(e.target.value)}
                      required
                      className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                    />
                  </div>
                  <div>
                    <label className="text-[10px] text-slate-400">Postal Code *</label>
                    <input
                      type="text"
                      placeholder="e.g. 600001"
                      value={recipientPostalCode}
                      onChange={(e) => setRecipientPostalCode(e.target.value)}
                      required
                      className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                    />
                  </div>
                </div>
              </div>
            </div>

            {/* Package Specifications */}
            <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800 space-y-3">
              <span className="text-[11px] font-bold text-amber-400 uppercase tracking-wider flex items-center gap-1.5">
                <Scale className="w-3.5 h-3.5" /> Package Specifications &amp; Value
              </span>
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                <div>
                  <label className="text-[10px] text-slate-400">Weight (KG)</label>
                  <input
                    type="number"
                    step="0.1"
                    min="0.1"
                    value={weightKg}
                    onChange={(e) => setWeightKg(e.target.value)}
                    className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-[10px] text-slate-400">Length (CM)</label>
                  <input
                    type="number"
                    value={lengthCm}
                    onChange={(e) => setLengthCm(e.target.value)}
                    className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-[10px] text-slate-400">Width (CM)</label>
                  <input
                    type="number"
                    value={widthCm}
                    onChange={(e) => setWidthCm(e.target.value)}
                    className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
                <div>
                  <label className="text-[10px] text-slate-400">Declared Value (₹)</label>
                  <input
                    type="number"
                    value={declaredValue}
                    onChange={(e) => setDeclaredValue(e.target.value)}
                    className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-xs text-white"
                  />
                </div>
              </div>

              <div className="flex items-center gap-2 pt-1">
                <input
                  type="checkbox"
                  id="fragile-check"
                  checked={isFragile}
                  onChange={(e) => setIsFragile(e.target.checked)}
                  className="rounded bg-slate-950 border-slate-800 text-indigo-500 focus:ring-0"
                />
                <label htmlFor="fragile-check" className="text-xs text-slate-300">
                  Fragile / High-Care Package
                </label>
              </div>
            </div>

            {/* Bottom Bar: Price Quote + Submit */}
            <div className="flex items-center justify-between pt-2">
              <div className="flex items-center gap-2">
                <span className="text-xs text-slate-400">Estimated Cost:</span>
                <span className="text-lg font-black text-emerald-400">₹{calculateEstimatedCost()}</span>
                <span className="text-[10px] text-slate-500">incl. GST</span>
              </div>
              <div className="flex gap-2.5">
                <button
                  type="button"
                  onClick={onClose}
                  className="px-4 py-2.5 rounded-xl text-xs font-bold text-slate-400 hover:text-white bg-slate-800/60 transition-all"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  className="px-6 py-2.5 rounded-xl text-xs font-bold bg-indigo-600 hover:bg-indigo-500 text-white shadow-lg shadow-indigo-600/30 flex items-center gap-2 transition-all disabled:opacity-50"
                >
                  {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <ArrowRight className="w-4 h-4" />}
                  Confirm Booking
                </button>
              </div>
            </div>
          </form>
        )}
      </div>
    </div>
  );
};
