import { useState, useEffect } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';
import { BackButton } from '../../components/common/BackButton';

function FinanceDashboard() {
  const [summary, setSummary] = useState({
    totalIncome: 0,
    totalExpenses: 0,
    balance: 0,
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const unitId = '00000000-0000-0000-0000-000000000000';
    fetchSummary(unitId);
  }, []);

  const fetchSummary = async (id: string) => {
    try {
      const response = await axios.get(`/api/v1/transactions/units/${id}/report`);
      setSummary(response.data);
    } catch (error: any) {
      toast.error('Failed to fetch summary');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold text-blue-600 mb-6">💰 Finance Dashboard</h1>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
          <div className="bg-white rounded-lg shadow p-6 text-center">
            <p className="text-gray-500 text-sm">Total Income</p>
            <p className="text-2xl font-bold text-green-600">₦{summary.totalIncome?.toFixed(2) || '0.00'}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6 text-center">
            <p className="text-gray-500 text-sm">Total Expenses</p>
            <p className="text-2xl font-bold text-red-600">₦{summary.totalExpenses?.toFixed(2) || '0.00'}</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6 text-center">
            <p className="text-gray-500 text-sm">Balance</p>
            <p className="text-2xl font-bold text-blue-600">₦{summary.balance?.toFixed(2) || '0.00'}</p>
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <a href="/add-transaction" className="bg-green-600 text-white px-6 py-4 rounded-lg hover:bg-green-700 text-center font-semibold">➕ Add Transaction</a>
          <a href="/ledger" className="bg-blue-600 text-white px-6 py-4 rounded-lg hover:bg-blue-700 text-center font-semibold">📒 View Ledger</a>
        </div>
        <div className="mt-6">
          <BackButton to="/dashboard" label="← Back to Dashboard" />
        </div>
      </div>
    </div>
  );
}

export default FinanceDashboard;