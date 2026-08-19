import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import WalkieTalkie from '../../components/communication/WalkieTalkie';

export default function WalkieTalkiePage() {
  const [user, setUser] = useState<any>(null);
  const [userUnitId, setUserUnitId] = useState<string>('');

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    if (userStr) {
      const userData = JSON.parse(userStr);
      setUser(userData);
      setUserUnitId(userData.unitId || '');
    }
  }, []);

  if (!user) {
    return (
      <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
        <div className="text-center py-8">
          <p className="text-gray-500">Please login to use Walkie-Talkie</p>
          <Link to="/login" className="text-blue-600 hover:underline">Login</Link>
        </div>
      </div>
    );
  }

  if (!userUnitId) {
    return (
      <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
        <div className="max-w-2xl mx-auto bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 text-center">
          <div className="text-4xl mb-4">📻</div>
          <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200 mb-2">
            Join a Unit to Use Walkie-Talkie
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mb-4">
            You need to be part of a security unit to use the walkie-talkie feature.
          </p>
          <Link
            to="/units"
            className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition"
          >
            Find Units
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-2xl mx-auto">
        <div className="mb-4">
          <Link to="/home" className="text-gray-600 dark:text-gray-400 hover:underline">
            ← Back to Home
          </Link>
        </div>
        <WalkieTalkie
          unitId={userUnitId}
          userId={user.id}
          username={`${user.firstName} ${user.lastName}`}
        />
      </div>
    </div>
  );
}