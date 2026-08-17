import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { useEffect } from 'react';
import { ThemeProvider } from './context/ThemeContext';
import './index.css';

// Auth Pages
import Login from './pages/auth/Login';
import Register from './pages/auth/Register';
import ForgotPassword from './pages/auth/ForgotPassword';
import ResetPassword from './pages/auth/ResetPassword';
import RegisterUnit from './pages/auth/RegisterUnit';

// Landing Page (public)
import LandingPage from './pages/public/LandingPage';

// Dashboard
import Dashboard from './pages/dashboard/Dashboard';
import Profile from './pages/profile/Profile';
import ReportCase from './pages/cases/ReportCase';
import MyCases from './pages/cases/MyCases';
import CaseDetail from './pages/cases/CaseDetail';
import SOSPage from './pages/sos/SOSPage';
import SOSHistory from './pages/sos/SOSHistory';
import AdminDashboard from './pages/admin/AdminDashboard';
import Units from './pages/admin/Units';
import OfficerDashboard from './pages/officer/OfficerDashboard';
import FinanceDashboard from './pages/finance/FinanceDashboard';
import AddTransaction from './pages/finance/AddTransaction';
import Ledger from './pages/finance/Ledger';
import AISummary from './pages/ai/AISummary';
import AIMonitor from './pages/ai/AIMonitor';
import News from './pages/news/News';
import Alerts from './pages/alerts/Alerts';
import FindUnits from './pages/units/FindUnits';
import Analytics from './pages/admin/Analytics';
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
  if (userRole !== 'unit_admin') return <Navigate to="/dashboard" />;
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
  if (userRole !== 'officer') return <Navigate to="/dashboard" />;
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
          <Route path="/" element={<LandingPage />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/forgot-password" element={<ForgotPassword />} />
          <Route path="/reset-password" element={<ResetPassword />} />
          <Route path="/alerts" element={<Alerts />} />
          <Route path="/register-unit" element={<RegisterUnit />} />

          <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
          <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
          <Route path="/report" element={<ProtectedRoute><ReportCase /></ProtectedRoute>} />
          <Route path="/my-cases" element={<ProtectedRoute><MyCases /></ProtectedRoute>} />
          <Route path="/case/:id" element={<ProtectedRoute><CaseDetail /></ProtectedRoute>} />
          <Route path="/sos" element={<ProtectedRoute><SOSPage /></ProtectedRoute>} />
          <Route path="/sos-history" element={<ProtectedRoute><SOSHistory /></ProtectedRoute>} />
          <Route path="/finance" element={<ProtectedRoute><FinanceDashboard /></ProtectedRoute>} />
          <Route path="/add-transaction" element={<ProtectedRoute><AddTransaction /></ProtectedRoute>} />
          <Route path="/ledger" element={<ProtectedRoute><Ledger /></ProtectedRoute>} />
          <Route path="/ai" element={<ProtectedRoute><AISummary /></ProtectedRoute>} />
          <Route path="/ai/monitor" element={<ProtectedRoute><AIMonitor /></ProtectedRoute>} />
          <Route path="/news" element={<ProtectedRoute><News /></ProtectedRoute>} />
          <Route path="/units" element={<ProtectedRoute><FindUnits /></ProtectedRoute>} />
          <Route path="/leaderboard" element={<ProtectedRoute><Leaderboard /></ProtectedRoute>} />

          <Route path="/admin" element={<AdminRoute><AdminDashboard /></AdminRoute>} />
          <Route path="/admin/units" element={<AdminRoute><Units /></AdminRoute>} />
          <Route path="/admin/analytics" element={<AdminRoute><Analytics /></AdminRoute>} />

          <Route path="/officer" element={<OfficerRoute><OfficerDashboard /></OfficerRoute>} />
        </Routes>
      </Router>
    </ThemeProvider>
  );
}

export default App;