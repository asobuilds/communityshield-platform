import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function SOSPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [location, setLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [description, setDescription] = useState('');
  const [locationError, setLocationError] = useState('');
  const [manualLat, setManualLat] = useState('');
  const [manualLng, setManualLng] = useState('');
  const [useManual, setUseManual] = useState(false);
  const [nearestUnit, setNearestUnit] = useState<any>(null);
  const [findingUnit, setFindingUnit] = useState(false);
  const [anonymous, setAnonymous] = useState(false);
  const [deviceId, setDeviceId] = useState('');

  useEffect(() => {
    let storedDeviceId = localStorage.getItem('deviceId');
    if (!storedDeviceId) {
      storedDeviceId = 'device_' + Math.random().toString(36).substring(2, 15) + Date.now().toString(36);
      localStorage.setItem('deviceId', storedDeviceId);
    }
    setDeviceId(storedDeviceId);
    getLocation();
  }, []);

  const getLocation = () => {
    if (!navigator.geolocation) {
      setLocationError('Geolocation not supported. Please enter manually.');
      setUseManual(true);
      return;
    }
    setLocationError('Getting your location...');
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setLocation({ lat: position.coords.latitude, lng: position.coords.longitude });
        setLocationError('');
        toast.success('📍 Location captured');
        findNearestUnit(position.coords.latitude, position.coords.longitude);
      },
      (error) => {
        console.error('Geolocation error:', error);
        let errorMsg = 'Unable to get your location. ';
        if (error.code === 1) {
          errorMsg = '❌ Location access denied. Please enter manually.';
        } else if (error.code === 2) {
          errorMsg = '❌ Location unavailable. Please enter manually.';
        } else if (error.code === 3) {
          errorMsg = '❌ Location request timed out. Please try again or enter manually.';
        }
        setLocationError(errorMsg);
        setUseManual(true);
        toast.error('Location access failed. Please enter manually.');
      },
      { enableHighAccuracy: true, timeout: 10000 }
    );
  };

  const findNearestUnit = async (lat: number, lng: number) => {
    setFindingUnit(true);
    try {
      const response = await axios.get(
        `/api/v1/units/nearby?lat=${lat}&lng=${lng}&radius=20`
      );
      const units = response.data.units || [];
      if (units.length > 0) {
        setNearestUnit(units[0]);
        toast.success(`Nearest unit: ${units[0].name} (${units[0].distance.toFixed(1)} km)`);
      } else {
        setNearestUnit(null);
        toast('No units found nearby. SOS will be sent to all units.', { icon: 'ℹ️' });
      }
    } catch (error) {
      console.error('Failed to find nearest unit:', error);
      setNearestUnit(null);
    } finally {
      setFindingUnit(false);
    }
  };

  const handleUseManual = () => {
    setUseManual(true);
  };

  const handleSendSOS = async () => {
    let lat: number, lng: number;

    if (useManual) {
      lat = parseFloat(manualLat);
      lng = parseFloat(manualLng);
      if (isNaN(lat) || isNaN(lng)) {
        toast.error('Please enter valid latitude and longitude');
        return;
      }
    } else {
      if (!location) {
        toast.error('Location not available. Please enter manually.');
        return;
      }
      lat = location.lat;
      lng = location.lng;
    }

    const userStr = localStorage.getItem('user');
    const user = userStr ? JSON.parse(userStr) : null;

    if (!anonymous && !user) {
      toast.error('Please login or select "Send Anonymously"');
      navigate('/login');
      return;
    }

    setLoading(true);
    try {
      if (!nearestUnit && !useManual) {
        await findNearestUnit(lat, lng);
      }

      const payload: any = {
        latitude: lat,
        longitude: lng,
        description: description.trim() || 'SOS alert! Please send help.',
        unitId: nearestUnit?.id || null,
        anonymous: anonymous,
        deviceId: deviceId,
      };

      if (!anonymous && user) {
        payload.userId = user.id;
      }

      await axios.post('/api/v1/sos', payload);

      const target = nearestUnit
        ? `to ${nearestUnit.name} (${nearestUnit.distance.toFixed(1)} km away)`
        : 'to all available units';
      const anonymity = anonymous ? ' (anonymous)' : '';
      toast.success(`🚨 SOS alert sent successfully ${target}!${anonymity} Help is on the way.`);

      navigate('/sos-history');
    } catch (error: any) {
      console.error('SOS error:', error);
      let errorMsg = error.response?.data?.error || 'Failed to send SOS';
      if (error.response?.data?.retry_after) {
        errorMsg += ` Please wait ${error.response.data.retry_after} seconds.`;
      }
      toast.error(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-xl p-8 max-w-md w-full">
        <div className="text-center mb-6">
          <div className="text-6xl mb-4">🚨</div>
          <h1 className="text-3xl font-bold text-red-600">SOS Emergency</h1>
          <p className="text-gray-600 mt-2">Send an emergency alert with your location</p>
        </div>

        {locationError && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
            <p className="text-yellow-800 text-sm">{locationError}</p>
          </div>
        )}

        {!useManual && location && !locationError && (
          <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-4">
            <p className="text-green-800 text-sm">
              ✅ Location captured: {location.lat.toFixed(6)}, {location.lng.toFixed(6)}
            </p>
          </div>
        )}

        {!useManual && nearestUnit && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
            <p className="text-blue-800 text-sm">
              📍 Nearest unit: <strong>{nearestUnit.name}</strong> ({nearestUnit.distance.toFixed(1)} km away)
            </p>
          </div>
        )}

        {!useManual && nearestUnit === null && !findingUnit && (
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-4">
            <p className="text-gray-600 text-sm">No units found nearby. SOS will be sent to all units.</p>
          </div>
        )}

        {findingUnit && (
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-4">
            <p className="text-gray-600 text-sm">🔍 Finding nearest unit...</p>
          </div>
        )}

        {!useManual && locationError?.includes('denied') && (
          <div className="mb-4">
            <button
              onClick={getLocation}
              className="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700"
            >
              🔄 Retry Location Access
            </button>
            <button
              onClick={handleUseManual}
              className="w-full mt-2 bg-gray-600 text-white py-2 rounded-lg hover:bg-gray-700"
            >
              📍 Enter Location Manually
            </button>
          </div>
        )}

        {useManual && (
          <div className="mb-4">
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-3">
              <p className="text-blue-800 text-sm">📍 Enter your location manually</p>
            </div>
            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="block text-sm font-medium mb-1">Latitude</label>
                <input
                  type="number"
                  step="any"
                  value={manualLat}
                  onChange={(e) => setManualLat(e.target.value)}
                  placeholder="e.g. 6.5244"
                  className="w-full px-3 py-2 border rounded-lg"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Longitude</label>
                <input
                  type="number"
                  step="any"
                  value={manualLng}
                  onChange={(e) => setManualLng(e.target.value)}
                  placeholder="e.g. 3.3792"
                  className="w-full px-3 py-2 border rounded-lg"
                />
              </div>
            </div>
            <p className="text-xs text-gray-500 mt-1">
              Tip: You can get your coordinates from Google Maps (right-click → "What's here?")
            </p>
          </div>
        )}

        <div className="mb-4">
          <label className="block text-sm font-medium mb-1">Message (optional)</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe your emergency..."
            className="w-full px-3 py-2 border rounded-lg"
            rows={3}
          />
        </div>

        <div className="mb-4 flex items-center gap-2">
          <input
            type="checkbox"
            id="anonymous"
            checked={anonymous}
            onChange={() => setAnonymous(!anonymous)}
            className="w-4 h-4"
          />
          <label htmlFor="anonymous" className="text-sm text-gray-700">
            Send anonymously (your identity will be hidden)
          </label>
        </div>

        <button
          onClick={handleSendSOS}
          disabled={loading}
          className="w-full bg-red-600 text-white py-4 rounded-lg hover:bg-red-700 disabled:opacity-50 text-lg font-semibold transition-all duration-200"
        >
          {loading ? 'Sending...' : '🚨 SEND SOS'}
        </button>

        {!useManual && !locationError?.includes('denied') && (
          <button
            onClick={handleUseManual}
            className="w-full mt-2 text-blue-600 hover:underline text-sm"
          >
            Enter location manually instead
          </button>
        )}

        <div className="mt-4 text-center">
          <a href="/dashboard" className="text-blue-600 hover:underline text-sm">
            ← Back to Dashboard
          </a>
        </div>
      </div>
    </div>
  );
}

export default SOSPage;