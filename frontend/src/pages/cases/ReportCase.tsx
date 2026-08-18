import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';
import Button from '../../components/ui/Button';
import Input from '../../components/ui/Input';
import Badge from '../../components/ui/Badge';
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/Card';

function ReportCase() {
  const navigate = useNavigate();
  const location = useLocation();
  const [loading, setLoading] = useState(false);
  const [locationLoading, setLocationLoading] = useState(true);
  const [locationState, setLocationState] = useState<{ lat: number; lng: number } | null>(null);
  const [locationError, setLocationError] = useState('');
  const [manualLat, setManualLat] = useState('');
  const [manualLng, setManualLng] = useState('');
  const [useManual, setUseManual] = useState(false);
  const [addressQuery, setAddressQuery] = useState('');
  const [addressLoading, setAddressLoading] = useState(false);

  const [units, setUnits] = useState<any[]>([]);
  const [unitsLoading, setUnitsLoading] = useState(false);
  const [selectedUnitId, setSelectedUnitId] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState<'nearest' | 'popular'>('nearest');

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [isAnonymous, setIsAnonymous] = useState(false);
  const [locationAddress, setLocationAddress] = useState('');
  const [evidenceFiles, setEvidenceFiles] = useState<File[]>([]);
  const [uploadingEvidence, setUploadingEvidence] = useState(false);

  useEffect(() => {
    const params = new URLSearchParams(location.search);
    const unitId = params.get('unit');
    if (unitId) {
      setSelectedUnitId(unitId);
      toast('Unit pre-selected from search', { icon: 'ℹ️' });
    }
  }, [location.search]);

  useEffect(() => {
    getLocation();
  }, []);

  useEffect(() => {
    if (locationState) {
      fetchNearbyUnits();
    }
  }, [locationState]);

  const getLocation = () => {
    setLocationLoading(true);
    if (!navigator.geolocation) {
      setLocationError('Geolocation not supported. Please enter manually or search address.');
      setUseManual(true);
      setLocationLoading(false);
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLocationState({ lat: pos.coords.latitude, lng: pos.coords.longitude });
        setLocationLoading(false);
        toast.success('📍 Location captured');
      },
      () => {
        setLocationError('Unable to get location. Please enter manually or search address.');
        setUseManual(true);
        setLocationLoading(false);
        toast.error('Location access denied. Use manual entry or address search.');
      }
    );
  };

  const fetchNearbyUnits = async () => {
    if (!locationState) return;
    setUnitsLoading(true);
    try {
      const response = await axios.get(
        `/api/v1/units/nearby?lat=${locationState.lat}&lng=${locationState.lng}&radius=20`
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
      toast.error('Failed to load nearby units');
    } finally {
      setUnitsLoading(false);
    }
  };

  const searchAddress = async () => {
    if (!addressQuery.trim()) {
      toast.error('Please enter a location');
      return;
    }
    setAddressLoading(true);
    try {
      const response = await axios.get(
        `https://nominatim.openstreetmap.org/search?q=${encodeURIComponent(addressQuery)}&format=json&limit=1`
      );
      if (response.data && response.data.length > 0) {
        const result = response.data[0];
        const lat = parseFloat(result.lat);
        const lng = parseFloat(result.lon);
        if (!isNaN(lat) && !isNaN(lng)) {
          setLocationState({ lat, lng });
          setUseManual(false);
          setLocationError('');
          toast.success(`Location found: ${result.display_name}`);
          setLocationAddress(result.display_name);
        } else {
          toast.error('Invalid coordinates from address');
        }
      } else {
        toast.error('Address not found. Please try a different search.');
      }
    } catch (error) {
      toast.error('Failed to search address. Please try again.');
    } finally {
      setAddressLoading(false);
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      const files = Array.from(e.target.files);
      setEvidenceFiles(files);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title || !description) {
      toast.error('Please fill in title and description');
      return;
    }
    setLoading(true);
    try {
      const userStr = localStorage.getItem('user');
      if (!userStr) {
        toast.error('Please login first');
        navigate('/login');
        return;
      }
      const user = JSON.parse(userStr);

      const payload: any = {
        title,
        description,
        location: locationAddress || (locationState ? `${locationState.lat},${locationState.lng}` : ''),
        latitude: locationState?.lat || 0,
        longitude: locationState?.lng || 0,
        status: 'pending',
        priority: 'medium',
        isAnonymous: isAnonymous,
        reportedBy: user.id,
      };

      if (selectedUnitId) {
        payload.unitId = selectedUnitId;
      }

      const response = await axios.post('/api/v1/cases', payload);
      const caseId = response.data.case.id;
      toast.success('Case reported successfully!');

      if (evidenceFiles.length > 0) {
        setUploadingEvidence(true);
        for (const file of evidenceFiles) {
          const formData = new FormData();
          formData.append('file', file);
          formData.append('description', file.name);
          formData.append('uploadedBy', user.id);
          await axios.post(`/api/v1/cases/${caseId}/evidence`, formData, {
            headers: { 'Content-Type': 'multipart/form-data' },
          });
        }
        toast.success(`${evidenceFiles.length} file(s) uploaded as evidence`);
        setUploadingEvidence(false);
      }

      navigate('/my-cases');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to report case');
    } finally {
      setLoading(false);
    }
  };

  const handleUnitSelect = (unitId: string) => {
    setSelectedUnitId(unitId === selectedUnitId ? null : unitId);
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

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-gray-900 p-4 md:p-8">
      <div className="max-w-4xl mx-auto bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
        <h1 className="text-3xl font-bold text-blue-600 dark:text-blue-400 mb-6">📝 Report a Case</h1>

        {/* Location Section */}
        <div className="mb-6 p-4 bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600">
          <div className="flex justify-between items-center flex-wrap gap-2">
            <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200">📍 Your Location</h2>
            <div className="flex gap-2">
              {!useManual && (
                <button
                  onClick={() => { setUseManual(true); setLocationState(null); }}
                  className="text-sm text-blue-600 dark:text-blue-400 hover:underline"
                >
                  Enter manually
                </button>
              )}
              <button
                onClick={getLocation}
                className="text-sm bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 px-3 py-1 rounded hover:bg-blue-200 dark:hover:bg-blue-800"
              >
                Refresh GPS
              </button>
            </div>
          </div>

          {locationLoading ? (
            <p className="text-gray-500 dark:text-gray-400 mt-2">Detecting location...</p>
          ) : locationState && !useManual ? (
            <div className="mt-2 text-green-700 dark:text-green-400">
              <p>Lat: {locationState.lat.toFixed(6)}, Lng: {locationState.lng.toFixed(6)}</p>
            </div>
          ) : (
            <div className="mt-2">
              <div className="flex flex-wrap gap-2 items-end">
                <div className="flex-1 min-w-[150px]">
                  <Input
                    type="text"
                    value={addressQuery}
                    onChange={(e) => setAddressQuery(e.target.value)}
                    placeholder="Search address, bus stop, street"
                  />
                </div>
                <Button
                  onClick={searchAddress}
                  disabled={addressLoading}
                  variant="primary"
                  size="sm"
                >
                  {addressLoading ? 'Searching...' : 'Search'}
                </Button>
                <Button
                  onClick={() => {
                    const lat = parseFloat(manualLat);
                    const lng = parseFloat(manualLng);
                    if (!isNaN(lat) && !isNaN(lng)) {
                      setLocationState({ lat, lng });
                      setUseManual(false);
                      toast.success('Location set manually');
                    } else {
                      toast.error('Enter valid numbers');
                    }
                  }}
                  variant="primary"
                  size="sm"
                >
                  Set Coords
                </Button>
              </div>
              <div className="flex gap-2 mt-2">
                <Input
                  type="number"
                  step="any"
                  placeholder="Latitude"
                  value={manualLat}
                  onChange={(e) => setManualLat(e.target.value)}
                />
                <Input
                  type="number"
                  step="any"
                  placeholder="Longitude"
                  value={manualLng}
                  onChange={(e) => setManualLng(e.target.value)}
                />
              </div>
            </div>
          )}
          {locationError && <p className="text-red-500 text-sm mt-2">{locationError}</p>}
        </div>

        {/* Nearby Units Section */}
        <div className="mb-6">
          <div className="flex justify-between items-center flex-wrap gap-2">
            <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200">🏢 Nearby Security Units</h2>
            <div className="flex gap-2">
              <a
                href="/units"
                className="px-3 py-1 rounded text-sm bg-blue-600 text-white hover:bg-blue-700"
              >
                Find Units →
              </a>
            </div>
          </div>
          {unitsLoading ? (
            <p className="text-gray-500 dark:text-gray-400 mt-2">Loading units...</p>
          ) : units.length === 0 ? (
            <div className="mt-2 text-gray-600 dark:text-gray-400 bg-yellow-50 dark:bg-yellow-900 p-3 rounded border border-yellow-200 dark:border-yellow-700">
              <p>No units found nearby. You can still report – it will be sent to all available units.</p>
              <button
                onClick={() => setSelectedUnitId(null)}
                className="mt-2 text-sm bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 px-3 py-1 rounded hover:bg-blue-200 dark:hover:bg-blue-800"
              >
                Report to All Units
              </button>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mt-3">
              {units.map((unit) => (
                <div
                  key={unit.id}
                  onClick={() => handleUnitSelect(unit.id)}
                  className={`p-4 border rounded-lg cursor-pointer transition ${
                    selectedUnitId === unit.id
                      ? 'border-blue-600 bg-blue-50 dark:bg-blue-900 dark:border-blue-400 shadow-md'
                      : 'border-gray-200 dark:border-gray-600 hover:border-blue-300 dark:hover:border-blue-500 hover:shadow'
                  }`}
                >
                  <div className="flex justify-between">
                    <h3 className="font-semibold text-blue-700 dark:text-blue-300">{unit.name}</h3>
                    <span className="text-sm text-gray-500 dark:text-gray-400">{unit.distance.toFixed(1)} km</span>
                  </div>
                  <p className="text-sm text-gray-600 dark:text-gray-400">Type: {unit.type}</p>
                  {unit.contactPerson && (
                    <p className="text-sm text-gray-600 dark:text-gray-400">Contact: {unit.contactPerson}</p>
                  )}
                  <div className="mt-1 flex items-center gap-1">
                    <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{unit.averageRating?.toFixed(1) || 'New'}</span>
                    <span className="text-yellow-500">
                      {renderStars(unit.averageRating || 0)}
                    </span>
                    <span className="text-xs text-gray-400">({unit.ratingCount || 0} reviews)</span>
                  </div>
                  {selectedUnitId === unit.id && (
                    <div className="mt-2 text-green-600 dark:text-green-400 text-sm font-medium">✓ Selected</div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Case Details Form */}
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Case Title *</label>
            <Input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              required
            />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description *</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={4}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600 dark:text-white"
              required
            />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Location Address (optional)</label>
            <Input
              type="text"
              value={locationAddress}
              onChange={(e) => setLocationAddress(e.target.value)}
              placeholder="e.g., Lagos, Nigeria"
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Upload Evidence (optional)</label>
            <input
              type="file"
              multiple
              onChange={handleFileChange}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-600 dark:text-white"
              accept="image/*,video/*,audio/*,.pdf,.doc,.docx,.txt"
            />
            {evidenceFiles.length > 0 && (
              <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                {evidenceFiles.length} file(s) selected: {evidenceFiles.map(f => f.name).join(', ')}
              </p>
            )}
          </div>

          <div className="mb-4 flex items-center gap-2">
            <input
              type="checkbox"
              id="anonymous"
              checked={isAnonymous}
              onChange={() => setIsAnonymous(!isAnonymous)}
              className="w-4 h-4 accent-blue-600"
            />
            <label htmlFor="anonymous" className="text-sm text-gray-700 dark:text-gray-300">
              Report anonymously (hide my name from the unit)
            </label>
          </div>

          <Button
            type="submit"
            disabled={loading || uploadingEvidence}
            variant="primary"
            size="lg"
            className="w-full"
          >
            {loading ? 'Submitting...' : uploadingEvidence ? 'Uploading evidence...' : 'Submit Case'}
          </Button>
        </form>

        <div className="mt-6">
          <a href="/dashboard" className="text-blue-600 dark:text-blue-400 hover:underline">← Back to Dashboard</a>
        </div>
      </div>
    </div>
  );
}

export default ReportCase;