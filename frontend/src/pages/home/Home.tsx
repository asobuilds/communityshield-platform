import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';
import IncidentMap from '../../components/map/IncidentMap';

interface Incident {
  id: string;
  title: string;
  description: string;
  latitude: number;
  longitude: number;
  status: string;
  type: 'case' | 'sos' | 'unit';
}

export default function Home() {
  const [user, setUser] = useState<any>(null);
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState({ totalCases: 0, pendingCases: 0, totalUnits: 0 });

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      setUser(JSON.parse(userStr));
    }
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [casesRes, unitsRes] = await Promise.all([
        axios.get('/api/v1/cases').catch(() => ({ data: { cases: [] } })),
        axios.get('/api/v1/units').catch(() => ({ data: { units: [] } }))
      ]);

      const cases = casesRes.data.cases || [];
      const units = unitsRes.data.units || [];

      setStats({
        totalCases: cases.length,
        pendingCases: cases.filter((c: any) => c.status === 'pending').length,
        totalUnits: units.length
      });

      const mappedCases = cases.map((c: any) => ({
        id: c.id,
        title: c.title,
        description: c.description,
        latitude: c.latitude || 6.5244,
        longitude: c.longitude || 3.3792,
        status: c.status,
        type: 'case' as const
      }));

      setIncidents(mappedCases);
    } catch (error) {
      console.error('Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const isAdmin = user?.role === 'unit_admin' || user?.role === 'super_admin';
  const isOfficer = user?.role === 'officer';
  const canViewSuspects = isAdmin || isOfficer;

  return (
    <div>
      {/* Welcome Section */}
      <div className="flex flex-wrap justify-between items-start gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
            👋 Welcome back, {user?.firstName || 'User'}!
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Here's what's happening in your community
          </p>
        </div>
        <div className="flex gap-2">
          <Link to="/report">
            <button className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition">
              📝 Report Case
            </button>
          </Link>
          <Link to="/sos">
            <button className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition">
              🆘 SOS
            </button>
          </Link>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <div className="flex items-center gap-3">
            <div className="text-3xl">📋</div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Total Cases</p>
              <p className="text-2xl font-bold text-gray-800 dark:text-gray-200">{stats.totalCases}</p>
            </div>
          </div>
        </div>
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <div className="flex items-center gap-3">
            <div className="text-3xl">⏳</div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Pending Cases</p>
              <p className="text-2xl font-bold text-yellow-600">{stats.pendingCases}</p>
            </div>
          </div>
        </div>
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <div className="flex items-center gap-3">
            <div className="text-3xl">🏢</div>
            <div>
              <p className="text-sm text-gray-500 dark:text-gray-400">Security Units</p>
              <p className="text-2xl font-bold text-green-600">{stats.totalUnits}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Map Section */}
      <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow mb-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
            📍 Nearby Incidents
          </h2>
          <Link to="/map" className="text-blue-600 hover:underline text-sm">
            View Full Map →
          </Link>
        </div>
        {loading ? (
          <div className="flex justify-center items-center h-[300px]">
            <p className="text-gray-500 dark:text-gray-400">Loading map...</p>
          </div>
        ) : (
          <IncidentMap
            incidents={incidents.slice(0, 10)}
            center={[6.5244, 3.3792]}
            zoom={12}
            showRadius={false}
          />
        )}
      </div>

      {/* Quick Actions - Conditional based on role */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <Link to="/my-cases">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow text-center hover:shadow-md transition cursor-pointer">
            <div className="text-3xl">📋</div>
            <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">My Cases</p>
          </div>
        </Link>
        <Link to="/map">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow text-center hover:shadow-md transition cursor-pointer">
            <div className="text-3xl">🗺️</div>
            <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">View Map</p>
          </div>
        </Link>
        <Link to="/news">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow text-center hover:shadow-md transition cursor-pointer">
            <div className="text-3xl">📰</div>
            <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">News</p>
          </div>
        </Link>
        <Link to="/ai">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow text-center hover:shadow-md transition cursor-pointer">
            <div className="text-3xl">🤖</div>
            <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">AI Insights</p>
          </div>
        </Link>
        <Link to="/alerts">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow text-center hover:shadow-md transition cursor-pointer">
            <div className="text-3xl">🔔</div>
            <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">Alerts</p>
          </div>
        </Link>
        
        {/* Only show suspects if admin or officer */}
        {canViewSuspects && (
          <Link to="/suspects">
            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow text-center hover:shadow-md transition cursor-pointer">
              <div className="text-3xl">🕵️</div>
              <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">Suspects</p>
            </div>
          </Link>
        )}
        
        <Link to="/profile">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow text-center hover:shadow-md transition cursor-pointer">
            <div className="text-3xl">👤</div>
            <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">Profile</p>
          </div>
        </Link>
        
        {isAdmin && (
          <Link to="/admin">
            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow text-center hover:shadow-md transition cursor-pointer">
              <div className="text-3xl">⚙️</div>
              <p className="text-sm text-gray-700 dark:text-gray-300 mt-1">Admin</p>
            </div>
          </Link>
        )}
      </div>
    </div>
  );
}