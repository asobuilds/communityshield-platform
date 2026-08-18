import { useState } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';

interface AISecurityWarningProps {
  location: string;
  onWarningGenerated: (warning: string) => void;
}

export default function AISecurityWarning({ location, onWarningGenerated }: AISecurityWarningProps) {
  const [loading, setLoading] = useState(false);
  const [warning, setWarning] = useState('');
  const [incidentIds, setIncidentIds] = useState<string[]>([]);
  const [showForm, setShowForm] = useState(false);

  const handleGenerateWarning = async () => {
    if (!location) {
      toast.error('Please enter a location');
      return;
    }

    setLoading(true);
    try {
      const response = await axios.post('/api/v1/ai/security-warning', {
        location,
        incidentIds: incidentIds,
      });
      setWarning(response.data.warning);
      onWarningGenerated(response.data.warning);
      toast.success('Security warning generated');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to generate warning');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <div className="flex justify-between items-center mb-3">
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
          ⚠️ AI Security Warning Generator
        </h3>
        <button
          onClick={() => setShowForm(!showForm)}
          className="text-blue-600 hover:underline text-sm"
        >
          {showForm ? 'Hide' : 'Configure'}
        </button>
      </div>

      {showForm && (
        <div className="space-y-3 mb-3">
          <div>
            <input
              type="text"
              placeholder="Location"
              value={location}
              onChange={(e) => setLocation(e.target.value)}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white text-sm"
            />
          </div>
          <div>
            <input
              type="text"
              placeholder="Incident IDs (comma separated)"
              value={incidentIds.join(',')}
              onChange={(e) => setIncidentIds(e.target.value.split(',').filter(id => id.trim()))}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white text-sm"
            />
            <p className="text-xs text-gray-500 mt-1">Leave empty to auto-fetch recent incidents</p>
          </div>
        </div>
      )}

      <button
        onClick={handleGenerateWarning}
        disabled={loading}
        className="w-full bg-red-600 text-white py-2 rounded-lg hover:bg-red-700 transition disabled:opacity-50 text-sm font-medium"
      >
        {loading ? 'Generating...' : '🚨 Generate Security Warning'}
      </button>

      {warning && (
        <div className="mt-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-3 rounded-lg max-h-60 overflow-y-auto">
          <pre className="whitespace-pre-wrap text-xs text-red-800 dark:text-red-300">
            {warning}
          </pre>
        </div>
      )}
    </div>
  );
}