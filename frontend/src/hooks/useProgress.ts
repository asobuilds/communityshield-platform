import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/apiClient'
import { caseKeys } from './useCases'
import type { Progress, ProgressResponse } from '@/types/api'

export const progressKeys = {
  forCase: (caseId: string) => ['progress', caseId] as const,
}

export function useCaseProgress(caseId: string | undefined) {
  return useQuery({
    queryKey: progressKeys.forCase(caseId ?? ''),
    queryFn: () => api.get<ProgressResponse>(`/cases/${caseId}/progress`),
    enabled: Boolean(caseId),
    select: (data) => data.progress,
  })
}

/**
 * Add a progress update. The backend only accepts this while the case is
 * `dispatched` or `on_scene`; the UI gates the form on `canAddProgress`.
 */
export function useAddProgress(caseId: string | undefined) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: { action: string; description?: string }) =>
      api.post<{ message: string; progress: Progress }>(`/cases/${caseId}/progress`, input),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: progressKeys.forCase(caseId ?? '') }),
        queryClient.invalidateQueries({ queryKey: caseKeys.detail(caseId ?? '') }),
      ])
    },
  })
}
