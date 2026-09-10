import { QueryClient } from '@tanstack/react-query'
import { ApiError } from './apiClient'

/**
 * Shared React Query client.
 *
 * Retries: never retry auth/permission/validation failures; retry twice on
 * network or 5xx. `staleTime` keeps field data fresh for a short window so
 * tab switches don't refetch needlessly.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: false,
      retry: (failureCount, error) => {
        if (error instanceof ApiError) {
          if (error.status >= 400 && error.status < 500) return false
        }
        return failureCount < 2
      },
    },
    mutations: {
      retry: 0,
    },
  },
})
