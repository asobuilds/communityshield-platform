/**
 * The mock API route table.
 *
 * These deliberately mirror the *real* backend contract — envelopes, status
 * codes and authorization — so turning `VITE_USE_MOCKS` off changes only where
 * the data comes from, never how the frontend talks to it.
 *
 * Guards the UI depends on, mirrored here:
 *  - progress may only be added while a case is `dispatched` or `on_scene`;
 *  - evidence may not be attached to a closed case;
 *  - `GET /cases` is scoped by role server-side, never by the client.
 */

import { sleep, type MockRoute } from './adapter'
import { USERS, seedDatabase, type MockDatabase } from './seed'
import { MOCK_ACCOUNTS, mockTokenFor, userIdFromToken } from './config'
import type { Case, Progress, User } from '@/types/api'

const db: MockDatabase = seedDatabase()

function currentUser(request: Request): User | null {
  const header = request.headers.get('Authorization') ?? ''
  const userId = userIdFromToken(header.replace(/^Bearer\s+/i, ''))
  return userId ? USERS[userId] ?? null : null
}

const unauthorized = { status: 401, body: { error: 'authentication required' } }
const forbidden = (message: string) => ({ status: 403, body: { error: message } })
const notFound = (message: string) => ({ status: 404, body: { error: message } })

/**
 * The unit an officer belongs to, and the unit the administrator runs.
 *
 * The real `/auth/profile` carries no `unitId` yet (a documented contract gap),
 * so the mock stands in for that scoping: an officer sees their own unit's work,
 * and a unit administrator sees the unit they run — Surulere Central, whose
 * contact person is the `admin@shield.ng` account. A super administrator is the
 * only role that sees across units.
 */
const OFFICER_UNIT = seedDatabase().cases[0].unitId
const ADMIN_UNIT = OFFICER_UNIT

function casesVisibleTo(user: User): Case[] {
  switch (user.role) {
    case 'citizen':
      return db.cases.filter((c) => c.reportedBy === user.id)
    case 'officer':
      return db.cases.filter((c) => c.unitId === OFFICER_UNIT || c.assignedTo === user.id)
    case 'unit_admin':
      return db.cases.filter((c) => c.unitId === ADMIN_UNIT)
    case 'super_admin':
    default:
      return db.cases
  }
}

function canSeeCase(user: User, caseItem: Case): boolean {
  return casesVisibleTo(user).some((c) => c.id === caseItem.id)
}

function haversineKm(aLat: number, aLng: number, bLat: number, bLng: number): number {
  const R = 6371
  const toRad = (deg: number) => (deg * Math.PI) / 180
  const dLat = toRad(bLat - aLat)
  const dLng = toRad(bLng - aLng)
  const lat1 = toRad(aLat)
  const lat2 = toRad(bLat)
  const h = Math.sin(dLat / 2) ** 2 + Math.cos(lat1) * Math.cos(lat2) * Math.sin(dLng / 2) ** 2
  return 2 * R * Math.asin(Math.sqrt(h))
}

const nowIso = () => new Date().toISOString()

export const handlers: MockRoute[] = [
  /* ---------------------------------------------------------------- auth */

  {
    method: 'POST',
    path: '/auth/login',
    async respond({ request }) {
      await sleep(250)
      const body = (await request.json().catch(() => ({}))) as {
        email?: string
        password?: string
      }
      const email = (body.email ?? '').trim().toLowerCase()
      const account = MOCK_ACCOUNTS.find((a) => a.email === email)

      if (!account || account.password !== body.password) {
        return { status: 401, body: { error: 'invalid credentials' } }
      }

      const user = USERS[account.userId]
      return {
        body: {
          token: mockTokenFor(user.id),
          user: {
            id: user.id,
            email: user.email,
            firstName: user.firstName,
            lastName: user.lastName,
            role: user.role,
          },
        },
      }
    },
  },

  {
    method: 'GET',
    path: '/auth/profile',
    async respond({ request }) {
      await sleep(120)
      const user = currentUser(request)
      if (!user) return unauthorized
      return { body: { user } }
    },
  },

  /* --------------------------------------------------------------- cases */

  {
    method: 'GET',
    path: '/cases',
    async respond({ request }) {
      await sleep(220)
      const user = currentUser(request)
      if (!user) return unauthorized

      const cases = casesVisibleTo(user).map((c) => ({
        ...c,
        evidence: db.evidence.filter((e) => e.caseId === c.id),
        progress: db.progress.filter((p) => p.caseId === c.id),
      }))
      return { body: { cases } }
    },
  },

  {
    method: 'GET',
    path: '/cases/:id',
    async respond({ request, params }) {
      await sleep(200)
      const user = currentUser(request)
      if (!user) return unauthorized

      const caseItem = db.cases.find((c) => c.id === params.id)
      if (!caseItem) return notFound('case not found')
      if (!canSeeCase(user, caseItem)) {
        return forbidden('you are not authorized to view this case')
      }

      return {
        body: {
          case: {
            ...caseItem,
            evidence: db.evidence.filter((e) => e.caseId === caseItem.id),
            progress: db.progress.filter((p) => p.caseId === caseItem.id),
          },
          timeline: db.timeline
            .filter((t) => t.caseId === caseItem.id)
            .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()),
          feedback: db.feedback.filter((f) => f.caseId === caseItem.id),
        },
      }
    },
  },

  /* ------------------------------------------------------- case workflow */

  {
    method: 'POST',
    path: '/cases/:id/dispatch',
    async respond({ request, params }) {
      await sleep(200)
      const user = currentUser(request)
      if (!user) return unauthorized

      const caseItem = db.cases.find((c) => c.id === params.id)
      if (!caseItem) return notFound('case not found')
      if (caseItem.status !== 'assigned') {
        return {
          status: 409,
          body: {
            error: 'case must be assigned before it can be dispatched',
            status: caseItem.status,
          },
        }
      }

      caseItem.status = 'dispatched'
      caseItem.dispatchedAt = nowIso()
      caseItem.updatedAt = caseItem.dispatchedAt
      db.timeline.unshift({
        id: `tl-${Date.now()}`,
        caseId: caseItem.id,
        userId: user.id,
        action: 'dispatched',
        description: 'Officer dispatched.',
        status: 'dispatched',
        createdAt: caseItem.dispatchedAt,
        user,
      })

      return { body: { message: 'case dispatched successfully', case: caseItem } }
    },
  },

  {
    method: 'POST',
    path: '/cases/:id/arrive',
    async respond({ request, params }) {
      await sleep(200)
      const user = currentUser(request)
      if (!user) return unauthorized

      const caseItem = db.cases.find((c) => c.id === params.id)
      if (!caseItem) return notFound('case not found')
      if (caseItem.status !== 'dispatched') {
        return {
          status: 409,
          body: {
            error: 'case must be dispatched before arrival can be recorded',
            status: caseItem.status,
          },
        }
      }

      caseItem.status = 'on_scene'
      caseItem.arrivedAt = nowIso()
      caseItem.updatedAt = caseItem.arrivedAt
      db.timeline.unshift({
        id: `tl-${Date.now()}`,
        caseId: caseItem.id,
        userId: user.id,
        action: 'arrived',
        description: 'Officer arrived on scene.',
        status: 'on_scene',
        createdAt: caseItem.arrivedAt,
        user,
      })

      return { body: { message: 'arrival recorded successfully', case: caseItem } }
    },
  },

  {
    method: 'POST',
    path: '/cases/:id/close',
    async respond({ request, params }) {
      await sleep(250)
      const user = currentUser(request)
      if (!user) return unauthorized

      const caseItem = db.cases.find((c) => c.id === params.id)
      if (!caseItem) return notFound('case not found')
      if (caseItem.status === 'closed') {
        return { status: 409, body: { error: 'case is already closed' } }
      }

      const body = (await request.json().catch(() => ({}))) as { finalReport?: string }
      const finalReport = (body.finalReport ?? '').trim()
      if (!finalReport) {
        return { status: 400, body: { error: 'a final report is required' } }
      }

      caseItem.status = 'closed'
      caseItem.closedAt = nowIso()
      caseItem.closedBy = user.id
      caseItem.finalReport = finalReport
      caseItem.updatedAt = caseItem.closedAt
      db.timeline.unshift({
        id: `tl-${Date.now()}`,
        caseId: caseItem.id,
        userId: user.id,
        action: 'closed',
        description: 'Case closed with a final report.',
        status: 'closed',
        createdAt: caseItem.closedAt,
        user,
      })

      return { body: { message: 'case closed successfully', case: caseItem } }
    },
  },

  {
    method: 'POST',
    path: '/cases/:id/assign',
    async respond({ request, params }) {
      await sleep(200)
      const user = currentUser(request)
      if (!user) return unauthorized

      const caseItem = db.cases.find((c) => c.id === params.id)
      if (!caseItem) return notFound('case not found')

      const body = (await request.json().catch(() => ({}))) as {
        officerId?: string
        role?: string
      }
      if (!body.officerId) {
        return { status: 400, body: { error: 'officerId is required' } }
      }

      // Mirrors AssignCase: the id is an *officers* id, and it must belong to
      // the case's own unit, or the assignment is refused outright.
      const officer = db.officers.find((o) => o.id === body.officerId)
      if (!officer) return notFound('officer not found')
      if (officer.unitId !== caseItem.unitId) {
        return { status: 400, body: { error: 'officer does not belong to the case unit' } }
      }

      const role = body.role || 'primary'
      const existing = db.assignments.find(
        (a) => a.caseId === caseItem.id && a.officerId === officer.id,
      )

      let assignment = existing
      if (assignment) {
        assignment.role = role
      } else {
        assignment = {
          id: `as-${Date.now()}`,
          caseId: caseItem.id,
          officerId: officer.id,
          role,
          createdAt: nowIso(),
        }
        db.assignments.push(assignment)
      }

      // Only a primary assignment takes ownership of the case (backend rule).
      if (!caseItem.assignedTo || role === 'primary') {
        caseItem.assignedTo = officer.id
        caseItem.assignedAt = nowIso()
        if (caseItem.status === 'pending') caseItem.status = 'assigned'
        caseItem.updatedAt = caseItem.assignedAt
      }

      db.timeline.unshift({
        id: `tl-${Date.now()}`,
        caseId: caseItem.id,
        userId: user.id,
        action: 'case_assigned',
        description: `Assigned to ${officer.rank} ${officer.name} (${officer.badgeNumber}) as ${role}.`,
        status: caseItem.status,
        createdAt: nowIso(),
        user,
      })

      return {
        body: { message: 'case assigned successfully', assignment, case: caseItem },
      }
    },
  },

  {
    method: 'GET',
    path: '/cases/:id/assignments',
    async respond({ request, params }) {
      await sleep(150)
      const user = currentUser(request)
      if (!user) return unauthorized

      const caseItem = db.cases.find((c) => c.id === params.id)
      if (!caseItem) return notFound('case not found')

      const assignments = db.assignments
        .filter((a) => a.caseId === caseItem.id)
        .sort((a, b) => new Date(a.createdAt ?? 0).getTime() - new Date(b.createdAt ?? 0).getTime())

      return { body: { assignments } }
    },
  },

  /* ------------------------------------------------------------ progress */

  {
    method: 'GET',
    path: '/cases/:id/progress',
    async respond({ request, params }) {
      await sleep(180)
      const user = currentUser(request)
      if (!user) return unauthorized

      const caseItem = db.cases.find((c) => c.id === params.id)
      if (!caseItem) return notFound('case not found')

      // Mirrors the real handler: ordered oldest-first.
      const progress = db.progress
        .filter((p) => p.caseId === caseItem.id)
        .sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())

      return { body: { caseId: caseItem.id, progress } }
    },
  },

  {
    method: 'POST',
    path: '/cases/:id/progress',
    async respond({ request, params }) {
      await sleep(220)
      const user = currentUser(request)
      if (!user) return unauthorized

      const caseItem = db.cases.find((c) => c.id === params.id)
      if (!caseItem) return notFound('case not found')

      const body = (await request.json().catch(() => ({}))) as {
        action?: string
        description?: string
      }
      if (!body.action) {
        return { status: 400, body: { error: 'action is required' } }
      }

      if (caseItem.status !== 'dispatched' && caseItem.status !== 'on_scene') {
        return {
          status: 409,
          body: {
            error: 'progress cannot be added from the current case status',
            status: caseItem.status,
          },
        }
      }

      const progress: Progress = {
        id: `pg-${Date.now()}`,
        caseId: caseItem.id,
        officerId: user.id,
        action: body.action,
        description: body.description,
        createdAt: nowIso(),
      }
      db.progress.push(progress)

      return { status: 201, body: { message: 'case progress added successfully', progress } }
    },
  },

  /* ------------------------------------------------------------ evidence */

  {
    method: 'GET',
    path: '/evidence/case/:caseId',
    async respond({ request, params }) {
      await sleep(180)
      const user = currentUser(request)
      if (!user) return unauthorized

      const evidence = db.evidence
        .filter((e) => e.caseId === params.caseId)
        .sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime())

      return { body: { evidence } }
    },
  },

  {
    method: 'POST',
    path: '/evidence/upload',
    async respond({ request }) {
      await sleep(250)
      const user = currentUser(request)
      if (!user) return unauthorized

      const body = (await request.json().catch(() => ({}))) as {
        caseId?: string
        type?: string
        fileUrl?: string
        description?: string
        latitude?: number
        longitude?: number
      }

      if (!body.caseId || !body.type || !body.fileUrl) {
        return { status: 400, body: { error: 'caseId, type and fileUrl are required' } }
      }

      const caseItem = db.cases.find((c) => c.id === body.caseId)
      if (!caseItem) return notFound('case not found')
      if (caseItem.status === 'closed') {
        return { status: 409, body: { error: 'evidence cannot be uploaded to a closed case' } }
      }

      const evidence = {
        id: `ev-${Date.now()}`,
        caseId: caseItem.id,
        uploadedBy: user.id,
        type: body.type,
        fileUrl: body.fileUrl,
        description: body.description,
        latitude: body.latitude,
        longitude: body.longitude,
        isVerified: false,
        uploadedAt: nowIso(),
        createdAt: nowIso(),
      }
      db.evidence.push(evidence)

      return { status: 201, body: { message: 'evidence uploaded successfully', evidence } }
    },
  },

  {
    method: 'PATCH',
    path: '/evidence/:id/verify',
    async respond({ request, params }) {
      await sleep(200)
      const user = currentUser(request)
      if (!user) return unauthorized

      const evidence = db.evidence.find((e) => e.id === params.id)
      if (!evidence) return notFound('evidence not found')

      evidence.isVerified = true
      return { body: { message: 'evidence verified successfully', evidence } }
    },
  },

  /* --------------------------------------------------------------- units */

  {
    method: 'GET',
    path: '/units',
    async respond() {
      await sleep(150)
      return { body: { units: db.units } }
    },
  },

  {
    method: 'GET',
    path: '/units/:unitId/officers',
    async respond({ request, params }) {
      await sleep(150)
      const user = currentUser(request)
      if (!user) return unauthorized

      // Mirrors handlers.GetOfficersByUnit — a handler that exists in the
      // backend but is NOT registered in routes/routes.go, so this 404s against
      // the live API. AssignOfficerDialog treats that as a contract gap.
      const officers = db.officers.filter((o) => o.unitId === params.unitId)
      return { body: { officers } }
    },
  },

  {
    method: 'GET',
    path: '/units/nearby',
    async respond({ url }) {
      await sleep(150)
      const lat = Number(url.searchParams.get('lat'))
      const lng = Number(url.searchParams.get('lng'))
      const radius = Number(url.searchParams.get('radius') ?? 50)

      const units = db.units
        .map((unit) => {
          const distance = haversineKm(lat, lng, unit.latitude, unit.longitude)
          return {
            ...unit,
            distance: Math.round(distance * 100) / 100,
            isInRange: distance <= radius,
          }
        })
        .sort((a, b) => a.distance - b.distance)

      return { body: { units } }
    },
  },

  /* ------------------------------------------------------- notifications */

  {
    method: 'GET',
    path: '/mobile/notifications',
    async respond({ request }) {
      await sleep(150)
      const user = currentUser(request)
      if (!user) return unauthorized

      const notifications = db.notifications
        .filter((n) => n.userId === user.id)
        .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())

      return {
        body: {
          notifications,
          unreadCount: notifications.filter((n) => n.status === 'unread').length,
        },
      }
    },
  },

  {
    method: 'PUT',
    path: '/mobile/notifications/:id/read',
    async respond({ request, params }) {
      await sleep(120)
      const user = currentUser(request)
      if (!user) return unauthorized

      const notification = db.notifications.find((n) => n.id === params.id && n.userId === user.id)
      if (!notification) return notFound('notification not found')

      notification.status = 'read'
      return { body: { message: 'notification marked as read' } }
    },
  },

  {
    method: 'PUT',
    path: '/mobile/notifications/read-all',
    async respond({ request }) {
      await sleep(150)
      const user = currentUser(request)
      if (!user) return unauthorized

      db.notifications.forEach((n) => {
        if (n.userId === user.id) n.status = 'read'
      })
      return { body: { message: 'all notifications marked as read' } }
    },
  },
]
