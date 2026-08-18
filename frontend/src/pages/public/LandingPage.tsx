import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import axios from 'axios';
import IncidentMap from '../../components/map/IncidentMap';

interface Incident {
  id: string;
  title: string;
  description: string;
  latitude: number;
  longitude: number;
  status: string;
  type: 'case' | 'sos' | 'unit';
}

export default function LandingPage() {
  const navigate = useNavigate();
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [loading, setLoading] = useState(true);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [userLocation, setUserLocation] = useState<[number, number]>([6.5244, 3.3792]);

  useEffect(() => {
    const token = localStorage.getItem('token');
    setIsLoggedIn(!!token);
    fetchIncidents();

    // Get user location
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          setUserLocation([pos.coords.latitude, pos.coords.longitude]);
        },
        () => console.log('Using default location')
      );
    }
  }, []);

  const fetchIncidents = async () => {
    setLoading(true);
    try {
      const [casesRes, unitsRes] = await Promise.all([
        axios.get('/api/v1/cases').catch(() => ({ data: { cases: [] } })),
        axios.get('/api/v1/units').catch(() => ({ data: { units: [] } }))
      ]);

      const cases = casesRes.data.cases || [];
      const units = unitsRes.data.units || [];

      const mappedCases = cases.map((c: any) => ({
        id: c.id,
        title: c.title,
        description: c.description,
        latitude: c.latitude || 6.5244,
        longitude: c.longitude || 3.3792,
        status: c.status,
        type: 'case' as const
      }));

      const mappedUnits = units.map((u: any) => ({
        id: u.id,
        title: u.name,
        description: `Contact: ${u.contactPerson || 'N/A'}`,
        latitude: u.latitude || 6.5244,
        longitude: u.longitude || 3.3792,
        status: 'active',
        type: 'unit' as const
      }));

      setIncidents([...mappedCases, ...mappedUnits]);
    } catch (error) {
      console.error('Failed to load incidents');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Header/Navigation */}
      <header className="bg-white dark:bg-gray-800 shadow-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-2">
              <span className="text-2xl">🛡️</span>
              <h1 className="text-xl font-bold text-blue-600 dark:text-blue-400">
                CommunityShield
              </h1>
            </div>
            <nav className="flex items-center gap-4">
              {isLoggedIn ? (
                <>
                  <Link to="/home" className="text-gray-700 dark:text-gray-300 hover:text-blue-600">
                    Dashboard
                  </Link>
                  <Link to="/map" className="text-gray-700 dark:text-gray-300 hover:text-blue-600">
                    Map
                  </Link>
                  <button
                    onClick={() => {
                      localStorage.clear();
                      window.location.href = '/';
                    }}
                    className="text-red-600 hover:text-red-700"
                  >
                    Logout
                  </button>
                </>
              ) : (
                <>
                  <Link to="/login" className="text-gray-700 dark:text-gray-300 hover:text-blue-600">
                    Login
                  </Link>
                  <Link
                    to="/register"
                    className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition"
                  >
                    Get Started
                  </Link>
                </>
              )}
            </nav>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="bg-gradient-to-r from-blue-600 to-blue-800 text-white py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center max-w-3xl mx-auto">
            <h1 className="text-4xl md:text-5xl font-bold mb-4">
              Community Safety & Security Platform
            </h1>
            <p className="text-xl md:text-2xl mb-6 text-blue-100">
              Report incidents, track cases, and connect with local security units in real-time
            </p>
            <p className="text-lg text-blue-200 mb-8">
              Empowering communities in rural Nigeria with digital security solutions
            </p>
            <div className="flex flex-wrap justify-center gap-4">
              <Link
                to={isLoggedIn ? "/report" : "/register"}
                className="bg-white text-blue-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-100 transition"
              >
                {isLoggedIn ? "Report Incident" : "Get Started Now"}
              </Link>
              <Link
                to="/map"
                className="bg-transparent border-2 border-white text-white px-8 py-3 rounded-lg font-semibold hover:bg-white/10 transition"
              >
                View Map
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-16 bg-white dark:bg-gray-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-3xl font-bold text-center text-gray-800 dark:text-gray-200 mb-12">
            How CommunityShield Works
          </h2>
          <div className="grid md:grid-cols-3 gap-8">
            <div className="text-center p-6">
              <div className="text-5xl mb-4">📱</div>
              <h3 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-2">1. Report</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Report incidents with location, description, and evidence. Choose to report anonymously or publicly.
              </p>
            </div>
            <div className="text-center p-6">
              <div className="text-5xl mb-4">📍</div>
              <h3 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-2">2. Track</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Track your case progress in real-time. See which security unit is handling your case and their updates.
              </p>
            </div>
            <div className="text-center p-6">
              <div className="text-5xl mb-4">🛡️</div>
              <h3 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-2">3. Resolve</h3>
              <p className="text-gray-600 dark:text-gray-400">
                Get timely response from verified security units. Receive updates and resolution notifications.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Map Section */}
      <section className="py-16 bg-gray-50 dark:bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
              📍 Live Incident Map
            </h2>
            <div className="flex gap-3">
              <button
                onClick={fetchIncidents}
                className="text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-3 py-1 rounded hover:bg-gray-300 dark:hover:bg-gray-600 transition"
              >
                🔄 Refresh
              </button>
              <Link
                to={isLoggedIn ? "/map" : "/login"}
                className="text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 transition"
              >
                View Full Map →
              </Link>
            </div>
          </div>
          {loading ? (
            <div className="flex justify-center items-center h-[400px] bg-white dark:bg-gray-800 rounded-lg">
              <p className="text-gray-500 dark:text-gray-400">Loading map...</p>
            </div>
          ) : (
            <IncidentMap
              incidents={incidents.slice(0, 15)}
              center={userLocation}
              zoom={12}
              showRadius={true}
              onMarkerClick={(id) => {
                if (isLoggedIn) {
                  navigate(`/case/${id}`);
                } else {
                  navigate('/login');
                }
              }}
            />
          )}
        </div>
      </section>

      {/* Stats Section */}
      <section className="py-16 bg-white dark:bg-gray-800">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 text-center">
            <div>
              <div className="text-4xl font-bold text-blue-600">500+</div>
              <p className="text-gray-600 dark:text-gray-400">Cases Reported</p>
            </div>
            <div>
              <div className="text-4xl font-bold text-green-600">50+</div>
              <p className="text-gray-600 dark:text-gray-400">Security Units</p>
            </div>
            <div>
              <div className="text-4xl font-bold text-yellow-600">1000+</div>
              <p className="text-gray-600 dark:text-gray-400">Citizens Served</p>
            </div>
            <div>
              <div className="text-4xl font-bold text-purple-600">95%</div>
              <p className="text-gray-600 dark:text-gray-400">Response Rate</p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-16 bg-gradient-to-r from-blue-600 to-blue-800 text-white">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold mb-4">
            Ready to Make Your Community Safer?
          </h2>
          <p className="text-xl text-blue-100 mb-8">
            Join CommunityShield today and be part of the digital security solution
          </p>
          <div className="flex flex-wrap justify-center gap-4">
            <Link
              to={isLoggedIn ? "/home" : "/register"}
              className="bg-white text-blue-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-100 transition"
            >
              {isLoggedIn ? "Go to Dashboard" : "Get Started Now"}
            </Link>
            <Link
              to="/map"
              className="bg-transparent border-2 border-white text-white px-8 py-3 rounded-lg font-semibold hover:bg-white/10 transition"
            >
              Explore Map
            </Link>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-800 text-white py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div>
              <h3 className="font-bold text-lg mb-2">CommunityShield</h3>
              <p className="text-gray-400 text-sm">
                Digital solution for local security units and vigilante organizations in Nigeria.
              </p>
            </div>
            <div>
              <h4 className="font-semibold mb-2">Quick Links</h4>
              <ul className="space-y-1 text-sm text-gray-400">
                <li><Link to="/map" className="hover:text-white">View Map</Link></li>
                <li><Link to="/login" className="hover:text-white">Login</Link></li>
                <li><Link to="/register" className="hover:text-white">Register</Link></li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold mb-2">Contact</h4>
              <p className="text-gray-400 text-sm">
                Email: support@communityshield.org<br />
                Phone: +234 800 123 4567
              </p>
            </div>
          </div>
          <div className="border-t border-gray-700 mt-8 pt-4 text-center text-gray-400 text-sm">
            © 2026 CommunityShield. All rights reserved.
          </div>
        </div>
      </footer>
    </div>
  );
}