import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

interface TransferRequest {
  id: string;
  targetId: string;
  targetType: string;
  reason: string;
  status: string;
  approvalCount: number;
  requiredApprovals: number;
  fromUnit: { id: string; name: string };
  toUnit: { id: string; name: string };
  requestedByUser: { firstName: string; lastName: string };
  createdAt: string;
  approvedAt: string;
  completedAt: string;
}

export default function TransfersList() {
  const [transfers, setTransfers] = useState<TransferRequest[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchTransfers();
  }, []);

  const fetchTransfers = async () => {
    setLoading(true);
    try {
      const response = await axios.get('/api/v1/transfers');
      setTransfers(response.data.transferRequests || []);
    } catch (error) {
      toast.error('Failed to fetch transfers');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
      case 'approved': return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400';
      case 'rejected': return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
      case 'completed': return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'pending': return '⏳';
      case 'approved': return '✅';
      case 'rejected': return '❌';
      case 'completed': return '🎉';
      default: return '📌';
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-4xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
            🔄 Transfer Requests
          </h1>
          <button
            onClick={fetchTransfers}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition"
          >
            🔄 Refresh
          </button>
        </div>

        {loading ? (
          <div className="text-center py-8 text-gray-500">Loading...</div>
        ) : transfers.length === 0 ? (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center">
            <p className="text-gray-500 dark:text-gray-400">No transfer requests</p>
          </div>
        ) : (
          <div className="space-y-4">
            {transfers.map((transfer) => (
              <div
                key={transfer.id}
                className="bg-white dark:bg-gray-800 rounded-lg shadow p-6"
              >
                <div className="flex justify-between items-start flex-wrap gap-2">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-lg">{getStatusIcon(transfer.status)}</span>
                      <h3 className="font-semibold text-gray-800 dark:text-gray-200">
                        {transfer.targetType.toUpperCase()}: {transfer.targetId}
                      </h3>
                    </div>
                    <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                      From: {transfer.fromUnit?.name || 'Unknown'} → To: {transfer.toUnit?.name || 'Unknown'}
                    </p>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      Requested by: {transfer.requestedByUser?.firstName} {transfer.requestedByUser?.lastName}
                    </p>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                      Reason: {transfer.reason}
                    </p>
                  </div>
                  <div className="text-right">
                    <span className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(transfer.status)}`}>
                      {transfer.status.toUpperCase()}
                    </span>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      {transfer.approvalCount}/{transfer.requiredApprovals} approvals
                    </p>
                    <p className="text-xs text-gray-400">
                      {new Date(transfer.createdAt).toLocaleDateString()}
                    </p>
                  </div>
                </div>

                {/* Progress Bar */}
                <div className="mt-3">
                  <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                    <div
                      className="bg-blue-600 h-2 rounded-full transition-all"
                      style={{
                        width: `${Math.min((transfer.approvalCount / transfer.requiredApprovals) * 100, 100)}%`
                      }}
                    />
                  </div>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    {transfer.approvalCount} of {transfer.requiredApprovals} approvals received
                  </p>
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