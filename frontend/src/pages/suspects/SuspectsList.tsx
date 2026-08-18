import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

interface Suspect {
  id: string;
  firstName: string;
  lastName: string;
  alias: string;
  gender: string;
  description: string;
  status: string;
  dangerLevel: string;
  wanted: boolean;
  unit: { id: string; name: string };
  createdAt: string;
}

export default function SuspectsList() {
  const navigate = useNavigate();
  const [user, setUser] = useState<any>(null);
  const [suspects, setSuspects] = useState<Suspect[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      const userData = JSON.parse(userStr);
      setUser(userData);
      // Redirect citizens away from suspects page
      if (userData.role === 'citizen') {
        toast.error('You do not have permission to view suspects');
        navigate('/home');
        return;
      }
    }
    fetchSuspects();
  }, []);

  const fetchSuspects = async () => {
    setLoading(true);
    try {
      const response = await axios.get('/api/v1/suspects');
      setSuspects(response.data.suspects || []);
    } catch (error) {
      toast.error('Failed to fetch suspects');
    } finally {
      setLoading(false);
    }
  };

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
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const filteredSuspects = suspects.filter(s => {
    if (filter === 'all') return true;
    if (filter === 'wanted') return s.wanted;
    return s.status === filter;
  });

  // If user is citizen, redirect (already handled in useEffect)
  if (user?.role === 'citizen') {
    return null;
  }

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
            🕵️ Suspect Tracking
          </h1>
          {user?.role !== 'officer' && (
            <Link
              to="/suspects/create"
              className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition"
            >
              + Add Suspect
            </Link>
          )}
        </div>

        {/* Filters */}
        <div className="flex flex-wrap gap-2 mb-4">
          {['all', 'active', 'captured', 'cleared', 'wanted'].map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`px-4 py-1 rounded-full text-sm transition ${
                filter === f
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300'
              }`}
            >
              {f.charAt(0).toUpperCase() + f.slice(1)}
            </button>
          ))}
          <button
            onClick={fetchSuspects}
            className="px-4 py-1 rounded-full text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300"
          >
            🔄 Refresh
          </button>
        </div>

        {loading ? (
          <div className="text-center py-8 text-gray-500">Loading...</div>
        ) : filteredSuspects.length === 0 ? (
          <div className="text-center py-8 bg-white dark:bg-gray-800 rounded-lg shadow">
            <p className="text-gray-500 dark:text-gray-400">No suspects found</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredSuspects.map((suspect) => (
              <Link
                key={suspect.id}
                to={`/suspects/${suspect.id}`}
                className="bg-white dark:bg-gray-800 rounded-lg shadow hover:shadow-lg transition p-4"
              >
                <div className="flex justify-between items-start">
                  <div>
                    <h3 className="font-semibold text-gray-800 dark:text-gray-200">
                      {suspect.firstName} {suspect.lastName}
                    </h3>
                    {suspect.alias && (
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        AKA: {suspect.alias}
                      </p>
                    )}
                    {suspect.unit && (
                      <p className="text-xs text-gray-400 dark:text-gray-500">
                        Unit: {suspect.unit.name}
                      </p>
                    )}
                  </div>
                  {suspect.wanted && (
                    <span className="text-xs bg-red-600 text-white px-2 py-1 rounded-full">
                      WANTED
                    </span>
                  )}
                </div>

                <div className="mt-2 flex flex-wrap gap-2">
                  <span className={`text-xs px-2 py-1 rounded-full ${getStatusColor(suspect.status)}`}>
                    {suspect.status}
                  </span>
                  <span className={`text-xs px-2 py-1 rounded-full ${getDangerColor(suspect.dangerLevel)}`}>
                    ⚠️ {suspect.dangerLevel}
                  </span>
                </div>

                {suspect.description && (
                  <p className="text-sm text-gray-600 dark:text-gray-400 mt-2 line-clamp-2">
                    {suspect.description}
                  </p>
                )}

                <div className="mt-2 text-xs text-gray-500 dark:text-gray-400">
                  Added: {new Date(suspect.createdAt).toLocaleDateString()}
                </div>
              </Link>
            ))}
          </div>
        )}

        <div className="mt-6">
          <Link to="/home" className="text-gray-600 dark:text-gray-400 hover:underline">
            ← Back to Home
          </Link>
        </div>
      </div>
    </div>
  );
}