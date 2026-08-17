import { useState, useEffect } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement,
  PointElement,
  LineElement,
} from 'chart.js';
import { Bar, Pie, Line } from 'react-chartjs-2';

ChartJS.register(
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
  ArcElement,
  PointElement,
  LineElement
);

function Analytics() {
  const [loading, setLoading] = useState(true);
  const [caseData, setCaseData] = useState<any>(null);
  const [sosData, setSosData] = useState<any>(null);
  const [unitData, setUnitData] = useState<any>(null);

  useEffect(() => {
    fetchAnalytics();
  }, []);

  const fetchAnalytics = async () => {
    try {
      const [casesRes, sosRes, unitsRes] = await Promise.all([
        axios.get('/api/v1/admin/analytics/cases'),
        axios.get('/api/v1/admin/analytics/sos'),
        axios.get('/api/v1/admin/analytics/units'),
      ]);
      setCaseData(casesRes.data);
      setSosData(sosRes.data);
      setUnitData(unitsRes.data);
    } catch (error) {
      toast.error('Failed to fetch analytics');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="text-center py-12 text-gray-500">Loading analytics...</div>;
  }

  const statusLabels = caseData?.statusCounts?.map((s: any) => s.status) || [];
  const statusCounts = caseData?.statusCounts?.map((s: any) => s.count) || [];
  const priorityLabels = caseData?.priorityCounts?.map((p: any) => p.priority) || [];
  const priorityCounts = caseData?.priorityCounts?.map((p: any) => p.count) || [];
  const priorityColors = {
    low: '#22c55e',
    medium: '#eab308',
    high: '#f97316',
    critical: '#ef4444',
  };
  const sosLabels = sosData?.dailySOS?.map((d: any) => d.date) || [];
  const sosCounts = sosData?.dailySOS?.map((d: any) => d.count) || [];

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400 mb-6">📊 Analytics Dashboard</h1>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <p className="text-sm text-gray-500 dark:text-gray-400">Total Cases</p>
            <p className="text-2xl font-bold text-gray-800 dark:text-white">{caseData?.totalCases || 0}</p>
          </div>
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <p className="text-sm text-gray-500 dark:text-gray-400">Total SOS</p>
            <p className="text-2xl font-bold text-gray-800 dark:text-white">{sosData?.totalSOS || 0}</p>
          </div>
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <p className="text-sm text-gray-500 dark:text-gray-400">Active Cases</p>
            <p className="text-2xl font-bold text-gray-800 dark:text-white">
              {caseData?.statusCounts?.filter((s: any) => s.status === 'pending' || s.status === 'investigating')
                .reduce((acc: any, s: any) => acc + s.count, 0) || 0}
            </p>
          </div>
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <p className="text-sm text-gray-500 dark:text-gray-400">Resolved Cases</p>
            <p className="text-2xl font-bold text-green-600">
              {caseData?.statusCounts?.find((s: any) => s.status === 'resolved')?.count || 0}
            </p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <h2 className="text-lg font-semibold text-gray-800 dark:text-white mb-4">Cases by Status</h2>
            <Pie
              data={{
                labels: statusLabels,
                datasets: [
                  {
                    data: statusCounts,
                    backgroundColor: ['#eab308', '#3b82f6', '#22c55e', '#6b7280'],
                  },
                ],
              }}
              options={{
                plugins: {
                  legend: { position: 'bottom' },
                },
              }}
            />
          </div>

          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <h2 className="text-lg font-semibold text-gray-800 dark:text-white mb-4">Cases by Priority</h2>
            <Bar
              data={{
                labels: priorityLabels,
                datasets: [
                  {
                    label: 'Cases',
                    data: priorityCounts,
                    backgroundColor: priorityLabels.map((p: string) => priorityColors[p as keyof typeof priorityColors] || '#9ca3af'),
                  },
                ],
              }}
              options={{
                plugins: {
                  legend: { display: false },
                },
                scales: {
                  y: { beginAtZero: true },
                },
              }}
            />
          </div>

          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow md:col-span-2">
            <h2 className="text-lg font-semibold text-gray-800 dark:text-white mb-4">SOS Alerts (Last 7 Days)</h2>
            <Line
              data={{
                labels: sosLabels,
                datasets: [
                  {
                    label: 'SOS Count',
                    data: sosCounts,
                    borderColor: '#ef4444',
                    backgroundColor: 'rgba(239, 68, 68, 0.2)',
                    fill: true,
                    tension: 0.3,
                  },
                ],
              }}
              options={{
                responsive: true,
                plugins: {
                  legend: { display: false },
                },
                scales: {
                  y: { beginAtZero: true },
                },
              }}
            />
          </div>

          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow md:col-span-2">
            <h2 className="text-lg font-semibold text-gray-800 dark:text-white mb-4">Unit Performance</h2>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600">
                    <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600 dark:text-gray-300">Unit Name</th>
                    <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600 dark:text-gray-300">Cases Handled</th>
                    <th className="text-left py-3 px-4 text-sm font-semibold text-gray-600 dark:text-gray-300">Avg Rating</th>
                  </tr>
                </thead>
                <tbody>
                  {unitData?.unitStats?.map((unit: any) => (
                    <tr key={unit.unitId} className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{unit.name}</td>
                      <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{unit.caseCount}</td>
                      <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{unit.avgRating?.toFixed(2) || 'N/A'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div className="mt-6">
          <a href="/admin" className="text-blue-600 dark:text-blue-400 hover:underline">← Back to Admin Dashboard</a>
        </div>
      </div>
    </div>
  );
}

export default Analytics;