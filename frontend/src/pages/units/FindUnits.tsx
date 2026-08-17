import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function FindUnits() {
  const navigate = useNavigate();
  const [units, setUnits] = useState<any[]>([]);
  const [usedUnits, setUsedUnits] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState<'nearest' | 'popular'>('nearest');
  const [userLocation, setUserLocation] = useState<{ lat: number; lng: number } | null>(null);
  const user = JSON.parse(localStorage.getItem('user') || '{}');

  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          setUserLocation({ lat: pos.coords.latitude, lng: pos.coords.longitude });
        },
        () => {
          toast.error('Location access denied – showing all units');
        }
      );
    }
  }, []);

  useEffect(() => {
    if (userLocation) {
      fetchUnits();
      fetchUsedUnits();
    }
  }, [userLocation, sortBy]);

  const fetchUnits = async () => {
    if (!userLocation) return;
    setLoading(true);
    try {
      const response = await axios.get(
        `/api/v1/units/nearby?lat=${userLocation.lat}&lng=${userLocation.lng}&radius=50`
      );
      let unitList = response.data.units || [];

      const unitsWithRatings = await Promise.all(
        unitList.map(async (unit: any) => {
          try {
            const ratingRes = await axios.get(`/api/v1/ratings/units/${unit.id}`);
            return { ...unit, averageRating: ratingRes.data.average || 0, ratingCount: ratingRes.data.count || 0 };
          } catch {
            return { ...unit, averageRating: 0, ratingCount: 0 };
          }
        })
      );

      if (sortBy === 'popular') {
        unitsWithRatings.sort((a, b) => b.averageRating - a.averageRating);
      } else {
        unitsWithRatings.sort((a, b) => a.distance - b.distance);
      }

      setUnits(unitsWithRatings);
    } catch (error) {
      toast.error('Failed to load units');
    } finally {
      setLoading(false);
    }
  };

  const fetchUsedUnits = async () => {
    if (!user.id) return;
    try {
      const response = await axios.get(`/api/v1/users/${user.id}/used-units`);
      setUsedUnits(response.data.usedUnits || []);
    } catch (error) {
      console.error('Failed to fetch used units');
    }
  };

  const renderStars = (rating: number) => {
    const fullStars = Math.floor(rating);
    const hasHalf = rating % 1 >= 0.5;
    const emptyStars = 5 - fullStars - (hasHalf ? 1 : 0);
    return (
      <span className="text-yellow-500">
        {'★'.repeat(fullStars)}
        {hasHalf && '☆'}
        {'☆'.repeat(emptyStars)}
      </span>
    );
  };

  const filteredUnits = units.filter(unit =>
    unit.name.toLowerCase().includes(search.toLowerCase()) ||
    unit.type.toLowerCase().includes(search.toLowerCase())
  );

  const usedUnitIds = usedUnits.map(u => u.unit.id);

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-8">
      <div className="max-w-6xl mx-auto">
        {/* Back Button */}
        <button
          onClick={() => navigate('/dashboard')}
          className="text-blue-600 dark:text-blue-400 hover:underline mb-4 flex items-center gap-1"
        >
          ← Back to Dashboard
        </button>

        <div className="flex justify-between items-center mb-6 flex-wrap gap-4">
          <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400">🏢 Find Security Units</h1>
          <div className="flex gap-2">
            <button
              onClick={() => setSortBy('nearest')}
              className={`px-4 py-2 rounded ${sortBy === 'nearest' ? 'bg-blue-600 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}`}
            >
              Nearest
            </button>
            <button
              onClick={() => setSortBy('popular')}
              className={`px-4 py-2 rounded ${sortBy === 'popular' ? 'bg-blue-600 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}`}
            >
              Most Popular
            </button>
          </div>
        </div>

        {/* Search bar */}
        <div className="mb-6">
          <input
            type="text"
            placeholder="Search by name or type..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full px-4 py-3 border rounded-lg shadow-sm focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-700 dark:text-white"
          />
        </div>

        {/* Used Units Section */}
        {usedUnits.length > 0 && (
          <div className="mb-8">
            <h2 className="text-xl font-semibold text-gray-700 dark:text-gray-300 mb-3">🔄 Recently Used Units</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {usedUnits.slice(0, 3).map((item) => (
                <div key={item.unit.id} className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow border-l-4 border-green-500">
                  <h3 className="font-bold text-blue-600 dark:text-blue-400">{item.unit.name}</h3>
                  <p className="text-sm text-gray-500 dark:text-gray-400">{item.unit.type}</p>
                  <p className="text-sm text-gray-600 dark:text-gray-300">Used {item.count} time(s)</p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Units List */}
        {loading ? (
          <div className="text-center py-12 text-gray-500">Loading units...</div>
        ) : filteredUnits.length === 0 ? (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-8 text-center text-gray-500 dark:text-gray-400">
            <p>No units found matching your criteria.</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredUnits.map((unit) => (
              <div
                key={unit.id}
                className={`bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 border-2 transition-all ${usedUnitIds.includes(unit.id) ? 'border-green-400' : 'border-gray-200'}`}
              >
                <h3 className="text-xl font-semibold text-blue-600 dark:text-blue-400">{unit.name}</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">Type: {unit.type}</p>
                <p className="text-sm text-gray-600 dark:text-gray-300">📍 {unit.distance.toFixed(1)} km away</p>
                <div className="mt-2 flex items-center gap-2">
                  <span className="text-sm font-medium">{unit.averageRating?.toFixed(1) || 'New'}</span>
                  {renderStars(unit.averageRating || 0)}
                  <span className="text-xs text-gray-400">({unit.ratingCount || 0} reviews)</span>
                </div>
                {unit.contactPerson && (
                  <p className="text-sm text-gray-600 dark:text-gray-300 mt-1">Contact: {unit.contactPerson}</p>
                )}
                {usedUnitIds.includes(unit.id) && (
                  <div className="mt-2 inline-block bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200 text-xs px-2 py-1 rounded">
                    ✅ Previously used
                  </div>
                )}
                <button
                  onClick={() => navigate(`/report?unit=${unit.id}`)}
                  className="mt-4 w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 transition"
                >
                  Select This Unit
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default FindUnits;