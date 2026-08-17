import NotificationBell from '../../components/NotificationBell';
import { Card, CardContent } from '../../components/ui/Card';
import Button from '../../components/ui/Button';
import Badge from '../../components/ui/Badge';

function Dashboard() {
  const user = JSON.parse(localStorage.getItem('user') || '{}');
  const isAdmin = user.role === 'unit_admin';

  const actions = [
    { label: 'SOS', icon: '🚨', color: 'bg-red-500 hover:bg-red-600', link: '/sos' },
    { label: 'Report', icon: '📝', color: 'bg-green-500 hover:bg-green-600', link: '/report' },
    { label: 'My Cases', icon: '📋', color: 'bg-blue-500 hover:bg-blue-600', link: '/my-cases' },
    { label: 'SOS History', icon: '📜', color: 'bg-orange-500 hover:bg-orange-600', link: '/sos-history' },
    { label: 'Finance', icon: '💰', color: 'bg-purple-500 hover:bg-purple-600', link: '/finance' },
    { label: 'Find Units', icon: '🏢', color: 'bg-indigo-500 hover:bg-indigo-600', link: '/units' },
  ];

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-5xl mx-auto">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold text-gray-800 dark:text-white">Dashboard</h1>
          <NotificationBell />
        </div>

        <Card className="mb-8">
          <CardContent className="pt-6">
            <p className="text-gray-700 dark:text-gray-300">Welcome back, <strong>{user.firstName || 'User'}</strong>!</p>
            <p className="text-sm text-gray-500 dark:text-gray-400">You are logged in as <Badge variant="info">{user.role || 'citizen'}</Badge></p>
            {isAdmin && (
              <div className="mt-3 p-3 bg-blue-50 dark:bg-blue-900/30 rounded-lg border border-blue-200 dark:border-blue-800">
                <p className="text-blue-800 dark:text-blue-200">🔑 Admin access available</p>
                <Button variant="primary" size="sm" className="mt-2" onClick={() => window.location.href = '/admin'}>
                  Go to Admin Dashboard
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {actions.map((action) => (
            <a key={action.label} href={action.link} className="block">
              <Card className="hover:shadow-lg transition-shadow duration-200 border-0">
                <CardContent className="p-6 text-center">
                  <div className="text-4xl mb-2">{action.icon}</div>
                  <p className="font-medium text-gray-800 dark:text-white">{action.label}</p>
                </CardContent>
              </Card>
            </a>
          ))}
        </div>

        <div className="mt-8">
          <Button
            variant="danger"
            onClick={() => { localStorage.removeItem('user'); window.location.href = '/login'; }}
          >
            Logout
          </Button>
        </div>
      </div>
    </div>
  );
}

export default Dashboard;