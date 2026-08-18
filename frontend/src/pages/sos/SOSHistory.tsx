import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

interface SOSAlert {
  id: string;
  description: string;
  latitude: number;
  longitude: number;
  status: string;
  priority: string;
  createdAt: string;
  unit?: { name: string };
}

export default function SOSHistory() {
  const [alerts, setAlerts] = useState<SOSAlert[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAlerts();
  }, []);

  const fetchAlerts = async () => {
    setLoading(true);
    try {
      const response = await axios.get('/api/v1/sos/my');
      setAlerts(response.data.alerts || []);
    } catch (error) {
      toast.error('Failed to fetch SOS history');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
      case 'dispatched': return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400';
      case 'resolved': return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
      case 'cancelled': return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-400';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusEmoji = (status: string) => {
    switch (status) {
      case 'pending': return '⏳';
      case 'dispatched': return '🚨';
      case 'resolved': return '✅';
      case 'cancelled': return '❌';
      default: return '📌';
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-3xl mx-auto">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
          <div className="flex justify-between items-center mb-6">
            <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
              🆘 SOS History
            </h1>
            <Link
              to="/sos"
              className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition text-sm"
            >
              New SOS
            </Link>
          </div>

          {loading ? (
            <div className="text-center py-8 text-gray-500">Loading...</div>
          ) : alerts.length === 0 ? (
            <div className="text-center py-8">
              <p className="text-gray-500 dark:text-gray-400">No SOS alerts sent yet</p>
              <Link to="/sos" className="text-blue-600 hover:underline mt-2 inline-block">
                Send your first SOS →
              </Link>
            </div>
          ) : (
            <div className="space-y-4">
              {alerts.map((alert) => (
                <div
                  key={alert.id}
                  className="border dark:border-gray-700 rounded-lg p-4 hover:shadow-md transition"
                >
                  <div className="flex justify-between items-start">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="text-lg">{getStatusEmoji(alert.status)}</span>
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(alert.status)}`}>
                          {alert.status.toUpperCase()}
                        </span>
                        <span className="text-xs text-gray-500 dark:text-gray-400">
                          {new Date(alert.createdAt).toLocaleString()}
                        </span>
                      </div>
                      <p className="text-gray-700 dark:text-gray-300">{alert.description}</p>
                      <div className="mt-2 text-sm text-gray-500 dark:text-gray-400">
                        📍 {alert.latitude.toFixed(6)}, {alert.longitude.toFixed(6)}
                        {alert.unit && (
                          <span className="ml-2">🏢 {alert.unit.name}</span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          <div className="mt-6 text-center">
            <Link to="/home" className="text-gray-600 dark:text-gray-400 hover:underline">
              ← Back to Home
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}