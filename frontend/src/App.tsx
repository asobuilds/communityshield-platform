import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Toaster } from 'react-hot-toast';
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

// Dashboard & Features
import Home from './pages/home/Home';
import Profile from './pages/profile/Profile';
import MapPage from './pages/map/MapPage';
import ReportCase from './pages/cases/ReportCase';
import MyCases from './pages/cases/MyCases';
import CaseDetail from './pages/cases/CaseDetail';
import SOSPage from './pages/sos/SOSPage';
import SOSHistory from './pages/sos/SOSHistory';
import News from './pages/news/News';
import Alerts from './pages/alerts/Alerts';

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

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const user = localStorage.getItem('user');
  if (!user) return <Navigate to="/login" />;
  return <>{children}</>;
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
  return <>{children}</>;
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
  return <>{children}</>;
}

function App() {
  return (
    <Router>
      <Toaster position="top-right" />
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
        <Route path="/map" element={<ProtectedRoute><MapPage /></ProtectedRoute>} />
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

        {/* FALLBACK */}
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </Router>
  );
}

export default App;