/**
 * CommunityShield API types.
 *
 * Hand-written from the verified Go contract in backend/ (handlers + models).
 * Envelopes are intentional: the API wraps collections, e.g. `{ cases: [...] }`.
 * Keep this file in sync with the backend — it is the frontend's contract.
 */

export type Role = 'citizen' | 'officer' | 'unit_admin' | 'super_admin'

/** The five states the backend workflow actually sets, in order. */
export type CaseStatus = 'pending' | 'assigned' | 'dispatched' | 'on_scene' | 'closed'

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
