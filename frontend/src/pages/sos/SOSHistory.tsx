import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';
import ReportButton from '../../components/ReportButton';

function SOSHistory() {
  const navigate = useNavigate();
  const [sosList, setSosList] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const user = JSON.parse(localStorage.getItem('user') || '{}');

  useEffect(() => {
    fetchSOSHistory();
  }, []);

  const fetchSOSHistory = async () => {
    try {
      const userStr = localStorage.getItem('user');
      if (!userStr) {
        toast.error('Please login');
        navigate('/login');
        return;
      }
      const user = JSON.parse(userStr);
      const response = await axios.get(`/api/v1/sos/history?userId=${user.id}`);
      setSosList(response.data.sos || []);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to fetch SOS history');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-200 text-yellow-800';
      case 'dispatched': return 'bg-blue-200 text-blue-800';
      case 'resolved': return 'bg-green-200 text-green-800';
      default: return 'bg-gray-200 text-gray-800';
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-4xl mx-auto bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
        <button
          onClick={() => navigate('/dashboard')}
          className="text-blue-600 dark:text-blue-400 hover:underline mb-4 flex items-center gap-1"
        >
          ← Back to Dashboard
        </button>

        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold text-red-600 dark:text-red-400">🚨 SOS History</h1>
          {user.id && (
            <ReportButton
              targetType="sos"
              targetId=""
              userId={user.id}
            />
          )}
        </div>

        {loading ? (
          <p className="text-gray-500">Loading...</p>
        ) : sosList.length === 0 ? (
          <div className="text-center py-8">
            <p className="text-gray-500 text-lg">You haven't sent any SOS alerts.</p>
            <a href="/sos" className="mt-4 inline-block bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700">Send SOS</a>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="bg-gray-50 border-b">
                  <th className="text-left py-3 px-4 font-semibold text-sm text-gray-600">Date</th>
                  <th className="text-left py-3 px-4 font-semibold text-sm text-gray-600">Location</th>
                  <th className="text-left py-3 px-4 font-semibold text-sm text-gray-600">Status</th>
                  <th className="text-left py-3 px-4 font-semibold text-sm text-gray-600">Message</th>
                  <th className="text-left py-3 px-4 font-semibold text-sm text-gray-600">Action</th>
                </tr>
              </thead>
              <tbody>
                {sosList.map((sos) => (
                  <tr key={sos.id} className="border-b hover:bg-gray-50">
                    <td className="py-3 px-4 text-sm">{formatDate(sos.createdAt)}</td>
                    <td className="py-3 px-4 text-sm">{sos.latitude?.toFixed(6)}, {sos.longitude?.toFixed(6)}</td>
                    <td className="py-3 px-4">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(sos.status)}`}>
                        {sos.status.charAt(0).toUpperCase() + sos.status.slice(1)}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-sm">{sos.description || 'No message'}</td>
                    <td className="py-3 px-4">
                      {user.id && (
                        <ReportButton
                          targetType="sos"
                          targetId={sos.id}
                          userId={user.id}
                        />
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <div className="mt-6">
          <a href="/dashboard" className="text-blue-600 dark:text-blue-400 hover:underline">← Back to Dashboard</a>
        </div>
      </div>
    </div>
  );
}

export default SOSHistory;