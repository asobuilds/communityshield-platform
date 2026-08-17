import { useState } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function ForgotPassword() {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [resetLink, setResetLink] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const response = await axios.post('/api/v1/auth/forgot-password', { email });
      setResetLink(response.data.resetLink || '');
      setSubmitted(true);
      toast.success('If your email is registered, you will receive a reset link.');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };

  const shareOnWhatsApp = () => {
    if (!resetLink) {
      toast.error('No reset link available. Please request again.');
      return;
    }
    const message = `🔐 Reset your CommunityShield password using this link: ${resetLink}`;
    const url = `https://wa.me/?text=${encodeURIComponent(message)}`;
    window.open(url, '_blank');
  };

  const copyLink = () => {
    if (!resetLink) {
      toast.error('No reset link available. Please request again.');
      return;
    }
    navigator.clipboard.writeText(resetLink).then(() => {
      toast.success('Reset link copied to clipboard!');
    }).catch(() => {
      toast.error('Failed to copy link');
    });
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100 dark:bg-gray-900 p-4">
      <div className="bg-white dark:bg-gray-800 p-8 rounded-lg shadow-md w-full max-w-md">
        <div className="flex items-center mb-6">
          <Link to="/login" className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 mr-4">← Back</Link>
          <h2 className="text-2xl font-bold text-center flex-1 text-gray-800 dark:text-gray-100">Reset Password</h2>
        </div>

        {submitted ? (
          <div className="text-center py-4">
            <p className="text-green-600 dark:text-green-400 mb-4">✅ If your email is registered, you will receive a password reset link.</p>
            <div className="flex flex-col gap-2">
              <button
                onClick={copyLink}
                className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
              >
                📋 Copy Reset Link
              </button>
              <button
                onClick={shareOnWhatsApp}
                className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
              >
                💬 Share on WhatsApp
              </button>
              <Link to="/login" className="text-blue-600 dark:text-blue-400 hover:underline mt-2">← Back to Login</Link>
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit}>
            <p className="text-gray-600 dark:text-gray-300 mb-4">Enter your email address and we'll send you a link to reset your password.</p>
            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                required
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Sending...' : 'Send Reset Link'}
            </button>
            <p className="mt-4 text-center text-sm">
              <Link to="/login" className="text-blue-600 dark:text-blue-400">← Back to Login</Link>
            </p>
          </form>
        )}
      </div>
    </div>
  );
}

export default ForgotPassword;