import type { CaseStatus } from '@/types/api'

/* Case lifecycle metadata — the single source of truth for how a status is
 * labelled and coloured. Colours come from the Tailwind v4 tokens in index.css
 * (--color-status-*). Keep class strings literal so Tailwind can see them. */

export const CASE_STATUS_ORDER: CaseStatus[] = [
  'pending',
  'assigned',
  'dispatched',
  'on_scene',
  'closed',
]

export interface StatusMeta {
  label: string
  /** Plain-language meaning, shown to explain a state to citizens/officers. */
  description: string
  dot: string
  text: string
  bg: string
  ring: string
}

export const CASE_STATUS_META: Record<CaseStatus, StatusMeta> = {
  pending: {
    label: 'Pending',
    description: 'Reported and awaiting triage or assignment.',
    dot: 'bg-status-pending',
    text: 'text-status-pending',
    bg: 'bg-status-pending/10',
    ring: 'ring-status-pending/30',
  },
  assigned: {
    label: 'Assigned',
    description: 'An officer has been assigned and will be dispatched.',
    dot: 'bg-status-assigned',
    text: 'text-status-assigned',
    bg: 'bg-status-assigned/10',
    ring: 'ring-status-assigned/30',
  },
  dispatched: {
    label: 'Dispatched',
    description: 'The assigned officer is en route.',
    dot: 'bg-status-dispatched',
    text: 'text-status-dispatched',
    bg: 'bg-status-dispatched/10',
    ring: 'ring-status-dispatched/30',
  },
  on_scene: {
    label: 'On scene',
    description: 'The officer has arrived and is working the case.',
    dot: 'bg-status-onscene',
    text: 'text-status-onscene',
    bg: 'bg-status-onscene/10',
    ring: 'ring-status-onscene/30',
  },
  closed: {
    label: 'Closed',
    description: 'The case has been resolved and closed with a final report.',
    dot: 'bg-status-closed',
    text: 'text-status-closed',
    bg: 'bg-status-closed/10',
    ring: 'ring-status-closed/30',
  },
}

export function statusMeta(status: string | undefined): StatusMeta {
  return CASE_STATUS_META[(status as CaseStatus) ?? 'pending'] ?? CASE_STATUS_META.pending
}

export function statusIndex(status: string | undefined): number {
  const i = CASE_STATUS_ORDER.indexOf((status as CaseStatus) ?? 'pending')
  return i < 0 ? 0 : i
}

/* Priority bands (backend PriorityLevel: P1 critical, P2 high, P3 routine). */

export interface PriorityMeta {
  label: string
  text: string
  bg: string
  ring: string
}

export const PRIORITY_META: Record<string, PriorityMeta> = {
  P1: { label: 'P1 · Critical', text: 'text-emergency', bg: 'bg-emergency/10', ring: 'ring-emergency/40' },
  P2: { label: 'P2 · High', text: 'text-warn', bg: 'bg-warn/10', ring: 'ring-warn/40' },
  P3: { label: 'P3 · Routine', text: 'text-ink-muted', bg: 'bg-ink-muted/10', ring: 'ring-border-hi' },
}

export function priorityMeta(level: string | undefined): PriorityMeta {
  return PRIORITY_META[level ?? 'P3'] ?? PRIORITY_META.P3
}

/* The single next operational action available from a status, mirroring the
 * backend's transition guards (see case_workflow_handler.go). */

export type CaseAction = 'assign' | 'dispatch' | 'arrive' | 'close'

export const CASE_ACTION_META: Record<CaseAction, { label: string; path: string }> = {
  assign: { label: 'Assign officer', path: 'assign' },
  dispatch: { label: 'Dispatch', path: 'dispatch' },
  arrive: { label: 'Mark on scene', path: 'arrive' },
  close: { label: 'Close case', path: 'close' },
}

/** Whether progress may be added from this status (backend allows dispatched | on_scene). */
export function canAddProgress(status: string | undefined): boolean {
  return status === 'dispatched' || status === 'on_scene'
}
