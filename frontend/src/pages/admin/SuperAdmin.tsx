import { useState, useEffect } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';
import { BackButton } from '../../components/common/BackButton';


interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: string;
  status: string;
  phone: string;
  createdAt: string;
  lastLogin: string;
}

interface Stats {
  totalUsers: number;
  totalCases: number;
  totalUnits: number;
  totalOfficers: number;
  totalSOS: number;
  totalSuspects: number;
  dailyActive: number;
  pendingCases: number;
  resolvedCases: number;
}

export default function SuperAdmin() {
  const [users, setUsers] = useState<User[]>([]);
  const [stats, setStats] = useState<Stats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      if (!token) {
        toast.error('Please login first');
        setLoading(false);
        return;
      }

      const [usersRes, statsRes] = await Promise.all([
        axios.get('/api/v1/admin/users', {
          headers: { Authorization: `Bearer ${token}` }
        }),
        axios.get('/api/v1/admin/stats', {
          headers: { Authorization: `Bearer ${token}` }
        })
      ]);
      setUsers(usersRes.data.users || []);
      setStats(statsRes.data);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to fetch data');
    } finally {
      setLoading(false);
    }
  };

  const handleRoleChange = async (userId: string, role: string) => {
    try {
      const token = localStorage.getItem('token');
      await axios.put(`/api/v1/admin/users/${userId}/role`, { role }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      toast.success('Role updated successfully');
      fetchData();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update role');
    }
  };

  const handleSuspend = async (userId: string) => {
    if (!confirm('Are you sure you want to suspend this user?')) return;
    try {
      const token = localStorage.getItem('token');
      await axios.post(`/api/v1/admin/users/${userId}/suspend`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      toast.success('User suspended');
      fetchData();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to suspend user');
    }
  };

  const handleActivate = async (userId: string) => {
    try {
      const token = localStorage.getItem('token');
      await axios.post(`/api/v1/admin/users/${userId}/activate`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      toast.success('User activated');
      fetchData();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to activate user');
    }
  };

  const handleImpersonate = async (userId: string) => {
    if (!confirm('You are about to impersonate this user. Continue?')) return;
    try {
      const token = localStorage.getItem('token');
      const response = await axios.post(`/api/v1/admin/users/${userId}/impersonate`, {}, {
        headers: { Authorization: `Bearer ${token}` }
      });
      localStorage.setItem('token', response.data.token);
      localStorage.setItem('user', JSON.stringify(response.data.user));
      localStorage.setItem('impersonating', 'true');
      toast.success(`Now impersonating user`);
      window.location.href = '/home';
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to impersonate');
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'text-green-600 bg-green-100';
      case 'suspended': return 'text-red-600 bg-red-100';
      case 'pending': return 'text-yellow-600 bg-yellow-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'super_admin': return 'text-purple-600 bg-purple-100';
      case 'unit_admin': return 'text-blue-600 bg-blue-100';
      case 'officer': return 'text-green-600 bg-green-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4 flex justify-center items-center">
        <p className="text-gray-500">Loading...</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-7xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
            🛡️ Super Admin Panel
          </h1>
          <button
            onClick={fetchData}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition"
          >
            🔄 Refresh
          </button>
        </div>
        <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
          <div className="max-w-7xl mx-auto">
            <BackButton />
            {/* rest of content */}
          </div>
        </div>

        {stats && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
              <p className="text-sm text-gray-500 dark:text-gray-400">Total Users</p>
              <p className="text-2xl font-bold">{stats.totalUsers}</p>
            </div>
            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
              <p className="text-sm text-gray-500 dark:text-gray-400">Total Cases</p>
              <p className="text-2xl font-bold">{stats.totalCases}</p>
            </div>
            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
              <p className="text-sm text-gray-500 dark:text-gray-400">Daily Active</p>
              <p className="text-2xl font-bold">{stats.dailyActive}</p>
            </div>
            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
              <p className="text-sm text-gray-500 dark:text-gray-400">Pending Cases</p>
              <p className="text-2xl font-bold text-red-600">{stats.pendingCases}</p>
            </div>
          </div>
        )}

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">User</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Email</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Role</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Status</th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {users.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-4 py-3">
                      <div>
                        <p className="font-medium text-gray-900 dark:text-gray-100">
                          {user.firstName} {user.lastName}
                        </p>
                        <p className="text-xs text-gray-500">{user.phone}</p>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
                      {user.email}
                    </td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getRoleColor(user.role)}`}>
                        {user.role}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(user.status)}`}>
                        {user.status}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-2">
                        <select
                          value={user.role}
                          onChange={(e) => handleRoleChange(user.id, e.target.value)}
                          className="text-xs border rounded px-2 py-1 dark:bg-gray-600"
                        >
                          <option value="citizen">Citizen</option>
                          <option value="officer">Officer</option>
                          <option value="unit_admin">Unit Admin</option>
                          <option value="super_admin">Super Admin</option>
                        </select>
                        {user.status === 'active' ? (
                          <button
                            onClick={() => handleSuspend(user.id)}
                            className="text-xs bg-red-100 text-red-600 px-2 py-1 rounded hover:bg-red-200"
                          >
                            Suspend
                          </button>
                        ) : (
                          <button
                            onClick={() => handleActivate(user.id)}
                            className="text-xs bg-green-100 text-green-600 px-2 py-1 rounded hover:bg-green-200"
                          >
                            Activate
                          </button>
                        )}
                        <button
                          onClick={() => handleImpersonate(user.id)}
                          className="text-xs bg-purple-100 text-purple-600 px-2 py-1 rounded hover:bg-purple-200"
                        >
                          Impersonate
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}