# WardGuard — Frontend Build Handoff

> Everything the backend offers today. Endpoint reference, auth flow,
> role matrix, and the patterns the frontend must follow.

## Auth flow

1. POST /api/v1/auth/login → { token, refreshToken, user }
2. Store both. Attach `Authorization: Bearer <token>` to every authed call.
3. On 401 → POST /api/v1/auth/refresh { refreshToken } → new pair. Retry once.
4. On second 401 → hard logout, clear storage, redirect to /login.
5. POST /api/v1/auth/logout revokes the current token server-side.

**Token lifetime:** access = 24h, refresh = 30 days. Refresh tokens rotate
on use — save the new one every time.

## Sessions (device list)

- GET  /api/v1/auth/sessions          → list active devices
- DELETE /api/v1/auth/sessions/:jti   → revoke one device
- DELETE /api/v1/auth/sessions        → revoke all

The `jti` comes from the JWT payload — decode client-side or return it
from the sessions list.

## Role matrix

| Role | Sees |
|---|---|
| citizen | Own cases, public map, community, alerts, own ratings |
| officer | Own unit's cases + roster + finance ledger + officer leaderboard |
| unit_admin | Own unit's cases + all admin actions + finance + revocation |
| head_admin | Above + all cases in unit + archive + UnitAuth + news |
| super_admin | Everything |

## Key endpoint groups

### Public (no auth)
- GET /api/v1/public/cases
- GET /api/v1/public/units
- GET /api/v1/public/units/:unitId/bank-accounts
- GET /api/v1/public/units/:unitId/ledger
- GET /api/v1/public/units/:unitId/financial-years
- GET /api/v1/public/units/:unitId/financial-summary
- GET /api/v1/public/platform/donation-info
- POST /api/v1/public/platform/donations
- GET /api/v1/public/platform/supporters
- GET /api/v1/public/leaderboard
- GET /api/v1/public/units/suggest
- GET /api/v1/public/officers/:id/rating
- GET /api/v1/public/units/:unitId/rating
- POST /api/v1/invites/validate

### Auth (any logged-in user)
- GET  /api/v1/auth/profile
- POST /api/v1/auth/change-password
- POST /api/v1/auth/refresh
- GET/DELETE /api/v1/auth/sessions
- POST /api/v1/cases
- GET  /api/v1/cases
- GET  /api/v1/cases/:id
- POST /api/v1/suspects/me/cases
- POST /api/v1/ratings
- POST /api/v1/ratings/:id/flag
- POST /api/v1/users/me/avatar  (multipart: file)
- POST /api/v1/users/me/cover   (multipart: file)

### Case workflow (with CanAccessCase middleware)
Full list in `routes.go`. Key ones:
- POST /api/v1/cases/:id/assign
- POST /api/v1/cases/:id/dispatch
- POST /api/v1/cases/:id/arrive
- POST /api/v1/cases/:id/progress
- POST /api/v1/cases/:id/submit-review
- POST /api/v1/cases/:id/review/approve
- POST /api/v1/cases/:id/review/request-changes
- POST /api/v1/cases/:id/weekly-update
- GET  /api/v1/cases/:id/weekly-updates

### Governance (admin/head only)
- POST /api/v1/units/:unitId/elections
- POST /api/v1/units/:unitId/head-admin-elections
- POST /api/v1/units/:unitId/revocations
- GET/PUT /api/v1/units/:unitId/auth
- POST /api/v1/elections/:id/vote
- POST /api/v1/elections/:id/close
- GET  /api/v1/elections/:id/results
- POST /api/v1/revocations/:id/vote
- POST /api/v1/revocations/:id/close
- GET  /api/v1/revocations/:id

### Evidence (with CanAccessCase)
- POST /api/v1/evidence/upload          (legacy — takes URL)
- POST /api/v1/evidence/case/:caseId/file  (multipart: file)
- GET  /api/v1/evidence/case/:caseId
- PATCH /api/v1/evidence/:id/verify
- DELETE /api/v1/evidence/:id

### File serving (authed)
- GET /api/v1/files/:category/:hash
  - category: avatars | covers | evidence | gov_ids | voice_notes
  - evidence requires ?caseId=<uuid>
  - Returns raw file stream with correct Content-Type

### Finance (unit-scoped)
- POST /api/v1/bank/accounts
- GET  /api/v1/bank/:unitId/accounts
- PATCH /api/v1/bank/accounts/:id/public
- GET  /api/v1/bank/units/:unitId/ledger
- POST /api/v1/finance/transactions
- POST /api/v1/finance/transactions/:id/approve
- POST /api/v1/finance/transactions/:id/reject
- POST /api/v1/finance/budgets
- POST /api/v1/finance/reports

### Admin (super_admin only)
- GET  /api/v1/admin/users
- PUT  /api/v1/admin/users/:id/role
- POST /api/v1/admin/users/:id/suspend
- GET  /api/v1/admin/platform-donations
- POST /api/v1/admin/platform-donations/:id/confirm

## Rate limits

- Auth/register/login/logout: 10/min per IP
- OTP send/verify/resend: 5/5min per IP
- Vote endpoints: 20/min per user
- Invite creation: 10/hour per user
- General: 200/min per IP

On 429, read `Retry-After` header and back off.

## Error shapes

All errors are JSON: `{ "error": "human-readable message" }`.
Never display the raw message to end users for 500s — show a generic
"something went wrong" and log the actual body.

## Idempotency

Endpoints that accept `Idempotency-Key` header:
- POST /cases/:id/assign
- POST /elections/:id/vote, /close
- POST /revocations/:id/vote, /close
- POST /units/:unitId/elections, /head-admin-elections, /revocations

Send a UUID per user action. Retry with the same key → replays
the original response instead of re-executing.

## Ratings UI notes

- Bayesian score is 1.0–5.0. Display with 1 decimal: `4.6★`
- Tier values: `unranked` | `bronze` | `silver` | `gold` | `platinum` | `diamond`
- Only the reporter can rate, only after the case closes via the platform
- 30-day window after closure
- Officer leaderboard is **within-unit only** — don't render it on the public map
- Platform-wide leaderboard (`/public/leaderboard`) shows **units**, not officers

## Multipart uploads

File fields must be named exactly `file`. Max sizes:
- avatar/cover: 5 MB
- evidence: 100 MB
- gov_id: 10 MB

Validation happens server-side using magic bytes. The client should
also check MIME and extension — but never trust the client alone.

## What's NOT built yet (frontend must not call)

- AI streaming — current endpoints are request/response only
- WebSocket push — `/ws` exists but no room auth in place
- Payment processor integration — donors see bank details, no in-app pay
- Blueprint public docs — endpoint pending, coming soon

## Frontend gaps to flag

- `on_scene → investigating` has no backend route — don't build a button
- SMS webhooks are HMAC-gated — frontend doesn't touch them
- Public map should use only `/public/*` endpoints for unauthenticated users