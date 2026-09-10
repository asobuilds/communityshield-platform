import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  CalendarRange,
  FileText,
  Image as ImageIcon,
  Info,
  NotebookPen,
  TriangleAlert,
  UserPlus,
  Users,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Chips'
import { TabPanel, Tabs } from '@/components/ui/Tabs'
import { useToast } from '@/components/ui/Toast'
import { ErrorState, OfflineBanner, Skeleton } from '@/components/ui/States'
import { CaseHeader } from '@/components/case/CaseHeader'
import { CaseFacts } from '@/components/case/CaseFacts'
import { ProgressTimeline } from '@/components/case/ProgressTimeline'
import { WeeklyUpdates } from '@/components/case/WeeklyUpdates'
import { EvidenceGallery } from '@/components/case/EvidenceGallery'
import { AssignOfficerDialog } from '@/components/admin/AssignOfficerDialog'
import { useCaseDetail } from '@/hooks/useCases'
import { useCaseProgress } from '@/hooks/useProgress'
import { useCaseEvidence, useVerifyEvidence } from '@/hooks/useEvidence'
import { useAssignOfficer, useCaseAssignments, useUnitOfficers } from '@/hooks/useOfficers'
import { ApiError } from '@/lib/apiClient'
import { roleLabel } from '@/lib/assignment'
import { relativeTime } from '@/lib/format'

type TabId = 'overview' | 'progress' | 'weekly' | 'evidence'

/**
 * The administrator's view of a single case.
 *
 * Built for a decision, not for a browse: who is on this case, is anyone at all,
 * and does the record support what was done. It reads the same case facts as the
 * officer workspace (one shared component), and the only lifecycle action it
 * owns is assignment — the transition officers cannot perform for themselves.
 */
export function AdminCaseReviewPage() {
  const { id } = useParams<{ id: string }>()
  const { notify } = useToast()

  const detail = useCaseDetail(id)
  const progressQuery = useCaseProgress(id)
  const evidenceQuery = useCaseEvidence(id)
  const assignmentsQuery = useCaseAssignments(id)
  const assign = useAssignOfficer(id)
  const verifyEvidence = useVerifyEvidence(id)

  const caseItem = detail.data?.case
  const directory = useUnitOfficers(caseItem?.unitId)
  // Stable identity while the roster loads — see AdminCaseQueuePage.
  const officers = useMemo(() => directory.data ?? [], [directory.data])

  const [tab, setTab] = useState<TabId>('overview')
  const [assignOpen, setAssignOpen] = useState(false)

  const progress = progressQuery.data ?? []
  const evidence = evidenceQuery.data ?? []
  const assignments = assignmentsQuery.data ?? []

  const officerById = useMemo(() => new Map(officers.map((o) => [o.id, o])), [officers])

  const assignedOfficerName = useMemo(() => {
    if (!caseItem?.assignedTo) return undefined
    const officer = officerById.get(caseItem.assignedTo)
    if (officer) return `${officer.rank} ${officer.name}`
    const entry = detail.data?.timeline.find((t) => t.user && t.user.id === caseItem.assignedTo)
    if (entry?.user) return `${entry.user.firstName} ${entry.user.lastName}`.trim()
    return undefined
  }, [caseItem, officerById, detail.data])

  const networkIssue = ApiError.isNetwork(detail.error)

  const tabs = useMemo(
    () => [
      { id: 'overview', label: 'Overview', icon: <FileText className="size-4" aria-hidden /> },
      {
        id: 'progress',
        label: 'Progress',
        count: progress.length,
        icon: <NotebookPen className="size-4" aria-hidden />,
      },
      { id: 'weekly', label: 'Weekly', icon: <CalendarRange className="size-4" aria-hidden /> },
      {
        id: 'evidence',
        label: 'Evidence',
        count: evidence.length,
        icon: <ImageIcon className="size-4" aria-hidden />,
      },
    ],
    [progress.length, evidence.length],
  )

  if (detail.isLoading) {
    return (
      <div className="mx-auto w-full max-w-4xl space-y-4 p-4 sm:p-6">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-64 w-full rounded-panel" />
      </div>
    )
  }

  if (detail.isError || !caseItem) {
    return (
      <div className="mx-auto w-full max-w-4xl p-4 sm:p-6">
        <BackLink />
        <Card className="mt-4">
          <ErrorState
            title={networkIssue ? 'You are offline' : 'Case unavailable'}
            description={
              networkIssue
                ? 'Reconnect to load this case.'
                : 'This case does not exist, or your unit does not have access to it.'
            }
            offline={networkIssue}
            onRetry={() => void detail.refetch()}
          />
        </Card>
      </div>
    )
  }

  const unassigned = !caseItem.assignedTo && caseItem.status !== 'closed'

  return (
    <div className="mx-auto w-full max-w-4xl p-4 sm:p-6">
      <BackLink />

      {networkIssue ? (
        <div className="mb-3">
          <OfflineBanner />
        </div>
      ) : null}

      <div className="mt-3 flex flex-col gap-4">
        <CaseHeader
          caseItem={caseItem}
          assignedOfficerName={assignedOfficerName}
          actions={
            <Button
              variant={unassigned ? 'primary' : 'secondary'}
              icon={<UserPlus className="size-4" aria-hidden />}
              onClick={() => setAssignOpen(true)}
            >
              {unassigned ? 'Assign officer' : 'Reassign'}
            </Button>
          }
        />

        {unassigned ? (
          <div className="flex flex-wrap items-center justify-between gap-3 rounded-panel border border-warn/40 bg-warn/10 p-3">
            <p className="flex items-start gap-2 text-sm text-warn">
              <TriangleAlert className="mt-0.5 size-4 shrink-0" aria-hidden />
              This case has no officer assigned, so no one can dispatch it. Assigning an officer
              moves it from <span className="font-medium">pending</span> to{' '}
              <span className="font-medium">assigned</span>.
            </p>
            <Button
              size="sm"
              variant="primary"
              icon={<UserPlus className="size-4" aria-hidden />}
              onClick={() => setAssignOpen(true)}
            >
              Assign officer
            </Button>
          </div>
        ) : null}

        <AssignmentPanel
          assignments={assignments}
          state={
            assignmentsQuery.isLoading
              ? 'loading'
              : assignmentsQuery.isError
                ? assignmentsQuery.error instanceof ApiError &&
                  assignmentsQuery.error.status === 404
                  ? 'missing'
                  : 'failed'
                : 'ready'
          }
          officerName={(officerId) => {
            const officer = officerById.get(officerId)
            return officer ? `${officer.rank} ${officer.name} · ${officer.badgeNumber}` : undefined
          }}
          onAssign={() => setAssignOpen(true)}
        />

        <Card>
          <Tabs items={tabs} activeId={tab} onChange={(next) => setTab(next as TabId)} />

          {tab === 'overview' ? (
            <TabPanel id="overview">
              <CaseFacts
                caseItem={caseItem}
                timeline={detail.data?.timeline ?? []}
                feedback={detail.data?.feedback ?? []}
              />
            </TabPanel>
          ) : null}

          {tab === 'progress' ? (
            <TabPanel id="progress">
              {progressQuery.isLoading ? (
                <ProgressTimeline progress={[]} isLoading />
              ) : progressQuery.isError ? (
                <ErrorState
                  title="Could not load progress"
                  description="The updates for this case did not load."
                  onRetry={() => void progressQuery.refetch()}
                />
              ) : (
                <ProgressTimeline progress={progress} />
              )}
            </TabPanel>
          ) : null}

          {tab === 'weekly' ? (
            <TabPanel id="weekly">
              {/* Read-only for administrators: filing the weekly narrative is the
                  responding officer's job, and the endpoint is gated to it. */}
              <WeeklyUpdates progress={progress} timeline={detail.data?.timeline ?? []} />
            </TabPanel>
          ) : null}

          {tab === 'evidence' ? (
            <TabPanel id="evidence">
              {evidenceQuery.isLoading ? (
                <EvidenceGallery evidence={[]} isLoading />
              ) : evidenceQuery.isError ? (
                <ErrorState
                  title="Could not load evidence"
                  description="The evidence for this case did not load."
                  onRetry={() => void evidenceQuery.refetch()}
                />
              ) : (
                <EvidenceGallery
                  evidence={evidence}
                  canVerify={caseItem.status !== 'closed'}
                  verifyingId={verifyEvidence.isPending ? verifyEvidence.variables : null}
                  onVerify={(evidenceId) =>
                    verifyEvidence.mutate(evidenceId, {
                      onSuccess: () => notify('Evidence verified.', 'success'),
                      onError: (cause) =>
                        notify(
                          cause instanceof ApiError
                            ? cause.message
                            : 'Could not verify that item.',
                          'error',
                        ),
                    })
                  }
                />
              )}
            </TabPanel>
          ) : null}
        </Card>
      </div>

      {assignOpen ? (
        <AssignOfficerDialog
          caseItem={caseItem}
          assignments={assignments}
          submitting={assign.isPending}
          onClose={() => setAssignOpen(false)}
          onSubmit={({ officerId, role }) =>
            assign.mutate(
              { officerId, role },
              {
                onSuccess: () => {
                  setAssignOpen(false)
                  notify('Officer assigned.', 'success')
                },
                onError: (cause) =>
                  notify(
                    ApiError.isNetwork(cause)
                      ? 'No connection — the assignment was not saved.'
                      : cause instanceof ApiError
                        ? cause.message
                        : 'Could not assign that officer.',
                    'error',
                  ),
              },
            )
          }
        />
      ) : null}
    </div>
  )
}

/**
 * Who is on the case.
 *
 * `missing` is distinct from `failed` on purpose: the assignments route is part
 * of the API, so a 404 means this build genuinely does not expose it, whereas a
 * network failure is temporary. Conflating the two would let a dropped
 * connection read as "nobody is assigned".
 */
function AssignmentPanel({
  assignments,
  state,
  officerName,
  onAssign,
}: {
  assignments: { id: string; officerId: string; role: string; createdAt?: string }[]
  state: 'loading' | 'ready' | 'missing' | 'failed'
  officerName: (officerId: string) => string | undefined
  onAssign: () => void
}) {
  return (
    <Card>
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border px-4 py-3">
        <h2 className="flex items-center gap-2 text-sm font-semibold text-ink">
          <Users className="size-4 text-ink-muted" aria-hidden />
          Assigned officers
        </h2>
        <Button size="sm" variant="secondary" onClick={onAssign}>
          Manage
        </Button>
      </div>

      <div className="p-4">
        {state === 'loading' ? (
          <div className="flex flex-col gap-2">
            {[0, 1].map((i) => (
              <Skeleton key={i} className="h-10 w-full rounded-lg" />
            ))}
          </div>
        ) : state === 'missing' ? (
          <p className="flex items-start gap-2 text-xs text-ink-muted">
            <Info className="mt-0.5 size-3.5 shrink-0 text-signal" aria-hidden />
            This API build does not expose a case-assignment list, so the officers attached to this
            case cannot be listed here. The case log below records each assignment as it happened.
          </p>
        ) : state === 'failed' ? (
          <p className="flex items-start gap-2 text-xs text-warn">
            <Info className="mt-0.5 size-3.5 shrink-0" aria-hidden />
            The assignment list could not be loaded. This does not mean the case is unassigned — use
            the case log below to see what is on record.
          </p>
        ) : assignments.length === 0 ? (
          <p className="text-xs text-ink-muted">
            No officer is attached to this case yet. Assigning one is what moves it out of pending.
          </p>
        ) : (
          <ul className="flex flex-col gap-2">
            {assignments.map((assignment) => (
              <li
                key={assignment.id}
                className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-border bg-surface-hi px-3 py-2"
              >
                <span className="min-w-0 text-sm text-ink">
                  {officerName(assignment.officerId) ?? shortId(assignment.officerId)}
                </span>
                <span className="flex items-center gap-2">
                  {assignment.createdAt ? (
                    <span className="text-[11px] text-ink-faint">
                      since {relativeTime(assignment.createdAt)}
                    </span>
                  ) : null}
                  <Badge tone={assignment.role === 'primary' ? 'signal' : 'neutral'}>
                    {roleLabel(assignment.role)}
                  </Badge>
                </span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </Card>
  )
}

function BackLink() {
  return (
    <Link
      to="/admin/cases"
      className="inline-flex items-center gap-1.5 text-xs text-ink-muted transition-colors hover:text-ink"
    >
      <ArrowLeft className="size-3.5" aria-hidden />
      Back to case review
    </Link>
  )
}

function shortId(id: string | undefined | null): string {
  if (!id) return '—'
  return id.length > 10 ? `${id.slice(0, 8)}…` : id
}
