import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';
import IncidentMap from '../../components/map/IncidentMap';
import { useLanguage } from '../../context/LanguageContext';
import { BackButton } from '../../components/common/BackButton';

interface Incident {
  id: string;
  title: string;
  description: string;
  latitude: number;
  longitude: number;
  status: string;
  priority: string;
  type: 'case' | 'sos' | 'unit';
  createdAt: string;
}

export default function MapPage() {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [loading, setLoading] = useState(true);
  const [userLocation, setUserLocation] = useState<[number, number]>([6.5244, 3.3792]);
  const [showRadius, setShowRadius] = useState(true);
  const [showHeatmap, setShowHeatmap] = useState(false);
  const [filter, setFilter] = useState<'all' | 'cases' | 'sos' | 'units'>('all');
  const [statusFilter, setStatusFilter] = useState<'all' | 'pending' | 'investigating' | 'resolved'>('all');

  useEffect(() => {
    fetchIncidents();
  }, []);

  const fetchIncidents = async () => {
    setLoading(true);
    try {
      if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
          (pos) => {
            setUserLocation([pos.coords.latitude, pos.coords.longitude]);
          },
          () => {
            console.log('Using default location');
          }
        );
      }

      const [casesRes, unitsRes] = await Promise.all([
        axios.get('/api/v1/cases').catch(() => ({ data: { cases: [] } })),
        axios.get('/api/v1/units').catch(() => ({ data: { units: [] } }))
      ]);

      const cases = casesRes.data.cases || [];
      const units = unitsRes.data.units || [];

      const mappedCases = cases.map((c: any) => ({
        id: c.id,
        title: c.title,
        description: c.description,
        latitude: c.latitude || 6.5244,
        longitude: c.longitude || 3.3792,
        status: c.status || 'pending',
        priority: c.priority || 'medium',
        type: 'case' as const,
        createdAt: c.createdAt
      }));

      const mappedUnits = units.map((u: any) => ({
        id: u.id,
        title: u.name,
        description: `📍 ${u.city || u.state || 'Nigeria'} | Contact: ${u.contactPerson || 'N/A'}`,
        latitude: u.latitude || 6.5244,
        longitude: u.longitude || 3.3792,
        status: 'active',
        priority: 'low',
        type: 'unit' as const,
        createdAt: u.createdAt
      }));

      setIncidents([...mappedCases, ...mappedUnits]);
    } catch (error) {
      toast.error(t('errors.generic'));
    } finally {
      setLoading(false);
    }
  };

  const handleMarkerClick = (id: string) => {
    navigate(`/case/${id}`);
  };

  const filteredIncidents = incidents.filter(incident => {
    if (filter === 'all') return true;
    if (filter === 'cases') return incident.type === 'case';
    if (filter === 'sos') return incident.type === 'sos';
    if (filter === 'units') return incident.type === 'unit';
    return true;
  }).filter(incident => {
    if (statusFilter === 'all') return true;
    return incident.status === statusFilter;
  });

  return (
    <div>
      <div className="flex flex-wrap justify-between items-center gap-4 mb-6">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-200">
          🗺️ {t('map.title')}
        </h1>
        <div>
          <BackButton />
          {/* rest of content */}
        </div>
        <div className="flex flex-wrap gap-2">
          <Link to="/report">
            <button className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition text-sm">
              + {t('home.quickActions.reportCase')}
            </button>
          </Link>
          <Link to="/sos">
            <button className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition text-sm">
              🆘 {t('home.quickActions.sos')}
            </button>
          </Link>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-2 mb-4">
        <button
          onClick={() => setFilter('all')}
          className={`px-3 py-1 rounded-full text-sm transition ${
            filter === 'all' 
              ? 'bg-blue-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          {t('map.all')}
        </button>
        <button
          onClick={() => setFilter('cases')}
          className={`px-3 py-1 rounded-full text-sm transition ${
            filter === 'cases' 
              ? 'bg-blue-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          📋 {t('map.cases')}
        </button>
        <button
          onClick={() => setFilter('sos')}
          className={`px-3 py-1 rounded-full text-sm transition ${
            filter === 'sos' 
              ? 'bg-red-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          🆘 {t('map.sos')}
        </button>
        <button
          onClick={() => setFilter('units')}
          className={`px-3 py-1 rounded-full text-sm transition ${
            filter === 'units' 
              ? 'bg-blue-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          🏢 {t('map.units')}
        </button>

        <div className="w-px h-6 bg-gray-300 dark:bg-gray-600 mx-1"></div>

        <button
          onClick={() => setStatusFilter('all')}
          className={`px-3 py-1 rounded-full text-sm transition ${
            statusFilter === 'all' 
              ? 'bg-blue-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          {t('map.all')}
        </button>
        <button
          onClick={() => setStatusFilter('pending')}
          className={`px-3 py-1 rounded-full text-sm transition ${
            statusFilter === 'pending' 
              ? 'bg-yellow-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          ⏳ {t('cases.pending')}
        </button>
        <button
          onClick={() => setStatusFilter('investigating')}
          className={`px-3 py-1 rounded-full text-sm transition ${
            statusFilter === 'investigating' 
              ? 'bg-orange-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          🔍 {t('cases.investigating')}
        </button>
        <button
          onClick={() => setStatusFilter('resolved')}
          className={`px-3 py-1 rounded-full text-sm transition ${
            statusFilter === 'resolved' 
              ? 'bg-green-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          ✅ {t('cases.resolved')}
        </button>

        <button
          onClick={() => setShowRadius(!showRadius)}
          className={`px-3 py-1 rounded-full text-sm transition ${
            showRadius 
              ? 'bg-green-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          {showRadius ? '📍 Radius On' : '📍 Radius Off'}
        </button>
        <button
          onClick={() => setShowHeatmap(!showHeatmap)}
          className={`px-3 py-1 rounded-full text-sm transition ${
            showHeatmap 
              ? 'bg-red-600 text-white' 
              : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
          }`}
        >
          {showHeatmap ? '🔥 Heatmap On' : '🔥 Heatmap Off'}
        </button>
        <button
          onClick={fetchIncidents}
          className="px-3 py-1 rounded-full text-sm bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600 transition"
        >
          🔄 {t('map.refresh')}
        </button>
      </div>

      {/* Map */}
      {loading ? (
        <div className="flex justify-center items-center h-[500px] bg-white dark:bg-gray-800 rounded-lg">
          <p className="text-gray-500 dark:text-gray-400">{t('map.loading')}</p>
        </div>
      ) : (
        <IncidentMap
          incidents={filteredIncidents}
          center={userLocation}
          zoom={12}
          onMarkerClick={handleMarkerClick}
          showRadius={showRadius}
          showHeatmap={showHeatmap}
          filterType={filter}
          filterStatus={statusFilter}
        />
      )}

      {/* Legend */}
      <div className="mt-4 flex flex-wrap gap-4 bg-white dark:bg-gray-800 p-3 rounded-lg shadow">
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-red-500 rounded-full"></div>
          <span className="text-sm text-gray-700 dark:text-gray-300">{t('map.cases')}</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-orange-500 rounded-full"></div>
          <span className="text-sm text-gray-700 dark:text-gray-300">{t('cases.investigating')}</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-green-500 rounded-full"></div>
          <span className="text-sm text-gray-700 dark:text-gray-300">{t('cases.resolved')}</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-blue-500 rounded-full"></div>
          <span className="text-sm text-gray-700 dark:text-gray-300">{t('map.units')}</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-yellow-500 rounded-full"></div>
          <span className="text-sm text-gray-700 dark:text-gray-300">{t('cases.pending')}</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="w-4 h-4 bg-red-700 rounded-full"></div>
          <span className="text-sm text-gray-700 dark:text-gray-300">{t('map.sos')}</span>
        </div>
      </div>
    </div>
  );
}