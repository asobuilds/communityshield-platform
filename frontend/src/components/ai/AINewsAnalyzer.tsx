import { useState } from 'react';
import axios from 'axios';
import toast from 'react-hot-toast';

interface AINewsAnalyzerProps {
  onAnalysisComplete: (analysis: string) => void;
}

export default function AINewsAnalyzer({ onAnalysisComplete }: AINewsAnalyzerProps) {
  const [content, setContent] = useState('');
  const [source, setSource] = useState('');
  const [loading, setLoading] = useState(false);
  const [analysis, setAnalysis] = useState('');

  const handleAnalyze = async () => {
    if (!content.trim()) {
      toast.error('Please enter news content');
      return;
    }

    setLoading(true);
    try {
      const response = await axios.post('/api/v1/ai/analyze-news', {
        content,
        source: source || 'Manual Entry',
      });
      setAnalysis(response.data.analysis);
      onAnalysisComplete(response.data.analysis);
      toast.success('News analyzed successfully');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Analysis failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">
        📰 AI News Sentiment Analyzer
      </h3>
      <div className="space-y-3">
        <div>
          <input
            type="text"
            placeholder="Source (optional)"
            value={source}
            onChange={(e) => setSource(e.target.value)}
            className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white text-sm"
          />
        </div>
        <div>
          <textarea
            placeholder="Paste news content here..."
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={5}
            className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white text-sm"
          />
        </div>
        <button
          onClick={handleAnalyze}
          disabled={loading}
          className="w-full bg-purple-600 text-white py-2 rounded-lg hover:bg-purple-700 transition disabled:opacity-50 text-sm font-medium"
        >
          {loading ? 'Analyzing...' : 'Analyze News'}
        </button>
        {analysis && (
          <div className="bg-gray-50 dark:bg-gray-700 p-3 rounded-lg max-h-60 overflow-y-auto">
            <pre className="whitespace-pre-wrap text-xs text-gray-700 dark:text-gray-300">
              {analysis}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}