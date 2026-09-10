import type { CaseTimelineEntry, Progress } from '@/types/api'

/**
 * A single dated thing that happened on a case, normalised from either the
 * progress feed (`/cases/:id/progress`, written by officers) or the audit
 * timeline (`/cases/:id/timeline`, written by the system).
 */
export interface ActivityItem {
  id: string
  createdAt: string
  action: string
  description?: string
  status?: string
  source: 'progress' | 'timeline'
}

export function progressToActivity(items: Progress[]): ActivityItem[] {
  return items.map((p) => ({
    id: `progress-${p.id}`,
    createdAt: p.createdAt,
    action: p.action,
    description: p.description,
    source: 'progress',
  }))
}

export function timelineToActivity(items: CaseTimelineEntry[]): ActivityItem[] {
  return items.map((t) => ({
    id: `timeline-${t.id}`,
    createdAt: t.createdAt,
    action: t.action,
    description: t.description,
    status: t.status,
    source: 'timeline',
  }))
}

/** Both feeds merged, newest first. */
export function mergeActivity(
  progress: Progress[],
  timeline: CaseTimelineEntry[],
): ActivityItem[] {
  return [...progressToActivity(progress), ...timelineToActivity(timeline)].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
  )
}

const ACTION_LABELS: Record<string, string> = {
  case_created: 'Case reported',
  case_assigned: 'Officer assigned',
  officer_assigned: 'Officer assigned',
  dispatched: 'Officer dispatched',
  arrived: 'Officer on scene',
  on_scene: 'Officer on scene',
  closed: 'Case closed',
  status_updated: 'Status updated',
  evidence_added: 'Evidence added',
  evidence_uploaded: 'Evidence added',
  evidence_verified: 'Evidence verified',
  note: 'Note',
  update: 'Update',
  progress: 'Progress update',
  weekly_summary: 'Weekly summary',
}

/** Action code used when an officer files the week's narrative summary. */
export const WEEKLY_SUMMARY_ACTION = 'weekly_summary'

/** `on_scene` → "Officer on scene"; unknown codes fall back to Title Case. */
export function humanizeAction(action: string | undefined): string {
  if (!action) return 'Update'
  const known = ACTION_LABELS[action]
  if (known) return known
  return action
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/^\w/, (c) => c.toUpperCase())
}
