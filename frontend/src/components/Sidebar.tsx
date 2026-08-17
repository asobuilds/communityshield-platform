import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useTheme } from '../context/ThemeContext';

function Sidebar() {
  const navigate = useNavigate();
  const location = useLocation();
  const { theme, toggleTheme, fontSize, setFontSize } = useTheme();

  const user = JSON.parse(localStorage.getItem('user') || '{}');
  const isAdmin = user.role === 'unit_admin';

  const handleLogout = () => {
    localStorage.removeItem('user');
    navigate('/login');
  };

  const navItems = [
    { path: '/dashboard', label: 'Dashboard', icon: '📊' },
    { path: '/units', label: 'Find Units', icon: '🏢' }, // NEW
    { path: '/alerts', label: 'Alerts', icon: '📢' },
    { path: '/news', label: 'News', icon: '📰' },
    { path: '/report', label: 'Report Case', icon: '📝' },
    { path: '/my-cases', label: 'My Cases', icon: '📋' },
    { path: '/sos', label: 'SOS', icon: '🚨' },
    { path: '/sos-history', label: 'SOS History', icon: '📜' },
    { path: '/finance', label: 'Finance', icon: '💰' },
    { path: '/profile', label: 'Profile', icon: '👤' },
    { path: '/leaderboard', label: 'Leaderboard', icon: '🏆' },
  ];

  if (isAdmin) {
    navItems.push({ path: '/admin', label: 'Admin', icon: '⚙️' });
    navItems.push({ path: '/admin/units', label: 'Units', icon: '🏢' });
  }

  return (
    <aside className="w-64 bg-white shadow-md dark:bg-gray-800 dark:text-white h-screen sticky top-0 flex flex-col">
      <div className="p-4 border-b dark:border-gray-700">
        <h1 className="text-xl font-bold text-blue-600 dark:text-blue-400">CommunityShield</h1>
      </div>
      <nav className="flex-1 p-4 space-y-1">
        {navItems.map((item) => (
          <Link
            key={item.path}
            to={item.path}
            className={`flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-blue-50 dark:hover:bg-gray-700 transition ${
              location.pathname === item.path ? 'bg-blue-50 dark:bg-gray-700 text-blue-600 dark:text-blue-400' : 'text-gray-700 dark:text-gray-300'
            }`}
          >
            <span>{item.icon}</span>
            <span>{item.label}</span>
          </Link>
        ))}
      </nav>
      <div className="p-4 border-t dark:border-gray-700 space-y-3">
        <button
          onClick={toggleTheme}
          className="flex items-center gap-3 w-full px-4 py-2 rounded-lg hover:bg-blue-50 dark:hover:bg-gray-700 transition text-gray-700 dark:text-gray-300"
        >
          <span>{theme === 'light' ? '🌙' : '☀️'}</span>
          <span>{theme === 'light' ? 'Dark Mode' : 'Light Mode'}</span>
        </button>
        <div className="flex items-center gap-3 px-4 py-2">
          <span className="text-gray-700 dark:text-gray-300">🔤 Font:</span>
          <button
            onClick={() => setFontSize('small')}
            className={`px-2 py-1 rounded ${fontSize === 'small' ? 'bg-blue-500 text-white' : 'bg-gray-200 dark:bg-gray-700'}`}
          >
            A
          </button>
          <button
            onClick={() => setFontSize('medium')}
            className={`px-2 py-1 rounded ${fontSize === 'medium' ? 'bg-blue-500 text-white' : 'bg-gray-200 dark:bg-gray-700'}`}
          >
            A
          </button>
          <button
            onClick={() => setFontSize('large')}
            className={`px-2 py-1 rounded text-lg ${fontSize === 'large' ? 'bg-blue-500 text-white' : 'bg-gray-200 dark:bg-gray-700'}`}
          >
            A
          </button>
        </div>
        <button
          onClick={handleLogout}
          className="flex items-center gap-3 w-full px-4 py-2 rounded-lg hover:bg-red-50 dark:hover:bg-red-900 transition text-red-600 dark:text-red-400"
        >
          <span>🚪</span>
          <span>Logout</span>
        </button>
      </div>
    </aside>
  );
}

export default Sidebar;