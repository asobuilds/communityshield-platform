import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';
import { useLanguage } from '../../context/LanguageContext';
import { BackButton } from '../../components/common/BackButton';

interface Unit {
  id: string;
  name: string;
  distance?: number;
}

export default function SOSPage() {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [loading, setLoading] = useState(false);
  const [location, setLocation] = useState<{ lat: number; lng: number } | null>(null);
  const [locationLoading, setLocationLoading] = useState(true);
  const [description, setDescription] = useState('');
  const [selectedUnit, setSelectedUnit] = useState('');
  const [units, setUnits] = useState<Unit[]>([]);
  const [countdown, setCountdown] = useState(0);

  useEffect(() => {
    getLocation();
    fetchUnits();
  }, []);

  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => setCountdown(countdown - 1), 1000);
      return () => clearTimeout(timer);
    }
  }, [countdown]);

  const getLocation = () => {
    setLocationLoading(true);
    if (!navigator.geolocation) {
      toast.error('Geolocation not supported');
      setLocationLoading(false);
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLocation({ lat: pos.coords.latitude, lng: pos.coords.longitude });
        setLocationLoading(false);
        toast.success('📍 Location captured');
      },
      () => {
        toast.error('Unable to get location. Please enter manually.');
        setLocationLoading(false);
      }
    );
  };

  const fetchUnits = async () => {
    try {
      const response = await axios.get('/api/v1/units');
      setUnits(response.data.units || []);
    } catch (error) {
      console.error('Failed to fetch units');
    }
  };

  const handleSendSOS = async () => {
    if (!location) {
      toast.error('Please enable location');
      return;
    }

    if (!description.trim()) {
      toast.error('Please describe your emergency');
      return;
    }

    setLoading(true);
    setCountdown(5);

    try {
      const payload = {
        latitude: location.lat,
        longitude: location.lng,
        description: description,
        unitId: selectedUnit || undefined,
      };

      const response = await axios.post('/api/v1/sos/send', payload);
      toast.success('🚨 SOS alert sent successfully!');

      setTimeout(() => {
        navigate('/sos-history');
      }, 2000);
    } catch (error: any) {
      toast.error(error.response?.data?.error || t('errors.sos_failed'));
    } finally {
      setLoading(false);
      setCountdown(0);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
      <div className="max-w-2xl mx-auto">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
          <div className="text-center mb-6">
            <div className="text-6xl mb-4">🚨</div>
            <h1 className="text-3xl font-bold text-red-600 dark:text-red-400">
              {t('sos.title')}
            </h1>
            <p className="text-gray-600 dark:text-gray-400">
              {t('sos.description')}
            </p>
          </div>
          <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4">
            <div className="max-w-2xl mx-auto">
              <BackButton />
              {/* rest of content */}
            </div>
          </div>

          {/* Location Status */}
          <div className="mb-4 p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                📍 Location:
              </span>
              {locationLoading ? (
                <span className="text-sm text-gray-500">{t('common.loading')}</span>
              ) : location ? (
                <span className="text-sm text-green-600">
                  ✅ {location.lat.toFixed(6)}, {location.lng.toFixed(6)}
                </span>
              ) : (
                <span className="text-sm text-red-600">❌ No location</span>
              )}
              <button
                onClick={getLocation}
                className="text-sm text-blue-600 hover:underline"
              >
                {t('common.refresh') || 'Refresh'}
              </button>
            </div>
          </div>

          {/* Description */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('sos.description')} *
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={4}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600 dark:text-white"
              placeholder={t('sos.description')}
            />
          </div>

          {/* Unit Selection */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {t('sos.selectUnit') || 'Select Specific Unit (Optional)'}
            </label>
            <select
              value={selectedUnit}
              onChange={(e) => setSelectedUnit(e.target.value)}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600 dark:text-white"
            >
              <option value="">{t('sos.allUnits') || 'All Nearby Units'}</option>
              {units.map((unit) => (
                <option key={unit.id} value={unit.id}>
                  {unit.name} {unit.distance ? `(${unit.distance.toFixed(1)}km)` : ''}
                </option>
              ))}
            </select>
          </div>

          {/* Warning */}
          <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <p className="text-sm text-red-600 dark:text-red-400">
              ⚠️ Only use this for genuine emergencies. False alarms will be penalized.
            </p>
          </div>

          {/* Send Button */}
          <button
            onClick={handleSendSOS}
            disabled={loading || !location || countdown > 0}
            className="w-full bg-red-600 text-white py-4 rounded-lg text-xl font-bold hover:bg-red-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? (
              countdown > 0 ? `Sending in ${countdown}...` : 'Sending...'
            ) : (
              `🚨 ${t('sos.send')}`
            )}
          </button>

          <div className="mt-4 text-center">
            <Link to="/home" className="text-gray-600 dark:text-gray-400 hover:underline">
              ← {t('common.back')}
            </Link>
          </div>
        </div>

        {/* Quick Tips */}
        <div className="mt-4 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h3 className="font-semibold text-gray-800 dark:text-gray-200 mb-2">
            💡 When to use SOS:
          </h3>
          <ul className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
            <li>• Life-threatening emergencies</li>
            <li>• Physical attacks or violence</li>
            <li>• Medical emergencies</li>
            <li>• Fire or natural disasters</li>
            <li>• Any situation requiring immediate response</li>
          </ul>
        </div>
      </div>
    </div>
  );
}