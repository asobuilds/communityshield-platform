import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function Profile() {
  const navigate = useNavigate();
  const [user, setUser] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [editing, setEditing] = useState(false);
  const [receiveEmail, setReceiveEmail] = useState(true);

  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    phone: '',
  });

  useEffect(() => {
    fetchProfile();
  }, []);

  const fetchProfile = async () => {
    try {
      const userStr = localStorage.getItem('user');
      if (!userStr) {
        toast.error('Please login');
        navigate('/login');
        return;
      }
      const userData = JSON.parse(userStr);
      const response = await axios.get(`/api/v1/users/${userData.id}`);
      setUser(response.data.user);
      setReceiveEmail(response.data.user.receiveEmail !== undefined ? response.data.user.receiveEmail : true);
      setFormData({
        firstName: response.data.user.firstName || '',
        lastName: response.data.user.lastName || '',
        phone: response.data.user.phone || '',
      });
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to load profile');
    } finally {
      setLoading(false);
    }
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length === 0) return;
    const file = e.target.files[0];
    setUploading(true);
    const formData = new FormData();
    formData.append('profilePicture', file);
    formData.append('userId', user.id);
    try {
      const response = await axios.put('/api/v1/users/profile-picture', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      setUser({ ...user, profileImage: response.data.profileImage });
      // Update localStorage
      const userStr = localStorage.getItem('user');
      if (userStr) {
        const storedUser = JSON.parse(userStr);
        storedUser.profileImage = response.data.profileImage;
        localStorage.setItem('user', JSON.stringify(storedUser));
      }
      toast.success('Profile picture updated!');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Upload failed');
    } finally {
      setUploading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSaveProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await axios.put('/api/v1/users/profile', {
        userId: user.id,
        firstName: formData.firstName,
        lastName: formData.lastName,
        phone: formData.phone,
      });
      toast.success('Profile updated!');
      setEditing(false);
      fetchProfile(); // refresh
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update');
    } finally {
      setLoading(false);
    }
  };

  const handleEmailToggle = async () => {
    try {
      const newValue = !receiveEmail;
      await axios.put('/api/v1/users/profile', {
        userId: user.id,
        receiveEmail: newValue,
      });
      setReceiveEmail(newValue);
      toast.success(`Email notifications ${newValue ? 'enabled' : 'disabled'}`);
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update preference');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <p className="text-gray-500">Loading profile...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <p className="text-red-500">User not found</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-2xl mx-auto bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400">My Profile</h1>
          <button
            onClick={() => navigate(-1)}
            className="text-blue-600 dark:text-blue-400 hover:underline"
          >
            ← Back
          </button>
        </div>

        {/* Profile Picture */}
        <div className="flex flex-col items-center mb-6">
          <div className="relative">
            {user.profileImage ? (
              <img
                src={`http://localhost:8080${user.profileImage}`}
                alt="Profile"
                className="w-32 h-32 rounded-full object-cover border-4 border-blue-500"
              />
            ) : (
              <div className="w-32 h-32 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center text-4xl text-gray-500 dark:text-gray-400 border-4 border-blue-500">
                {user.firstName?.[0]}{user.lastName?.[0]}
              </div>
            )}
            <label
              htmlFor="profilePicture"
              className="absolute bottom-0 right-0 bg-blue-600 text-white p-2 rounded-full cursor-pointer hover:bg-blue-700"
            >
              📷
            </label>
            <input
              id="profilePicture"
              type="file"
              accept="image/*"
              onChange={handleFileChange}
              className="hidden"
              disabled={uploading}
            />
          </div>
          {uploading && <p className="text-sm text-gray-500 dark:text-gray-400 mt-2">Uploading...</p>}
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Click the camera icon to change picture</p>
        </div>

        {/* Profile Info */}
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Email</label>
            <p className="text-gray-800 dark:text-gray-200 bg-gray-50 dark:bg-gray-700 p-2 rounded">{user.email}</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Role</label>
            <p className="text-gray-800 dark:text-gray-200 bg-gray-50 dark:bg-gray-700 p-2 rounded capitalize">{user.role}</p>
          </div>

          {editing ? (
            <form onSubmit={handleSaveProfile} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">First Name</label>
                <input
                  type="text"
                  name="firstName"
                  value={formData.firstName}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border rounded dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Last Name</label>
                <input
                  type="text"
                  name="lastName"
                  value={formData.lastName}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border rounded dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Phone</label>
                <input
                  type="text"
                  name="phone"
                  value={formData.phone}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border rounded dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="submit"
                  disabled={loading}
                  className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 disabled:opacity-50"
                >
                  {loading ? 'Saving...' : 'Save Changes'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setEditing(false);
                    setFormData({
                      firstName: user.firstName || '',
                      lastName: user.lastName || '',
                      phone: user.phone || '',
                    });
                  }}
                  className="bg-gray-300 dark:bg-gray-600 text-gray-800 dark:text-gray-200 px-4 py-2 rounded hover:bg-gray-400 dark:hover:bg-gray-500"
                >
                  Cancel
                </button>
              </div>
            </form>
          ) : (
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">First Name</label>
                <p className="text-gray-800 dark:text-gray-200 bg-gray-50 dark:bg-gray-700 p-2 rounded">{user.firstName}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Last Name</label>
                <p className="text-gray-800 dark:text-gray-200 bg-gray-50 dark:bg-gray-700 p-2 rounded">{user.lastName}</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 dark:text-gray-300">Phone</label>
                <p className="text-gray-800 dark:text-gray-200 bg-gray-50 dark:bg-gray-700 p-2 rounded">{user.phone || 'Not provided'}</p>
              </div>
              <button
                onClick={() => setEditing(true)}
                className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
              >
                Edit Profile
              </button>
            </div>
          )}

          {/* Email Subscription Toggle */}
          <div className="mt-4 pt-4 border-t dark:border-gray-700">
            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                id="receiveEmail"
                checked={receiveEmail}
                onChange={handleEmailToggle}
                className="w-5 h-5 accent-blue-600"
              />
              <label htmlFor="receiveEmail" className="text-gray-700 dark:text-gray-300">
                Receive email notifications for case updates and alerts
              </label>
            </div>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              {receiveEmail ? 'You will receive email notifications.' : 'Email notifications are disabled.'}
            </p>
          </div>
        </div>

        <div className="mt-6 pt-4 border-t dark:border-gray-700">
          <a href="/dashboard" className="text-blue-600 dark:text-blue-400 hover:underline">← Back to Dashboard</a>
        </div>
      </div>
    </div>
  );
}

export default Profile;