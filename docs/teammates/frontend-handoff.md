# Frontend Handoff — Wave 8

This is a reconciliation doc, not a spec. Your existing frontend work is
taken as given. This file tells you what changed under you, what the
backend guarantees, and where your freedom lives versus where you must
conform.

## 1. What you've already built (observed in repo)

- Map stack is Leaflet-based: `leaflet ^1.9.4`, `react-leaflet ^5.0.0`,
  `@types/leaflet ^1.9.15` (package.json).
- Map components: `src/components/map/MapView.tsx`,
  `src/components/map/IncidentMap.tsx`, `src/pages/MapPage.tsx`.
  Registered route: `/map` (App.tsx:97).
- Map code is mid-refactor — `src/components/map/recovery/*` contains
  stage snapshots. Expect further shape changes.
- API layer: check `src/api/` and `src/hooks/` for your client. Key file
  observed: `src/hooks/useOfficers.ts`.

## 2. What changed since your last sync

Wave 8b renamed a route parameter across the backend: `:unitId` → `:id`.
This affects any URL your client builds with `:unitId`.

Two specific mismatches to check on your side:

1. **MSW mock drift.** Your mock handlers file
   (`handlers.ts:948`) registers `/units/:unitId/officers`. The backend
   uses `:id`. Update the mock to match, or mock-vs-live will disagree.

2. **Missing backend route.** `useOfficers.ts:19` calls
   `GET /units/:id/officers`. That route is NOT registered in
   `backend/routes/routes.go`. The only officers route present is
   `GET /units/:id/officers/ranking`. Either the frontend should call the
   ranking route, or the backend needs to expose the plain officers list
   — flag which one you want and we'll wire it.

3. **Legacy doc.** `TEAM_HANDOFF.md:41` still shows
   `/public/units/:unitId/bank-accounts` paths. Those are pre-Wave-8b.
   Treat that section as outdated.

## 3. What the backend guarantees

- Auth: JWT access token + refresh token + revocation list. Endpoints
  under `/api/v1/*` (authenticated group) require
  `Authorization: Bearer <token>`. `/api/v1/public/*` endpoints are open.
- Case list pagination: `GET /cases` now caps `limit` at 100 and defaults
  to 50 (commit a5d7e9f). If you were relying on unbounded results,
  update the client.
- CORS origins are configured in `backend/main.go` — to add a new origin
  (e.g. a preview deploy URL), ask; do not assume.
- Error envelope shape: see `backend/handlers/*_test.go` for live
  examples of success and failure responses.

## 4. What's still in motion

- Email delivery is a TODO (SMS path is what's wired). Any "forgot
  password" flow should route through SMS today.
- More list endpoints will get pagination caps in Wave 8f.2
  (`units`, `notifications`, `finance/transactions`, `bank-accounts`,
  `ratings`, `public/units`). Do not hard-code "get all" assumptions.
- Rate limiting on map and location endpoints lands in Wave 8g.
- Location field visibility may tighten in Wave 8h (which lat/lng fields
  are public vs authed).

## 5. Where your creative freedom lives

Everything above the API line is yours:

- Component structure, state management, styling, motion, accessibility
- Data density choices (table / card / hybrid) per surface
- Onboarding flows, empty states, error states, skeleton design
- Copy, tone, interaction patterns

None of the above is specified by the backend. Build as you've been
building.

## 6. Where you must conform

- Use `:id` in route params, not `:unitId`.
- Respect the auth token lifecycle. The refresh + revocation flow is not
  negotiable.
- Parse the backend's error envelope as the backend emits it. Do not
  invent fields.
- CORS: request new origins; don't assume they exist.

## 7. How to test against the real thing

- `backend/handlers/*_test.go` are executable examples of expected
  request/response shapes. Read them as your contract reference.
- CI (`.github/workflows/integration-race.yml`) runs `-race` + fuzz on
  every push. If your integration breaks the suite, CI goes red on the
  same commit.
