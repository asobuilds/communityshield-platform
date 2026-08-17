import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';
import Chat from '../../components/Chat';
import ReportButton from '../../components/ReportButton';

function CaseDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [caseData, setCaseData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [evidence, setEvidence] = useState<any[]>([]);
  const [progress, setProgress] = useState<any[]>([]);
  const [officers, setOfficers] = useState<any[]>([]);
  const [assignments, setAssignments] = useState<any[]>([]);
  const [uploading, setUploading] = useState(false);
  const [files, setFiles] = useState<File[]>([]);
  const [description, setDescription] = useState('');
  const [progressAction, setProgressAction] = useState('');
  const [progressDescription, setProgressDescription] = useState('');
  const [addingProgress, setAddingProgress] = useState(false);
  const [availableOfficers, setAvailableOfficers] = useState<any[]>([]);
  const [selectedOfficerId, setSelectedOfficerId] = useState('');
  const [assigning, setAssigning] = useState(false);
  const user = JSON.parse(localStorage.getItem('user') || '{}');
  const isAdmin = user.role === 'unit_admin';
  const isReporter = user.id && caseData?.reportedBy === user.id;

  const [showTransferModal, setShowTransferModal] = useState(false);
  const [transferOfficerId, setTransferOfficerId] = useState('');
  const [transferUnitId, setTransferUnitId] = useState('');
  const [transferReason, setTransferReason] = useState('');
  const [transferLoading, setTransferLoading] = useState(false);
  const [pendingTransfers, setPendingTransfers] = useState<any[]>([]);

  useEffect(() => {
    fetchCase();
    fetchEvidence();
    fetchProgress();
    fetchAssignedOfficers();
    if (isAdmin) {
      fetchAvailableOfficers();
      fetchPendingTransfers();
    }
  }, [id]);

  const fetchCase = async () => {
    try {
      const response = await axios.get(`/api/v1/cases/${id}`);
      setCaseData(response.data.case);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load case');
      navigate('/my-cases');
    } finally {
      setLoading(false);
    }
  };

  const fetchEvidence = async () => {
    try {
      const response = await axios.get(`/api/v1/cases/${id}/evidence`);
      setEvidence(response.data.evidence || []);
    } catch (error) {
      console.error('Failed to fetch evidence', error);
    }
  };

  const fetchProgress = async () => {
    try {
      const response = await axios.get(`/api/v1/cases/${id}/progress`);
      setProgress(response.data.progress || []);
    } catch (error) {
      console.error('Failed to fetch progress', error);
    }
  };

  const fetchAssignedOfficers = async () => {
    try {
      const response = await axios.get(`/api/v1/cases/${id}/officers`);
      setOfficers(response.data.officers || []);
      setAssignments(response.data.assignments || []);
    } catch (error) {
      console.error('Failed to fetch assigned officers', error);
    }
  };

  const fetchAvailableOfficers = async () => {
    try {
      const userStr = localStorage.getItem('user');
      if (!userStr) return;
      const userData = JSON.parse(userStr);
      if (!userData.unitId) return;
      const response = await axios.get(`/api/v1/admin/units/${userData.unitId}/officers`);
      const allOfficers = response.data.officers || [];
      const assignedIds = officers.map((o: any) => o.id);
      const available = allOfficers.filter((o: any) => !assignedIds.includes(o.id));
      setAvailableOfficers(available);
    } catch (error) {
      console.error('Failed to fetch available officers', error);
    }
  };

  const fetchPendingTransfers = async () => {
    if (!user.unitId) return;
    try {
      const response = await axios.get(`/api/v1/cases/transfers/pending?unitId=${user.unitId}`);
      setPendingTransfers(response.data.transfers || []);
    } catch (error) {
      console.error('Failed to fetch pending transfers', error);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const selectedFiles = Array.from(e.target.files);
      setFiles(selectedFiles);
    }
  };

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (files.length === 0) {
      toast.error('Please select at least one file');
      return;
    }
    setUploading(true);
    let successCount = 0;
    for (const file of files) {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('description', description || file.name);
      formData.append('uploadedBy', user.id);
      try {
        await axios.post(`/api/v1/cases/${id}/evidence`, formData, {
          headers: { 'Content-Type': 'multipart/form-data' },
        });
        successCount++;
      } catch (error: any) {
        console.error('Upload failed for file:', file.name, error);
        toast.error(`Failed to upload ${file.name}`);
      }
    }
    if (successCount > 0) {
      toast.success(`Uploaded ${successCount} file(s) successfully`);
      setFiles([]);
      setDescription('');
      fetchEvidence();
    }
    setUploading(false);
  };

  const handleAddProgress = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!progressAction) {
      toast.error('Action is required');
      return;
    }
    setAddingProgress(true);
    try {
      await axios.post(`/api/v1/cases/${id}/progress`, {
        officerId: user.id,
        action: progressAction,
        description: progressDescription,
      });
      toast.success('Progress added');
      setProgressAction('');
      setProgressDescription('');
      fetchProgress();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to add progress');
    } finally {
      setAddingProgress(false);
    }
  };

  const handleAssignOfficer = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedOfficerId) {
      toast.error('Please select an officer');
      return;
    }
    setAssigning(true);
    try {
      await axios.post(`/api/v1/cases/${id}/assign`, {
        officerId: selectedOfficerId,
        role: 'investigator',
      });
      toast.success('Officer assigned to case');
      setSelectedOfficerId('');
      fetchAssignedOfficers();
      fetchAvailableOfficers();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to assign officer');
    } finally {
      setAssigning(false);
    }
  };

  const handleRemoveOfficer = async (officerId: string) => {
    if (!confirm('Remove this officer from the case?')) return;
    try {
      await axios.delete(`/api/v1/cases/${id}/assign?officerId=${officerId}`);
      toast.success('Officer removed from case');
      fetchAssignedOfficers();
      fetchAvailableOfficers();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to remove officer');
    }
  };

  const handleRequestTransfer = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!transferOfficerId && !transferUnitId) {
      toast.error('Please select a new officer or unit');
      return;
    }
    setTransferLoading(true);
    try {
      await axios.post(
        `/api/v1/cases/${id}/request-transfer`,
        {
          toOfficerId: transferOfficerId || undefined,
          toUnitId: transferUnitId || undefined,
          reason: transferReason || 'No reason provided',
        },
        {
          headers: { 'X-User-ID': user.id },
        }
      );
      toast.success('Transfer requested successfully');
      setShowTransferModal(false);
      setTransferOfficerId('');
      setTransferUnitId('');
      setTransferReason('');
      if (isAdmin) {
        fetchPendingTransfers();
      }
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Transfer request failed');
    } finally {
      setTransferLoading(false);
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

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'low': return 'text-green-600';
      case 'medium': return 'text-yellow-600';
      case 'high': return 'text-orange-600';
      case 'critical': return 'text-red-600';
      default: return 'text-gray-600';
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 dark:bg-gray-900 flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400 text-xl">Loading...</p>
      </div>
    );
  }

  if (!caseData) {
    return (
      <div className="min-h-screen bg-gray-100 dark:bg-gray-900 flex items-center justify-center">
        <p className="text-red-500 text-xl">Case not found</p>
      </div>
    );
  }

  const isClosed = caseData.status === 'closed';
  const canRequestTransfer = isReporter && !isClosed && caseData.status !== 'resolved';

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-4xl mx-auto bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
        <div className="flex justify-between items-start mb-6 flex-wrap gap-2">
          <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400">{caseData.title}</h1>
          <div className="flex items-center gap-3">
            <span className={`px-3 py-1 rounded-full text-sm font-medium ${getStatusColor(caseData.status)}`}>
              {caseData.status.charAt(0).toUpperCase() + caseData.status.slice(1)}
            </span>
            {user.id && (
              <ReportButton
                targetType="case"
                targetId={id || ''}
                userId={user.id}
              />
            )}
          </div>
        </div>

        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Description</label>
          <p className="mt-1 text-gray-800 dark:text-gray-200 bg-gray-50 dark:bg-gray-700 p-3 rounded">{caseData.description}</p>
        </div>

        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Location</label>
            <p className="mt-1 text-gray-800 dark:text-gray-200">{caseData.location || 'Not specified'}</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Priority</label>
            <p className={`mt-1 font-semibold ${getPriorityColor(caseData.priority)}`}>
              {caseData.priority.charAt(0).toUpperCase() + caseData.priority.slice(1)}
            </p>
          </div>
        </div>

        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Reported On</label>
          <p className="mt-1 text-gray-700 dark:text-gray-300">{formatDate(caseData.createdAt)}</p>
        </div>

        {/* Assigned Officers Section */}
        <div className="mt-8 border-t dark:border-gray-700 pt-6">
          <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">👮 Assigned Officers</h2>
          {officers.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400">No officers assigned yet.</p>
          ) : (
            <ul className="space-y-2">
              {officers.map((officer: any, index: number) => {
                const assignment = assignments[index];
                return (
                  <li key={officer.id} className="flex justify-between items-center bg-gray-50 dark:bg-gray-700 p-3 rounded">
                    <span className="text-gray-800 dark:text-gray-200">
                      <strong>{officer.firstName} {officer.lastName}</strong>
                      <span className="text-sm text-gray-500 dark:text-gray-400 ml-2">({assignment?.role || 'investigator'})</span>
                    </span>
                    {isAdmin && !isClosed && (
                      <button
                        onClick={() => handleRemoveOfficer(officer.id)}
                        className="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 text-sm"
                      >
                        Remove
                      </button>
                    )}
                  </li>
                );
              })}
            </ul>
          )}
          {isAdmin && !isClosed && availableOfficers.length > 0 && (
            <form onSubmit={handleAssignOfficer} className="mt-4 flex gap-2">
              <select
                value={selectedOfficerId}
                onChange={(e) => setSelectedOfficerId(e.target.value)}
                className="flex-1 px-3 py-2 border rounded dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                required
              >
                <option value="">Select an officer...</option>
                {availableOfficers.map((officer: any) => (
                  <option key={officer.id} value={officer.id}>
                    {officer.firstName} {officer.lastName}
                  </option>
                ))}
              </select>
              <button
                type="submit"
                disabled={assigning}
                className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
              >
                {assigning ? 'Assigning...' : 'Assign Officer'}
              </button>
            </form>
          )}
        </div>

        {/* Evidence Section */}
        {!isClosed && (
          <div className="mt-8 border-t dark:border-gray-700 pt-6">
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">📎 Evidence</h2>
            <form onSubmit={handleUpload} className="mb-4">
              <div className="flex flex-wrap gap-2">
                <input
                  type="file"
                  multiple
                  onChange={handleFileChange}
                  className="flex-1 px-3 py-2 border rounded dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  accept="image/*,video/*,audio/*,.pdf,.doc,.docx,.txt,.zip,.rar"
                />
                <input
                  type="text"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Description (shared for all files)"
                  className="flex-1 px-3 py-2 border rounded dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                />
                <button
                  type="submit"
                  disabled={uploading || files.length === 0}
                  className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
                >
                  {uploading ? 'Uploading...' : `Upload ${files.length} file(s)`}
                </button>
              </div>
              {files.length > 0 && (
                <ul className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                  {files.map((f, i) => (
                    <li key={i}>{f.name} ({Math.round(f.size / 1024)} KB)</li>
                  ))}
                </ul>
              )}
            </form>
            {evidence.length > 0 && (
              <ul className="space-y-2">
                {evidence.map((ev) => (
                  <li key={ev.id} className="flex justify-between items-center bg-gray-50 dark:bg-gray-700 p-3 rounded">
                    <span className="text-gray-800 dark:text-gray-200"><strong>{ev.fileType}</strong> – {ev.description || 'No description'}</span>
                    <a
                      href={`http://localhost:8080${ev.fileUrl}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 dark:text-blue-400 hover:underline"
                    >
                      View
                    </a>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}

        {/* Progress Timeline */}
        <div className="mt-8 border-t dark:border-gray-700 pt-6">
          <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">📋 Progress Timeline</h2>
          {isAdmin && !isClosed && (
            <form onSubmit={handleAddProgress} className="mb-6 p-4 bg-gray-50 dark:bg-gray-700 rounded border dark:border-gray-600">
              <h3 className="font-medium text-gray-800 dark:text-gray-200 mb-2">Add Progress Update</h3>
              <div className="flex flex-wrap gap-2">
                <input
                  type="text"
                  value={progressAction}
                  onChange={(e) => setProgressAction(e.target.value)}
                  placeholder="Action (e.g., Assigned, Investigated)"
                  className="flex-1 px-3 py-2 border rounded dark:bg-gray-800 dark:border-gray-600 dark:text-white"
                  required
                />
                <input
                  type="text"
                  value={progressDescription}
                  onChange={(e) => setProgressDescription(e.target.value)}
                  placeholder="Details (optional)"
                  className="flex-1 px-3 py-2 border rounded dark:bg-gray-800 dark:border-gray-600 dark:text-white"
                />
                <button
                  type="submit"
                  disabled={addingProgress}
                  className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 disabled:opacity-50"
                >
                  {addingProgress ? 'Adding...' : 'Add'}
                </button>
              </div>
            </form>
          )}
          {progress.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400">No progress updates yet.</p>
          ) : (
            <div className="space-y-4">
              {progress.map((p) => (
                <div key={p.id} className="border-l-4 border-blue-600 pl-4 py-2 bg-gray-50 dark:bg-gray-700 rounded-r">
                  <p className="font-medium text-blue-700 dark:text-blue-300">{p.action}</p>
                  {p.description && <p className="text-gray-700 dark:text-gray-300">{p.description}</p>}
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{formatDate(p.createdAt)}</p>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Transfer Section */}
        <div className="mt-8 border-t dark:border-gray-700 pt-6">
          <div className="flex justify-between items-center mb-4">
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200">🔄 Case Transfer</h2>
            {canRequestTransfer && (
              <button
                onClick={() => {
                  console.log('Transfer button clicked');
                  setShowTransferModal(true);
                }}
                className="bg-yellow-600 text-white px-4 py-2 rounded hover:bg-yellow-700 text-sm"
              >
                Request Transfer
              </button>
            )}
          </div>

          {isAdmin && pendingTransfers.length > 0 && (
            <div className="mb-4">
              <h3 className="font-medium text-gray-700 dark:text-gray-300 mb-2">Pending Transfers</h3>
              <ul className="space-y-2">
                {pendingTransfers.map((t: any) => (
                  <li key={t.transfer.id} className="flex justify-between items-center bg-gray-50 dark:bg-gray-700 p-3 rounded">
                    <div>
                      <p className="text-sm font-medium text-gray-800 dark:text-gray-200">{t.case.title}</p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">{t.transfer.reason}</p>
                    </div>
                    <div className="flex gap-2">
                      <button
                        onClick={async () => {
                          try {
                            await axios.put(`/api/v1/cases/${t.transfer.id}/approve`, {}, {
                              headers: { 'X-Admin-ID': user.id }
                            });
                            toast.success('Transfer approved');
                            fetchPendingTransfers();
                            fetchCase();
                          } catch (error: any) {
                            toast.error(error.response?.data?.error || 'Failed to approve');
                          }
                        }}
                        className="bg-green-600 text-white px-3 py-1 rounded text-sm hover:bg-green-700"
                      >
                        Approve
                      </button>
                      <button
                        onClick={async () => {
                          try {
                            await axios.put(`/api/v1/cases/${t.transfer.id}/reject`, {}, {
                              headers: { 'X-Admin-ID': user.id }
                            });
                            toast.success('Transfer rejected');
                            fetchPendingTransfers();
                          } catch (error: any) {
                            toast.error(error.response?.data?.error || 'Failed to reject');
                          }
                        }}
                        className="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700"
                      >
                        Reject
                      </button>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div>
            <h3 className="font-medium text-gray-700 dark:text-gray-300 mb-2">Transfer History</h3>
            {pendingTransfers.length === 0 ? (
              <p className="text-gray-500 dark:text-gray-400 text-sm">No transfer requests.</p>
            ) : (
              <ul className="space-y-1">
                {pendingTransfers.map((t: any) => (
                  <li key={t.transfer.id} className="text-sm text-gray-600 dark:text-gray-400">
                    <span className="font-medium">{t.transfer.status}</span> – {t.transfer.reason} ({new Date(t.transfer.createdAt).toLocaleDateString()})
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        {/* Chat Section */}
        <div className="mt-8 border-t dark:border-gray-700 pt-6">
          <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">💬 Case Chat</h2>
          <Chat
            room={`case:${id}`}
            userId={user.id}
            userName={`${user.firstName} ${user.lastName}`}
          />
        </div>

        {/* Transfer Modal */}
        {showTransferModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full p-6">
              <h2 className="text-xl font-bold text-gray-800 dark:text-white mb-4">Request Case Transfer</h2>
              <form onSubmit={handleRequestTransfer}>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Transfer to Officer (optional)</label>
                  <input
                    type="text"
                    value={transferOfficerId}
                    onChange={(e) => setTransferOfficerId(e.target.value)}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                    placeholder="Enter officer ID"
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Enter the ID of the officer you want to transfer to.</p>
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Transfer to Unit (optional)</label>
                  <input
                    type="text"
                    value={transferUnitId}
                    onChange={(e) => setTransferUnitId(e.target.value)}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                    placeholder="Enter unit ID"
                  />
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">Enter the ID of the unit you want to transfer to.</p>
                </div>
                <div className="mb-4">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Reason</label>
                  <textarea
                    value={transferReason}
                    onChange={(e) => setTransferReason(e.target.value)}
                    className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                    rows={3}
                    required
                  />
                </div>
                <div className="flex gap-3">
                  <button
                    type="submit"
                    disabled={transferLoading}
                    className="flex-1 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
                  >
                    {transferLoading ? 'Requesting...' : 'Request Transfer'}
                  </button>
                  <button
                    type="button"
                    onClick={() => setShowTransferModal(false)}
                    className="flex-1 bg-gray-300 dark:bg-gray-600 text-gray-800 dark:text-white px-4 py-2 rounded hover:bg-gray-400 dark:hover:bg-gray-500"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* Navigation */}
        <div className="mt-8 flex flex-wrap gap-4">
          <button
            onClick={() => navigate('/my-cases')}
            className="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700"
          >
            ← Back to My Cases
          </button>
          <button
            onClick={() => navigate('/report')}
            className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
          >
            Report Another Case
          </button>
          {isAdmin && (
            <button
              onClick={() => navigate('/admin')}
              className="bg-purple-600 text-white px-4 py-2 rounded hover:bg-purple-700"
            >
              Admin Dashboard
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

export default CaseDetail;