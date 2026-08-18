import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';
import api from '../../services/api';

export default function VerifyOTP() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);
  const [otp, setOtp] = useState('');
  const [userId, setUserId] = useState('');

  useEffect(() => {
    const storedUserId = localStorage.getItem('userId');
    if (!storedUserId) {
      navigate('/register');
      return;
    }
    setUserId(storedUserId);
  }, [navigate]);

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    if (otp.length !== 6) {
      toast.error('Please enter a valid 6-digit OTP');
      return;
    }

    setLoading(true);
    try {
      await api.post('/api/v1/otp/verify', {
        userId: userId,
        code: otp,
      });

      toast.success('OTP verified successfully!');
      navigate('/login');
    } catch (error: any) {
      const message = error.response?.data?.error || 'OTP verification failed';
      toast.error(message);
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async () => {
    setResending(true);
    try {
      await api.post('/api/v1/otp/resend', {
        userId: userId,
        type: 'email',
      });
      toast.success('OTP resent successfully!');
    } catch (error: any) {
      const message = error.response?.data?.error || 'Failed to resend OTP';
      toast.error(message);
    } finally {
      setResending(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Verify your email
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            We've sent a 6-digit code to your email. Please enter it below.
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleVerify}>
          <div>
            <input
              type="text"
              required
              maxLength={6}
              className="appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-primary-500 focus:border-primary-500 focus:z-10 sm:text-sm text-center text-2xl tracking-widest"
              placeholder="Enter 6-digit code"
              value={otp}
              onChange={(e) => setOtp(e.target.value.replace(/\D/g, ''))}
            />
          </div>

          <div>
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-primary-600 hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50"
            >
              {loading ? 'Verifying...' : 'Verify OTP'}
            </button>
          </div>

          <div className="text-sm text-center">
            <button
              type="button"
              onClick={handleResend}
              disabled={resending}
              className="font-medium text-primary-600 hover:text-primary-500 disabled:opacity-50"
            >
              {resending ? 'Sending...' : 'Resend OTP'}
            </button>
          </div>

          <div className="text-sm text-center">
            <a href="/login" className="font-medium text-gray-600 hover:text-gray-500">
              Back to login
            </a>
          </div>
        </form>
      </div>
    </div>
  );
}