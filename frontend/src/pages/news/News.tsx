import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

interface NewsItem {
  id: string;
  title: string;
  content: string;
  source: string;
  author: string;
  category: string;
  location: string;
  sentiment: string;
  sentimentScore: number;
  threatLevel: string;
  publishedAt: string;
}

interface NewsAlert {
  id: string;
  message: string;
  severity: string;
  alertType: string;
  location: string;
  createdAt: string;
}

export default function News() {
  const [news, setNews] = useState<NewsItem[]>([]);
  const [alerts, setAlerts] = useState<NewsAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [user, setUser] = useState<any>(null);

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      setUser(JSON.parse(userStr));
    }
    fetchNews();
    fetchAlerts();
  }, []);

  const fetchNews = async () => {
    setLoading(true);
    try {
      const response = await axios.get('/api/v1/news');
      setNews(response.data.news || []);
    } catch (error) {
      toast.error('Failed to fetch news');
    } finally {
      setLoading(false);
    }
  };

  const fetchAlerts = async () => {
    try {
      const response = await axios.get('/api/v1/alerts/news');
      setAlerts(response.data.alerts || []);
    } catch (error) {
      console.error('Failed to fetch alerts');
    }
  };

  const getThreatColor = (level: string) => {
    switch (level) {
      case 'low': return 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400';
      case 'medium': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400';
      case 'high': return 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400';
      case 'critical': return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getSentimentIcon = (sentiment: string) => {
    switch (sentiment) {
      case 'positive': return '😊';
      case 'negative': return '😟';
      default: return '😐';
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'low': return 'bg-blue-100 text-blue-800';
      case 'medium': return 'bg-yellow-100 text-yellow-800';
      case 'high': return 'bg-orange-100 text-orange-800';
      case 'critical': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const categories = ['all', 'general', 'security', 'community', 'advisory'];
  const filteredNews = selectedCategory === 'all' 
    ? news 
    : news.filter(item => item.category === selectedCategory);

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
            📰 Security News & Alerts
          </h1>
          <button
            onClick={fetchNews}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition"
          >
            🔄 Refresh
          </button>
        </div>

        {/* Alerts Section */}
        {alerts.length > 0 && (
          <div className="mb-6">
            <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-3">
              ⚠️ Active Alerts
            </h2>
            <div className="space-y-2">
              {alerts.map((alert) => (
                <div
                  key={alert.id}
                  className={`p-4 rounded-lg border ${
                    alert.severity === 'critical' 
                      ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800'
                      : alert.severity === 'high'
                      ? 'bg-orange-50 dark:bg-orange-900/20 border-orange-200 dark:border-orange-800'
                      : 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800'
                  }`}
                >
                  <div className="flex justify-between items-start">
                    <div>
                      <span className={`inline-block px-2 py-1 rounded-full text-xs font-medium ${getSeverityColor(alert.severity)}`}>
                        {alert.severity.toUpperCase()}
                      </span>
                      <span className="ml-2 text-sm text-gray-500">{alert.alertType}</span>
                      {alert.location && (
                        <span className="ml-2 text-sm text-gray-500">📍 {alert.location}</span>
                      )}
                    </div>
                    <span className="text-xs text-gray-400">
                      {new Date(alert.createdAt).toLocaleString()}
                    </span>
                  </div>
                  <p className="mt-2 text-sm text-gray-700 dark:text-gray-300 whitespace-pre-wrap">
                    {alert.message}
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Category Filter */}
        <div className="flex flex-wrap gap-2 mb-6">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setSelectedCategory(cat)}
              className={`px-4 py-1 rounded-full text-sm transition ${
                selectedCategory === cat
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
              }`}
            >
              {cat.charAt(0).toUpperCase() + cat.slice(1)}
            </button>
          ))}
        </div>

        {/* News List */}
        {loading ? (
          <div className="text-center py-8 text-gray-500">Loading news...</div>
        ) : filteredNews.length === 0 ? (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center">
            <p className="text-gray-500 dark:text-gray-400">No news available</p>
          </div>
        ) : (
          <div className="grid gap-4">
            {filteredNews.map((item) => (
              <div
                key={item.id}
                className="bg-white dark:bg-gray-800 rounded-lg shadow hover:shadow-md transition p-6"
              >
                <div className="flex justify-between items-start flex-wrap gap-2">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200">
                        {item.title}
                      </h3>
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getThreatColor(item.threatLevel)}`}>
                        {item.threatLevel.toUpperCase()}
                      </span>
                      <span className="text-sm">
                        {getSentimentIcon(item.sentiment)}
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 dark:text-gray-400 mt-2">
                      {item.content.length > 300 
                        ? item.content.substring(0, 300) + '...' 
                        : item.content}
                    </p>
                    <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
                      {item.source && <span>📰 {item.source}</span>}
                      {item.author && <span>✍️ {item.author}</span>}
                      {item.location && <span>📍 {item.location}</span>}
                      <span>📅 {new Date(item.publishedAt).toLocaleDateString()}</span>
                      <span className={`px-2 py-0.5 rounded-full ${item.category === 'security' ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-800'}`}>
                        {item.category}
                      </span>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Link
                      to={`/news/${item.id}`}
                      className="text-blue-600 hover:underline text-sm"
                    >
                      Read More →
                    </Link>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        <div className="mt-6">
          <Link to="/home" className="text-gray-600 dark:text-gray-400 hover:underline">
            ← Back to Home
          </Link>
        </div>
      </div>
    </div>
  );
}