import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';
import WalkieTalkie from '../../components/WalkieTalkie';

function AdminDashboard() {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('sos');
  const [sosAlerts, setSosAlerts] = useState<any[]>([]);
  const [cases, setCases] = useState<any[]>([]);
  const [archivedCases, setArchivedCases] = useState<any[]>([]);
  const [officers, setOfficers] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [officerLoading, setOfficerLoading] = useState(false);
  const [newOfficerEmail, setNewOfficerEmail] = useState('');
  const [newOfficerRank, setNewOfficerRank] = useState('officer');
  const [unitId, setUnitId] = useState('');
  const [unitInfo, setUnitInfo] = useState<any>(null);
  const [unitPictureLoading, setUnitPictureLoading] = useState(false);
  const [officersOfWeek, setOfficersOfWeek] = useState<any[]>([]);
  const user = JSON.parse(localStorage.getItem('user') || '{}');

  // Filter states
  const [sosSearch, setSosSearch] = useState('');
  const [sosStatusFilter, setSosStatusFilter] = useState('');
  const [sosStartDate, setSosStartDate] = useState('');
  const [sosEndDate, setSosEndDate] = useState('');
  const [caseSearch, setCaseSearch] = useState('');
  const [caseStatusFilter, setCaseStatusFilter] = useState('');
  const [casePriorityFilter, setCasePriorityFilter] = useState('');
  const [caseStartDate, setCaseStartDate] = useState('');
  const [caseEndDate, setCaseEndDate] = useState('');

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      try {
        const userData = JSON.parse(userStr);
        if (userData.unitId) {
          setUnitId(userData.unitId);
          fetchUnitInfo(userData.unitId);
          fetchOfficersOfWeek(userData.unitId);
        }
      } catch (e) {}
    }
    fetchSOSAlerts();
    fetchCases();
    fetchArchivedCases();
    fetchOfficers();
  }, []);

  const fetchUnitInfo = async (id: string) => {
    try {
      const response = await axios.get(`/api/v1/admin/units/${id}`);
      setUnitInfo(response.data.unit);
    } catch (error) {
      console.error('Failed to fetch unit info', error);
    }
  };

  const fetchOfficersOfWeek = async (id: string) => {
    try {
      const response = await axios.get(`/api/v1/officers/week?unitId=${id}`);
      setOfficersOfWeek(response.data.officers || []);
    } catch (error) {
      console.error('Failed to fetch officers of the week', error);
    }
  };

  const fetchSOSAlerts = async () => {
    try {
      let url = '/api/v1/admin/sos?';
      if (sosSearch) url += `search=${encodeURIComponent(sosSearch)}&`;
      if (sosStatusFilter) url += `status=${sosStatusFilter}&`;
      if (sosStartDate) url += `startDate=${sosStartDate}&`;
      if (sosEndDate) url += `endDate=${sosEndDate}&`;
      const response = await axios.get(url);
      setSosAlerts(response.data.sos || []);
    } catch (error) {
      toast.error('Failed to fetch SOS alerts');
    }
  };

  const fetchCases = async () => {
    try {
      let url = '/api/v1/admin/cases?';
      if (caseSearch) url += `search=${encodeURIComponent(caseSearch)}&`;
      if (caseStatusFilter) url += `status=${caseStatusFilter}&`;
      if (casePriorityFilter) url += `priority=${casePriorityFilter}&`;
      if (caseStartDate) url += `startDate=${caseStartDate}&`;
      if (caseEndDate) url += `endDate=${caseEndDate}&`;
      const response = await axios.get(url);
      setCases(response.data.cases || []);
    } catch (error) {
      toast.error('Failed to fetch cases');
    }
  };

  const fetchArchivedCases = async () => {
    try {
      const response = await axios.get('/api/v1/admin/cases/archived');
      setArchivedCases(response.data.cases || []);
    } catch (error) {
      toast.error('Failed to fetch archived cases');
    }
  };

  const fetchOfficers = async () => {
    if (!unitId) return;
    try {
      const response = await axios.get(`/api/v1/admin/units/${unitId}/officers`);
      setOfficers(response.data.officers || []);
    } catch (error) {
      toast.error('Failed to fetch officers');
    }
  };

  const updateSOSStatus = async (id: string, status: string) => {
    try {
      await axios.patch(`/api/v1/admin/sos/${id}/status`, { status });
      toast.success('SOS status updated');
      fetchSOSAlerts();
    } catch (error) {
      toast.error('Failed to update');
    }
  };

  const updateCaseStatus = async (id: string, status: string) => {
    try {
      await axios.patch(`/api/v1/admin/cases/${id}/status`, { status });
      toast.success('Case status updated');
      fetchCases();
    } catch (error) {
      toast.error('Failed to update');
    }
  };

  const addOfficer = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!unitId || !newOfficerEmail) {
      toast.error('Email is required');
      return;
    }
    setOfficerLoading(true);
    try {
      await axios.post(`/api/v1/admin/units/${unitId}/officers`, {
        email: newOfficerEmail,
        firstName: 'Officer',
        lastName: 'User',
        phone: '',
        rank: newOfficerRank,
        role: 'officer'
      });
      toast.success('Officer added');
      setNewOfficerEmail('');
      setNewOfficerRank('officer');
      fetchOfficers();
    } catch (error) {
      toast.error('Failed to add officer');
    } finally {
      setOfficerLoading(false);
    }
  };

  const removeOfficer = async (userId: string) => {
    if (!confirm('Remove this officer?')) return;
    try {
      await axios.delete(`/api/v1/admin/officers/${userId}`);
      toast.success('Officer removed');
      fetchOfficers();
    } catch (error) {
      toast.error('Failed to remove');
    }
  };

  const handleUnitPictureUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length === 0) return;
    const file = e.target.files[0];
    setUnitPictureLoading(true);
    const formData = new FormData();
    formData.append('profilePicture', file);
    formData.append('unitId', unitId);
    try {
      const response = await axios.put(`/api/v1/admin/units/${unitId}/profile-picture`, formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      toast.success('Unit profile picture updated!');
      fetchUnitInfo(unitId);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Upload failed');
    } finally {
      setUnitPictureLoading(false);
    }
  };

  const restoreArchivedCase = async (id: string) => {
    try {
      await axios.put(`/api/v1/admin/cases/${id}/restore`);
      toast.success('Case restored');
      fetchArchivedCases();
    } catch (error) {
      toast.error('Failed to restore');
    }
  };

  const deleteArchivedCasePermanent = async (id: string) => {
    if (!confirm('Delete this case permanently? This cannot be undone.')) return;
    try {
      await axios.delete(`/api/v1/admin/cases/${id}/permanent`);
      toast.success('Case deleted permanently');
      fetchArchivedCases();
    } catch (error) {
      toast.error('Failed to delete');
    }
  };

  const resetSOSFilters = () => {
    setSosSearch('');
    setSosStatusFilter('');
    setSosStartDate('');
    setSosEndDate('');
    fetchSOSAlerts();
  };

  const resetCaseFilters = () => {
    setCaseSearch('');
    setCaseStatusFilter('');
    setCasePriorityFilter('');
    setCaseStartDate('');
    setCaseEndDate('');
    fetchCases();
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-200 text-yellow-800';
      case 'dispatched': return 'bg-blue-200 text-blue-800';
      case 'resolved': return 'bg-green-200 text-green-800';
      case 'investigating': return 'bg-purple-200 text-purple-800';
      case 'closed': return 'bg-gray-200 text-gray-800';
      default: return 'bg-gray-200 text-gray-800';
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900">
      <div className="bg-white dark:bg-gray-800 shadow-md p-4">
        <div className="max-w-7xl mx-auto flex justify-between items-center flex-wrap gap-2">
          <div className="flex items-center gap-4">
            <button onClick={() => navigate(-1)} className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium">← Back</button>
            <h1 className="text-2xl font-bold text-blue-600 dark:text-blue-400">🛡️ Admin Dashboard</h1>
          </div>
          <div className="flex items-center space-x-4">
            <span className="text-sm text-gray-600 dark:text-gray-300">Admin</span>
            <button onClick={() => { localStorage.removeItem('user'); window.location.href = '/login'; }} className="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 text-sm">Logout</button>
          </div>
        </div>
      </div>

      {unitInfo && (
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 flex flex-col md:flex-row items-center md:items-start gap-4">
            <div className="relative">
              {unitInfo.profileImage ? (
                <img
                  src={`http://localhost:8080${unitInfo.profileImage}`}
                  alt={unitInfo.name}
                  className="w-24 h-24 rounded-full object-cover border-4 border-blue-500"
                />
              ) : (
                <div className="w-24 h-24 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center text-4xl font-bold text-blue-600 dark:text-blue-300 border-4 border-blue-500">
                  {unitInfo.name?.[0]?.toUpperCase() || 'U'}
                </div>
              )}
              <label htmlFor="unitPicture" className="absolute bottom-0 right-0 bg-blue-600 text-white p-1 rounded-full cursor-pointer hover:bg-blue-700 text-xs">📷</label>
              <input id="unitPicture" type="file" accept="image/*" onChange={handleUnitPictureUpload} className="hidden" disabled={unitPictureLoading} />
              {unitPictureLoading && <p className="text-xs text-gray-500 mt-1">Uploading...</p>}
            </div>
            <div className="flex-1 text-center md:text-left">
              <h2 className="text-2xl font-bold text-gray-800 dark:text-white">{unitInfo.name}</h2>
              {unitInfo.motto && <p className="text-sm text-blue-600 dark:text-blue-400 italic">"{unitInfo.motto}"</p>}
              {unitInfo.registrationNumber && <p className="text-sm text-gray-500 dark:text-gray-400">Reg: {unitInfo.registrationNumber}</p>}
              <p className="text-sm text-gray-600 dark:text-gray-300">Type: {unitInfo.type}</p>
              <p className="text-sm text-gray-600 dark:text-gray-300">Contact: {unitInfo.contactPerson} ({unitInfo.contactPhone})</p>
            </div>
          </div>
        </div>
      )}

      <div className="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-7xl mx-auto flex space-x-8 px-4 overflow-x-auto">
          <button onClick={() => setActiveTab('sos')} className={`py-3 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${activeTab === 'sos' ? 'border-red-600 text-red-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}`}>🚨 SOS Alerts ({sosAlerts.length})</button>
          <button onClick={() => setActiveTab('cases')} className={`py-3 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${activeTab === 'cases' ? 'border-blue-600 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}`}>📋 Cases ({cases.length})</button>
          <button onClick={() => setActiveTab('officers')} className={`py-3 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${activeTab === 'officers' ? 'border-green-600 text-green-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}`}>👮 Officers ({officers.length})</button>
          <button onClick={() => setActiveTab('units')} className={`py-3 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${activeTab === 'units' ? 'border-purple-600 text-purple-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}`}>🏢 Units</button>
          <button onClick={() => setActiveTab('walkie')} className={`py-3 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${activeTab === 'walkie' ? 'border-yellow-600 text-yellow-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}`}>📻 Walkie</button>
          <button onClick={() => setActiveTab('analytics')} className={`py-3 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${activeTab === 'analytics' ? 'border-indigo-600 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}`}>📊 Analytics</button>
          <button onClick={() => setActiveTab('officers-week')} className={`py-3 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${activeTab === 'officers-week' ? 'border-pink-600 text-pink-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}`}>⭐ Officers of Week</button>
          <button onClick={() => setActiveTab('archived')} className={`py-3 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${activeTab === 'archived' ? 'border-gray-600 text-gray-600' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}`}>📦 Archived ({archivedCases.length})</button>
        </div>
      </div>

      <div className="max-w-7xl mx-auto p-4">
        {loading ? <div className="text-center py-12 text-gray-500">Loading...</div>
        : activeTab === 'sos' ? (
          <div>
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">SOS Alerts</h2>
            <div className="flex flex-wrap gap-4 mb-4">
              <input type="text" placeholder="Search description..." value={sosSearch} onChange={(e) => setSosSearch(e.target.value)} className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
              <select value={sosStatusFilter} onChange={(e) => setSosStatusFilter(e.target.value)} className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                <option value="">All Statuses</option>
                <option value="pending">Pending</option>
                <option value="dispatched">Dispatched</option>
                <option value="resolved">Resolved</option>
              </select>
              <input type="date" value={sosStartDate} onChange={(e) => setSosStartDate(e.target.value)} className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
              <input type="date" value={sosEndDate} onChange={(e) => setSosEndDate(e.target.value)} className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
              <button onClick={fetchSOSAlerts} className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">Apply Filters</button>
              <button onClick={resetSOSFilters} className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700">Reset</button>
            </div>
            {sosAlerts.length === 0 ? <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center text-gray-500 dark:text-gray-400">No SOS alerts</div>
            : <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-x-auto">
              <table className="w-full min-w-[600px]">
                <thead><tr className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600"><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Date</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Location</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Status</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Action</th></tr></thead>
                <tbody>
                  {sosAlerts.map(sos => (
                    <tr key={sos.id} className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{formatDate(sos.createdAt)}</td>
                      <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{sos.latitude?.toFixed(4)}, {sos.longitude?.toFixed(4)}</td>
                      <td className="py-3 px-4"><span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(sos.status)}`}>{sos.status}</span></td>
                      <td className="py-3 px-4">
                        <select value={sos.status} onChange={(e) => updateSOSStatus(sos.id, e.target.value)} className="text-sm border rounded px-2 py-1 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                          <option value="pending">Pending</option>
                          <option value="dispatched">Dispatched</option>
                          <option value="resolved">Resolved</option>
                        </select>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>}
          </div>
        ) : activeTab === 'cases' ? (
          <div>
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">Cases</h2>
            <div className="flex flex-wrap gap-4 mb-4">
              <input type="text" placeholder="Search title or description..." value={caseSearch} onChange={(e) => setCaseSearch(e.target.value)} className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
              <select value={caseStatusFilter} onChange={(e) => setCaseStatusFilter(e.target.value)} className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                <option value="">All Statuses</option>
                <option value="pending">Pending</option>
                <option value="investigating">Investigating</option>
                <option value="resolved">Resolved</option>
                <option value="closed">Closed</option>
              </select>
              <select value={casePriorityFilter} onChange={(e) => setCasePriorityFilter(e.target.value)} className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                <option value="">All Priorities</option>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
              <input type="date" value={caseStartDate} onChange={(e) => setCaseStartDate(e.target.value)} className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
              <input type="date" value={caseEndDate} onChange={(e) => setCaseEndDate(e.target.value)} className="px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
              <button onClick={fetchCases} className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">Apply Filters</button>
              <button onClick={resetCaseFilters} className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700">Reset</button>
            </div>
            {cases.length === 0 ? <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center text-gray-500 dark:text-gray-400">No cases</div>
            : <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-x-auto">
              <table className="w-full min-w-[600px]">
                <thead><tr className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600"><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Date</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Title</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Status</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Action</th></tr></thead>
                <tbody>
                  {cases.map(c => (
                    <tr key={c.id} className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                      <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{formatDate(c.createdAt)}</td>
                      <td className="py-3 px-4 text-sm font-medium text-gray-800 dark:text-gray-200">{c.title}</td>
                      <td className="py-3 px-4"><span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(c.status)}`}>{c.status}</span></td>
                      <td className="py-3 px-4">
                        <select value={c.status} onChange={(e) => updateCaseStatus(c.id, e.target.value)} className="text-sm border rounded px-2 py-1 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                          <option value="pending">Pending</option>
                          <option value="investigating">Investigating</option>
                          <option value="resolved">Resolved</option>
                          <option value="closed">Closed</option>
                        </select>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>}
          </div>
        ) : activeTab === 'officers' ? (
          <div>
            <div className="flex justify-between items-center mb-4 flex-wrap gap-2">
              <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200">👮 Officers</h2>
              <form onSubmit={addOfficer} className="flex gap-2">
                <input type="email" value={newOfficerEmail} onChange={(e) => setNewOfficerEmail(e.target.value)} placeholder="officer@email.com" className="px-3 py-2 border rounded-lg text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-white" required />
                <select value={newOfficerRank} onChange={(e) => setNewOfficerRank(e.target.value)} className="px-3 py-2 border rounded-lg text-sm dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                  <option value="officer">Officer</option>
                  <option value="sergeant">Sergeant</option>
                  <option value="captain">Captain</option>
                  <option value="director">Director</option>
                  <option value="commander">Commander</option>
                </select>
                <button type="submit" disabled={officerLoading} className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 disabled:opacity-50 text-sm">{officerLoading ? 'Adding...' : '+ Add Officer'}</button>
              </form>
            </div>
            {officers.length === 0 ? <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center text-gray-500 dark:text-gray-400">No officers assigned.</div>
            : <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-x-auto">
              <table className="w-full"><thead><tr className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600"><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Name</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Email</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Rank</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Action</th></tr></thead>
              <tbody>
                {officers.map(o => (
                  <tr key={o.id} className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{o.firstName} {o.lastName}</td>
                    <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{o.email}</td>
                    <td className="py-3 px-4 text-sm capitalize text-gray-800 dark:text-gray-200">{o.rank || 'officer'}</td>
                    <td className="py-3 px-4"><button onClick={() => removeOfficer(o.id)} className="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700">Remove</button></td>
                  </tr>
                ))}
              </tbody>
            </table>
            </div>}
          </div>
        ) : activeTab === 'units' ? (
          <div>
            <div className="flex justify-between items-center mb-4"><h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200">🏢 Security Units</h2><a href="/admin/units" className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">+ Manage Units</a></div>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center text-gray-500 dark:text-gray-400"><p>Manage your security units from the dedicated page.</p><a href="/admin/units" className="mt-4 inline-block text-blue-600 dark:text-blue-400 hover:underline">Go to Unit Management →</a></div>
          </div>
        ) : activeTab === 'walkie' ? (
          <div>
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">📻 Unit Walkie-Talkie</h2>
            {unitId ? (
              <WalkieTalkie unitId={unitId} userId={user.id} userName={`${user.firstName} ${user.lastName}`} />
            ) : (
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center text-gray-500 dark:text-gray-400">You need to be assigned to a unit to use the Walkie-Talkie.</div>
            )}
          </div>
        ) : activeTab === 'analytics' ? (
          <div>
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">📊 Analytics</h2>
            <a href="/admin/analytics" className="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700">View Full Analytics Dashboard</a>
          </div>
        ) : activeTab === 'officers-week' ? (
          <div>
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">⭐ Officers of the Week</h2>
            {officersOfWeek.length === 0 ? (
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center text-gray-500 dark:text-gray-400">No officers have resolved cases this week.</div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {officersOfWeek.map((officer) => (
                  <div key={officer.id} className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-4 border border-gray-200 dark:border-gray-700">
                    <div className="flex items-center gap-3">
                      {officer.profileImage ? (
                        <img src={`http://localhost:8080${officer.profileImage}`} alt={officer.firstName} className="w-12 h-12 rounded-full object-cover border-2 border-blue-500" />
                      ) : (
                        <div className="w-12 h-12 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center text-lg font-bold text-blue-600 dark:text-blue-300">{officer.firstName?.[0]}{officer.lastName?.[0]}</div>
                      )}
                      <div>
                        <p className="font-semibold text-gray-800 dark:text-white">{officer.firstName} {officer.lastName}</p>
                        <p className="text-sm text-gray-500 dark:text-gray-400">{officer.rank || 'officer'}</p>
                        <p className="text-sm text-green-600 dark:text-green-400">✅ {officer.resolvedCount} cases resolved</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        ) : activeTab === 'archived' ? (
          <div>
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">📦 Archived Cases</h2>
            {archivedCases.length === 0 ? (
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center text-gray-500 dark:text-gray-400">No archived cases.</div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full min-w-[600px]">
                  <thead><tr className="bg-gray-50 dark:bg-gray-700 border-b dark:border-gray-600"><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Date</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Title</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Status</th><th className="text-left py-3 px-4 text-gray-600 dark:text-gray-300">Actions</th></tr></thead>
                  <tbody>
                    {archivedCases.map(c => (
                      <tr key={c.id} className="border-b dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                        <td className="py-3 px-4 text-sm text-gray-800 dark:text-gray-200">{formatDate(c.createdAt)}</td>
                        <td className="py-3 px-4 text-sm font-medium text-gray-800 dark:text-gray-200">{c.title}</td>
                        <td className="py-3 px-4"><span className="px-2 py-1 rounded-full text-xs font-medium bg-gray-200 text-gray-800">Archived</span></td>
                        <td className="py-3 px-4 flex gap-2">
                          <button onClick={() => restoreArchivedCase(c.id)} className="bg-green-600 text-white px-2 py-1 rounded text-sm hover:bg-green-700">Restore</button>
                          <button onClick={() => deleteArchivedCasePermanent(c.id)} className="bg-red-600 text-white px-2 py-1 rounded text-sm hover:bg-red-700">Delete</button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        ) : null}
      </div>
    </div>
  );
}

export default AdminDashboard;