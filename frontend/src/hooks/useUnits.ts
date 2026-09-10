import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/apiClient'
import type { NearbyUnitsResponse, SecurityUnit, UnitsResponse } from '@/types/api'

const unitKeys = {
  all: ['units'] as const,
  nearby: (lat: number, lng: number, radius: number) =>
    ['units', 'nearby', lat, lng, radius] as const,
}

/** All active security units (public endpoint, no auth required). */
export function useUnits() {
  return useQuery({
    queryKey: unitKeys.all,
    queryFn: () => api.get<UnitsResponse>('/units'),
    select: (data) => data.units,
    staleTime: 5 * 60_000,
  })
}

/**
 * Units near a coordinate, enriched with `distance` and `isInRange`.
 * Disabled until a coordinate is supplied.
 */
export function useNearbyUnits(
  lat: number | undefined,
  lng: number | undefined,
  radiusKm = 50,
) {
  return useQuery({
    queryKey: unitKeys.nearby(lat ?? 0, lng ?? 0, radiusKm),
    queryFn: () =>
      api.get<NearbyUnitsResponse>(`/units/nearby?lat=${lat}&lng=${lng}&radius=${radiusKm}`),
    enabled: typeof lat === 'number' && typeof lng === 'number',
    select: (data): SecurityUnit[] => data.units,
  })
}
