import { useParams, Link } from 'react-router-dom';

const featureData: Record<string, { title: string; icon: string; description: string; benefits: string[]; howItWorks: string[] }> = {
  'case-reporting': {
    title: 'Case Reporting',
    icon: '📝',
    description: 'Report security incidents quickly and easily to the nearest security unit',
    benefits: [
      '24/7 incident reporting',
      'Real-time case tracking',
      'Evidence upload (photos, videos, audio)',
      'Anonymous reporting option',
      'Direct communication with security units'
    ],
    howItWorks: [
      'Fill in incident details (title, description, location)',
      'Upload supporting evidence',
      'Select your preferred security unit',
      'Submit and receive case reference number',
      'Track case progress in real-time'
    ]
  },
  'sos-alerts': {
    title: 'SOS Emergency Alerts',
    icon: '🆘',
    description: 'One-touch emergency alert system for immediate response',
    benefits: [
      'Instant emergency notification',
      'Automatic location sharing',
      'Emergency contacts notification',
      'Real-time response tracking',
      'Medical information sharing'
    ],
    howItWorks: [
      'Press SOS button in emergency',
      'Location automatically captured',
      'Alert sent to nearest security units',
      'Unit dispatched to your location',
      'Track responder status in real-time'
    ]
  },
  'map-view': {
    title: 'Interactive Map',
    icon: '🗺️',
    description: 'Real-time incident visualization with unit locations',
    benefits: [
      'Live incident markers',
      'Security unit locations',
      'Crime heatmap overlay',
      'Route planning to incidents',
      'Unit coverage areas'
    ],
    howItWorks: [
      'View all incidents on interactive map',
      'See nearby security units',
      'Filter by incident type and status',
      'Click markers for detailed information',
      'Report incidents directly from map'
    ]
  },
  'suspect-tracking': {
    title: 'Suspect Tracking',
    icon: '🕵️',
    description: 'Advanced suspect management with multi-unit coordination',
    benefits: [
      'Centralized suspect database',
      'Multi-unit suspect transfers',
      'Sighting reporting with location',
      'Risk assessment scoring',
      'Suspect association mapping'
    ],
    howItWorks: [
      'Admin creates suspect profile',
      'Officers report sightings with location',
      'Suspect transfer requires admin approvals',
      'AI-powered risk assessment',
      'Link suspects to cases and other suspects'
    ]
  },
  'ai-intelligence': {
    title: 'AI Intelligence',
    icon: '🤖',
    description: 'AI-powered security insights and predictions',
    benefits: [
      'Predictive crime hotspot analysis',
      'Real-time security warnings',
      'News sentiment analysis',
      'Smart safety tips',
      'Anomaly detection'
    ],
    howItWorks: [
      'AI analyzes incident patterns',
      'Generates risk predictions',
      'Monitors news for threats',
      'Provides personalized safety tips',
      'Detects unusual patterns'
    ]
  },
  'peacebuilding': {
    title: 'Peacebuilding',
    icon: '🕊️',
    description: 'Community mediation and trust building',
    benefits: [
      'Peace committee formation',
      'Conflict resolution mediation',
      'Community trust scores',
      'Peace metrics monitoring',
      'Community feedback collection'
    ],
    howItWorks: [
      'Form community peace committees',
      'Mediate conflicts with trained mediators',
      'Track peace metrics monthly',
      'Build community trust scores',
      'Collect regular community feedback'
    ]
  },
  'communication': {
    title: 'Walkie-Talkie',
    icon: '📻',
    description: 'Real-time communication for security officers',
    benefits: [
      'Instant voice and text messaging',
      'Group chat for units',
      'Emergency priority messages',
      'Offline message sync',
      'File sharing'
    ],
    howItWorks: [
      'Join unit communication room',
      'Send text messages to group',
      'Initiate voice calls',
      'Mark messages as emergency',
      'Sync messages when offline'
    ]
  },
  'finance': {
    title: 'Financial Management',
    icon: '💰',
    description: 'Complete financial tracking for security units',
    benefits: [
      'Multi-signature transaction approvals',
      'Budget tracking',
      'Financial reports',
      'Category breakdown',
      'Budget alerts'
    ],
    howItWorks: [
      'Create transactions with type and amount',
      'Multiple admins approve based on amount',
      'Track income and expenses',
      'Set budget for different categories',
      'Generate period reports'
    ]
  }
};

export default function FeatureDetail() {
  const { featureId } = useParams();
  const feature = featureData[featureId || ''];

  if (!feature) {
    return (
      <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
        <div className="max-w-4xl mx-auto bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center">
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">Feature not found</h1>
          <Link to="/" className="text-blue-600 hover:underline mt-4 inline-block">← Back to Home</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-4xl mx-auto">
        <div className="mb-4">
          <Link to="/" className="text-gray-600 dark:text-gray-400 hover:underline">
            ← Back to Home
          </Link>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-8">
          <div className="text-center mb-8">
            <div className="text-6xl mb-4">{feature.icon}</div>
            <h1 className="text-3xl font-bold text-gray-800 dark:text-gray-200">
              {feature.title}
            </h1>
            <p className="text-lg text-gray-600 dark:text-gray-400 mt-2">
              {feature.description}
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-8">
            <div>
              <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">
                ✅ Key Benefits
              </h2>
              <ul className="space-y-2">
                {feature.benefits.map((benefit, index) => (
                  <li key={index} className="flex items-start gap-2 text-gray-700 dark:text-gray-300">
                    <span className="text-green-500 mt-1">✓</span>
                    {benefit}
                  </li>
                ))}
              </ul>
            </div>

            <div>
              <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-4">
                🔄 How It Works
              </h2>
              <ul className="space-y-2">
                {feature.howItWorks.map((step, index) => (
                  <li key={index} className="flex items-start gap-2 text-gray-700 dark:text-gray-300">
                    <span className="text-blue-500 font-bold mt-1">{index + 1}.</span>
                    {step}
                  </li>
                ))}
              </ul>
            </div>
          </div>

          <div className="mt-8 pt-8 border-t border-gray-200 dark:border-gray-700 text-center">
            <Link
              to="/register"
              className="bg-blue-600 text-white px-8 py-3 rounded-lg hover:bg-blue-700 transition inline-block"
            >
              Get Started with {feature.title}
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}