import { useEffect, useState } from 'react';
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

interface IncidentMapProps {
  incidents?: Array<{
    id: string;
    title: string;
    description: string;
    latitude: number;
    longitude: number;
    status: string;
    type: 'case' | 'sos' | 'unit';
  }>;
  center?: [number, number];
  zoom?: number;
  onMarkerClick?: (id: string) => void;
  showRadius?: boolean;
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

function ChangeView({ center, zoom }: { center: [number, number]; zoom: number }) {
  const map = useMap();
  useEffect(() => {
    map.setView(center, zoom);
  }, [center, zoom, map]);
  return null;
}

export default function IncidentMap({
  incidents = [],
  center = [6.5244, 3.3792], // Default: Lagos
  zoom = 12,
  onMarkerClick,
  showRadius = false
}: IncidentMapProps) {
  const [userLocation, setUserLocation] = useState<[number, number] | null>(null);

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
      case 'investigating': return 'text-blue-600';
      case 'resolved': return 'text-green-600';
      case 'closed': return 'text-gray-600';
      default: return 'text-gray-600';
    }
  };

  const getIcon = (type: string) => {
    switch (type) {
      case 'sos': return sosIcon;
      case 'unit': return unitIcon;
      default: return caseIcon;
    }
  };

  return (
    <div className="w-full h-[500px] rounded-lg overflow-hidden shadow-lg border border-gray-200 dark:border-gray-700">
      <MapContainer
        center={center}
        zoom={zoom}
        style={{ height: '100%', width: '100%' }}
        className="z-0"
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
        {incidents.map((incident) => (
          <Marker
            key={incident.id}
            position={[incident.latitude, incident.longitude]}
            icon={getIcon(incident.type)}
            eventHandlers={{
              click: () => onMarkerClick && onMarkerClick(incident.id)
            }}
          >
            <Popup>
              <div className="p-2 min-w-[200px]">
                <h3 className="font-bold text-gray-800 dark:text-gray-200">{incident.title}</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{incident.description}</p>
                <div className="mt-2 flex items-center gap-2">
                  <span className={`text-xs font-medium ${getStatusColor(incident.status)}`}>
                    {incident.status.toUpperCase()}
                  </span>
                  <span className="text-xs text-gray-500">
                    {incident.type.toUpperCase()}
                  </span>
                </div>
                {incident.type === 'unit' && (
                  <button 
                    onClick={() => window.location.href = `/report?unit=${incident.id}`}
                    className="mt-2 w-full text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700"
                  >
                    Report to this Unit
                  </button>
                )}
              </div>
            </Popup>
          </Marker>
        ))}
      </MapContainer>
    </div>
  );
}