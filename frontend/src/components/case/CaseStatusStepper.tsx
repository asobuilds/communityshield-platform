import { Check, Dot } from 'lucide-react'
import { cn } from '@/lib/cn'
import { CASE_STATUS_ORDER, statusIndex, statusMeta } from '@/lib/status'
import { formatDateTime, relativeTime } from '@/lib/format'
import type { CaseStatus } from '@/types/api'

/**
 * The case lifecycle rail: pending → assigned → dispatched → on scene → closed.
 *
 * Renders as a labelled rail on wider screens and collapses to a compact
 * progress bar on phones, where an officer is usually holding the device in
 * one hand.
 */
export function CaseStatusStepper({
  status,
  times = {},
  className,
}: {
  status: string | undefined
  /** Timestamp per status, e.g. `{ dispatched: caseItem.dispatchedAt }`. */
  times?: Partial<Record<CaseStatus, string | null | undefined>>
  className?: string
}) {
  const current = statusIndex(status)
  const currentMeta = statusMeta(status)

  return (
    <div className={className}>
      {/* Wide: full rail */}
      <ol className="hidden sm:flex" aria-label="Case progress">
        {CASE_STATUS_ORDER.map((state, index) => {
          const meta = statusMeta(state)
          const done = index < current
          const isCurrent = index === current
          const time = times[state]

          return (
            <li key={state} className="flex flex-1 flex-col gap-1.5">
              <div className="flex items-center">
                <span
                  aria-hidden
                  className={cn(
                    'grid size-6 shrink-0 place-items-center rounded-full ring-1',
                    done && 'bg-ok/15 text-ok ring-ok/40',
                    isCurrent && cn(meta.bg, meta.text, meta.ring),
                    !done && !isCurrent && 'bg-surface-hi text-ink-faint ring-border-hi',
                  )}
                >
                  {done ? (
                    <Check className="size-3.5" />
                  ) : isCurrent ? (
                    <span className={cn('size-2 rounded-full', meta.dot)} />
                  ) : (
                    <Dot className="size-3.5" />
                  )}
                </span>
                {index < CASE_STATUS_ORDER.length - 1 ? (
                  <span
                    aria-hidden
                    className={cn(
                      'mx-1 h-px flex-1',
                      index < current ? 'bg-ok/40' : 'bg-border-hi',
                    )}
                  />
                ) : null}
              </div>

              <div className="pr-2">
                <p
                  className={cn(
                    'text-xs font-medium',
                    isCurrent ? 'text-ink' : done ? 'text-ink-muted' : 'text-ink-faint',
                  )}
                  aria-current={isCurrent ? 'step' : undefined}
                >
                  {meta.label}
                </p>
                <p className="text-[10px] text-ink-faint">
                  {time ? relativeTime(time) : isCurrent ? 'in progress' : '—'}
                </p>
              </div>
            </li>
          )
        })}
      </ol>

      {/* Narrow: compact bar */}
      <div className="sm:hidden">
        <div className="flex items-baseline justify-between">
          <p className="text-sm font-semibold text-ink" aria-current="step">
            {currentMeta.label}
          </p>
          <p className="text-[11px] text-ink-muted tabular-nums">
            Step {current + 1} of {CASE_STATUS_ORDER.length}
          </p>
        </div>
        <div className="mt-2 flex gap-1" aria-hidden>
          {CASE_STATUS_ORDER.map((state, index) => (
            <span
              key={state}
              className={cn(
                'h-1.5 flex-1 rounded-full',
                index <= current ? statusMeta(state).dot : 'bg-surface-hi',
              )}
            />
          ))}
        </div>
        <p className="mt-2 text-[11px] text-ink-muted">{currentMeta.description}</p>
        {times[CASE_STATUS_ORDER[current]] ? (
          <p className="mt-0.5 text-[11px] text-ink-faint">
            Since {formatDateTime(times[CASE_STATUS_ORDER[current]])}
          </p>
        ) : null}
      </div>
    </div>
  )
}
