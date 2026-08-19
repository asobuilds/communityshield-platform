import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { useEffect } from 'react';
import { ThemeProvider } from './context/ThemeContext';
import './index.css';

// Auth Pages
import Login from './pages/auth/Login';
import Register from './pages/auth/Register';
import VerifyOTP from './pages/auth/VerifyOTP';
import ForgotPassword from './pages/auth/ForgotPassword';
import ResetPassword from './pages/auth/ResetPassword';
import RegisterUnit from './pages/auth/RegisterUnit';

// Landing Page (public)
import LandingPage from './pages/public/LandingPage';

// Feature Detail
import FeatureDetail from './pages/public/features/FeatureDetail';

// Home (was Dashboard)
import Home from './pages/home/Home';
import Profile from './pages/profile/Profile';
import ReportCase from './pages/cases/ReportCase';
import MyCases from './pages/cases/MyCases';
import CaseDetail from './pages/cases/CaseDetail';

// Map
import MapPage from './pages/map/MapPage';

// SOS
import SOSPage from './pages/sos/SOSPage';
import SOSHistory from './pages/sos/SOSHistory';

// Admin
import SuperAdmin from './pages/admin/SuperAdmin';
import AdminDashboard from './pages/admin/AdminDashboard';
import Units from './pages/admin/Units';
import Analytics from './pages/admin/Analytics';

// Officer
import OfficerDashboard from './pages/officer/OfficerDashboard';

// Finance
import FinanceDashboard from './pages/finance/FinanceDashboard';
import AddTransaction from './pages/finance/AddTransaction';
import Ledger from './pages/finance/Ledger';

// AI
import AISummary from './pages/ai/AISummary';
import AIMonitor from './pages/ai/AIMonitor';

// News
import News from './pages/news/News';

// Alerts
import Alerts from './pages/alerts/Alerts';

// Suspects
import SuspectsList from './pages/suspects/SuspectsList';
import SuspectDetail from './pages/suspects/SuspectDetail';
import CreateSuspect from './pages/suspects/CreateSuspect';

// Transfers
import TransfersList from './pages/transfers/TransfersList';

// Walkie-Talkie
import WalkieTalkiePage from './pages/walkie-talkie/WalkieTalkiePage';

// Other
import FindUnits from './pages/units/FindUnits';
import Leaderboard from './pages/leaderboard/Leaderboard';

import AppLayout from './components/AppLayout';
import SafetyExit from './components/SafetyExit';

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const user = localStorage.getItem('user');
  if (!user) return <Navigate to="/login" />;
  return <AppLayout>{children}</AppLayout>;
}

function AdminRoute({ children }: { children: React.ReactNode }) {
  const user = localStorage.getItem('user');
  if (!user) return <Navigate to="/login" />;
  let userRole = 'citizen';
  try {
    const userData = JSON.parse(user);
    userRole = userData?.role || 'citizen';
  } catch (e) {}
  if (userRole !== 'unit_admin') return <Navigate to="/home" />;
  return <AppLayout>{children}</AppLayout>;
}

function OfficerRoute({ children }: { children: React.ReactNode }) {
  const user = localStorage.getItem('user');
  if (!user) return <Navigate to="/login" />;
  let userRole = 'citizen';
  try {
    const userData = JSON.parse(user);
    userRole = userData?.role || 'citizen';
  } catch (e) {}
  if (userRole !== 'officer') return <Navigate to="/home" />;
  return <AppLayout>{children}</AppLayout>;
}

function App() {
  useEffect(() => {
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission().then(permission => {
        console.log('Notification permission:', permission);
      });
    }
  }, []);

  return (
    <ThemeProvider>
      <Router>
        <Toaster position="top-right" />
        <SafetyExit />
        <Routes>
          {/* PUBLIC ROUTES */}
          <Route path="/" element={<LandingPage />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/verify-otp" element={<VerifyOTP />} />
          <Route path="/forgot-password" element={<ForgotPassword />} />
          <Route path="/reset-password" element={<ResetPassword />} />
          <Route path="/alerts" element={<Alerts />} />
          <Route path="/register-unit" element={<RegisterUnit />} />
          <Route path="/features/:featureId" element={<FeatureDetail />} />

          {/* PROTECTED ROUTES */}
          <Route path="/home" element={<ProtectedRoute><Home /></ProtectedRoute>} />
          <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
          <Route path="/report" element={<ProtectedRoute><ReportCase /></ProtectedRoute>} />
          <Route path="/my-cases" element={<ProtectedRoute><MyCases /></ProtectedRoute>} />
          <Route path="/case/:id" element={<ProtectedRoute><CaseDetail /></ProtectedRoute>} />
          
          {/* Map */}
          <Route path="/map" element={<ProtectedRoute><MapPage /></ProtectedRoute>} />
          
          {/* SOS */}
          <Route path="/sos" element={<ProtectedRoute><SOSPage /></ProtectedRoute>} />
          <Route path="/sos-history" element={<ProtectedRoute><SOSHistory /></ProtectedRoute>} />
          
          {/* Finance */}
          <Route path="/finance" element={<ProtectedRoute><FinanceDashboard /></ProtectedRoute>} />
          <Route path="/add-transaction" element={<ProtectedRoute><AddTransaction /></ProtectedRoute>} />
          <Route path="/ledger" element={<ProtectedRoute><Ledger /></ProtectedRoute>} />
          
          {/* AI */}
          <Route path="/ai" element={<ProtectedRoute><AISummary /></ProtectedRoute>} />
          <Route path="/ai/monitor" element={<ProtectedRoute><AIMonitor /></ProtectedRoute>} />
          
          {/* News */}
          <Route path="/news" element={<ProtectedRoute><News /></ProtectedRoute>} />
          
          {/* Alerts */}
          <Route path="/alerts" element={<ProtectedRoute><Alerts /></ProtectedRoute>} />
          
          {/* Suspects */}
          <Route path="/suspects" element={<ProtectedRoute><SuspectsList /></ProtectedRoute>} />
          <Route path="/suspects/:id" element={<ProtectedRoute><SuspectDetail /></ProtectedRoute>} />
          <Route path="/suspects/create" element={<ProtectedRoute><CreateSuspect /></ProtectedRoute>} />
          
          {/* Transfers */}
          <Route path="/transfers" element={<ProtectedRoute><TransfersList /></ProtectedRoute>} />
          
          {/* Walkie-Talkie */}
          <Route path="/walkie-talkie" element={<ProtectedRoute><WalkieTalkiePage /></ProtectedRoute>} />
          
          {/* Super Admin */}
          <Route path="/super-admin" element={<ProtectedRoute><SuperAdmin /></ProtectedRoute>} />
          
          {/* Other */}
          <Route path="/units" element={<ProtectedRoute><FindUnits /></ProtectedRoute>} />
          <Route path="/leaderboard" element={<ProtectedRoute><Leaderboard /></ProtectedRoute>} />

          {/* Admin */}
          <Route path="/admin" element={<AdminRoute><AdminDashboard /></AdminRoute>} />
          <Route path="/admin/units" element={<AdminRoute><Units /></AdminRoute>} />
          <Route path="/admin/analytics" element={<AdminRoute><Analytics /></AdminRoute>} />

          {/* Officer */}
          <Route path="/officer" element={<OfficerRoute><OfficerDashboard /></OfficerRoute>} />

          {/* Fallback */}
          <Route path="*" element={<Navigate to="/" />} />
        </Routes>
      </Router>
    </ThemeProvider>
  );
}

export default App;