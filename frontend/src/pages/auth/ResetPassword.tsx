import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function ResetPassword() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');

  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);

  useEffect(() => {
    if (!token) {
      toast.error('Missing reset token. Please request a new password reset.');
      navigate('/forgot-password');
    }
  }, [token, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword.length < 6) {
      toast.error('Password must be at least 6 characters.');
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error('Passwords do not match.');
      return;
    }

    setLoading(true);
    try {
      await axios.post('/api/v1/auth/reset-password', {
        token,
        newPassword,
      });
      toast.success('Password reset successful! Please login with your new password.');
      setSubmitted(true);
      setTimeout(() => navigate('/login'), 3000);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to reset password. The link may have expired.');
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100 p-4">
        <div className="bg-white p-8 rounded-lg shadow-md w-full max-w-md text-center">
          <p className="text-red-600 mb-4">Invalid or missing reset token.</p>
          <Link to="/forgot-password" className="text-blue-600 hover:underline">Request a new reset link</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100 p-4">
      <div className="bg-white p-8 rounded-lg shadow-md w-full max-w-md">
        <div className="flex items-center mb-6">
          <Link to="/login" className="text-blue-600 hover:text-blue-800 mr-4">← Back</Link>
          <h2 className="text-2xl font-bold text-center flex-1">Reset Password</h2>
        </div>

        {submitted ? (
          <div className="text-center py-4">
            <p className="text-green-600 mb-4">✅ Your password has been reset successfully.</p>
            <p className="text-gray-600">Redirecting to login...</p>
            <Link to="/login" className="text-blue-600 hover:underline">← Back to Login</Link>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            <p className="text-gray-600 mb-4">Enter your new password below.</p>
            <div className="mb-4">
              <label className="block text-sm font-medium mb-1">New Password</label>
              <input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="w-full px-3 py-2 border rounded-lg"
                required
                minLength={6}
              />
            </div>
            <div className="mb-6">
              <label className="block text-sm font-medium mb-1">Confirm Password</label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full px-3 py-2 border rounded-lg"
                required
                minLength={6}
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Resetting...' : 'Reset Password'}
            </button>
            <p className="mt-4 text-center text-sm">
              <Link to="/login" className="text-blue-600">← Back to Login</Link>
            </p>
          </form>
        )}
      </div>
    </div>
  );
}

export default ResetPassword;