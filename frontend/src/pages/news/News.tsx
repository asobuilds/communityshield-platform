import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function News() {
  const navigate = useNavigate();
  const [articles, setArticles] = useState<any[]>([]);
  const [announcements, setAnnouncements] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [location, setLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [useLocation, setUseLocation] = useState(false);
  const [activeTab, setActiveTab] = useState('all');

  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          setLocation({ lat: pos.coords.latitude, lng: pos.coords.longitude });
          setUseLocation(true);
        },
        () => {
          toast.error('Location access denied – showing all news');
          setUseLocation(false);
        }
      );
    }
  }, []);

  useEffect(() => {
    if (useLocation && location) {
      fetchNews();
      fetchAnnouncements();
    } else {
      fetchNews();
      fetchAnnouncements();
    }
  }, [location, useLocation]);

  const fetchNews = async () => {
    try {
      let url = '/api/v1/news/location';
      if (useLocation && location) {
        url += `?lat=${location.lat}&lng=${location.lng}&radius=50`;
      }
      const response = await axios.get(url);
      setArticles(response.data.articles || []);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to fetch news');
    }
  };

  const fetchAnnouncements = async () => {
    try {
      const response = await axios.get('/api/v1/announcements?public=true');
      setAnnouncements(response.data.announcements || []);
    } catch (error: any) {
      toast.error('Failed to fetch announcements');
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    if (!dateStr) return 'Unknown date';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' });
  };

  const getSeverityBadge = (severity: string) => {
    const colors = {
      low: 'bg-blue-100 text-blue-800',
      medium: 'bg-yellow-100 text-yellow-800',
      high: 'bg-orange-100 text-orange-800',
      critical: 'bg-red-100 text-red-800',
    };
    return colors[severity as keyof typeof colors] || 'bg-gray-100 text-gray-800';
  };

  const allItems = [
    ...articles.map(a => ({ ...a, type: 'news' })),
    ...announcements.map(a => ({ ...a, type: 'announcement', title: a.title, description: a.content, published: a.createdAt, source: 'CommunityShield Admin' }))
  ].sort((a, b) => new Date(b.published || b.createdAt).getTime() - new Date(a.published || a.createdAt).getTime());

  const filteredItems = activeTab === 'all' ? allItems : activeTab === 'news' ? articles.map(a => ({ ...a, type: 'news' })) : announcements.map(a => ({ ...a, type: 'announcement', title: a.title, description: a.content, published: a.createdAt, source: 'CommunityShield Admin' }));

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-4xl mx-auto">
        {/* Back Button */}
        <button
          onClick={() => navigate('/dashboard')}
          className="text-blue-600 dark:text-blue-400 hover:underline mb-4 flex items-center gap-1"
        >
          ← Back to Dashboard
        </button>

        <div className="flex justify-between items-center mb-6 flex-wrap gap-4">
          <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400">📰 News & Alerts</h1>
          <div className="flex gap-2">
            <button
              onClick={() => setUseLocation(!useLocation)}
              className={`px-4 py-2 rounded ${useLocation ? 'bg-blue-600 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-800 dark:text-gray-300'}`}
            >
              {useLocation ? '📍 Near Me' : '📍 Use Location'}
            </button>
            <button
              onClick={() => { fetchNews(); fetchAnnouncements(); }}
              className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
            >
              🔄 Refresh
            </button>
          </div>
        </div>

        {/* Tabs */}
        <div className="flex space-x-4 mb-6 border-b dark:border-gray-700">
          <button
            onClick={() => setActiveTab('all')}
            className={`py-2 px-4 font-medium ${activeTab === 'all' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-500 dark:text-gray-400'}`}
          >
            All
          </button>
          <button
            onClick={() => setActiveTab('news')}
            className={`py-2 px-4 font-medium ${activeTab === 'news' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-500 dark:text-gray-400'}`}
          >
            News
          </button>
          <button
            onClick={() => setActiveTab('announcements')}
            className={`py-2 px-4 font-medium ${activeTab === 'announcements' ? 'border-b-2 border-blue-600 text-blue-600' : 'text-gray-500 dark:text-gray-400'}`}
          >
            Announcements
          </button>
        </div>

        {loading ? (
          <div className="text-center py-12 text-gray-500">Loading...</div>
        ) : filteredItems.length === 0 ? (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center text-gray-500 dark:text-gray-400">
            <p>No news or announcements found.</p>
          </div>
        ) : (
          <div className="space-y-4">
            {filteredItems.map((item, index) => (
              <div
                key={index}
                className={`bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border-l-4 ${
                  item.type === 'announcement'
                    ? item.severity === 'critical' ? 'border-red-500' : item.severity === 'high' ? 'border-orange-500' : 'border-blue-500'
                    : 'border-gray-300'
                }`}
              >
                <div className="flex justify-between items-start">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-gray-500 dark:text-gray-400 uppercase">{item.type}</span>
                      {item.severity && (
                        <span className={`text-xs px-2 py-1 rounded-full ${getSeverityBadge(item.severity)}`}>
                          {item.severity.toUpperCase()}
                        </span>
                      )}
                    </div>
                    <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-1">
                      {item.type === 'announcement' ? item.title : item.title}
                    </h2>
                    <p className="text-gray-700 dark:text-gray-300 mt-2">{item.type === 'announcement' ? item.description : item.description}</p>
                    <div className="mt-2 flex flex-wrap gap-3 text-sm text-gray-500 dark:text-gray-400">
                      <span>📅 {formatDate(item.published || item.createdAt)}</span>
                      {item.source && <span>📡 {item.source}</span>}
                      {item.link && (
                        <a href={item.link} target="_blank" rel="noopener noreferrer" className="text-blue-600 dark:text-blue-400 hover:underline">
                          Read more
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default News;