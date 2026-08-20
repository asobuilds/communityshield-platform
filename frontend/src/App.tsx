import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
import { useEffect } from 'react';
import { ThemeProvider } from './context/ThemeContext';
import { LanguageProvider } from './context/LanguageContext';
import { OfflineProvider } from './context/OfflineContext';
import { lazyLoad } from './components/common/LazyLoad';
import { OfflineStatus } from './components/common/OfflineStatus';
import './index.css';

// Auth Pages
import Login from './pages/auth/Login';
import Register from './pages/auth/Register';
import VerifyOTP from './pages/auth/VerifyOTP';
import ForgotPassword from './pages/auth/ForgotPassword';
import ResetPassword from './pages/auth/ResetPassword';
import RegisterUnit from './pages/auth/RegisterUnit';

// Landing Page
import LandingPage from './pages/public/LandingPage';
import FeatureDetail from './pages/public/features/FeatureDetail';

// Lazy load heavy components
const Home = lazyLoad(() => import('./pages/home/Home'));
const Profile = lazyLoad(() => import('./pages/profile/Profile'));
const ReportCase = lazyLoad(() => import('./pages/cases/ReportCase'));
const MyCases = lazyLoad(() => import('./pages/cases/MyCases'));
const CaseDetail = lazyLoad(() => import('./pages/cases/CaseDetail'));
const MapPage = lazyLoad(() => import('./pages/map/MapPage'));
const SOSPage = lazyLoad(() => import('./pages/sos/SOSPage'));
const SOSHistory = lazyLoad(() => import('./pages/sos/SOSHistory'));
const SuperAdmin = lazyLoad(() => import('./pages/admin/SuperAdmin'));
const AdminDashboard = lazyLoad(() => import('./pages/admin/AdminDashboard'));
const Units = lazyLoad(() => import('./pages/admin/Units'));
const Analytics = lazyLoad(() => import('./pages/admin/Analytics'));
const OfficerDashboard = lazyLoad(() => import('./pages/officer/OfficerDashboard'));
const FinanceDashboard = lazyLoad(() => import('./pages/finance/FinanceDashboard'));
const AddTransaction = lazyLoad(() => import('./pages/finance/AddTransaction'));
const Ledger = lazyLoad(() => import('./pages/finance/Ledger'));
const AISummary = lazyLoad(() => import('./pages/ai/AISummary'));
const AIMonitor = lazyLoad(() => import('./pages/ai/AIMonitor'));
const News = lazyLoad(() => import('./pages/news/News'));
const Alerts = lazyLoad(() => import('./pages/alerts/Alerts'));
const SuspectsList = lazyLoad(() => import('./pages/suspects/SuspectsList'));
const SuspectDetail = lazyLoad(() => import('./pages/suspects/SuspectDetail'));
const CreateSuspect = lazyLoad(() => import('./pages/suspects/CreateSuspect'));
const TransfersList = lazyLoad(() => import('./pages/transfers/TransfersList'));
const WalkieTalkiePage = lazyLoad(() => import('./pages/walkie-talkie/WalkieTalkiePage'));
const FindUnits = lazyLoad(() => import('./pages/units/FindUnits'));
const Leaderboard = lazyLoad(() => import('./pages/leaderboard/Leaderboard'));

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
      <LanguageProvider>
        <OfflineProvider>
          <Router>
            <Toaster position="top-right" />
            <SafetyExit />
            <OfflineStatus />
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
              
              <Route path="/map" element={<ProtectedRoute><MapPage /></ProtectedRoute>} />
              
              <Route path="/sos" element={<ProtectedRoute><SOSPage /></ProtectedRoute>} />
              <Route path="/sos-history" element={<ProtectedRoute><SOSHistory /></ProtectedRoute>} />
              
              <Route path="/finance" element={<ProtectedRoute><FinanceDashboard /></ProtectedRoute>} />
              <Route path="/add-transaction" element={<ProtectedRoute><AddTransaction /></ProtectedRoute>} />
              <Route path="/ledger" element={<ProtectedRoute><Ledger /></ProtectedRoute>} />
              
              <Route path="/ai" element={<ProtectedRoute><AISummary /></ProtectedRoute>} />
              <Route path="/ai/monitor" element={<ProtectedRoute><AIMonitor /></ProtectedRoute>} />
              
              <Route path="/news" element={<ProtectedRoute><News /></ProtectedRoute>} />
              <Route path="/alerts" element={<ProtectedRoute><Alerts /></ProtectedRoute>} />
              
              <Route path="/suspects" element={<ProtectedRoute><SuspectsList /></ProtectedRoute>} />
              <Route path="/suspects/:id" element={<ProtectedRoute><SuspectDetail /></ProtectedRoute>} />
              <Route path="/suspects/create" element={<ProtectedRoute><CreateSuspect /></ProtectedRoute>} />
              
              <Route path="/transfers" element={<ProtectedRoute><TransfersList /></ProtectedRoute>} />
              
              <Route path="/walkie-talkie" element={<ProtectedRoute><WalkieTalkiePage /></ProtectedRoute>} />
              
              <Route path="/super-admin" element={<ProtectedRoute><SuperAdmin /></ProtectedRoute>} />
              
              <Route path="/units" element={<ProtectedRoute><FindUnits /></ProtectedRoute>} />
              <Route path="/leaderboard" element={<ProtectedRoute><Leaderboard /></ProtectedRoute>} />

              <Route path="/admin" element={<AdminRoute><AdminDashboard /></AdminRoute>} />
              <Route path="/admin/units" element={<AdminRoute><Units /></AdminRoute>} />
              <Route path="/admin/analytics" element={<AdminRoute><Analytics /></AdminRoute>} />

              <Route path="/officer" element={<OfficerRoute><OfficerDashboard /></OfficerRoute>} />

              <Route path="*" element={<Navigate to="/" />} />
            </Routes>
          </Router>
        </OfflineProvider>
      </LanguageProvider>
    </ThemeProvider>
  );
}

export default App;