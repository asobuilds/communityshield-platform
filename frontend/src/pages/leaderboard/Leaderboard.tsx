import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function Leaderboard() {
  const navigate = useNavigate();
  const [units, setUnits] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchLeaderboard();
  }, []);

  const fetchLeaderboard = async () => {
    try {
      const response = await axios.get('/api/v1/leaderboard/units');
      setUnits(response.data.leaderboard || []);
    } catch (error) {
      toast.error('Failed to fetch leaderboard');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-4xl mx-auto">
        <button
          onClick={() => navigate('/dashboard')}
          className="text-blue-600 dark:text-blue-400 hover:underline mb-4 flex items-center gap-1"
        >
          ← Back to Dashboard
        </button>

        <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400 mb-6">🏆 Unit Leaderboard</h1>

        {loading ? (
          <div className="text-center py-12 text-gray-500">Loading...</div>
        ) : units.length === 0 ? (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center text-gray-500 dark:text-gray-400">
            No units found.
          </div>
        ) : (
          <div className="space-y-4">
            {units.map((unit, index) => (
              <div
                key={unit.unitId}
                className="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border-l-8 border-blue-500"
              >
                <div className="flex items-center gap-4">
                  <div className="text-3xl font-bold text-gray-400 dark:text-gray-500 w-12 text-center">
                    #{index + 1}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-3">
                      {unit.profileImage ? (
                        <img
                          src={`http://localhost:8080${unit.profileImage}`}
                          alt={unit.name}
                          className="w-12 h-12 rounded-full object-cover border-2 border-blue-500"
                        />
                      ) : (
                        <div className="w-12 h-12 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center text-lg font-bold text-blue-600 dark:text-blue-300">
                          {unit.name?.[0]?.toUpperCase() || 'U'}
                        </div>
                      )}
                      <div>
                        <h2 className="text-xl font-bold text-gray-800 dark:text-white">{unit.name}</h2>
                        {unit.motto && (
                          <p className="text-sm text-blue-600 dark:text-blue-400 italic">"{unit.motto}"</p>
                        )}
                        <p className="text-sm text-gray-500 dark:text-gray-400">{unit.type}</p>
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-2xl font-bold text-green-600 dark:text-green-400">
                      {unit.resolvedCount}
                    </div>
                    <div className="text-sm text-gray-500 dark:text-gray-400">Resolved</div>
                    <div className="text-sm text-yellow-500">⭐ {unit.avgRating?.toFixed(1) || 'N/A'}</div>
                  </div>
                </div>
                <div className="mt-2 text-sm text-gray-600 dark:text-gray-300">
                  Total cases: {unit.totalCases} • {unit.coverageArea || 'N/A'}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default Leaderboard;