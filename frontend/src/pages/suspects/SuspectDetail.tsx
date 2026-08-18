import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

interface Suspect {
  id: string;
  firstName: string;
  lastName: string;
  alias: string;
  gender: string;
  dateOfBirth: string;
  nationality: string;
  idNumber: string;
  phone: string;
  email: string;
  address: string;
  description: string;
  status: string;
  dangerLevel: string;
  wanted: boolean;
  unitId: string;
  transferStatus: string;
  transferReason: string;
  transferToUnit: string;
  transferToUnitObj: { id: string; name: string } | null;
  createdByUser: { firstName: string; lastName: string };
  unit: { id: string; name: string } | null;
  createdAt: string;
}

interface Sighting {
  id: string;
  location: string;
  description: string;
  latitude: number;
  longitude: number;
  timestamp: string;
  reporter: { firstName: string; lastName: string };
  unit: { name: string };
}

interface Unit {
  id: string;
  name: string;
}

export default function SuspectDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [user, setUser] = useState<any>(null);
  const [suspect, setSuspect] = useState<Suspect | null>(null);
  const [sightings, setSightings] = useState<Sighting[]>([]);
  const [units, setUnits] = useState<Unit[]>([]);
  const [loading, setLoading] = useState(true);
  const [showSightingForm, setShowSightingForm] = useState(false);
  const [transferUnitId, setTransferUnitId] = useState('');
  const [transferReason, setTransferReason] = useState('');
  const [sightingData, setSightingData] = useState({
    latitude: '',
    longitude: '',
    location: '',
    description: '',
  });

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      setUser(JSON.parse(userStr));
    }
    fetchSuspect();
    fetchSightings();
    fetchUnits();
  }, [id]);

  const fetchSuspect = async () => {
    try {
      const response = await axios.get(`/api/v1/suspects/${id}`);
      setSuspect(response.data.suspect);
    } catch (error) {
      toast.error('Failed to fetch suspect');
      navigate('/suspects');
    }
  };

  const fetchSightings = async () => {
    try {
      const response = await axios.get(`/api/v1/suspects/${id}/sightings`);
      setSightings(response.data.sightings || []);
    } catch (error) {
      console.error('Failed to fetch sightings');
    } finally {
      setLoading(false);
    }
  };

  const fetchUnits = async () => {
    try {
      const response = await axios.get('/api/v1/units');
      setUnits(response.data.units || []);
    } catch (error) {
      console.error('Failed to fetch units');
    }
  };

  const handleSightingSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await axios.post(`/api/v1/suspects/${id}/sighting`, {
        latitude: parseFloat(sightingData.latitude),
        longitude: parseFloat(sightingData.longitude),
        location: sightingData.location,
        description: sightingData.description,
      });
      toast.success('Sighting reported successfully');
      setShowSightingForm(false);
      setSightingData({ latitude: '', longitude: '', location: '', description: '' });
      fetchSightings();
    } catch (error) {
      toast.error('Failed to report sighting');
    }
  };

  const handleRequestTransfer = async () => {
    if (!transferUnitId) {
      toast.error('Please select a unit to transfer to');
      return;
    }
    if (!transferReason.trim()) {
      toast.error('Please provide a reason for transfer');
      return;
    }

    try {
      await axios.post(`/api/v1/suspects/${id}/transfer/request`, {
        transferToUnitId: transferUnitId,
        reason: transferReason,
      });
      toast.success('Transfer request submitted successfully');
      setTransferUnitId('');
      setTransferReason('');
      fetchSuspect();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to request transfer');
    }
  };

  const handleApproveTransfer = async () => {
    try {
      await axios.post(`/api/v1/suspects/${id}/transfer/approve`);
      toast.success('Transfer approved successfully');
      fetchSuspect();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to approve transfer');
    }
  };

  const handleRejectTransfer = async () => {
    try {
      await axios.post(`/api/v1/suspects/${id}/transfer/reject`);
      toast.success('Transfer rejected');
      fetchSuspect();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to reject transfer');
    }
  };

  const isAdmin = user?.role === 'unit_admin' || user?.role === 'super_admin';
  const isOfficer = user?.role === 'officer';

  const getDangerColor = (level: string) => {
    switch (level) {
      case 'low': return 'text-green-600 bg-green-100 dark:bg-green-900/30';
      case 'medium': return 'text-yellow-600 bg-yellow-100 dark:bg-yellow-900/30';
      case 'high': return 'text-orange-600 bg-orange-100 dark:bg-orange-900/30';
      case 'extreme': return 'text-red-600 bg-red-100 dark:bg-red-900/30';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
      case 'captured': return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
      case 'cleared': return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400';
      case 'transferred': return 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4 flex justify-center items-center">
        <p className="text-gray-500">Loading...</p>
      </div>
    );
  }

  if (!suspect) {
    return (
      <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
        <div className="text-center py-8">
          <p className="text-gray-500">Suspect not found</p>
          <Link to="/suspects" className="text-blue-600 hover:underline">← Back to Suspects</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-4xl mx-auto">
        <div className="mb-4">
          <Link to="/suspects" className="text-gray-600 dark:text-gray-400 hover:underline">
            ← Back to Suspects
          </Link>
        </div>

        {/* Suspect Info */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 mb-6">
          <div className="flex justify-between items-start flex-wrap gap-4">
            <div>
              <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
                {suspect.firstName} {suspect.lastName}
              </h1>
              {suspect.alias && (
                <p className="text-gray-500 dark:text-gray-400">AKA: {suspect.alias}</p>
              )}
            </div>
            <div className="flex flex-wrap gap-2">
              {suspect.wanted && (
                <span className="bg-red-600 text-white px-3 py-1 rounded-full text-sm font-bold">
                  WANTED
                </span>
              )}
              <span className={`px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(suspect.status)}`}>
                {suspect.status.toUpperCase()}
              </span>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mt-4">
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Gender</p>
              <p className="text-gray-800 dark:text-gray-200">{suspect.gender || 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Nationality</p>
              <p className="text-gray-800 dark:text-gray-200">{suspect.nationality || 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Phone</p>
              <p className="text-gray-800 dark:text-gray-200">{suspect.phone || 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Email</p>
              <p className="text-gray-800 dark:text-gray-200">{suspect.email || 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Unit</p>
              <p className="text-gray-800 dark:text-gray-200">{suspect.unit?.name || 'Not assigned'}</p>
            </div>
            <div className="md:col-span-2">
              <p className="text-sm text-gray-500 dark:text-gray-400">Address</p>
              <p className="text-gray-800 dark:text-gray-200">{suspect.address || 'N/A'}</p>
            </div>
            <div className="md:col-span-2">
              <p className="text-sm text-gray-500 dark:text-gray-400">Description</p>
              <p className="text-gray-800 dark:text-gray-200">{suspect.description || 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Danger Level</p>
              <span className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${getDangerColor(suspect.dangerLevel)}`}>
                ⚠️ {suspect.dangerLevel.toUpperCase()}
              </span>
            </div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Added</p>
              <p className="text-gray-800 dark:text-gray-200">{new Date(suspect.createdAt).toLocaleDateString()}</p>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="mt-4 flex flex-wrap gap-2">
            {(isAdmin || isOfficer) && (
              <button
                onClick={() => setShowSightingForm(!showSightingForm)}
                className="bg-yellow-600 text-white px-4 py-2 rounded-lg hover:bg-yellow-700 transition"
              >
                📍 Report Sighting
              </button>
            )}
            {isAdmin && (
              <>
                <Link
                  to={`/suspects/${suspect.id}/edit`}
                  className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition"
                >
                  ✏️ Edit
                </Link>
              </>
            )}
          </div>

          {/* Transfer Section - Admin Only */}
          {isAdmin && (
            <div className="mt-4 p-4 bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600">
              <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">
                🔄 Suspect Transfer
              </h3>
              {suspect.transferStatus === 'pending' ? (
                <div className="space-y-2">
                  <div className="flex flex-wrap gap-2">
                    <button
                      onClick={handleApproveTransfer}
                      className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition text-sm"
                    >
                      ✅ Approve Transfer
                    </button>
                    <button
                      onClick={handleRejectTransfer}
                      className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition text-sm"
                    >
                      ❌ Reject Transfer
                    </button>
                  </div>
                  <p className="text-xs text-gray-500 dark:text-gray-400">
                    Transfer to: {suspect.transferToUnitObj?.name || 'Unknown unit'}
                  </p>
                  {suspect.transferReason && (
                    <p className="text-xs text-gray-500 dark:text-gray-400">
                      Reason: {suspect.transferReason}
                    </p>
                  )}
                </div>
              ) : (
                <div className="space-y-3">
                  <div className="flex flex-wrap gap-2">
                    <select
                      value={transferUnitId}
                      onChange={(e) => setTransferUnitId(e.target.value)}
                      className="px-3 py-2 border rounded-lg text-sm dark:bg-gray-600 dark:border-gray-500 dark:text-white flex-1 min-w-[150px]"
                    >
                      <option value="">Select unit...</option>
                      {units.map((u) => (
                        <option key={u.id} value={u.id}>{u.name}</option>
                      ))}
                    </select>
                    <input
                      type="text"
                      placeholder="Reason for transfer"
                      value={transferReason}
                      onChange={(e) => setTransferReason(e.target.value)}
                      className="px-3 py-2 border rounded-lg text-sm dark:bg-gray-600 dark:border-gray-500 dark:text-white flex-1 min-w-[150px]"
                    />
                  </div>
                  <button
                    onClick={handleRequestTransfer}
                    className="bg-yellow-600 text-white px-4 py-2 rounded-lg hover:bg-yellow-700 transition text-sm"
                  >
                    📤 Request Transfer
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Sighting Form */}
        {showSightingForm && (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 mb-6">
            <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-4">
              📍 Report Sighting
            </h2>
            <form onSubmit={handleSightingSubmit}>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Latitude *
                  </label>
                  <input
                    type="number"
                    step="any"
                    required
                    value={sightingData.latitude}
                    onChange={(e) => setSightingData({ ...sightingData, latitude: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Longitude *
                  </label>
                  <input
                    type="number"
                    step="any"
                    required
                    value={sightingData.longitude}
                    onChange={(e) => setSightingData({ ...sightingData, longitude: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Location
                  </label>
                  <input
                    type="text"
                    value={sightingData.location}
                    onChange={(e) => setSightingData({ ...sightingData, location: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                    placeholder="e.g., Lagos, Nigeria"
                  />
                </div>
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Description
                  </label>
                  <textarea
                    value={sightingData.description}
                    onChange={(e) => setSightingData({ ...sightingData, description: e.target.value })}
                    rows={3}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                    placeholder="Describe what you saw..."
                  />
                </div>
              </div>
              <div className="mt-4 flex gap-2">
                <button
                  type="submit"
                  className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition"
                >
                  Submit Sighting
                </button>
                <button
                  type="button"
                  onClick={() => setShowSightingForm(false)}
                  className="bg-gray-500 text-white px-4 py-2 rounded-lg hover:bg-gray-600 transition"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Sightings List */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-4">
            📋 Sightings ({sightings.length})
          </h2>
          {sightings.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400">No sightings reported yet</p>
          ) : (
            <div className="space-y-3">
              {sightings.map((sighting) => (
                <div key={sighting.id} className="border dark:border-gray-700 rounded-lg p-4">
                  <div className="flex justify-between items-start">
                    <div>
                      <p className="text-gray-800 dark:text-gray-200">{sighting.description}</p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        📍 {sighting.location || `${sighting.latitude}, ${sighting.longitude}`}
                      </p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        Reported by: {sighting.reporter?.firstName} {sighting.reporter?.lastName}
                        {sighting.unit && ` • ${sighting.unit.name}`}
                      </p>
                    </div>
                    <span className="text-xs text-gray-500">
                      {new Date(sighting.timestamp).toLocaleString()}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}