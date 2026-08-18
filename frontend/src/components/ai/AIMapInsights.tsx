import { useState } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';

interface AIMapInsightsProps {
  latitude: number;
  longitude: number;
  onAnalysis: (analysis: string) => void;
}

export default function AIMapInsights({ latitude, longitude, onAnalysis }: AIMapInsightsProps) {
  const [loading, setLoading] = useState(false);
  const [analysis, setAnalysis] = useState('');

  const handleGetInsights = async () => {
    if (!latitude || !longitude) {
      toast.error('No location selected');
      return;
    }

    setLoading(true);
    try {
      const response = await axios.post('/api/v1/ai/map-insights', {
        latitude,
        longitude,
      });
      setAnalysis(response.data.analysis);
      onAnalysis(response.data.analysis);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to get insights');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <div className="flex justify-between items-center mb-3">
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
          🤖 AI Map Insights
        </h3>
        <button
          onClick={handleGetInsights}
          disabled={loading}
          className="bg-purple-600 text-white px-3 py-1 rounded text-sm hover:bg-purple-700 transition disabled:opacity-50"
        >
          {loading ? 'Analyzing...' : 'Get Insights'}
        </button>
      </div>
      {analysis && (
        <div className="bg-gray-50 dark:bg-gray-700 p-3 rounded-lg max-h-40 overflow-y-auto">
          <pre className="whitespace-pre-wrap text-xs text-gray-700 dark:text-gray-300">
            {analysis}
          </pre>
        </div>
      )}
    </div>
  );
}