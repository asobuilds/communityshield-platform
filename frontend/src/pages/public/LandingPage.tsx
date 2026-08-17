import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import Button from '../../components/ui/Button';
import { Card, CardHeader, CardTitle, CardDescription } from '../../components/ui/Card';
import Badge from '../../components/ui/Badge';

function LandingPage() {
  const navigate = useNavigate();
  const [challenge, setChallenge] = useState('');
  const user = localStorage.getItem('user');
  const isLoggedIn = !!user;

  const handleReportNow = () => {
    if (isLoggedIn) {
      navigate('/report');
    } else {
      if (challenge.trim()) sessionStorage.setItem('pendingReport', challenge);
      navigate('/register');
    }
  };

  const features = [
    {
      icon: '📱',
      title: 'Easy Reporting',
      desc: 'Report incidents in seconds with our simple, intuitive form. Add evidence with one tap.',
    },
    {
      icon: '📍',
      title: 'Location-Aware',
      desc: 'Automatically find the nearest security unit. No need to search – we connect you instantly.',
    },
    {
      icon: '🛡️',
      title: 'Accountability',
      desc: 'Track case progress, rate units, and ensure transparency. Build trust between citizens and security.',
    },
    {
      icon: '🚨',
      title: 'SOS Alerts',
      desc: 'Send emergency alerts with your location to the nearest unit. Help is just a tap away.',
    },
    {
      icon: '💰',
      title: 'Financial Transparency',
      desc: 'Track donations, bail, and expenses. Ensure funds are used properly for community safety.',
    },
    {
      icon: '🤝',
      title: 'Community Reviews',
      desc: 'See ratings and reviews from other citizens. Choose the most trusted units for your safety.',
    },
  ];

  return (
    <div className="min-h-screen bg-white dark:bg-gray-900">
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-blue-700 via-blue-800 to-indigo-900 text-white overflow-hidden">
        <div className="absolute inset-0 bg-[url('https://images.unsplash.com/photo-1582139329536-e7284fece509?w=1600')] bg-cover bg-center opacity-10" />
        <div className="absolute inset-0 bg-gradient-to-b from-transparent to-blue-900/50" />
        <div className="relative max-w-5xl mx-auto px-4 py-20 md:py-28 text-center">
          <Badge variant="info" className="mb-4 bg-blue-500/20 text-blue-200 border-blue-400/30">
            Nigeria's Security Platform
          </Badge>
          <h1 className="text-5xl md:text-7xl font-extrabold mb-4 tracking-tight">
            CommunityShield
          </h1>
          <p className="text-xl md:text-2xl mb-6 max-w-3xl mx-auto opacity-90">
            Your security, our priority. Report incidents, send SOS, and connect with local security units.
          </p>
          <div className="max-w-2xl mx-auto flex flex-col sm:flex-row gap-3 bg-white/10 backdrop-blur-sm rounded-xl p-2 shadow-lg border border-white/20">
            <input
              type="text"
              value={challenge}
              onChange={(e) => setChallenge(e.target.value)}
              placeholder="What is your security challenge? (e.g., I saw a suspicious person...)"
              className="flex-1 px-4 py-3 rounded-lg bg-white/90 text-gray-800 placeholder-gray-500 focus:outline-none"
            />
            <Button variant="primary" size="lg" onClick={handleReportNow} className="whitespace-nowrap">
              Report Now
            </Button>
          </div>
          <p className="text-sm mt-3 opacity-80">
            {isLoggedIn ? "You're logged in – report directly." : "Report anonymously or sign in to track your case."}
          </p>
          <div className="mt-6 flex flex-wrap justify-center gap-4">
            {!isLoggedIn && (
              <>
                <Link to="/register">
                  <Button variant="primary" size="lg">Sign Up</Button>
                </Link>
                <Link to="/login">
                  <Button variant="outline" size="lg" className="border-white text-white hover:bg-white/10">Sign In</Button>
                </Link>
              </>
            )}
            <Link to="/alerts">
              <Button variant="outline" size="lg" className="border-yellow-400 text-yellow-300 hover:bg-yellow-400/20">
                📢 View Alerts
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 px-4 max-w-6xl mx-auto">
        <div className="text-center mb-16">
          <Badge variant="info" className="mb-2">Why Choose Us</Badge>
          <h2 className="text-4xl font-bold text-gray-900 dark:text-white">Why CommunityShield?</h2>
          <p className="mt-2 text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
            Built for Nigerians, by Nigerians – to make your community safer.
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {features.map((feature, idx) => (
            <Card key={idx} className="hover:shadow-xl transition-all duration-300 hover:-translate-y-1">
              <CardHeader>
                <div className="text-4xl mb-2">{feature.icon}</div>
                <CardTitle>{feature.title}</CardTitle>
                <CardDescription>{feature.desc}</CardDescription>
              </CardHeader>
            </Card>
          ))}
        </div>
      </section>

      {/* CTA Section */}
      <section className="bg-gradient-to-r from-blue-700 to-indigo-800 text-white py-20 px-4 text-center">
        <div className="max-w-3xl mx-auto">
          <h2 className="text-4xl font-bold mb-4">Ready to make your community safer?</h2>
          <p className="text-xl mb-8 opacity-90">Join CommunityShield today and be part of the solution.</p>
          <Link to={isLoggedIn ? '/report' : '/register'}>
            <Button variant="primary" size="lg" className="bg-white text-blue-700 hover:bg-gray-100">
              {isLoggedIn ? 'Report a Case' : 'Create Account'}
            </Button>
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-8 px-4">
        <div className="max-w-6xl mx-auto text-center">
          <p>&copy; {new Date().getFullYear()} CommunityShield. All rights reserved.</p>
          <p className="text-sm opacity-75 mt-2">Empowering communities, one case at a time.</p>
        </div>
      </footer>
    </div>
  );
}

export default LandingPage;