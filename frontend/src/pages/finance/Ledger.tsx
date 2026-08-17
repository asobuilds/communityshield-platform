import { useState, useEffect } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';

function Ledger() {
  const [transactions, setTransactions] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [unitId] = useState('00000000-0000-0000-0000-000000000000');

  useEffect(() => {
    fetchTransactions();
  }, []);

  const fetchTransactions = async () => {
    try {
      const response = await axios.get(`/api/v1/transactions/units/${unitId}`);
      setTransactions(response.data.transactions || []);
    } catch (error) {
      toast.error('Failed to fetch transactions');
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
  };

  const getTypeColor = (type: string) => type === 'expense' ? 'text-red-600' : 'text-green-600';
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-200 text-yellow-800';
      case 'approved': return 'bg-green-200 text-green-800';
      case 'rejected': return 'bg-red-200 text-red-800';
      default: return 'bg-gray-200 text-gray-800';
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-6xl mx-auto bg-white rounded-lg shadow p-6">
        <h1 className="text-3xl font-bold text-blue-600 mb-6">📒 Transaction Ledger</h1>
        {loading ? (
          <p>Loading...</p>
        ) : transactions.length === 0 ? (
          <div className="text-center py-8">
            <p>No transactions found.</p>
            <a href="/add-transaction" className="mt-4 inline-block bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">Add First Transaction</a>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="bg-gray-50 border-b">
                  <th className="text-left py-3 px-4 text-sm font-semibold">Date</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold">Type</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold">Description</th>
                  <th className="text-right py-3 px-4 text-sm font-semibold">Amount</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold">Status</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold">Method</th>
                </tr>
              </thead>
              <tbody>
                {transactions.map((txn) => (
                  <tr key={txn.id} className="border-b hover:bg-gray-50">
                    <td className="py-3 px-4 text-sm">{formatDate(txn.transactionDate)}</td>
                    <td className="py-3 px-4 text-sm capitalize">{txn.type}</td>
                    <td className="py-3 px-4 text-sm">{txn.description}</td>
                    <td className={`py-3 px-4 text-sm font-medium text-right ${getTypeColor(txn.type)}`}>
                      {txn.type === 'expense' ? '-' : '+'} ₦{txn.amount.toFixed(2)}
                    </td>
                    <td className="py-3 px-4">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(txn.status)}`}>
                        {txn.status.charAt(0).toUpperCase() + txn.status.slice(1)}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-sm capitalize">{txn.paymentMethod || 'N/A'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        <div className="mt-6">
          <a href="/finance" className="text-blue-600 hover:underline">← Back to Finance Dashboard</a>
        </div>
      </div>
    </div>
  );
}

export default Ledger;