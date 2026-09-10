import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/apiClient'
import type { Case, CaseDetailResponse, CasesResponse } from '@/types/api'

export const caseKeys = {
  all: ['cases'] as const,
  list: () => [...caseKeys.all, 'list'] as const,
  detail: (id: string) => [...caseKeys.all, 'detail', id] as const,
}

/** All cases visible to the current user (the backend scopes by role). */
export function useCases() {
  return useQuery({
    queryKey: caseKeys.list(),
    queryFn: () => api.get<CasesResponse>('/cases'),
    select: (data) => data.cases,
  })
}

/** A single case with its timeline and feedback. */
export function useCaseDetail(id: string | undefined) {
  return useQuery({
    queryKey: caseKeys.detail(id ?? ''),
    queryFn: () => api.get<CaseDetailResponse>(`/cases/${id}`),
    enabled: Boolean(id),
  })
}

type CaseActionResponse = { message: string; case: Case }

async function runAction(
  id: string,
  action: 'dispatch' | 'arrive' | 'close' | 'assign',
  body?: unknown,
): Promise<CaseActionResponse> {
  return api.post<CaseActionResponse>(`/cases/${id}/${action}`, body)
}

/**
 * Case lifecycle mutations. Each invalidates the case detail + list so the
 * stepper, timeline and queue reflect the new state.
 */
export function useCaseActions(id: string | undefined) {
  const queryClient = useQueryClient()

  const invalidate = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: caseKeys.detail(id ?? '') }),
      queryClient.invalidateQueries({ queryKey: caseKeys.list() }),
    ])
  }

  const dispatch = useMutation({
    mutationFn: () => runAction(id as string, 'dispatch'),
    onSuccess: invalidate,
  })
  const arrive = useMutation({
    mutationFn: () => runAction(id as string, 'arrive'),
    onSuccess: invalidate,
  })
  const close = useMutation({
    mutationFn: (finalReport: string) => runAction(id as string, 'close', { finalReport }),
    onSuccess: invalidate,
  })
  const assign = useMutation({
    mutationFn: (vars: { officerId: string; role?: string }) =>
      runAction(id as string, 'assign', vars),
    onSuccess: invalidate,
  })

  return { dispatch, arrive, close, assign }
}
