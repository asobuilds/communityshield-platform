import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

interface Alert {
  id: string;
  title: string;
  content: string;
  type: string;
  location: string;
  severity: string;
  status: string;
  createdAt: string;
  author: { firstName: string; lastName: string };
}

export default function Alerts() {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');

  useEffect(() => {
    fetchAlerts();
  }, []);

  const fetchAlerts = async () => {
    setLoading(true);
    try {
      const response = await axios.get('/api/v1/alerts/news');
      setAlerts(response.data.alerts || []);
    } catch (error) {
      toast.error('Failed to fetch alerts');
    } finally {
      setLoading(false);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'low': return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400';
      case 'medium': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
      case 'high': return 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400';
      case 'critical': return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'low': return 'ℹ️';
      case 'medium': return '⚠️';
      case 'high': return '🚨';
      case 'critical': return '🔥';
      default: return '📌';
    }
  };

  const filteredAlerts = filter === 'all' 
    ? alerts 
    : alerts.filter(a => a.severity === filter);

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-4xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
            🔔 Security Alerts
          </h1>
          <button
            onClick={fetchAlerts}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition"
          >
            🔄 Refresh
          </button>
        </div>

        <div className="flex flex-wrap gap-2 mb-4">
          {['all', 'low', 'medium', 'high', 'critical'].map((s) => (
            <button
              key={s}
              onClick={() => setFilter(s)}
              className={`px-4 py-1 rounded-full text-sm transition ${
                filter === s
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300'
              }`}
            >
              {s.charAt(0).toUpperCase() + s.slice(1)}
            </button>
          ))}
        </div>

        {loading ? (
          <div className="text-center py-8 text-gray-500">Loading alerts...</div>
        ) : filteredAlerts.length === 0 ? (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center">
            <div className="text-4xl mb-2">🔔</div>
            <p className="text-gray-500 dark:text-gray-400">No alerts found</p>
          </div>
        ) : (
          <div className="space-y-4">
            {filteredAlerts.map((alert) => (
              <div
                key={alert.id}
                className={`bg-white dark:bg-gray-800 rounded-lg shadow p-6 border-l-4 ${
                  alert.severity === 'critical' ? 'border-red-500' :
                  alert.severity === 'high' ? 'border-orange-500' :
                  alert.severity === 'medium' ? 'border-yellow-500' :
                  'border-blue-500'
                }`}
              >
                <div className="flex justify-between items-start">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-lg">{getSeverityIcon(alert.severity)}</span>
                      <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
                        {alert.title}
                      </h3>
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getSeverityColor(alert.severity)}`}>
                        {alert.severity.toUpperCase()}
                      </span>
                    </div>
                    {alert.location && (
                      <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        📍 {alert.location}
                      </p>
                    )}
                  </div>
                  <span className="text-xs text-gray-400">
                    {new Date(alert.createdAt).toLocaleString()}
                  </span>
                </div>
                <p className="mt-3 text-gray-700 dark:text-gray-300 whitespace-pre-wrap">
                  {alert.content}
                </p>
                <div className="mt-3 flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
                  {alert.author && (
                    <span>👤 {alert.author.firstName} {alert.author.lastName}</span>
                  )}
                  <span>📌 {alert.type}</span>
                  <span className={`px-2 py-0.5 rounded-full ${
                    alert.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                  }`}>
                    {alert.status}
                  </span>
                </div>
              </div>
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