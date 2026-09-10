import { useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  ArrowLeft,
  BellRing,
  CalendarRange,
  CheckCircle2,
  FileText,
  Image as ImageIcon,
  Info,
  MapPin,
  NotebookPen,
  Paperclip,
} from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Chips'
import { Field, Input, Select, Textarea } from '@/components/ui/Field'
import { Modal } from '@/components/ui/Modal'
import { TabPanel, Tabs } from '@/components/ui/Tabs'
import { useToast } from '@/components/ui/Toast'
import { ErrorState, OfflineBanner, Skeleton } from '@/components/ui/States'
import { CaseHeader } from '@/components/case/CaseHeader'
import { CaseFacts } from '@/components/case/CaseFacts'
import { ProgressTimeline } from '@/components/case/ProgressTimeline'
import { WeeklyUpdates } from '@/components/case/WeeklyUpdates'
import { EvidenceGallery } from '@/components/case/EvidenceGallery'
import { EvidenceUpload } from '@/components/case/EvidenceUpload'
import { WEEKLY_SUMMARY_ACTION } from '@/components/case/activity'
import { useCaseActions, useCaseDetail } from '@/hooks/useCases'
import { useAddProgress, useCaseProgress } from '@/hooks/useProgress'
import { useCaseEvidence, useUploadEvidence, useVerifyEvidence } from '@/hooks/useEvidence'
import { ApiError } from '@/lib/apiClient'
import { canAddProgress } from '@/lib/status'

const PROGRESS_ACTIONS = [
  { value: 'progress', label: 'Progress update' },
  { value: 'checkpoint', label: 'Checkpoint reached' },
  { value: 'interview', label: 'Witness interview' },
  { value: 'patrol', label: 'Patrol / sweep' },
  { value: 'note', label: 'Note' },
]

type TabId = 'details' | 'progress' | 'weekly' | 'evidence'

/**
 * Officer case workspace.
 *
 * One case, four lenses: what it is (details), what has been done (progress),
 * how the work breaks down by week, and what proves it (evidence). Lifecycle
 * actions are gated to the transitions the backend actually permits, and each
 * disabled action says why.
 */
export function OfficerCasePage() {
  const { id } = useParams<{ id: string }>()
  const { notify } = useToast()

  const detail = useCaseDetail(id)
  const progressQuery = useCaseProgress(id)
  const evidenceQuery = useCaseEvidence(id)
  const { dispatch, arrive, close } = useCaseActions(id)
  const addProgress = useAddProgress(id)
  const uploadEvidence = useUploadEvidence(id)
  const verifyEvidence = useVerifyEvidence(id)

  const [tab, setTab] = useState<TabId>('details')
  const [closeOpen, setCloseOpen] = useState(false)
  const [finalReport, setFinalReport] = useState('')
  const [evidenceOpen, setEvidenceOpen] = useState(false)
  const [action, setAction] = useState(PROGRESS_ACTIONS[0].value)
  const [description, setDescription] = useState('')

  const caseItem = detail.data?.case
  const progress = progressQuery.data ?? []
  const evidence = evidenceQuery.data ?? []

  const networkIssue =
    ApiError.isNetwork(detail.error) ||
    ApiError.isNetwork(progressQuery.error) ||
    ApiError.isNetwork(evidenceQuery.error)

  const offline = networkIssue && Boolean(caseItem)

  const assignedOfficerName = useMemo(() => {
    const entry = detail.data?.timeline.find((t) => t.user && t.user.id === caseItem?.assignedTo)
    if (entry?.user) return `${entry.user.firstName} ${entry.user.lastName}`.trim()
    return undefined
  }, [detail.data, caseItem])

  const tabs = useMemo(
    () => [
      { id: 'details', label: 'Details', icon: <FileText className="size-4" aria-hidden /> },
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

  function reportError(cause: unknown, fallback: string) {
    if (ApiError.isNetwork(cause)) {
      notify('No connection — the change was not saved.', 'error')
      return
    }
    notify(cause instanceof ApiError ? cause.message : fallback, 'error')
  }

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
            title={ApiError.isNetwork(detail.error) ? 'You are offline' : 'Case unavailable'}
            description={
              ApiError.isNetwork(detail.error)
                ? 'Reconnect to load this case.'
                : 'This case does not exist, or your unit does not have access to it.'
            }
            offline={ApiError.isNetwork(detail.error)}
            onRetry={() => void detail.refetch()}
          />
        </Card>
      </div>
    )
  }

  const status = caseItem.status
  const closed = status === 'closed'
  const progressAllowed = canAddProgress(status)

  return (
    <div className="mx-auto w-full max-w-4xl p-4 sm:p-6">
      <BackLink />

      {offline ? (
        <div className="mb-3">
          <OfflineBanner />
        </div>
      ) : null}

      <div className="mt-3 flex flex-col gap-4">
        <CaseHeader
          caseItem={caseItem}
          assignedOfficerName={assignedOfficerName}
          actions={
            <>
              {status === 'assigned' ? (
                <Button
                  variant="primary"
                  loading={dispatch.isPending}
                  icon={<BellRing className="size-4" aria-hidden />}
                  onClick={() =>
                    dispatch.mutate(undefined, {
                      onSuccess: () => notify('Marked as dispatched.', 'success'),
                      onError: (cause) => reportError(cause, 'Could not dispatch this case.'),
                    })
                  }
                >
                  Dispatch
                </Button>
              ) : null}

              {status === 'dispatched' ? (
                <Button
                  variant="primary"
                  loading={arrive.isPending}
                  icon={<MapPin className="size-4" aria-hidden />}
                  onClick={() =>
                    arrive.mutate(undefined, {
                      onSuccess: () => notify('Marked as on scene.', 'success'),
                      onError: (cause) => reportError(cause, 'Could not update this case.'),
                    })
                  }
                >
                  Mark on scene
                </Button>
              ) : null}

              {status === 'on_scene' ? (
                <Button
                  variant="primary"
                  icon={<CheckCircle2 className="size-4" aria-hidden />}
                  onClick={() => setCloseOpen(true)}
                >
                  Close case
                </Button>
              ) : null}

              {status === 'pending' ? (
                <Badge tone="neutral" className="h-9 px-3">
                  Awaiting assignment by your unit admin
                </Badge>
              ) : null}
            </>
          }
        />

        <Card>
          <Tabs items={tabs} activeId={tab} onChange={(next) => setTab(next as TabId)} />

          {tab === 'details' ? (
            <TabPanel id="details">
              <CaseFacts
                caseItem={caseItem}
                timeline={detail.data?.timeline ?? []}
                feedback={detail.data?.feedback ?? []}
              />
            </TabPanel>
          ) : null}

          {tab === 'progress' ? (
            <TabPanel id="progress">
              <div className="border-b border-border p-4">
                {progressAllowed ? (
                  <form
                    className="flex flex-col gap-3"
                    onSubmit={(event) => {
                      event.preventDefault()
                      const text = description.trim()
                      if (!text) return
                      addProgress.mutate(
                        { action, description: text },
                        {
                          onSuccess: () => {
                            setDescription('')
                            notify('Progress update added.', 'success')
                          },
                          onError: (cause) => reportError(cause, 'Could not add the update.'),
                        },
                      )
                    }}
                  >
                    <div className="grid gap-3 sm:grid-cols-[minmax(0,12rem)_1fr]">
                      <Field label="Kind of update">
                        {(props) => (
                          <Select
                            {...props}
                            value={action}
                            onChange={(event) => setAction(event.target.value)}
                          >
                            {PROGRESS_ACTIONS.map((option) => (
                              <option key={option.value} value={option.value}>
                                {option.label}
                              </option>
                            ))}
                          </Select>
                        )}
                      </Field>
                      <Field label="What happened" required>
                        {(props) => (
                          <Input
                            {...props}
                            value={description}
                            onChange={(event) => setDescription(event.target.value)}
                            placeholder="Brief, specific, in plain language"
                          />
                        )}
                      </Field>
                    </div>
                    <div className="flex justify-end">
                      <Button
                        type="submit"
                        variant="primary"
                        size="sm"
                        loading={addProgress.isPending}
                        disabled={!description.trim()}
                      >
                        Add update
                      </Button>
                    </div>
                  </form>
                ) : (
                  <p className="flex items-start gap-2 text-xs text-ink-muted">
                    <Info className="mt-0.5 size-4 shrink-0 text-signal" aria-hidden />
                    {closed
                      ? 'This case is closed — the progress record is now read-only.'
                      : 'Progress updates can be added once the case is dispatched. This case is still ' +
                        `${status}.`}
                  </p>
                )}
              </div>

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
              <WeeklyUpdates
                progress={progress}
                timeline={detail.data?.timeline ?? []}
                submitting={addProgress.isPending}
                onSubmitSummary={(text) =>
                  addProgress.mutate(
                    { action: WEEKLY_SUMMARY_ACTION, description: text },
                    {
                      onSuccess: () => notify('Weekly summary filed.', 'success'),
                      onError: (cause) => reportError(cause, 'Could not file the summary.'),
                    },
                  )
                }
              />
            </TabPanel>
          ) : null}

          {tab === 'evidence' ? (
            <TabPanel id="evidence">
              <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
                <p className="text-xs text-ink-muted">
                  {closed
                    ? 'This case is closed — no further evidence can be attached.'
                    : 'Attach photos, recordings and documents, then verify them.'}
                </p>
                <Button
                  size="sm"
                  variant="secondary"
                  icon={<Paperclip className="size-4" aria-hidden />}
                  disabled={closed}
                  title={closed ? 'Evidence cannot be added to a closed case.' : undefined}
                  onClick={() => setEvidenceOpen(true)}
                >
                  Attach
                </Button>
              </div>

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
                  canVerify={!closed}
                  verifyingId={verifyEvidence.isPending ? verifyEvidence.variables : null}
                  onVerify={(evidenceId) =>
                    verifyEvidence.mutate(evidenceId, {
                      onSuccess: () => notify('Evidence verified.', 'success'),
                      onError: (cause) => reportError(cause, 'Could not verify that item.'),
                    })
                  }
                />
              )}
            </TabPanel>
          ) : null}
        </Card>
      </div>

      {/* Close case */}
      <Modal
        open={closeOpen}
        onClose={() => setCloseOpen(false)}
        title="Close this case"
        description="The final report becomes the permanent record. This cannot be undone."
        footer={
          <>
            <Button variant="ghost" onClick={() => setCloseOpen(false)}>
              Cancel
            </Button>
            <Button
              variant="primary"
              loading={close.isPending}
              disabled={!finalReport.trim()}
              onClick={() =>
                close.mutate(finalReport.trim(), {
                  onSuccess: () => {
                    setCloseOpen(false)
                    setFinalReport('')
                    notify('Case closed.', 'success')
                  },
                  onError: (cause) => reportError(cause, 'Could not close this case.'),
                })
              }
            >
              Close case
            </Button>
          </>
        }
      >
        <Field
          label="Final report"
          required
          hint="What was found, what was done, and how the case was resolved."
        >
          {(props) => (
            <Textarea
              {...props}
              rows={6}
              value={finalReport}
              onChange={(event) => setFinalReport(event.target.value)}
              placeholder="Summary of the investigation and its outcome…"
            />
          )}
        </Field>
      </Modal>

      <EvidenceUpload
        open={evidenceOpen}
        onClose={() => setEvidenceOpen(false)}
        submitting={uploadEvidence.isPending}
        caseLatitude={caseItem.latitude}
        caseLongitude={caseItem.longitude}
        caseLocationLabel={caseItem.location}
        onSubmit={(input) =>
          uploadEvidence.mutate(input, {
            onSuccess: () => {
              setEvidenceOpen(false)
              notify('Evidence attached.', 'success')
            },
            onError: (cause) => reportError(cause, 'Could not attach that evidence.'),
          })
        }
      />
    </div>
  )
}

function BackLink() {
  return (
    <Link
      to="/officer/queue"
      className="inline-flex items-center gap-1.5 text-xs text-ink-muted transition-colors hover:text-ink"
    >
      <ArrowLeft className="size-3.5" aria-hidden />
      Back to queue
    </Link>
  )
}

