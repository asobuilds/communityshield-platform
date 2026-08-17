import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function AISummary() {
  const navigate = useNavigate();
  const [text, setText] = useState('');
  const [summary, setSummary] = useState('');
  const [loading, setLoading] = useState(false);
  const [analysis, setAnalysis] = useState<any>(null);

  const handleSummarize = async () => {
    if (!text.trim()) {
      toast.error('Please enter text to summarize');
      return;
    }
    setLoading(true);
    try {
      const response = await axios.post('/api/v1/ai/summarize', { text });
      setSummary(response.data.summary);
      setAnalysis(response.data.analysis);
      toast.success('Summary generated!');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to summarize');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-4xl mx-auto bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
        <button
          onClick={() => navigate('/dashboard')}
          className="text-blue-600 dark:text-blue-400 hover:underline mb-4 flex items-center gap-1"
        >
          ← Back to Dashboard
        </button>

        <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400 mb-6">🤖 AI Security Intelligence</h1>

        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Enter text to analyze (case description, news, etc.)
          </label>
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            rows={6}
            className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            placeholder="Paste any text for AI analysis..."
          />
          <button
            onClick={handleSummarize}
            disabled={loading}
            className="mt-3 bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? 'Analyzing...' : '🔍 Analyze with AI'}
          </button>
        </div>

        {summary && (
          <div className="mt-6 border-t dark:border-gray-700 pt-6">
            <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-3">📊 AI Analysis</h2>
            <div className="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
              <p className="text-gray-700 dark:text-gray-300 whitespace-pre-wrap">{summary}</p>
            </div>

            {analysis && (
              <div className="mt-4 grid grid-cols-2 gap-4">
                <div className="bg-green-50 dark:bg-green-900 p-3 rounded">
                  <p className="text-sm text-gray-600 dark:text-gray-400">Sentiment</p>
                  <p className="font-semibold capitalize">{analysis.sentiment || 'N/A'}</p>
                </div>
                <div className="bg-red-50 dark:bg-red-900 p-3 rounded">
                  <p className="text-sm text-gray-600 dark:text-gray-400">Threat Level</p>
                  <p className={`font-semibold capitalize ${analysis.threatLevel === 'critical' ? 'text-red-600' : analysis.threatLevel === 'high' ? 'text-orange-600' : 'text-yellow-600'}`}>
                    {analysis.threatLevel || 'N/A'}
                  </p>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default AISummary;