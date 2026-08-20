import { useOffline } from '../../context/OfflineContext';
import { useLanguage } from '../../context/LanguageContext';

export function OfflineStatus() {
  const { isOnline, isOffline, pendingCount, syncData } = useOffline();
  const { t } = useLanguage();

  if (isOnline && pendingCount === 0) {
    return null;
  }

  return (
    <div className={`fixed bottom-4 left-4 right-4 md:left-auto md:right-4 z-50 max-w-md mx-auto md:mx-0 ${
      isOffline ? 'bg-red-600' : 'bg-yellow-600'
    } text-white rounded-lg shadow-lg p-4`}>
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {isOffline ? (
            <>
              <span className="text-2xl">📡</span>
              <div>
                <p className="font-semibold">You are offline</p>
                <p className="text-sm opacity-90">Data will sync when online</p>
              </div>
            </>
          ) : (
            <>
              <span className="text-2xl">🔄</span>
              <div>
                <p className="font-semibold">Syncing...</p>
                <p className="text-sm opacity-90">{pendingCount} items pending</p>
              </div>
            </>
          )}
        </div>
        {!isOffline && pendingCount > 0 && (
          <button
            onClick={syncData}
            className="bg-white text-gray-800 px-3 py-1 rounded text-sm font-medium hover:bg-gray-100"
          >
            Sync Now
          </button>
        )}
      </div>
    </div>
  );
}