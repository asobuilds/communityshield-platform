import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';

function Home() {
  return (
    <div style={{ padding: '20px', fontFamily: 'Arial' }}>
      <h1>🛡️ CommunityShield</h1>
      <p>Welcome to CommunityShield - Your Community Security Platform</p>
      <p style={{ color: '#666' }}>Connecting communities with security units in Nigeria.</p>
      <div style={{ marginTop: '20px' }}>
        <Link to="/about" style={{ marginRight: '10px', color: '#2563eb' }}>About</Link>
        <Link to="/dashboard" style={{ color: '#2563eb' }}>Dashboard</Link>
      </div>
    </div>
  );
}

function About() {
  return (
    <div style={{ padding: '20px', fontFamily: 'Arial' }}>
      <h1>About CommunityShield</h1>
      <p>CommunityShield is a digital security platform for Nigerian communities.</p>
      <p>We help citizens report incidents, track cases, and connect with local security units.</p>
      <Link to="/" style={{ color: '#2563eb' }}>← Back to Home</Link>
    </div>
  );
}

function Dashboard() {
  return (
    <div style={{ padding: '20px', fontFamily: 'Arial' }}>
      <h1>📊 Dashboard</h1>
      <p>Your security dashboard will appear here.</p>
      <p>Stay tuned for updates!</p>
      <Link to="/" style={{ color: '#2563eb' }}>← Back to Home</Link>
    </div>
  );
}

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/about" element={<About />} />
        <Route path="/dashboard" element={<Dashboard />} />
      </Routes>
    </Router>
  );
}

export default App;