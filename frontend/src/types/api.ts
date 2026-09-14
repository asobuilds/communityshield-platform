/**
 * CommunityShield API types.
 *
 * Hand-written from the verified Go contract in backend/ (handlers + models).
 * Envelopes are intentional: the API wraps collections, e.g. `{ cases: [...] }`.
 * Keep this file in sync with the backend — it is the frontend's contract.
 */

export type Role = 'citizen' | 'officer' | 'unit_admin' | 'super_admin'

/**
 * The eight states the backend case workflow actually sets.
 *
 * Declared as one const object so the literals live in exactly one place — a
 * contract correction is then a single edit rather than a hunt. They are
 * snake_case because the backend stores plain strings: `case_workflow_handler.go`
 * compares against `"pending"` / `"on_scene"` / … directly, and
 * `case_review_handler.go` uses `models.CaseStatus*` constants of the same shape.
 *
 * NOTE: the three review-phase literals are the ones frontReadme.md §0.4 flags as
 * needing confirmation against `backend/models` — see the verification command there.
 */
export const CASE_STATUS = {
  pending: 'pending',
  assigned: 'assigned',
  dispatched: 'dispatched',
  onScene: 'on_scene',
  investigating: 'investigating',
  pendingAdminReview: 'pending_admin_review',
  adminChangesRequested: 'admin_changes_requested',
  closed: 'closed',
} as const

export type CaseStatus = (typeof CASE_STATUS)[keyof typeof CASE_STATUS]

/** Backend priority banding (see CreateCase in handlers/case_handler.go). */
export type PriorityLevel = 'P1' | 'P2' | 'P3'

export interface User {
  id: string
  email: string
  phone?: string
  firstName: string
  lastName: string
  role: Role
  status?: string
  createdAt?: string
  updatedAt?: string
}

export interface Evidence {
  id: string
  caseId: string
  uploadedBy: string
  /** Free-form today (image | video | audio | document …). */
  type: string
  fileUrl: string
  description?: string
  latitude?: number
  longitude?: number
  isVerified: boolean
  uploadedAt: string
  createdAt: string
}

export interface Progress {
  id: string
  caseId: string
  officerId: string
  action: string
  description?: string
  createdAt: string
}

export interface CaseTimelineEntry {
  id: string
  caseId: string
  userId: string
  action: string
  description?: string
  status?: string
  metadata?: string
  createdAt: string
  updatedAt?: string
  user?: User
}

export interface CaseOfficer {
  id: string
  caseId: string
  officerId: string
  /** primary | investigator | support */
  role: string
  createdAt?: string
}

export interface CaseFeedback {
  id: string
  caseId: string
  userId: string
  rating: number
  comment?: string
  categories?: string
  isPublic?: boolean
  createdAt?: string
}

export interface Case {
  id: string
  unitId: string
  reportedBy: string
  assignedTo?: string | null
  title: string
  description: string
  incidentDate?: string
  location?: string
  latitude: number
  longitude: number
  status: CaseStatus
  priority?: string
  priorityLevel: PriorityLevel
  trackingId: string
  gisLatitude: number
  gisLongitude: number
  transferDetails?: string
  isPublic: boolean
  assignedAt?: string | null
  dispatchedAt?: string | null
  arrivedAt?: string | null
  closedAt?: string | null
  closedBy?: string | null
  approvedBy?: string | null
  finalReport?: string
  createdAt: string
  updatedAt: string
  evidence?: Evidence[]
  progress?: Progress[]
}

/**
 * One decision recorded against a case's closure review (`models.CaseReview`).
 *
 * `decision` is typed as a plain string on purpose: the backend's
 * `CaseReviewDecision*` literals are not yet verified against `backend/models`
 * (frontReadme.md §0.4), and a decision this build does not recognise must render
 * as itself rather than be coerced into one of the two it expects.
 */
export interface CaseReview {
  id: string
  caseId: string
  adminId: string
  /** Expected `approve` | `request_changes`. Not yet contract-verified. */
  decision: string
  comment: string
  createdAt: string
}

/**
 * An officer's narrative for one reporting week (`models.CaseWeeklyUpdate`).
 *
 * The week is Monday 00:00 UTC → the following Monday, computed **server-side**;
 * `weekStart` is the key the backend deduplicates on, so a second submission for
 * the same week is refused rather than merged.
 */
export interface CaseWeeklyUpdate {
  id: string
  caseId: string
  officerId: string
  weekStart: string
  weekEnd: string
  summary: string
  investigation: string
  actionsTaken?: string
  findings?: string
  evidenceSummary?: string
  outstandingActions?: string
  nextSteps?: string
  /**
   * Whether this update is exposed to the case reporter. The backend writes `true`
   * today and filters a reporter's read by it — so the *same* case returns fewer
   * updates to a citizen than to an officer.
   *
   * Optional because the underlying JSON tag is not yet contract-verified
   * (frontReadme.md §0.4). Read it as a label only, and only when it is `true`:
   * the server does the filtering, so a missing value must never be treated as
   * "visible" and rendered as a privacy claim.
   */
  citizenVisible?: boolean
  submittedAt?: string | null
  createdAt: string
}

/** Body of `POST /cases/:id/weekly-update` — the first two are required. */
export interface CaseWeeklyUpdateInput {
  summary: string
  investigation: string
  actionsTaken?: string
  findings?: string
  evidenceSummary?: string
  outstandingActions?: string
  nextSteps?: string
}

export interface SecurityUnit {
  id: string
  name: string
  type: string
  latitude: number
  longitude: number
  /** Operational coverage radius in kilometres. */
  operationalRadius: number
  state?: string
  lga?: string
  city?: string
  coverageArea?: string
  contactPerson?: string
  contactPhone?: string
  contactEmail?: string
  registrationNumber?: string
  status: string
  isVerified: boolean
  verificationStatus?: string
}

/** Enriched unit returned by GET /units/nearby. */
export interface UnitWithDistance extends SecurityUnit {
  distance: number
  isInRange: boolean
}

/**
 * A roster officer — the backend's `officers` table, which is a *different*
 * entity from the login `User`.
 *
 * This matters for assignment: `POST /cases/:id/assign` takes an `officers` id,
 * NOT a user id, and the backend rejects any officer whose unit differs from the
 * case's unit (`OfficerID` is checked against `Officer.UnitID`).
 */
export interface UnitOfficer {
  id: string
  unitId: string
  name: string
  rank: string
  badgeNumber: string
  /** Duty role: patrol, investigator, commander, dispatch … */
  role: string
  phone?: string
  email?: string
  joinedDate?: string
  status: string
}

export interface Notification {
  id: string
  userId: string
  title: string
  message: string
  type?: string
  status: 'unread' | 'read' | string
  createdAt: string
}

/* ------------------------------------------------------------------ */
/* Response envelopes                                                  */
/* ------------------------------------------------------------------ */

export interface LoginResponse {
  token: string
  user: Pick<User, 'id' | 'email' | 'firstName' | 'lastName' | 'role'>
}

export interface ProfileResponse {
  user: User
}

export interface CasesResponse {
  cases: Case[]
}

export interface CaseDetailResponse {
  case: Case
  timeline: CaseTimelineEntry[]
  feedback: CaseFeedback[]
}

export interface ProgressResponse {
  caseId: string
  progress: Progress[]
}

export interface EvidenceListResponse {
  evidence: Evidence[]
}

export interface EvidenceCreateResponse {
  message: string
  evidence: Evidence
}

export interface UnitsResponse {
  units: SecurityUnit[]
}

export interface NearbyUnitsResponse {
  units: UnitWithDistance[]
}

export interface UnitResponse {
  unit: SecurityUnit
}

export interface UnitOfficersResponse {
  officers: UnitOfficer[]
}

export interface CaseAssignmentsResponse {
  assignments: CaseOfficer[]
}

export interface AssignCaseResponse {
  message: string
  assignment: CaseOfficer
  case: Case
}

/** `GET /cases/:id/review` — the case's current state plus its decision history. */
export interface CaseReviewResponse {
  caseId: string
  status: CaseStatus
  reviews: CaseReview[]
}

/** `GET /cases/:id/weekly-updates` — already filtered for the caller's role. */
export interface CaseWeeklyUpdatesResponse {
  caseId: string
  updates: CaseWeeklyUpdate[]
}

/**
 * `POST /cases/:id/weekly-update`. A duplicate week is refused with **409** and a
 * body of `{ error, update }` — i.e. the existing update is returned so the UI can
 * show what is already on file instead of an empty failure.
 */
export interface CaseWeeklyUpdateCreateResponse {
  message: string
  update: CaseWeeklyUpdate
}

export interface CaseWeeklyUpdateConflictResponse {
  error: string
  update: CaseWeeklyUpdate
}

/**
 * `POST /cases/:id/submit-review`. The 409 that a wrong-state submission returns
 * carries the case's actual `status`, so the UI can correct itself instead of
 * guessing why it was refused.
 */
export interface SubmitReviewConflictResponse {
  error: string
  status: CaseStatus
}

export interface NotificationsResponse {
  notifications: Notification[]
  unreadCount: number
}

export interface CaseAnalyticsResponse {
  totalCases: number
  /**
   * NOTE: the backend counts `status = "resolved"`, a state the workflow never
   * sets, so this (and resolutionRate) is always 0 today. Do not present it as
   * truth — prefer deriving open/closed counts client-side.
   */
  resolvedCases: number
  pendingCases: number
  resolutionRate: number
}
