import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function MyCases() {
  const navigate = useNavigate();
  const [cases, setCases] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCases();
  }, []);

  const fetchCases = async () => {
    try {
      const userStr = localStorage.getItem('user');
      if (!userStr) {
        toast.error('Please login');
        navigate('/login');
        return;
      }
      const user = JSON.parse(userStr);
      const response = await axios.get(`/api/v1/cases?userId=${user.id}`);
      setCases(response.data.cases || []);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to fetch cases');
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

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-4xl mx-auto bg-white rounded-lg shadow-md p-6">
        <h1 className="text-3xl font-bold text-blue-600 mb-6">My Reported Cases</h1>
        {loading ? (
          <p className="text-gray-500">Loading...</p>
        ) : cases.length === 0 ? (
          <div className="text-center py-8">
            <p className="text-gray-500 text-lg">You haven't reported any cases yet.</p>
            <a href="/report" className="mt-4 inline-block bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">Report a Case</a>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="bg-gray-50 border-b">
                  <th className="text-left py-3 px-4 font-semibold text-sm text-gray-600">Title</th>
                  <th className="text-left py-3 px-4 font-semibold text-sm text-gray-600">Status</th>
                  <th className="text-left py-3 px-4 font-semibold text-sm text-gray-600">Date</th>
                  <th className="text-left py-3 px-4 font-semibold text-sm text-gray-600">Action</th>
                </tr>
              </thead>
              <tbody>
                {cases.map((caseItem) => (
                  <tr key={caseItem.id} className="border-b hover:bg-gray-50">
                    <td className="py-3 px-4">{caseItem.title}</td>
                    <td className="py-3 px-4">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(caseItem.status)}`}>
                        {caseItem.status.charAt(0).toUpperCase() + caseItem.status.slice(1)}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-sm text-gray-600">{formatDate(caseItem.createdAt)}</td>
                    <td className="py-3 px-4">
                      <a href={`/case/${caseItem.id}`} className="text-blue-600 hover:underline text-sm">View Details</a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <div className="mt-6">
          <a href="/dashboard" className="text-blue-600 hover:underline">← Back to Dashboard</a>
        </div>
      </div>
    </div>
  );
}

export default MyCases;