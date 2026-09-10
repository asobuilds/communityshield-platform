import { useMemo, useState } from 'react'
import { CalendarRange, ChevronDown, Sparkles } from 'lucide-react'
import { cn } from '@/lib/cn'
import { formatDateTime, relativeTime } from '@/lib/format'
import { groupByIsoWeek, isoWeekKey, startOfIsoWeek } from '@/lib/week'
import { Badge } from '@/components/ui/Chips'
import { Button } from '@/components/ui/Button'
import { Field, Textarea } from '@/components/ui/Field'
import { EmptyState } from '@/components/ui/States'
import { humanizeAction, WEEKLY_SUMMARY_ACTION, type ActivityItem } from './activity'
import type { CaseTimelineEntry, Progress } from '@/types/api'

/**
 * Weekly updates.
 *
 * The backend has no weekly-update entity, so this view is *derived*: it groups
 * the case's real progress + timeline timestamps into ISO weeks. Officers can
 * also file a narrative summary for the current week, which is stored as a
 * progress entry with the `weekly_summary` action — using an endpoint that
 * already exists rather than inventing a new one.
 */
export function WeeklyUpdates({
  progress,
  timeline,
  onSubmitSummary,
  submitting = false,
  className,
}: {
  progress: Progress[]
  timeline: CaseTimelineEntry[]
  onSubmitSummary?: (description: string) => void
  submitting?: boolean
  className?: string
}) {
  const [draft, setDraft] = useState('')
  const [expanded, setExpanded] = useState<Record<string, boolean>>({})

  const activity = useMemo<ActivityItem[]>(() => {
    const merged: ActivityItem[] = [
      ...progress.map((p) => ({
        id: `progress-${p.id}`,
        createdAt: p.createdAt,
        action: p.action,
        description: p.description,
        source: 'progress' as const,
      })),
      ...timeline.map((t) => ({
        id: `timeline-${t.id}`,
        createdAt: t.createdAt,
        action: t.action,
        description: t.description,
        status: t.status,
        source: 'timeline' as const,
      })),
    ]
    return merged.sort(
      (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
    )
  }, [progress, timeline])

  const weeks = useMemo(() => groupByIsoWeek(activity), [activity])

  const summaryByWeek = useMemo(() => {
    const map = new Map<string, string>()
    for (const p of progress) {
      if (p.action !== WEEKLY_SUMMARY_ACTION) continue
      map.set(isoWeekKey(p.createdAt), p.description ?? '')
    }
    return map
  }, [progress])

  if (weeks.length === 0) {
    return (
      <EmptyState
        icon={<CalendarRange className="size-5" aria-hidden />}
        title="Nothing recorded yet this case"
        description="Once officers file progress updates or the case changes status, each week's activity will be grouped here."
      />
    )
  }

  const currentBucket = weeks.find((w) => w.isCurrent)

  return (
    <div className={cn('flex flex-col gap-4 p-4', className)}>
      {/* Current-week digest */}
      <section className="rounded-panel border border-signal/25 bg-signal/5 p-3">
        <div className="flex items-center gap-2">
          <Sparkles className="size-4 text-signal" aria-hidden />
          <h3 className="text-sm font-semibold text-ink">This week at a glance</h3>
        </div>
        <dl className="mt-2 grid grid-cols-3 gap-3">
          <Stat label="Updates logged" value={String(currentBucket?.items.length ?? 0)} />
          <Stat
            label="Weeks on record"
            value={String(weeks.length)}
          />
          <Stat
            label="Latest activity"
            value={currentBucket?.items[0] ? relativeTime(currentBucket.items[0].createdAt) : '—'}
          />
        </dl>
      </section>

      {/* Officer's narrative summary for the current week */}
      {onSubmitSummary ? (
        <section className="rounded-panel border border-border bg-surface-hi p-3">
          <h3 className="text-sm font-semibold text-ink">File this week's summary</h3>
          <p className="mt-0.5 text-xs text-ink-muted">
            A short narrative for the week — what was done, what is blocked, what happens next.
            Stored on the case as a weekly summary update.
          </p>
          <form
            className="mt-3 flex flex-col gap-3"
            onSubmit={(event) => {
              event.preventDefault()
              const value = draft.trim()
              if (!value) return
              onSubmitSummary(value)
              setDraft('')
            }}
          >
            <Field label="Summary" hint="Visible to your unit and the reviewing admin.">
              {(props) => (
                <Textarea
                  {...props}
                  rows={3}
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                  placeholder="e.g. Visited the market and interviewed three traders; CCTV pulled; awaiting analysis."
                />
              )}
            </Field>
            <div className="flex justify-end">
              <Button
                type="submit"
                variant="primary"
                size="sm"
                loading={submitting}
                disabled={!draft.trim()}
              >
                File summary
              </Button>
            </div>
          </form>
        </section>
      ) : null}

      {/* Week-by-week record */}
      <div className="flex flex-col gap-3">
        {weeks.map((week) => {
          const isOpen = expanded[week.key] ?? week.isCurrent
          const summary = summaryByWeek.get(week.key)
          const statusChanges = [
            ...new Set(week.items.map((i) => i.status).filter(Boolean) as string[]),
          ]

          return (
            <section key={week.key} className="overflow-hidden rounded-panel border border-border">
              <button
                type="button"
                aria-expanded={isOpen}
                onClick={() => setExpanded((prev) => ({ ...prev, [week.key]: !isOpen }))}
                className="flex w-full items-center justify-between gap-3 bg-surface px-3 py-2.5 text-left transition-colors hover:bg-surface-hi"
              >
                <div className="min-w-0">
                  <p className="flex items-center gap-2 text-sm font-medium text-ink">
                    {week.label}
                    {week.isCurrent ? <Badge tone="signal">Current</Badge> : null}
                    <span className="text-[11px] font-normal text-ink-faint">{week.key}</span>
                  </p>
                  <p className="mt-0.5 text-[11px] text-ink-muted">
                    {week.items.length} update{week.items.length === 1 ? '' : 's'}
                    {statusChanges.length > 0
                      ? ` · status: ${statusChanges.join(' → ')}`
                      : ''}
                  </p>
                </div>
                <ChevronDown
                  className={cn('size-4 shrink-0 text-ink-muted transition-transform', isOpen && 'rotate-180')}
                  aria-hidden
                />
              </button>

              {isOpen ? (
                <div className="border-t border-border bg-base">
                  {summary ? (
                    <div className="border-b border-border bg-surface-hi px-3 py-2.5">
                      <p className="text-[11px] font-semibold uppercase tracking-wide text-signal">
                        Weekly summary
                      </p>
                      <p className="mt-1 whitespace-pre-line text-sm text-ink">{summary}</p>
                    </div>
                  ) : null}

                  <ul className="divide-y divide-border">
                    {week.items.map((item) => (
                      <li key={item.id} className="flex gap-3 px-3 py-2.5">
                        <span
                          className={cn(
                            'mt-1.5 size-1.5 shrink-0 rounded-full',
                            item.source === 'progress' ? 'bg-signal' : 'bg-border-hi',
                          )}
                          aria-hidden
                        />
                        <div className="min-w-0 flex-1">
                          <div className="flex flex-wrap items-baseline justify-between gap-x-3">
                            <p className="text-sm text-ink">
                              {humanizeAction(item.action)}
                              <span className="ml-2 text-[10px] uppercase tracking-wide text-ink-faint">
                                {item.source === 'progress' ? 'officer note' : 'system'}
                              </span>
                            </p>
                            <time
                              dateTime={item.createdAt}
                              title={formatDateTime(item.createdAt)}
                              className="text-[11px] text-ink-faint"
                            >
                              {relativeTime(item.createdAt)}
                            </time>
                          </div>
                          {item.description ? (
                            <p className="mt-0.5 text-xs text-ink-muted">{item.description}</p>
                          ) : null}
                        </div>
                      </li>
                    ))}
                  </ul>
                </div>
              ) : null}
            </section>
          )
        })}
      </div>

      <p className="text-[11px] text-ink-faint">
        Weeks run Monday–Sunday (ISO 8601). Grouping is derived from update timestamps; the
        current week starts {formatDateTime(startOfIsoWeek(new Date()))}.
      </p>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-[10px] uppercase tracking-wide text-ink-faint">{label}</dt>
      <dd className="mt-0.5 text-base font-semibold text-ink tabular-nums">{value}</dd>
    </div>
  )
}
