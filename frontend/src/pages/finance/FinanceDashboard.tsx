import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';
import { useLanguage } from '../../context/LanguageContext';
import { BackButton } from '../../components/common/BackButton';
import { Chart as ChartJS, ArcElement, Tooltip, Legend, CategoryScale, LinearScale, BarElement, Title } from 'chart.js';
import { Doughnut, Bar } from 'react-chartjs-2';

ChartJS.register(ArcElement, Tooltip, Legend, CategoryScale, LinearScale, BarElement, Title);

export default function FinanceDashboard() {
  const { t } = useLanguage();
  const [loading, setLoading] = useState(true);
  const [summary, setSummary] = useState<any>(null);
  const [transactions, setTransactions] = useState<any[]>([]);
  const [budgets, setBudgets] = useState<any[]>([]);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      // For demo, use a default unit ID or get from user
      const unitId = '13c545a1-dc45-4f61-9d07-4483ce629bb4'; // Replace with actual unit ID
      
      const [summaryRes, transactionsRes, budgetsRes] = await Promise.all([
        axios.get(`/api/v1/finance/units/${unitId}/summary`, {
          headers: { Authorization: `Bearer ${token}` }
        }).catch(() => ({ data: {} })),
        axios.get(`/api/v1/finance/units/${unitId}/transactions`, {
          headers: { Authorization: `Bearer ${token}` }
        }).catch(() => ({ data: { transactions: [] } })),
        axios.get(`/api/v1/finance/units/${unitId}/budgets`, {
          headers: { Authorization: `Bearer ${token}` }
        }).catch(() => ({ data: { budgets: [] } }))
      ]);

      setSummary(summaryRes.data);
      setTransactions(transactionsRes.data.transactions || []);
      setBudgets(budgetsRes.data.budgets || []);
    } catch (error) {
      toast.error('Failed to fetch financial data');
    } finally {
      setLoading(false);
    }
  };

  const doughnutData = {
    labels: ['Income', 'Expenses'],
    datasets: [
      {
        data: [summary?.totalIncome || 0, summary?.totalExpenses || 0],
        backgroundColor: ['#22c55e', '#ef4444'],
        borderColor: ['#16a34a', '#dc2626'],
        borderWidth: 1,
      },
    ],
  };

  const barData = {
    labels: ['Income', 'Expenses', 'Balance'],
    datasets: [
      {
        label: 'Financial Overview',
        data: [summary?.totalIncome || 0, summary?.totalExpenses || 0, summary?.balance || 0],
        backgroundColor: ['#22c55e', '#ef4444', '#3b82f6'],
        borderColor: ['#16a34a', '#dc2626', '#2563eb'],
        borderWidth: 1,
      },
    ],
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[400px]">
        <p className="text-gray-500">{t('common.loading')}</p>
      </div>
    );
  }

  return (
    <div>
      <BackButton />
      
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
          💰 {t('finance.title')}
        </h1>
        <Link to="/add-transaction">
          <button className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition">
            + {t('finance.addTransaction')}
          </button>
        </Link>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <p className="text-sm text-gray-500 dark:text-gray-400">{t('finance.income')}</p>
          <p className="text-2xl font-bold text-green-600">
            ₦{summary?.totalIncome?.toLocaleString() || 0}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <p className="text-sm text-gray-500 dark:text-gray-400">{t('finance.expenses')}</p>
          <p className="text-2xl font-bold text-red-600">
            ₦{summary?.totalExpenses?.toLocaleString() || 0}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <p className="text-sm text-gray-500 dark:text-gray-400">{t('finance.balance')}</p>
          <p className={`text-2xl font-bold ${(summary?.balance || 0) >= 0 ? 'text-blue-600' : 'text-red-600'}`}>
            ₦{summary?.balance?.toLocaleString() || 0}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <p className="text-sm text-gray-500 dark:text-gray-400">{t('finance.pending')}</p>
          <p className="text-2xl font-bold text-yellow-600">
            {summary?.pendingCount || 0}
          </p>
        </div>
      </div>

      {/* Charts */}
      <div className="grid md:grid-cols-2 gap-6 mb-6">
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-4">
            Income vs Expenses
          </h3>
          <div className="h-64 flex justify-center">
            <Doughnut data={doughnutData} options={{ responsive: true, maintainAspectRatio: false }} />
          </div>
        </div>
        <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
          <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-4">
            Financial Overview
          </h3>
          <div className="h-64">
            <Bar 
              data={barData} 
              options={{ 
                responsive: true, 
                maintainAspectRatio: false,
                scales: {
                  y: {
                    beginAtZero: true
                  }
                }
              }} 
            />
          </div>
        </div>
      </div>

      {/* Recent Transactions */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
            📋 {t('finance.transactions')}
          </h3>
          <Link to="/ledger" className="text-blue-600 hover:underline text-sm">
            View All →
          </Link>
        </div>
        {transactions.length === 0 ? (
          <p className="text-gray-500 dark:text-gray-400 text-center py-8">
            No transactions found
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b dark:border-gray-700">
                  <th className="text-left py-2 text-sm font-medium text-gray-500 dark:text-gray-400">Date</th>
                  <th className="text-left py-2 text-sm font-medium text-gray-500 dark:text-gray-400">Description</th>
                  <th className="text-left py-2 text-sm font-medium text-gray-500 dark:text-gray-400">Type</th>
                  <th className="text-right py-2 text-sm font-medium text-gray-500 dark:text-gray-400">Amount</th>
                  <th className="text-center py-2 text-sm font-medium text-gray-500 dark:text-gray-400">Status</th>
                </tr>
              </thead>
              <tbody>
                {transactions.slice(0, 5).map((tx) => (
                  <tr key={tx.id} className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="py-2 text-sm text-gray-600 dark:text-gray-400">
                      {new Date(tx.createdAt).toLocaleDateString()}
                    </td>
                    <td className="py-2 text-sm text-gray-800 dark:text-gray-200">{tx.description}</td>
                    <td className="py-2 text-sm">
                      <span className={`px-2 py-1 rounded-full text-xs ${
                        tx.type === 'donation' || tx.type === 'gift' 
                          ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                          : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                      }`}>
                        {tx.type}
                      </span>
                    </td>
                    <td className={`py-2 text-sm text-right font-medium ${
                      tx.type === 'donation' || tx.type === 'gift' 
                        ? 'text-green-600' 
                        : 'text-red-600'
                    }`}>
                      ₦{tx.amount?.toLocaleString() || 0}
                    </td>
                    <td className="py-2 text-center">
                      <span className={`px-2 py-1 rounded-full text-xs ${
                        tx.status === 'approved' 
                          ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                          : tx.status === 'pending'
                          ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400'
                          : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                      }`}>
                        {tx.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}