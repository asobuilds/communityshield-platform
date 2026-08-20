import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import * as offlineService from '../services/offlineService';

interface OfflineContextType {
  isOnline: boolean;
  isOffline: boolean;
  pendingCount: number;
  syncData: () => Promise<void>;
}

const OfflineContext = createContext<OfflineContextType | undefined>(undefined);

export function OfflineProvider({ children }: { children: ReactNode }) {
  const [isOnline, setIsOnline] = useState(offlineService.isOnline());
  const [pendingCount, setPendingCount] = useState(0);

  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      offlineService.syncPendingRequests();
    };
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // Check pending count periodically
    const interval = setInterval(async () => {
      const pending = await offlineService.getAllFromDB('pendingRequests');
      setPendingCount(pending.length);
    }, 30000);

    // Initial setup
    offlineService.setupOfflineListeners();

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
      clearInterval(interval);
    };
  }, []);

  const syncData = async () => {
    await offlineService.syncPendingRequests();
    const pending = await offlineService.getAllFromDB('pendingRequests');
    setPendingCount(pending.length);
  };

  return (
    <OfflineContext.Provider value={{
      isOnline,
      isOffline: !isOnline,
      pendingCount,
      syncData
    }}>
      {children}
    </OfflineContext.Provider>
  );
}

export function useOffline() {
  const context = useContext(OfflineContext);
  if (!context) {
    throw new Error('useOffline must be used within an OfflineProvider');
  }
  return context;
}