import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/apiClient'
import { caseKeys } from './useCases'
import type {
  AssignCaseResponse,
  CaseAssignmentsResponse,
  UnitOfficersResponse,
} from '@/types/api'

export const officerKeys = {
  byUnit: (unitId: string) => ['officers', 'unit', unitId] as const,
  assignments: (caseId: string) => ['officers', 'assignments', caseId] as const,
}

/**
 * The roster of a unit, used to pick someone to assign.
 *
 * CONTRACT GAP: the backend implements `GetOfficersByUnit` in
 * `handlers/officers_handler.go` but never registers it in `routes/routes.go`,
 * so against the live API this 404s. `retry: false` is deliberate — a 404 here
 * is a permanent contract gap, not a transient blip — and callers are expected
 * to treat a 404 as "no roster available" rather than as a failure
 * (see components/admin/AssignOfficerDialog.tsx).
 */
export function useUnitOfficers(unitId: string | undefined) {
  return useQuery({
    queryKey: officerKeys.byUnit(unitId ?? ''),
    queryFn: () => api.get<UnitOfficersResponse>(`/units/${unitId}/officers`),
    enabled: Boolean(unitId),
    select: (data) => data.officers,
    retry: false,
  })
}

/** Officers attached to a case (`GET /cases/:id/assignments`). */
export function useCaseAssignments(caseId: string | undefined) {
  return useQuery({
    queryKey: officerKeys.assignments(caseId ?? ''),
    queryFn: () => api.get<CaseAssignmentsResponse>(`/cases/${caseId}/assignments`),
    enabled: Boolean(caseId),
    select: (data) => data.assignments,
    retry: false,
  })
}

export interface AssignOfficerInput {
  officerId: string
  role?: string
}

/**
 * Assign or reassign an officer. Mirrors `POST /cases/:id/assign`, which is the
 * only transition that moves a case out of `pending`.
 */
export function useAssignOfficer(caseId: string | undefined) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: AssignOfficerInput) =>
      api.post<AssignCaseResponse>(`/cases/${caseId}/assign`, input),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: caseKeys.detail(caseId ?? '') }),
        queryClient.invalidateQueries({ queryKey: caseKeys.list() }),
        queryClient.invalidateQueries({ queryKey: officerKeys.assignments(caseId ?? '') }),
      ])
    },
  })
}
