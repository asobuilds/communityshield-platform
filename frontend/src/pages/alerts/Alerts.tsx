import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';
import ReportButton from '../../components/ReportButton';

function Alerts() {
  const navigate = useNavigate();
  const [alerts, setAlerts] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [location, setLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [useLocation, setUseLocation] = useState(false);
  const [filters, setFilters] = useState({
    severity: '',
    type: '',
    public: 'true',
  });
  const user = JSON.parse(localStorage.getItem('user') || '{}');

  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          setLocation({ lat: pos.coords.latitude, lng: pos.coords.longitude });
          setUseLocation(true);
        },
        () => {
          toast.error('Location access denied – showing all alerts');
          setUseLocation(false);
        }
      );
    }
  }, []);

  useEffect(() => {
    fetchAlerts();
  }, [location, filters]);

  const fetchAlerts = async () => {
    setLoading(true);
    try {
      let url = '/api/v1/announcements?';
      const params = new URLSearchParams();
      if (filters.severity) params.append('severity', filters.severity);
      if (filters.type) params.append('type', filters.type);
      params.append('public', filters.public);
      if (useLocation && location) {
        params.append('lat', location.lat.toString());
        params.append('lng', location.lng.toString());
        params.append('radius', '20');
      }
      const response = await axios.get(url + params.toString());
      setAlerts(response.data.announcements || []);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load alerts');
    } finally {
      setLoading(false);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'bg-red-600 text-white';
      case 'high': return 'bg-orange-500 text-white';
      case 'medium': return 'bg-yellow-500 text-black';
      case 'low': return 'bg-blue-500 text-white';
      default: return 'bg-gray-400 text-white';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'alert': return '🚨';
      case 'warning': return '⚠️';
      case 'news': return '📰';
      default: return '📢';
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-4xl mx-auto">
        <button
          onClick={() => navigate('/dashboard')}
          className="text-blue-600 dark:text-blue-400 hover:underline mb-4 flex items-center gap-1"
        >
          ← Back to Dashboard
        </button>

        <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400 mb-6">📢 Public Alerts & News</h1>

        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow mb-6 flex flex-wrap gap-4 items-end">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Severity</label>
            <select
              value={filters.severity}
              onChange={(e) => setFilters({ ...filters, severity: e.target.value })}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            >
              <option value="">All</option>
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Type</label>
            <select
              value={filters.type}
              onChange={(e) => setFilters({ ...filters, type: e.target.value })}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            >
              <option value="">All</option>
              <option value="news">News</option>
              <option value="alert">Alert</option>
              <option value="warning">Warning</option>
            </select>
          </div>
          <button
            onClick={() => setUseLocation(!useLocation)}
            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            {useLocation ? '📍 Near Me' : '📍 Use Location'}
          </button>
          <button
            onClick={fetchAlerts}
            className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
          >
            Refresh
          </button>
        </div>

        {loading ? (
          <div className="text-center py-12 text-gray-500">Loading alerts...</div>
        ) : alerts.length === 0 ? (
          <div className="text-center py-12 text-gray-500 dark:text-gray-400">
            No alerts found.
          </div>
        ) : (
          <div className="space-y-4">
            {alerts.map((alert) => (
              <div
                key={alert.id}
                className={`bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border-l-8 ${getSeverityColor(alert.severity)}`}
              >
                <div className="flex justify-between items-start">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-2xl">{getTypeIcon(alert.type)}</span>
                      <h2 className="text-xl font-bold text-gray-900 dark:text-white">{alert.title}</h2>
                    </div>
                    <p className="text-sm text-gray-600 dark:text-gray-300 mt-2">{alert.content}</p>
                    <div className="mt-2 flex flex-wrap gap-2 text-xs">
                      <span className="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded-full">
                        {alert.type.charAt(0).toUpperCase() + alert.type.slice(1)}
                      </span>
                      <span className={`px-2 py-1 rounded-full ${getSeverityColor(alert.severity)}`}>
                        {alert.severity.toUpperCase()}
                      </span>
                      {alert.latitude && alert.longitude && (
                        <span className="px-2 py-1 bg-blue-100 dark:bg-blue-900 rounded-full text-blue-800 dark:text-blue-200">
                          📍 {alert.latitude.toFixed(4)}, {alert.longitude.toFixed(4)}
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="flex flex-col items-end gap-2">
                    <div className="text-sm text-gray-500 dark:text-gray-400">
                      {formatDate(alert.createdAt)}
                    </div>
                    {user.id && (
                      <ReportButton
                        targetType="announcement"
                        targetId={alert.id}
                        userId={user.id}
                      />
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default Alerts;