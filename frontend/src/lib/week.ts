/**
 * ISO-week grouping for the "Weekly updates" interface.
 *
 * The backend has NO weekly-update entity or endpoint (verified: the only
 * "weekly" reference is a finance budget-period enum), so weekly views are
 * derived here, client-side, from the timestamps of progress / timeline items.
 *
 * Weeks are ISO-8601: Monday 00:00 (local) to Sunday 23:59:59.
 */

const DAY_MS = 86_400_000

export interface Dated {
  createdAt: string
}

export interface WeekBucket<T extends Dated> {
  /** Stable key, e.g. "2026-W36". */
  key: string
  /** Monday 00:00 of the week. */
  start: Date
  /** Sunday 23:59:59.999 of the week. */
  end: Date
  label: string
  isCurrent: boolean
  /** Newest first. */
  items: T[]
}

/** Monday 00:00 (local) of the week containing `input`. */
export function startOfIsoWeek(input: Date | string): Date {
  const d = new Date(input)
  d.setHours(0, 0, 0, 0)
  const mondayOffset = (d.getDay() + 6) % 7 // Sun=0 → 6, Mon=1 → 0
  d.setDate(d.getDate() - mondayOffset)
  return d
}

export function isoWeekParts(input: Date | string): { year: number; week: number } {
  // The ISO week's year is the year of its Thursday.
  const thursday = startOfIsoWeek(input)
  thursday.setDate(thursday.getDate() + 3)

  const year = thursday.getFullYear()
  const firstThursday = startOfIsoWeek(new Date(year, 0, 4))
  firstThursday.setDate(firstThursday.getDate() + 3)

  const week = 1 + Math.round((thursday.getTime() - firstThursday.getTime()) / (7 * DAY_MS))
  return { year, week }
}

export function isoWeekKey(input: Date | string): string {
  const { year, week } = isoWeekParts(input)
  return `${year}-W${String(week).padStart(2, '0')}`
}

function formatRange(start: Date): string {
  const end = new Date(start.getTime() + 6 * DAY_MS)
  const sameMonth = start.getMonth() === end.getMonth()
  const startLabel = start.toLocaleDateString(undefined, {
    day: 'numeric',
    ...(sameMonth ? {} : { month: 'short' }),
  })
  const endLabel = end.toLocaleDateString(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
  return `${startLabel} – ${endLabel}`
}

export function weekLabel(start: Date | string, now: Date = new Date()): string {
  const key = isoWeekKey(start)
  if (key === isoWeekKey(now)) return 'This week'

  const lastWeekStart = startOfIsoWeek(new Date(startOfIsoWeek(now).getTime() - 7 * DAY_MS))
  if (key === isoWeekKey(lastWeekStart)) return 'Last week'

  return `Week of ${formatRange(startOfIsoWeek(start))}`
}

/** Number of weeks between `now`'s week and `start`'s week (0 = current, -1 = last). */
export function weeksAgo(start: Date | string, now: Date = new Date()): number {
  const a = startOfIsoWeek(now).getTime()
  const b = startOfIsoWeek(start).getTime()
  return Math.round((b - a) / (7 * DAY_MS))
}

/**
 * Group dated items into ISO weeks, newest week first, and newest item first
 * within each week. Items with unparseable timestamps are skipped.
 */
export function groupByIsoWeek<T extends Dated>(
  items: T[],
  now: Date = new Date(),
): WeekBucket<T>[] {
  const buckets = new Map<string, WeekBucket<T>>()

  for (const item of items) {
    const created = new Date(item.createdAt)
    if (Number.isNaN(created.getTime())) continue

    const start = startOfIsoWeek(created)
    const key = isoWeekKey(created)

    let bucket = buckets.get(key)
    if (!bucket) {
      bucket = {
        key,
        start,
        end: new Date(start.getTime() + 7 * DAY_MS - 1),
        label: weekLabel(start, now),
        isCurrent: key === isoWeekKey(now),
        items: [],
      }
      buckets.set(key, bucket)
    }
    bucket.items.push(item)
  }

  for (const bucket of buckets.values()) {
    bucket.items.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
  }

  return [...buckets.values()].sort((a, b) => b.start.getTime() - a.start.getTime())
}
