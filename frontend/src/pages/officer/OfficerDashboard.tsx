import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function OfficerDashboard() {
  const navigate = useNavigate();
  const [assignedCases, setAssignedCases] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const user = JSON.parse(localStorage.getItem('user') || '{}');

  useEffect(() => {
    if (user.id) {
      fetchAssignedCases();
    }
  }, []);

  const fetchAssignedCases = async () => {
    try {
      const response = await axios.get('/api/v1/admin/cases');
      const allCases = response.data.cases || [];
      const filtered = allCases.filter((c: any) => c.assignedTo === user.id);
      setAssignedCases(filtered);
    } catch (error) {
      toast.error('Failed to fetch assigned cases');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-200 text-yellow-800';
      case 'investigating': return 'bg-blue-200 text-blue-800';
      case 'resolved': return 'bg-green-200 text-green-800';
      case 'closed': return 'bg-gray-200 text-gray-800';
      default: return 'bg-gray-200 text-gray-800';
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-4xl mx-auto bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
        {/* Back Button */}
        <button
          onClick={() => navigate('/dashboard')}
          className="text-blue-600 dark:text-blue-400 hover:underline mb-4 flex items-center gap-1"
        >
          ← Back to Dashboard
        </button>

        <div className="flex justify-between items-center mb-4">
          <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400">Officer Dashboard</h1>
          <button
            onClick={() => { localStorage.removeItem('user'); window.location.href = '/login'; }}
            className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 text-sm"
          >
            Logout
          </button>
        </div>
        <p className="text-gray-700 dark:text-gray-300">Welcome, <strong>{user.firstName || 'Officer'}</strong>!</p>
        <p className="text-gray-600 dark:text-gray-400 mt-2">You are logged in as: <span className="font-semibold">{user.role || 'officer'}</span></p>
        <div className="mt-6 grid grid-cols-1 md:grid-cols-2 gap-4">
          <a href="/profile" className="bg-blue-600 text-white px-6 py-4 rounded-lg hover:bg-blue-700 text-center font-semibold text-lg">👤 My Profile</a>
          <a href="/sos" className="bg-red-600 text-white px-6 py-4 rounded-lg hover:bg-red-700 text-center font-semibold text-lg">🚨 SOS</a>
          <a href="/report" className="bg-green-600 text-white px-6 py-4 rounded-lg hover:bg-green-700 text-center font-semibold text-lg">📝 Report Case</a>
          <a href="/my-cases" className="bg-orange-600 text-white px-6 py-4 rounded-lg hover:bg-orange-700 text-center font-semibold text-lg">📋 My Cases</a>
        </div>

        <div className="mt-8 border-t dark:border-gray-700 pt-6">
          <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">📌 Assigned Cases</h2>
          {loading ? (
            <p className="text-gray-500 dark:text-gray-400">Loading...</p>
          ) : assignedCases.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400">No cases assigned to you.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                    <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600 dark:text-gray-300">Title</th>
                    <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600 dark:text-gray-300">Status</th>
                    <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600 dark:text-gray-300">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {assignedCases.map((c) => (
                    <tr key={c.id} className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{c.title}</td>
                      <td className="py-3 px-4">
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(c.status)}`}>
                          {c.status.charAt(0).toUpperCase() + c.status.slice(1)}
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <a href={`/case/${c.id}`} className="text-blue-600 dark:text-blue-400 hover:underline text-sm">View Details</a>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default OfficerDashboard;