import axios from 'axios';

// IndexedDB for offline storage
const DB_NAME = 'CommunityShieldDB';
const DB_VERSION = 1;
const STORES = {
  CASES: 'cases',
  UNITS: 'units',
  SOS: 'sos',
  MESSAGES: 'messages',
  PENDING_REQUESTS: 'pendingRequests'
};

let db: IDBDatabase | null = null;

// Initialize IndexedDB
export function initDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    if (db) {
      resolve(db);
      return;
    }

    const request = indexedDB.open(DB_NAME, DB_VERSION);

    request.onupgradeneeded = (event) => {
      const database = (event.target as IDBOpenDBRequest).result;
      
      // Create object stores
      if (!database.objectStoreNames.contains(STORES.CASES)) {
        database.createObjectStore(STORES.CASES, { keyPath: 'id' });
      }
      if (!database.objectStoreNames.contains(STORES.UNITS)) {
        database.createObjectStore(STORES.UNITS, { keyPath: 'id' });
      }
      if (!database.objectStoreNames.contains(STORES.SOS)) {
        database.createObjectStore(STORES.SOS, { keyPath: 'id' });
      }
      if (!database.objectStoreNames.contains(STORES.MESSAGES)) {
        database.createObjectStore(STORES.MESSAGES, { keyPath: 'id', autoIncrement: true });
      }
      if (!database.objectStoreNames.contains(STORES.PENDING_REQUESTS)) {
        database.createObjectStore(STORES.PENDING_REQUESTS, { keyPath: 'id', autoIncrement: true });
      }
    };

    request.onsuccess = (event) => {
      db = (event.target as IDBOpenDBRequest).result;
      resolve(db);
    };

    request.onerror = (event) => {
      reject((event.target as IDBOpenDBRequest).error);
    };
  });
}

// Save data to IndexedDB
export async function saveToDB(storeName: string, data: any) {
  const database = await initDB();
  return new Promise((resolve, reject) => {
    const transaction = database.transaction(storeName, 'readwrite');
    const store = transaction.objectStore(storeName);
    const request = store.put(data);
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

// Get all data from a store
export async function getAllFromDB(storeName: string): Promise<any[]> {
  const database = await initDB();
  return new Promise((resolve, reject) => {
    const transaction = database.transaction(storeName, 'readonly');
    const store = transaction.objectStore(storeName);
    const request = store.getAll();
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

// Get single item from store
export async function getFromDB(storeName: string, id: string): Promise<any> {
  const database = await initDB();
  return new Promise((resolve, reject) => {
    const transaction = database.transaction(storeName, 'readonly');
    const store = transaction.objectStore(storeName);
    const request = store.get(id);
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

// Delete from store
export async function deleteFromDB(storeName: string, id: string) {
  const database = await initDB();
  return new Promise((resolve, reject) => {
    const transaction = database.transaction(storeName, 'readwrite');
    const store = transaction.objectStore(storeName);
    const request = store.delete(id);
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

// Clear store
export async function clearStore(storeName: string) {
  const database = await initDB();
  return new Promise((resolve, reject) => {
    const transaction = database.transaction(storeName, 'readwrite');
    const store = transaction.objectStore(storeName);
    const request = store.clear();
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

// Add pending request for offline sync
export async function addPendingRequest(url: string, method: string, data?: any) {
  const pendingRequest = {
    url,
    method,
    data,
    timestamp: new Date().toISOString(),
    retries: 0
  };
  await saveToDB(STORES.PENDING_REQUESTS, pendingRequest);
}

// Sync pending requests when online
export async function syncPendingRequests() {
  const pendingRequests = await getAllFromDB(STORES.PENDING_REQUESTS);
  
  for (const request of pendingRequests) {
    try {
      let response;
      switch (request.method) {
        case 'POST':
          response = await axios.post(request.url, request.data);
          break;
        case 'PUT':
          response = await axios.put(request.url, request.data);
          break;
        case 'DELETE':
          response = await axios.delete(request.url);
          break;
        default:
          response = await axios.get(request.url);
      }
      
      if (response.status === 200 || response.status === 201) {
        await deleteFromDB(STORES.PENDING_REQUESTS, request.id);
        console.log('✅ Synced:', request.url);
      }
    } catch (error) {
      console.error('❌ Failed to sync:', request.url);
      // Increment retries
      request.retries = (request.retries || 0) + 1;
      if (request.retries > 3) {
        await deleteFromDB(STORES.PENDING_REQUESTS, request.id);
      } else {
        await saveToDB(STORES.PENDING_REQUESTS, request);
      }
    }
  }
}

// Check if online
export function isOnline(): boolean {
  return navigator.onLine;
}

// Setup online/offline listeners
export function setupOfflineListeners() {
  window.addEventListener('online', () => {
    console.log('🟢 Online - Syncing data...');
    syncPendingRequests();
  });

  window.addEventListener('offline', () => {
    console.log('🔴 Offline - Data will be saved locally');
  });

  // Initial sync if online
  if (isOnline()) {
    syncPendingRequests();
  }
}