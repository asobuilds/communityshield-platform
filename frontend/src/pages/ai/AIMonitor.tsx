import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function AIMonitor() {
  const navigate = useNavigate();
  const [monitoring, setMonitoring] = useState(false);
  const [rssResults, setRssResults] = useState<any>(null);
  const [socialResults, setSocialResults] = useState<any>(null);
  const [twitterResults, setTwitterResults] = useState<any>(null);

  const monitorRSS = async () => {
    setMonitoring(true);
    try {
      const response = await axios.post('/api/v1/ai/monitor/rss');
      setRssResults(response.data);
      toast.success(`Found ${response.data.count} new security alerts from RSS`);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to monitor RSS');
    } finally {
      setMonitoring(false);
    }
  };

  const monitorSocial = async () => {
    setMonitoring(true);
    try {
      const response = await axios.post('/api/v1/ai/monitor/social');
      setSocialResults(response.data);
      toast.success(`Found ${response.data.count} new alerts from social media`);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to monitor social media');
    } finally {
      setMonitoring(false);
    }
  };

  const monitorTwitter = async () => {
    setMonitoring(true);
    try {
      const response = await axios.get('/api/v1/twitter/monitor');
      setTwitterResults(response.data);
      toast.success(`Found ${response.data.count} new security alerts from Twitter`);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to monitor Twitter');
    } finally {
      setMonitoring(false);
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

        <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400 mb-6">📡 AI & Social Security Monitor</h1>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <button
            onClick={monitorRSS}
            disabled={monitoring}
            className="bg-green-600 text-white px-6 py-4 rounded-lg hover:bg-green-700 disabled:opacity-50 text-center"
          >
            📰 Scan RSS Feeds
            <p className="text-sm opacity-75 mt-1">Check Nigerian security news</p>
          </button>
          <button
            onClick={monitorSocial}
            disabled={monitoring}
            className="bg-purple-600 text-white px-6 py-4 rounded-lg hover:bg-purple-700 disabled:opacity-50 text-center"
          >
            🌐 Scan Social Media
            <p className="text-sm opacity-75 mt-1">Monitor Twitter, Facebook (simulated)</p>
          </button>
          <button
            onClick={monitorTwitter}
            disabled={monitoring}
            className="bg-blue-600 text-white px-6 py-4 rounded-lg hover:bg-blue-700 disabled:opacity-50 text-center"
          >
            🐦 Scan Twitter (Real)
            <p className="text-sm opacity-75 mt-1">Real-time security tweets</p>
          </button>
        </div>

        {monitoring && (
          <div className="text-center py-8">
            <p className="text-gray-600 dark:text-gray-400">🔄 Scanning for threats...</p>
          </div>
        )}

        {rssResults && (
          <div className="mt-4 border-t dark:border-gray-700 pt-4">
            <h3 className="font-semibold text-gray-800 dark:text-gray-200">📰 RSS Feed Results</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">{rssResults.count} new alerts created</p>
          </div>
        )}

        {socialResults && (
          <div className="mt-4 border-t dark:border-gray-700 pt-4">
            <h3 className="font-semibold text-gray-800 dark:text-gray-200">🌐 Social Media Results</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">{socialResults.count} new alerts created</p>
          </div>
        )}

        {twitterResults && (
          <div className="mt-4 border-t dark:border-gray-700 pt-4">
            <h3 className="font-semibold text-gray-800 dark:text-gray-200">🐦 Twitter Results</h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">{twitterResults.count} new alerts created</p>
          </div>
        )}
      </div>
    </div>
  );
}

export default AIMonitor;