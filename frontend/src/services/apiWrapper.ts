import axios from 'axios';
import * as offlineService from './offlineService';

// API wrapper with offline support
export async function apiRequest<T>(
  method: 'GET' | 'POST' | 'PUT' | 'DELETE',
  url: string,
  data?: any
): Promise<T> {
  const isOnline = offlineService.isOnline();

  if (!isOnline) {
    // Offline: Save request and return cached data if available
    await offlineService.addPendingRequest(url, method, data);
    
    // Try to get cached data
    try {
      const cacheKey = url.replace(/[^a-zA-Z0-9]/g, '_');
      const cached = await offlineService.getFromDB('cases', cacheKey);
      if (cached) {
        return cached as T;
      }
    } catch (e) {
      console.log('No cached data available');
    }
    
    throw new Error('Offline - Request saved for later sync');
  }

  try {
    let response;
    switch (method) {
      case 'GET':
        response = await axios.get(url);
        break;
      case 'POST':
        response = await axios.post(url, data);
        break;
      case 'PUT':
        response = await axios.put(url, data);
        break;
      case 'DELETE':
        response = await axios.delete(url);
        break;
    }

    // Cache successful responses
    if (method === 'GET' && response.data) {
      const cacheKey = url.replace(/[^a-zA-Z0-9]/g, '_');
      await offlineService.saveToDB('cases', { id: cacheKey, ...response.data });
    }

    return response.data;
  } catch (error) {
    // On network error, save for later
    if (method !== 'GET') {
      await offlineService.addPendingRequest(url, method, data);
    }
    throw error;
  }
}

// Helper for offline case reporting
export async function reportCaseOffline(caseData: any) {
  const isOnline = offlineService.isOnline();
  
  if (isOnline) {
    return await apiRequest('POST', '/api/v1/cases', caseData);
  } else {
    // Offline: Save case locally with pending status
    await offlineService.addPendingRequest('/api/v1/cases', 'POST', caseData);
    
    // Create local case with offline flag
    const localCase = {
      ...caseData,
      id: 'offline_' + Date.now(),
      status: 'pending_offline',
      synced: false,
      createdAt: new Date().toISOString()
    };
    
    await offlineService.saveToDB('cases', localCase);
    return localCase;
  }
}