import React, { useState } from 'react';
import { Shield, AlertOctagon, Radio, MapPin, Eye, Camera, Mic, Activity } from 'lucide-react';

export default function App() {
  const [sosActive, setSosActive] = useState(false);
  const [currentLocation, setCurrentLocation] = useState(null);
  const [trackingId, setTrackingId] = useState("");
  const [loading, setLoading] = useState(false);

  // Trigger Native Browser Geolocation API
  const handleSOSClick = () => {
    if (sosActive) {
      setSosActive(false);
      return;
    }

    setLoading(true);
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (position) => {
          setCurrentLocation({
            lat: position.coords.latitude.toFixed(4),
            lng: position.coords.longitude.toFixed(4)
          });
          setSosActive(true);
          setLoading(false);
          setTrackingId("SOS-" + Math.floor(100000 + Math.random() * 900000));
        },
        () => {
          // Fallback static coordinate mocks if permissions are locally blocked
          setCurrentLocation({ lat: "6.6111", lng: "3.5074" });
          setSosActive(true);
          setLoading(false);
          setTrackingId("SOS-FALLBACK-8941");
        }
      );
    }
  };

  return (
    <div className="max-w-md mx-auto min-h-screen bg-tactical-darkest text-slate-100 flex flex-col justify-between p-4 border-x border-tactical-border shadow-2xl">
      
      {/* Contemporary Top Banner Navigation */}
      <header className="flex items-center justify-between border-b border-tactical-border pb-4 pt-2">
        <div className="flex items-center gap-2">
          <Shield className="w-7 h-7 text-red-500 animate-pulse" />
          <div>
            <h1 className="text-md font-bold tracking-wider text-slate-50">COMMUNITY SHIELD</h1>
            <p className="text-xs text-slate-400">Nigeria Local Security Grid</p>
          </div>
        </div>
        <div className="flex items-center gap-1.5 bg-emerald-950/40 border border-emerald-500/30 px-2.5 py-1 rounded-full">
          <Activity className="w-3.5 h-3.5 text-emerald-400" />
          <span className="text-[10px] font-mono text-emerald-400 tracking-widest uppercase">API Live</span>
        </div>
      </header>

      {/* Main Core View Area */}
      <main className="flex-1 flex flex-col justify-center items-center my-6 gap-8">
        
        {/* The One-Tap Instant SOS Button Controller */}
        <div className="relative flex items-center justify-center">
          <div className={`absolute w-64 h-64 rounded-full transition-all duration-1000 ${sosActive ? 'bg-red-600/20 animate-ping' : 'bg-red-500/5'}`}></div>
          <div className={`absolute w-52 h-52 rounded-full transition-all duration-700 ${sosActive ? 'bg-red-600/30 animate-pulse' : 'bg-red-500/10'}`}></div>
          
          <button 
            onClick={handleSOSClick}
            disabled={loading}
            className={`w-40 h-40 rounded-full flex flex-col items-center justify-center gap-2 font-black text-xl tracking-widest transition-all duration-300 border-4 active:scale-95 shadow-[0_0_40px_rgba(239,68,68,0.2)] ${
              sosActive 
                ? 'bg-gradient-to-b from-red-700 to-red-600 border-red-400 text-white' 
                : 'bg-gradient-to-b from-slate-900 to-tactical-card border-tactical-border text-red-500 hover:border-red-500/50'
            }`}
          >
            <AlertOctagon className={`w-10 h-10 ${sosActive ? 'animate-bounce' : ''}`} />
            <span>{loading ? "PINGING..." : sosActive ? "ACTIVE SOS" : "TAP SOS"}</span>
          </button>
        </div>

        {/* Dynamic Telemetry Status Block */}
        {sosActive && currentLocation && (
          <div className="w-full bg-red-950/20 border border-red-500/20 rounded-xl p-4 space-y-3">
            <div className="flex items-center gap-2 text-red-400 font-bold text-sm tracking-wide">
              <Radio className="w-4 h-4 animate-spin" />
              <span>TRANSMITTING EMERGENCY PROTOCOL...</span>
            </div>
            <div className="grid grid-cols-2 gap-2 text-xs font-mono bg-tactical-darkest/60 p-3 rounded-lg border border-tactical-border">
              <div>
                <span className="text-slate-500 block uppercase tracking-wider text-[10px]">Tracking Token</span>
                <span className="text-red-400 font-bold">{trackingId}</span>
              </div>
              <div>
                <span className="text-slate-500 block uppercase tracking-wider text-[10px]">GPS Broadcast</span>
                <span className="text-slate-200">{currentLocation.lat}N, {currentLocation.lng}E</span>
              </div>
            </div>
          </div>
        )}

        {/* Contemporary Media Evidence & Reporting Shortcuts */}
        <div className="w-full grid grid-cols-3 gap-3">
          <button className="bg-tactical-card border border-tactical-border p-3.5 rounded-xl flex flex-col items-center gap-1.5 hover:border-slate-700 transition">
            <Camera className="w-5 h-5 text-slate-400" />
            <span className="text-xs font-medium text-slate-400">Capture Video</span>
          </button>
          <button className="bg-tactical-card border border-tactical-border p-3.5 rounded-xl flex flex-col items-center gap-1.5 hover:border-slate-700 transition">
            <Mic className="w-5 h-5 text-slate-400" />
            <span className="text-xs font-medium text-slate-400">Voice Note</span>
          </button>
          <button className="bg-tactical-card border border-tactical-border p-3.5 rounded-xl flex flex-col items-center gap-1.5 hover:border-slate-700 transition">
            <Eye className="w-5 h-5 text-slate-400" />
            <span className="text-xs font-medium text-slate-400">Report Case</span>
          </button>
        </div>
      </main>

      {/* Strategic Footer Status */}
      <footer className="border-t border-tactical-border pt-4 pb-2 text-center">
        <div className="flex items-center justify-center gap-1.5 text-xs text-slate-500">
          <MapPin className="w-3.5 h-3.5" />
          <span>Tracking Active: Security Unit Proximity Lock</span>
        </div>
      </footer>
    </div>
  );
}
