import { useCallback, useEffect, useReducer } from 'react'
import { api } from '@/lib/apiClient'

/** Persisted between route changes so a prompt is not repeated within a tab. */
const STORAGE_KEY = 'cs.lastLocation'

export type LocationPermission = 'granted' | 'denied' | 'prompt' | 'unavailable'

export interface LocationState {
  latitude: number | null
  longitude: number | null
  accuracy: number | null
  permission: LocationPermission
  loading: boolean
  error: string | null
}

export interface LocationCoords {
  latitude: number
  longitude: number
  accuracy?: number
}

export interface UseLocationResult extends LocationState {
  /** Ask the browser for a position now. Resolves as the "prompt" path; the
   *  caller decides when to arm it. */
  request: () => void
}

interface CachedLocation {
  latitude: number
  longitude: number
  accuracy: number | null
  permission: LocationPermission
}

/**
 * A successful geolocation position.
 *
 * - `permission_resolved`: the Permissions API answered.
 * - `fetch_start`: `request()` (or the silent mount fetch) began resolving.
 * - `position_success`: the browser returned a fix.
 * - `position_error`: the browser refused or timed out.
 */
export type LocationEvent =
  | { type: 'permission_resolved'; permission: LocationPermission; error?: string }
  | { type: 'fetch_start' }
  | { type: 'position_success'; coords: LocationCoords }
  | { type: 'position_error'; message?: string }

/**
 * Single source of truth for location state.
 *
 * Extracted as a pure reducer so the logic backing `request()` and the mount
 * flow is pinned without a DOM — this repo ships no DOM test environment, so
 * the three behaviour tests exercise this function directly.
 */
export function locationReducer(state: LocationState, event: LocationEvent): LocationState {
  switch (event.type) {
    case 'permission_resolved':
      return {
        ...state,
        permission: event.permission,
        error: event.error ?? null,
        ...(event.permission === 'denied'
          ? { latitude: null, longitude: null, accuracy: null, loading: false }
          : {}),
      }
    case 'fetch_start':
      return { ...state, loading: true, error: null }
    case 'position_success':
      return {
        latitude: event.coords.latitude,
        longitude: event.coords.longitude,
        accuracy: event.coords.accuracy ?? null,
        permission: 'granted',
        loading: false,
        error: null,
      }
    case 'position_error':
      return {
        latitude: null,
        longitude: null,
        accuracy: null,
        permission: 'denied',
        loading: false,
        error: event.message ?? 'Location permission denied.',
      }
    default:
      return state
  }
}

/** Read the last cached fix, or `null` when none is trustworthy. */
export function readLocationCache(): CachedLocation | null {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as CachedLocation
    if (
      typeof parsed.latitude === 'number' &&
      typeof parsed.longitude === 'number' &&
      (parsed.permission === 'granted' || parsed.permission === 'denied')
    ) {
      return parsed
    }
    return null
  } catch {
    return null
  }
}

/** Persist a fix so a route change does not re-prompt. */
export function writeLocationCache(coords: CachedLocation): void {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(coords))
  } catch {
    /* storage unavailable (private mode) — cache stays in memory only */
  }
}

/** Initial state: a previous fix survives a route change so we don't re-prompt. */
export function initialLocationState(): LocationState {
  const cached = readLocationCache()
  if (cached) {
    return {
      latitude: cached.latitude,
      longitude: cached.longitude,
      accuracy: cached.accuracy,
      permission: cached.permission,
      loading: false,
      error: null,
    }
  }
  return {
    latitude: null,
    longitude: null,
    accuracy: null,
    permission: 'unavailable',
    loading: false,
    error: null,
  }
}

const HIGH_ACCURACY_OPTIONS: PositionOptions = {
  enableHighAccuracy: true,
  timeout: 10_000,
  maximumAge: 60_000,
}

function toPermission(state: PermissionState): LocationPermission {
  switch (state) {
    case 'granted':
      return 'granted'
    case 'denied':
      return 'denied'
    case 'prompt':
      return 'prompt'
    default:
      return 'unavailable'
  }
}

/**
 * Cache a fix locally and, only when the reporter has opted into sharing,
 * report it to `POST /location`. The position is cached either way, so the map
 * can still centre on the user even when the network call cannot be made.
 */
async function cacheAndShare(coords: LocationCoords): Promise<void> {
  writeLocationCache({
    latitude: coords.latitude,
    longitude: coords.longitude,
    accuracy: coords.accuracy ?? null,
    permission: 'granted',
  })

  try {
    const { enabled } = await api.get<{ enabled: boolean }>('/location/sharing')
    if (enabled) {
      try {
        await api.post('/location', {
          latitude: coords.latitude,
          longitude: coords.longitude,
          accuracy: coords.accuracy,
        })
      } catch {
        /* non-fatal: the fix is already cached locally */
      }
    }
  } catch {
    /* Not authenticated, offline, or sharing is off — keep the local cache. */
  }
}

/**
 * The reporter's live position.
 *
 * On mount it asks the Permissions API for the geolocation state:
 *  - `granted`   → fetches once silently, so the map centres on the user with no prompt.
 *  - `prompt`    → leaves the choice to the caller (`request()`); no auto-fetch.
 *  - `denied` / `unavailable` → records the state and never asks again from here.
 *
 * A successful fix is cached in `sessionStorage` so a route change does not re-prompt.
 */
export function useLocation(): UseLocationResult {
  const [state, dispatch] = useReducer(locationReducer, undefined, initialLocationState)

  const resolvePosition = useCallback(() => {
    dispatch({ type: 'fetch_start' })
    if (!('geolocation' in navigator)) {
      dispatch({
        type: 'permission_resolved',
        permission: 'unavailable',
        error: 'Geolocation is not supported on this device.',
      })
      return
    }
    navigator.geolocation.getCurrentPosition(
      (position) => {
        const coords: LocationCoords = {
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
          accuracy: position.coords.accuracy,
        }
        dispatch({ type: 'position_success', coords })
        void cacheAndShare(coords)
      },
      (error) => {
        dispatch({
          type: 'position_error',
          message: error?.message ?? 'Location request was denied.',
        })
      },
      HIGH_ACCURACY_OPTIONS,
    )
  }, [])

  useEffect(() => {
    let cancelled = false
    if (!('permissions' in navigator) || !navigator.permissions?.query) {
      // No Permissions API. We can still resolve geolocation directly, but we
      // cannot pre-know the state, so expose "prompt" where the capability exists.
      if ('geolocation' in navigator) {
        dispatch({ type: 'permission_resolved', permission: 'prompt' })
      } else {
        dispatch({
          type: 'permission_resolved',
          permission: 'unavailable',
          error: 'Geolocation is not supported on this device.',
        })
      }
      return
    }

    navigator.permissions
      .query({ name: 'geolocation' })
      .then((status) => {
        if (cancelled) return
        dispatch({ type: 'permission_resolved', permission: toPermission(status.state) })
        if (status.state === 'granted' || status.state === 'prompt') {
          resolvePosition()
        }
        status.onchange = () => {
          if (cancelled) return
          const resolved = toPermission(status.state)
          dispatch({ type: 'permission_resolved', permission: resolved })
          if (resolved === 'granted' || resolved === 'prompt') {
            resolvePosition()
          }
        }
      })
      .catch(() => {
        if (cancelled) return
        if ('geolocation' in navigator) {
          dispatch({ type: 'permission_resolved', permission: 'prompt' })
          resolvePosition()
        } else {
          dispatch({ type: 'permission_resolved', permission: 'unavailable' })
        }
      })

    return () => {
      cancelled = true
    }
  }, [resolvePosition])

  const request = useCallback(() => {
    resolvePosition()
  }, [resolvePosition])

  return { ...state, request }
}
