import { Link } from 'react-router-dom'
import { FolderSearch, MapPin } from 'lucide-react'
import { Card, CardBody, CardHeader } from '@/components/ui/Card'
import { StatusChip } from '@/components/ui/Chips'
import { EmptyState, ErrorState, Skeleton } from '@/components/ui/States'
import { CaseStatusStepper } from '@/components/case/CaseStatusStepper'
import { useCases } from '@/hooks/useCases'
import { relativeTime } from '@/lib/format'

/**
 * Citizen home: the reports *this* citizen filed.
 *
 * `GET /cases` is scoped server-side by role, so a citizen only ever receives
 * their own cases — this page renders that list. Case detail for citizens
 * (full timeline, feedback, ratings) is the next milestone.
 */
export function CitizenHomePage() {
  const { data, isLoading, isError, refetch } = useCases()

  return (
    <div className="mx-auto w-full max-w-3xl p-4 sm:p-6">
      <header className="mb-4">
        <h1 className="text-xl font-semibold text-ink">Your reports</h1>
        <p className="mt-1 text-sm text-ink-muted">
          Track what you have reported and where it has got to.
        </p>
      </header>

      {isLoading ? (
        <div className="flex flex-col gap-3">
          {[0, 1].map((i) => (
            <Card key={i}>
              <CardBody className="space-y-3">
                <Skeleton className="h-4 w-1/2" />
                <Skeleton className="h-3 w-3/4" />
                <Skeleton className="h-6 w-full" />
              </CardBody>
            </Card>
          ))}
        </div>
      ) : isError ? (
        <Card>
          <ErrorState
            title="Could not load your reports"
            description="Check your connection and try again."
            onRetry={() => void refetch()}
          />
        </Card>
      ) : !data || data.length === 0 ? (
        <Card>
          <EmptyState
            icon={<FolderSearch className="size-5" aria-hidden />}
            title="You haven't reported anything yet"
            description="When you report an incident, it will appear here with live status updates from the responding unit."
          />
        </Card>
      ) : (
        <ul className="flex flex-col gap-3">
          {data.map((caseItem) => (
            <li key={caseItem.id}>
              <Card as="article">
                <CardHeader
                  title={caseItem.title}
                  subtitle={
                    <span className="tabular-nums">
                      {caseItem.trackingId} · reported {relativeTime(caseItem.createdAt)}
                    </span>
                  }
                  actions={<StatusChip status={caseItem.status} />}
                />
                <CardBody className="flex flex-col gap-4">
                  <p className="text-sm text-ink-muted">{caseItem.description}</p>
                  <p className="flex items-center gap-1.5 text-xs text-ink-faint">
                    <MapPin className="size-3.5" aria-hidden />
                    {caseItem.location || 'Location recorded'}
                  </p>
                  <CaseStatusStepper status={caseItem.status} />
                </CardBody>
              </Card>
            </li>
          ))}
        </ul>
      )}

      <p className="mt-6 text-center text-[11px] text-ink-faint">
        Need to see incidents near you?{' '}
        <Link to="/map" className="text-signal underline-offset-2 hover:underline">
          Open the safety map
        </Link>
        .
      </p>
    </div>
  )
}
