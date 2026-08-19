import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';

interface AppLayoutProps {
  children: React.ReactNode;
}

export default function AppLayout({ children }: AppLayoutProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const [user, setUser] = useState<any>(null);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      setUser(JSON.parse(userStr));
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    toast.success('Logged out successfully');
    navigate('/login');
  };

  const isActive = (path: string) => {
    return location.pathname === path ? 'bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300' : '';
  };

  // Base navigation items for all users
  const baseNavItems = [
    { path: '/home', label: '🏠 Home', icon: '🏠' },
    { path: '/map', label: '🗺️ Map', icon: '🗺️' },
    { path: '/report', label: '📝 Report', icon: '📝' },
    { path: '/my-cases', label: '📋 My Cases', icon: '📋' },
    { path: '/sos', label: '🆘 SOS', icon: '🆘' },
    { path: '/units', label: '🏢 Units', icon: '🏢' },
    { path: '/news', label: '📰 News', icon: '📰' },
    { path: '/ai', label: '🤖 AI', icon: '🤖' },
    { path: '/alerts', label: '🔔 Alerts', icon: '🔔' },
  ];

  // Build navigation based on role
  let navItems = [...baseNavItems];

  // Only show suspects if user is admin or officer (NOT citizen)
  if (user?.role === 'unit_admin' || user?.role === 'super_admin' || user?.role === 'officer') {
    navItems.push({ path: '/suspects', label: '🕵️ Suspects', icon: '🕵️' });
  }

  // Super Admin only items
  if (user?.role === 'super_admin') {
    navItems.push({ path: '/super-admin', label: '🛡️ Admin', icon: '🛡️' });
  }

  // Admin only items
  if (user?.role === 'unit_admin' || user?.role === 'super_admin') {
    navItems.push({ path: '/admin', label: '⚙️ Manage', icon: '⚙️' });
  }

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900">
      {/* Top Navigation Bar */}
      <nav className="bg-white dark:bg-gray-800 shadow-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              {/* Logo - Goes to Landing Page */}
              <Link to="/" className="flex items-center gap-2">
                <span className="text-2xl">🛡️</span>
                <span className="text-xl font-bold text-blue-600 dark:text-blue-400">
                  CommunityShield
                </span>
              </Link>
            </div>

            {/* Desktop Navigation */}
            <div className="hidden md:flex items-center gap-1 overflow-x-auto">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`px-3 py-2 rounded-lg text-sm font-medium transition whitespace-nowrap ${isActive(item.path)} hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300`}
                >
                  {item.label}
                </Link>
              ))}
              <button
                onClick={handleLogout}
                className="px-3 py-2 rounded-lg text-sm font-medium text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 transition whitespace-nowrap"
              >
                🚪 Logout
              </button>
            </div>

            {/* Mobile menu button */}
            <button
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              className="md:hidden p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              <svg className="w-6 h-6 text-gray-700 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
          </div>
        </div>

        {/* Mobile Navigation */}
        {isMobileMenuOpen && (
          <div className="md:hidden bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
            <div className="px-4 py-2 space-y-1">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className={`block px-3 py-2 rounded-lg text-sm font-medium transition ${isActive(item.path)} hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300`}
                >
                  {item.label}
                </Link>
              ))}
              <button
                onClick={handleLogout}
                className="w-full text-left px-3 py-2 rounded-lg text-sm font-medium text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 transition"
              >
                🚪 Logout
              </button>
            </div>
          </div>
        )}
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Back Button */}
        {location.pathname !== '/home' && location.pathname !== '/' && (
          <button
            onClick={() => navigate(-1)}
            className="mb-4 inline-flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 transition"
          >
            ← Back
          </button>
        )}
        {children}
      </main>
    </div>
  );
}