import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function VerifyOTP() {
  const navigate = useNavigate();
  const location = useLocation();
  const [otp, setOtp] = useState('');
  const [loading, setLoading] = useState(false);
  const [userId, setUserId] = useState('');
  const [otpType, setOtpType] = useState('both');
  const [resendTimer, setResendTimer] = useState(0);
  const [timer, setTimer] = useState(300); // 5 minutes countdown

  useEffect(() => {
    // Get userId from location state or URL params
    const state = location.state as any;
    if (state?.userId) {
      setUserId(state.userId);
      setOtpType(state.type || 'both');
    } else {
      // Fallback: try to get from localStorage
      const user = JSON.parse(localStorage.getItem('user') || '{}');
      if (user.id) {
        setUserId(user.id);
      } else {
        toast.error('User not found. Please register first.');
        navigate('/register');
      }
    }
  }, [location]);

  useEffect(() => {
    if (timer > 0) {
      const interval = setInterval(() => {
        setTimer(prev => prev - 1);
      }, 1000);
      return () => clearInterval(interval);
    }
  }, [timer]);

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!otp || otp.length !== 6) {
      toast.error('Please enter a valid 6-digit OTP');
      return;
    }
    setLoading(true);
    try {
      const response = await axios.post('/api/v1/otp/verify', {
        userId,
        code: otp,
      });
      toast.success('OTP verified successfully!');
      // Update user in localStorage with verification status
      const user = JSON.parse(localStorage.getItem('user') || '{}');
      user.emailVerified = response.data.user.emailVerified;
      user.phoneVerified = response.data.user.phoneVerified;
      localStorage.setItem('user', JSON.stringify(user));
      navigate('/dashboard');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Invalid OTP');
    } finally {
      setLoading(false);
    }
  };

  const handleResendOTP = async () => {
    if (resendTimer > 0) return;
    try {
      await axios.post('/api/v1/otp/resend', {
        userId,
        type: otpType,
      });
      toast.success('OTP resent!');
      setResendTimer(60);
      setTimer(300);
      const interval = setInterval(() => {
        setResendTimer(prev => {
          if (prev <= 1) {
            clearInterval(interval);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to resend OTP');
    }
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900 p-4">
      <div className="bg-white dark:bg-gray-800 p-8 rounded-lg shadow-md w-full max-w-md">
        <div className="flex items-center mb-6">
          <button onClick={() => navigate('/')} className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 mr-4">← Back</button>
          <h2 className="text-2xl font-bold text-center flex-1 text-gray-800 dark:text-gray-100">Verify Your Account</h2>
        </div>
        <p className="text-gray-600 dark:text-gray-300 mb-4">
          We've sent a 6-digit OTP to your {otpType === 'email' ? 'email' : otpType === 'phone' ? 'phone' : 'email and phone'}.
          Please enter it below to verify your account.
        </p>
        <form onSubmit={handleVerify}>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Enter OTP</label>
            <input
              type="text"
              value={otp}
              onChange={(e) => setOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white text-center text-2xl tracking-widest"
              placeholder="000000"
              maxLength={6}
              required
            />
          </div>
          <div className="flex justify-between items-center text-sm mb-4">
            <span className="text-gray-500 dark:text-gray-400">Time remaining: {formatTime(timer)}</span>
            <button
              type="button"
              onClick={handleResendOTP}
              disabled={resendTimer > 0}
              className={`text-blue-600 dark:text-blue-400 hover:underline ${resendTimer > 0 ? 'opacity-50 cursor-not-allowed' : ''}`}
            >
              {resendTimer > 0 ? `Resend in ${resendTimer}s` : 'Resend OTP'}
            </button>
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? 'Verifying...' : 'Verify OTP'}
          </button>
        </form>
        <div className="mt-4 text-center">
          <a href="/login" className="text-sm text-blue-600 dark:text-blue-400 hover:underline">Already verified? Login</a>
        </div>
      </div>
    </div>
  );
}

export default VerifyOTP;