import React, { useState, useEffect } from 'react';
import { MapPin, Plus, Trash2, Loader2, AlertCircle, Building } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export interface SavedAddress {
  id: string;
  label: string;
  contact_name: string;
  contact_phone: string;
  contact_email?: string;
  street_line1: string;
  street_line2?: string;
  city: string;
  state: string;
  postal_code: string;
  country: string;
  is_default_pickup: boolean;
  is_default_delivery: boolean;
}

export const AddressBookView: React.FC = () => {
  const { tokens } = useAuth();
  const [addresses, setAddresses] = useState<SavedAddress[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAddForm, setShowAddForm] = useState(false);

  // New Address Form State
  const [label, setLabel] = useState('Office');
  const [contactName, setContactName] = useState('');
  const [contactPhone, setContactPhone] = useState('');
  const [streetLine1, setStreetLine1] = useState('');
  const [city, setCity] = useState('');
  const [state, setState] = useState('Karnataka');
  const [postalCode, setPostalCode] = useState('');
  const [isDefaultPickup, setIsDefaultPickup] = useState(false);

  const fetchAddresses = async () => {
    if (!tokens?.access_token) return;
    setLoading(true);
    try {
      const res = await fetch('/api/v1/customers/addresses', {
        headers: { Authorization: `Bearer ${tokens.access_token}` },
      });
      const data = await res.json();
      if (res.ok) {
        setAddresses(data.data || []);
      }
    } catch {
      setError('Failed to load address book');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAddresses();
  }, [tokens]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!contactName.trim() || !streetLine1.trim() || !city.trim() || !postalCode.trim()) {
      setError('Please fill in all required address fields.');
      return;
    }

    try {
      const res = await fetch('/api/v1/customers/addresses', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${tokens?.access_token}`,
        },
        body: JSON.stringify({
          label,
          contact_name: contactName,
          contact_phone: contactPhone,
          street_line1: streetLine1,
          city,
          state,
          postal_code: postalCode,
          is_default_pickup: isDefaultPickup,
        }),
      });

      if (res.ok) {
        setShowAddForm(false);
        setContactName('');
        setStreetLine1('');
        setPostalCode('');
        fetchAddresses();
      } else {
        const errData = await res.json();
        setError(errData.error?.message || 'Failed to save address');
      }
    } catch {
      setError('Failed to create address');
    }
  };

  const handleDelete = async (id: string) => {
    try {
      const res = await fetch(`/api/v1/customers/addresses/${id}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${tokens?.access_token}` },
      });
      if (res.ok) {
        setAddresses((prev) => prev.filter((a) => a.id !== id));
      }
    } catch {
      setError('Failed to delete address');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-base font-bold text-white">Saved Address Book</h2>
          <p className="text-xs text-slate-400">Manage saved pickup and delivery destinations</p>
        </div>
        <button
          onClick={() => setShowAddForm(!showAddForm)}
          className="px-4 py-2 rounded-xl text-xs font-bold bg-indigo-600 hover:bg-indigo-500 text-white flex items-center gap-1.5 shadow-md shadow-indigo-600/20 transition-all"
        >
          <Plus className="w-4 h-4" />
          Add Address
        </button>
      </div>

      {error && (
        <div className="p-3.5 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs flex items-center gap-2">
          <AlertCircle className="w-4 h-4 shrink-0 text-rose-400" />
          <span>{error}</span>
        </div>
      )}

      {showAddForm && (
        <form onSubmit={handleCreate} className="p-5 rounded-2xl bg-slate-900 border border-slate-800 space-y-4 animate-fade-in">
          <h3 className="text-xs font-bold uppercase tracking-wider text-indigo-400">New Saved Address</h3>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs">
            <div>
              <label className="text-slate-400 text-[10px]">Label</label>
              <input
                type="text"
                value={label}
                onChange={(e) => setLabel(e.target.value)}
                placeholder="e.g. Headquarters / Warehouse"
                className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-white"
              />
            </div>
            <div>
              <label className="text-slate-400 text-[10px]">Contact Person *</label>
              <input
                type="text"
                value={contactName}
                onChange={(e) => setContactName(e.target.value)}
                required
                placeholder="Contact Name"
                className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-white"
              />
            </div>
            <div>
              <label className="text-slate-400 text-[10px]">Phone Number</label>
              <input
                type="text"
                value={contactPhone}
                onChange={(e) => setContactPhone(e.target.value)}
                placeholder="+91..."
                className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-white"
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs">
            <div className="sm:col-span-2">
              <label className="text-slate-400 text-[10px]">Street Address *</label>
              <input
                type="text"
                value={streetLine1}
                onChange={(e) => setStreetLine1(e.target.value)}
                required
                placeholder="Building, street, suite"
                className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-white"
              />
            </div>
            <div>
              <label className="text-slate-400 text-[10px]">City *</label>
              <input
                type="text"
                value={city}
                onChange={(e) => setCity(e.target.value)}
                required
                placeholder="City"
                className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-white"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 text-xs">
            <div>
              <label className="text-slate-400 text-[10px]">State</label>
              <input
                type="text"
                value={state}
                onChange={(e) => setState(e.target.value)}
                className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-white"
              />
            </div>
            <div>
              <label className="text-slate-400 text-[10px]">Postal Code *</label>
              <input
                type="text"
                value={postalCode}
                onChange={(e) => setPostalCode(e.target.value)}
                required
                placeholder="Postal Code"
                className="w-full mt-1 px-3 py-2 bg-slate-950 border border-slate-800 rounded-xl text-white"
              />
            </div>
          </div>

          <div className="flex items-center justify-between pt-2">
            <label className="flex items-center gap-2 text-xs text-slate-300 cursor-pointer">
              <input
                type="checkbox"
                checked={isDefaultPickup}
                onChange={(e) => setIsDefaultPickup(e.target.checked)}
                className="rounded bg-slate-950 border-slate-800 text-indigo-600 focus:ring-0"
              />
              Set as default pickup location
            </label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setShowAddForm(false)}
                className="px-3 py-1.5 rounded-xl text-xs text-slate-400 hover:text-white bg-slate-800 transition-all"
              >
                Cancel
              </button>
              <button
                type="submit"
                className="px-4 py-1.5 rounded-xl text-xs font-bold bg-indigo-600 hover:bg-indigo-500 text-white transition-all"
              >
                Save Address
              </button>
            </div>
          </div>
        </form>
      )}

      {loading ? (
        <div className="py-12 flex justify-center text-slate-500">
          <Loader2 className="w-6 h-6 animate-spin" />
        </div>
      ) : addresses.length === 0 ? (
        <div className="p-8 text-center rounded-2xl bg-slate-900/40 border border-slate-800">
          <MapPin className="w-8 h-8 text-slate-600 mx-auto mb-2" />
          <p className="text-sm font-bold text-slate-400">No saved addresses yet</p>
          <p className="text-xs text-slate-500 mt-1">Add your warehouse, office, or regular dispatch locations</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {addresses.map((a) => (
            <div key={a.id} className="p-4 rounded-2xl bg-slate-900/70 border border-slate-800 relative group transition-all hover:border-slate-700">
              <div className="flex items-center justify-between mb-2">
                <span className="font-bold text-xs text-white flex items-center gap-1.5">
                  <Building className="w-3.5 h-3.5 text-indigo-400" />
                  {a.label}
                </span>
                {a.is_default_pickup && (
                  <span className="text-[10px] font-bold px-2 py-0.5 rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                    Default Pickup
                  </span>
                )}
              </div>
              <p className="text-xs text-slate-300 font-semibold">{a.contact_name}</p>
              <p className="text-xs text-slate-400 mt-0.5">{a.street_line1}, {a.city}, {a.state} - {a.postal_code}</p>
              <div className="flex items-center justify-between mt-3 pt-3 border-t border-slate-800 text-[11px] text-slate-500">
                <span>{a.contact_phone}</span>
                <button
                  onClick={() => handleDelete(a.id)}
                  className="text-slate-500 hover:text-rose-400 transition-colors p-1"
                  title="Delete address"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
