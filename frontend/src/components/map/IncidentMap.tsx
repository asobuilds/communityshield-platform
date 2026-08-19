import { useEffect, useState, useRef } from 'react';
import { MapContainer, TileLayer, Marker, Popup, Circle, useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';

// Fix default marker icons
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
});

interface Incident {
  id: string;
  title: string;
  description: string;
  latitude: number;
  longitude: number;
  status: string;
  priority?: string;
  type: 'case' | 'sos' | 'unit';
  createdAt?: string;
}

interface IncidentMapProps {
  incidents?: Incident[];
  center?: [number, number];
  zoom?: number;
  onMarkerClick?: (id: string) => void;
  showRadius?: boolean;
  showHeatmap?: boolean;
  filterType?: string;
  filterStatus?: string;
}

// Custom marker icons
const caseIcon = new L.Icon({
  iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-red.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41]
});

const sosIcon = new L.Icon({
  iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-red.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
  iconSize: [35, 41],
  iconAnchor: [17, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41]
});

const unitIcon = new L.Icon({
  iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-blue.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41]
});

const resolvedIcon = new L.Icon({
  iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-green.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41]
});

const investigatingIcon = new L.Icon({
  iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-orange.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41]
});

function ChangeView({ center, zoom }: { center: [number, number]; zoom: number }) {
  const map = useMap();
  useEffect(() => {
    map.setView(center, zoom);
  }, [center, zoom, map]);
  return null;
}

export default function IncidentMap({
  incidents = [],
  center = [6.5244, 3.3792],
  zoom = 12,
  onMarkerClick,
  showRadius = true,
  showHeatmap = false,
  filterType = 'all',
  filterStatus = 'all'
}: IncidentMapProps) {
  const [userLocation, setUserLocation] = useState<[number, number] | null>(null);
  const mapRef = useRef<any>(null);

  useEffect(() => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          setUserLocation([pos.coords.latitude, pos.coords.longitude]);
        },
        () => {
          console.log('Unable to get location');
        }
      );
    }
  }, []);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'text-yellow-600';
      case 'investigating': return 'text-orange-600';
      case 'resolved': return 'text-green-600';
      case 'closed': return 'text-gray-600';
      default: return 'text-gray-600';
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'critical': return 'bg-red-600 text-white';
      case 'high': return 'bg-orange-600 text-white';
      case 'medium': return 'bg-yellow-600 text-white';
      case 'low': return 'bg-green-600 text-white';
      default: return 'bg-gray-600 text-white';
    }
  };

  const getIcon = (incident: Incident) => {
    if (incident.type === 'sos') return sosIcon;
    if (incident.type === 'unit') return unitIcon;
    if (incident.status === 'resolved') return resolvedIcon;
    if (incident.status === 'investigating') return investigatingIcon;
    return caseIcon;
  };

  const getStatusEmoji = (status: string) => {
    switch (status) {
      case 'pending': return '⏳';
      case 'investigating': return '🔍';
      case 'resolved': return '✅';
      case 'closed': return '📌';
      default: return '📋';
    }
  };

  // Filter incidents
  const filteredIncidents = incidents.filter(incident => {
    if (filterType !== 'all' && incident.type !== filterType) return false;
    if (filterStatus !== 'all' && incident.status !== filterStatus) return false;
    return true;
  });

  return (
    <div className="w-full h-[500px] rounded-lg overflow-hidden shadow-lg border border-gray-200 dark:border-gray-700 relative">
      <MapContainer
        center={center}
        zoom={zoom}
        style={{ height: '100%', width: '100%' }}
        className="z-0"
        ref={mapRef}
      >
        <ChangeView center={center} zoom={zoom} />
        
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        {/* User location */}
        {userLocation && (
          <>
            <Marker position={userLocation} icon={unitIcon}>
              <Popup>
                <div className="text-sm font-semibold">📍 Your Location</div>
                <div className="text-xs text-gray-500">
                  {userLocation[0].toFixed(6)}, {userLocation[1].toFixed(6)}
                </div>
              </Popup>
            </Marker>
            {showRadius && (
              <Circle
                center={userLocation}
                radius={5000}
                pathOptions={{ color: 'blue', fillColor: 'blue', fillOpacity: 0.1 }}
              />
            )}
          </>
        )}

        {/* Incident markers */}
        {filteredIncidents.map((incident) => (
          <Marker
            key={incident.id}
            position={[incident.latitude, incident.longitude]}
            icon={getIcon(incident)}
            eventHandlers={{
              click: () => onMarkerClick && onMarkerClick(incident.id)
            }}
          >
            <Popup>
              <div className="p-2 min-w-[250px] max-w-[300px]">
                <div className="flex justify-between items-start">
                  <h3 className="font-bold text-gray-800 dark:text-gray-200 text-sm">
                    {incident.title}
                  </h3>
                  {incident.type === 'sos' && (
                    <span className="text-xs bg-red-600 text-white px-2 py-0.5 rounded-full">SOS</span>
                  )}
                </div>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1 line-clamp-2">
                  {incident.description}
                </p>
                <div className="mt-2 flex flex-wrap items-center gap-2">
                  <span className={`text-xs font-medium ${getStatusColor(incident.status)}`}>
                    {getStatusEmoji(incident.status)} {incident.status.toUpperCase()}
                  </span>
                  {incident.priority && (
                    <span className={`text-xs px-2 py-0.5 rounded-full ${getPriorityColor(incident.priority)}`}>
                      {incident.priority}
                    </span>
                  )}
                  <span className="text-xs text-gray-500">
                    {incident.type.toUpperCase()}
                  </span>
                </div>
                {incident.createdAt && (
                  <div className="mt-1 text-xs text-gray-400">
                    {new Date(incident.createdAt).toLocaleDateString()}
                  </div>
                )}
                {incident.type === 'unit' && (
                  <button 
                    onClick={() => window.location.href = `/report?unit=${incident.id}`}
                    className="mt-2 w-full text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 transition"
                  >
                    Report to this Unit
                  </button>
                )}
                {incident.type === 'case' && (
                  <button 
                    onClick={() => window.location.href = `/case/${incident.id}`}
                    className="mt-2 w-full text-sm bg-gray-600 text-white px-3 py-1 rounded hover:bg-gray-700 transition"
                  >
                    View Details
                  </button>
                )}
              </div>
            </Popup>
          </Marker>
        ))}
      </MapContainer>

      {/* Legend */}
      <div className="absolute bottom-4 right-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-3 z-[1000]">
        <div className="space-y-1 text-xs">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-red-500 rounded-full"></div>
            <span className="text-gray-700 dark:text-gray-300">Case</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-orange-500 rounded-full"></div>
            <span className="text-gray-700 dark:text-gray-300">Investigating</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-green-500 rounded-full"></div>
            <span className="text-gray-700 dark:text-gray-300">Resolved</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
            <span className="text-gray-700 dark:text-gray-300">Security Unit</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-yellow-500 rounded-full"></div>
            <span className="text-gray-700 dark:text-gray-300">Pending</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-red-700 rounded-full"></div>
            <span className="text-gray-700 dark:text-gray-300">SOS Alert</span>
          </div>
        </div>
      </div>

      {/* Heatmap toggle */}
      {showHeatmap && (
        <div className="absolute top-4 right-4 z-[1000]">
          <button className="bg-white dark:bg-gray-800 px-3 py-1 rounded-lg shadow text-xs font-medium">
            🔥 Heatmap
          </button>
        </div>
      )}

      {/* Refresh button */}
      <div className="absolute top-4 left-4 z-[1000]">
        <button 
          onClick={() => window.location.reload()}
          className="bg-white dark:bg-gray-800 px-3 py-1 rounded-lg shadow text-xs font-medium hover:bg-gray-100 dark:hover:bg-gray-700 transition"
        >
          🔄 Refresh Map
        </button>
      </div>

      {/* Incident count */}
      <div className="absolute bottom-4 left-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg px-3 py-1 z-[1000] text-xs text-gray-600 dark:text-gray-400">
        {filteredIncidents.length} incidents displayed
      </div>
    </div>
  );
}