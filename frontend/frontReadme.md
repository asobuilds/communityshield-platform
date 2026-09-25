# Nativity Guard — Frontend Feature & Build Specification

> **This is the frontend feature plan and mock-app inventory.**
> It lists every feature, screen, state, and API mapping needed to ship the product.
> Pair it with [`frontagent.md`](./frontagent.md) — the design agent that governs *how* the UI
> should look and feel. `frontReadme` says **what to build**; `frontagent` says **how to make it
> compelling**.

---

## Python backend contract — current source of truth

The backend in this repository is **Python (FastAPI and SQLAlchemy)**. Its mounted routes are
declared in `backend/app/main.py` and `backend/app/routers/`. The browser app currently targets a
different API shape: its client defaults to `/api/v1`, while the Python app mounts routes such as
`/incidents/`, `/locations/`, `/organizations/`, `/emergency-resources/`,
`/response-assignments/` and `/live-locations/` without that prefix. The mock API supplies
`/api/v1/cases`, `/api/v1/units`, `/api/v1/auth/*`, `/api/v1/sos/*` and
`/api/v1/mobile/notifications*` to the frontend; these are **not mounted Python routes**.

The Python `IncidentCreate` requires `incident_type`, `severity`, `location_id` and `reported_at`;
the current report wizard submits a different shape. The Python incident model uses integer IDs,
while the frontend's case model uses UUIDs. `/emergency-resources/` manages responder resources;
it is **not** an SOS submission route. No auth, SOS submission or notification router is mounted
in the Python app inspected here.

**Interpretation of the rest of this document:** its long-standing API examples, lifecycle notes,
mock authorization rules and Appendices A/B record a *legacy frontend target contract*. They are
useful design history and backlog, but no claim there about a live backend is verified against the
current Python service. The `x` marks describe browser/mock implementation only. Before using any
flow live, agree on the Python endpoint and DTO, implement authentication and server-side access
rules, adapt the frontend, then test it end to end. Setting `VITE_USE_MOCKS=false` alone is not an
integration step.

**Effect on the recent ten frontend items:** all ten UI items remain implemented against the mock:
seven SOS controls, map marker grouping, aggregated activity areas, and the notification centre.
The SOS send button is already disabled outside mock mode. Map and notification requests also need
Python adapters or backend endpoints before live use. The ten are **not ten completed live
integrations**. An SOS receipt in the mock never represents real dispatch.

| Python backend | Current frontend expectation | Work before live use |
|---|---|---|
| `/incidents/` with integer ID and `IncidentCreate` | `/api/v1/cases` with UUID and different fields | Agree on status and reporting DTOs; adapt or implement |
| `/locations/`, `/organizations/`, `/emergency-resources/` | `/api/v1/units` with coverage and nearby-unit envelope | Map models and scoping |
| `/response-assignments/` and `/live-locations/` | Case assignment and progress routes | Reconcile workflows and authorization |
| No mounted auth, SOS or notifications routers | `/api/v1/auth/*`, `/api/v1/sos/*`, `/api/v1/mobile/notifications*` | Define and implement secure contracts before live use |

**Next priority:** reconcile these API contracts before adding another flow that depends on the
legacy mock service. Keep the catalogue below as a feature plan, with the mock/live boundary visible.

---

## 0. Current reality (read first)

The original 126-line `App.jsx` SOS mock has been **replaced** (milestone **M0** is done). The
frontend is now a typed React 19 + Vite + Tailwind v4 application with routing, an auth/session
layer, a server-state cache, a design system, a reusable map, and the first real workflows — citizen
case tracking, officer operations, and unit-admin triage and assignment.

> ### Frontend stage — M4.5 complete, M2 (citizen reporting + tracking) complete
>
> M0 is **complete**. M1–M5 are **part-built** (§6): ten routes render working screens against the
> mock API. M6 has an aggregated map activity layer; F11 and F12 remain open.

### Current feature inventory (2026-09-24)

This is a code-level inventory of the frontend on `main` after the frontend-tree repair, before
the visual refresh in §0.5. A rendered route or mock is not proof of a live backend integration.
The detailed checklists in §3 explain the boundaries of each partial feature.

| Area | Implemented | Still to build or confirm |
|---|---|---|
| F1 Auth | Login, role guards, signup, password recovery screens, rotating token refresh, `/login` alias | OTP, officer/unit applications, onboarding, session management; live signup and recovery checks |
| F2 SOS | Citizen SOS console, confirm/cancel, map or manual location, optional responder details, status/history and persistent entry point (mock verified) | Verify request/response and dispatch semantics against the live SOS service; mock does not simulate dispatch |
| F3 Reporting | Four-step report wizard, unit/pin selection, evidence links, draft restore, receipt | Binary uploads require a backend route; offline submit queue remains open |
| F4 Tracking | Citizen case list/detail, status rail and review loop, shared weekly updates, staff evidence view, mock feedback submission | Push status updates, privacy review of live responses |
| F5 Awareness | Notification centre, demo alert feed/detail/confirmation, news, subscriptions and permission opt-in | Python APIs and real push delivery |
| F6 Community | Demo forum/replies, announcements, event RSVP, post reporting | Python APIs, durable moderation workflow and AI-assisted tips |
| F7 Officer | Queue, dispatch/arrive, progress, evidence, weekly narrative, review submission | Investigate transition (backend route required), team view and communications |
| F8 Unit admin | Triage/assignment, case review, closure decisions, evidence verification; demo overview, roster, analytics, finance and settings | Python integrations for demo screens; full verification |
| F9 Super admin | Demo overview, unit registry, audit, analytics/health and settings/export | Python integrations, user governance and impersonation |
| F10 Maps | Case map, marker grouping, aggregated activity areas, unit coverage, report pin picker and basic filters | Advanced filters and offline map fallback; activity overlay live data verification |
| F11 Communication | No frontend screens | Rooms, messaging, calls, presence and sync |
| F12 AI | No frontend screens | Assistant, tips, warnings and labelled case summaries |
| F13 Settings | Demo profile edit (name, phone and HTTPS photo URL) | Python profile write/photo upload, preferences, consent, language, export and deletion |
| F14 PWA | No installable app | Service worker, offline reads/writes, sync and low-bandwidth mode |

Also open from Appendix B: governance (T10), suspect self-view (T11), invites (T12),
session management (T9) and live backend verification. These are separate from visual polish.

**Next ten frontend checklist items, this branch:** F2's seven SOS items, F10 marker grouping,
F10 aggregated activity areas, and F5's full notification centre. The first seven are mock-verified
and require the live integration gate described under F2. This tally counts the clustering work
within the already partially implemented F10 marker item, rather than counting every sub-control
as a separate feature.

**Following ten frontend checklist items (`feature/frontend-awareness-community`):** F4 feedback
submission; F5 alert feed, alert detail/confirmation/share, news, subscriptions and push permission
opt-in; F6 forum/replies, announcements, event RSVP and reporting a post. All are browser/mock
features. Demo content is visibly labelled and stored in memory. The Python backend has no mounted
API for these flows; push permission does not imply delivery, and reporting a post does not reach a
live moderation team. The static preparedness copy is not AI-assisted, so F6 tips stay open.
>
> **M4.5 (mock lifecycle re-alignment) is done.** The former target contract moved the case lifecycle underneath this
> frontend — a post-frontend commit (`36a94ec feat: add accountable case review workflow`) added three
> case states, a submit-for-review / approve-or-request-changes loop, and real weekly-update
> endpoints, and removed the `POST /cases/:id/close` route the shipped **Close case** button called.
> The frontend now matches it: the eight-state vocabulary with unknown states made visible, the
> officer's submit and resubmit path, the administrator's decision surface with required comments and
> the explained self-approval guard, and the weekly narrative read from the real entity. The
> fabricated routes and record types that stood in for those are deleted, not deprecated.
>
> **The one piece left is not a frontend task:** `on_scene → investigating` has no registered route,
> so nothing in the UI performs it — reachable in the demo via the seed only.
>
> **M2 (citizen core) is now complete — reporting and tracking are both built.** Reporters previously
> saw their case *state* and nothing else, and could not create a case from inside the app at all.
> `/report` is now the one creating flow in the product: four steps (what happened → where → optional
> evidence → review) against `POST /cases`, ending in a receipt that carries the tracking ID and
> attaches the links the reporter added. `/cases/:id` is the curated record that follows it: the final
> report, the closure decisions with their comments, the weekly narratives an officer chose to share,
> and a log of status changes with descriptions and personal names withheld. It is a *subset* of the
> staff case page by omission, not by a mode flag — it does not fetch the progress feed or the evidence
> list, so there is no code path on that screen that could render them.
>
> **Two of the wizard's rules come from the contract, not from taste.** `POST /cases` has **no category
> field** — so the wizard cannot ask for one and nothing may read one back; a title and a description
> are what a unit triages on, which is why they are the only two free-text answers that are required.
> And `GetAllCases` scopes an officer's and a unit administrator's list by `unit_id`, so **a case with
> no unit is returned to nobody except its reporter and a super admin** — which is why choosing a unit
> is *required* here even though the endpoint treats it as optional, and why the wizard offers no
> "no preference" option. The location pin is required for a related reason: the contract defaults
> coordinates to zero, and a case pinned at 0,0 sends a unit to the Gulf of Guinea.
>
> **What the curated view did *not* fix, and cannot:** the API still hands the reporter the progress
> feed, the evidence list and timeline descriptions naming officers. The curated view is presentation.
> See the note at the end of F4.
>
> **Next on the list, in order:** (1) `on_scene → investigating`, which is blocked on a backend route
> rather than on frontend work; (2) the remaining admin surfaces still named as placeholders (M5/M7);
> (3) `MapView` cannot draw units or coverage while in `pick` mode, so a reporter choosing a unit reads
> the candidates as a list rather than seeing them on the map.

| Item | Today | Target |
|---|---|---|
| Screens | 10 routes rendering real screens | 40+ across 4 role consoles |
| Routing | ✅ React Router v7, role-gated | Complete |
| API layer | ✅ typed client, mock adapter, 401 handling | Complete |
| State | ✅ React Query (server) + auth context | Complete |
| Design system | ✅ primitives + tokens + 4 states | Grow with features |
| Tests | ✅ 6 unit modules (ISO weeks, API client, assignment rules, status vocabulary, reporter case-log disclosure, report draft) — 83 tests | Component + E2E |
| Mocks | ✅ in-browser mock API (`VITE_USE_MOCKS`), covering the full review loop **and case creation** | Contract tests |
| Case lifecycle | ✅ **all 8 states rendered; the review loop walks end-to-end** (officer submits → admin decides → resubmit → approve → closed). Weekly narratives are read from the real entity, and the reporter now reads their own case through a curated view | Full lifecycle (§0.4) |

### 0.1 What exists right now

**Run it:** `cd frontend && npm install && npm run dev` — mocks are on by default
(`.env.development`), so every screen below works with **no backend**. Sign in with any demo
account: `officer@nativityguard.ng`, `admin@nativityguard.ng`, `citizen@nativityguard.ng`, `super@nativityguard.ng`
(password `password`).

| Route | Screen | State |
|---|---|---|
| `/auth/login` | Sign-in (role quick-fill under mocks) | ✅ |
| `/` | Citizen: your reports + lifecycle stepper | ✅ |
| `/report` | Citizen: **report wizard** — what happened · where (pin + unit) · evidence links · review, then a receipt | ✅ |
| `/cases/:id` | Citizen: one report — curated record, case log, shared weekly narratives | ✅ |
| `/officer/queue` | Case queue — priority sort, filters, search, SLA badge | ✅ |
| `/officer/cases/:id` | Case workspace — Details (incl. closure review) · Progress · Weekly · Evidence | ✅ |
| `/admin/cases` | Admin triage board — attention counters, search, assign/reassign | ✅ |
| `/admin/cases/:id` | Admin case review — **closure decision** · Review history · facts · progress · weekly · evidence · assignment | ✅ |
| `/map` | Operations map — cases + unit coverage + filters | ✅ |
| `/admin/*`, `/super/*` | Remaining admin surfaces — named placeholders | ⏳ M5 / M7 |
| `*` | 404 | ✅ |

The admin console stops at case review on purpose: **the two lifecycle transitions an officer cannot
perform for themselves both live there.** Assignment (`pending → assigned`) is one, and it is what
makes the officer workspace reachable at all — without it no case can be dispatched. The closure
decision (`pending_admin_review → closed`, or back to `admin_changes_requested`) is the other, and it
is the one the accountability workflow exists for. That is why `/admin/cases` leads the unit-admin nav
and is where `homePathForRole('unit_admin')` lands.

> **Note the asymmetry in that pair.** Assignment is a *routing* decision — someone has to do the
> work. Closure approval is a *judgement* about work already done, which is why it carries a required
> comment and a self-approval refusal and assignment carries neither. If a future change makes those
> two feel symmetrical, the accountability has probably leaked out of the closure path.

**Key source files**

```
src/
  App.tsx                  route table (public → protected shell → role-gated)
  main.tsx                 installs the mock API, then renders
  types/api.ts             types for the legacy mock contract; not the Python schemas
  lib/apiClient.ts         typed fetch: bearer token, ApiError, 401 → logout
  lib/queryClient.ts       React Query defaults (no retry on 4xx)
  lib/status.ts            case-status + priority metadata, field rail vs review phase (single source)
  lib/week.ts              ISO-week display helpers: labelling a supplied weekStart, grouping
                           activity into weeks (the server owns the reporting week)
  lib/assignment.ts        assignment rules mirroring the backend guard (pure, tested)
  lib/caseLog.ts           what a reporter's case log shows and withholds (pure, tested)
  lib/reportDraft.ts       the report wizard's localStorage draft: parse, save, clear (pure, tested)
  auth/                    AuthContext (session restore) + RequireRole guard
  components/ui/           Button, BackLink, Card, Chips, Field, Tabs, Modal, Toast, States
  components/layout/       AppShell (sidebar ⇄ bottom nav) + NotificationBell
  components/map/MapView   the one map: view mode + location-pick mode
  components/case/         CaseHeader, CaseFacts, CaseStatusStepper, CaseReviewTrail,
                           WeeklyUpdates, ProgressTimeline, EvidenceGallery, EvidenceUpload,
                           CaseLog (the redacted reporter log)
  components/admin/        AssignOfficerDialog
  hooks/                   useCases, useCaseReview, useWeeklyUpdates, useProgress, useEvidence,
                           useUnits, useOfficers, useNotifications
  mocks/                   fetch-level mock API + Lagos seed data
```

`CaseReviewTrail` is shared the same way `CaseFacts` is: the officer's "what were you asked to
change?" notice and the administrator's decision history are two renderings of one record, and
splitting them would let the two roles disagree about what was decided.

`CaseFacts` is shared by the officer workspace and the admin review on purpose: **an administrator
reviewing a decision must not be looking at a different rendering of it.** Both consoles read the
same case facts from one component (`components/case/CaseFacts.tsx`), which also exports the
`SystemTimeline` used by both.

**The citizen view breaks that pattern deliberately, and the difference is worth stating.** The
officer and the administrator are the same audience — people accountable for a case — so sharing a
component makes them agree. The reporter is a different audience with different rights, and there the
safe default inverts: with a shared component, every field added later becomes visible to reporters
unless somebody remembers to exclude them. `/cases/:id` therefore composes its own header and log from
the shared primitives (`StatusChip`, `CaseStatusStepper`, `ReviewHistory`, `WeeklyUpdates`) and
composes a *subset* — the failure mode of forgetting to add something is a reporter seeing less than
they could, not more than they should.

**The wizard reuses rather than recomposes, and that is the same principle read the other way.** Every
step of `/report` is built out of existing primitives — `MapView` in `pick` mode (including its own
locate control), `Field` + `Input`/`Textarea`/`Select`, `Badge`, `Button`, `Card`, `Skeleton` — because
a reporter filing a report has no narrower set of rights than themselves. There is no second audience
to withhold from, so subset-by-omission does not apply and sharing is the safer default again.

### 0.2 Three decisions worth knowing

1. **The mock layer is a fetch adapter, not MSW.** MSW needs an npm install *and* a generated
   `public/mockServiceWorker.js`; that file cannot be produced without network access, which would
   have left "works offline against mocks" unverifiable. `src/mocks/adapter.ts` + `install.ts`
   match routes in MSW's shape (`method`, `path` with `:params`, `respond({ request, params })`),
   so porting later is mechanical. **Add `msw` back only when the worker file can be generated.**
2. **The Weekly interface is client-derived, and that is now a stopgap rather than the design.**
   When it was written the backend had no weekly entity, so `WeeklyUpdates` grouped the case's real
   progress + timeline timestamps into ISO weeks (`lib/week.ts`) and filed an officer's narrative as
   a progress entry with action `weekly_summary`. **The backend has since grown the real thing**
   (`POST|GET /cases/:id/weekly-update(s)`, `models.CaseWeeklyUpdate`), so this should be rewired —
   see §0.4. The derivation is kept working meanwhile, not deleted, so the tab never goes blank.
3. **The status vocabulary covers all eight states, and an unrecognised state is now *visible* — this
   was the biggest liability in the codebase and it is fixed.** `CaseStatus` is derived from a single
   `CASE_STATUS` const object in `types/api.ts`, so the literal strings live in exactly one place;
   `statusMeta()` returns a `known: false` "Unrecognised state" instead of a confident label, and
   `statusIndex()` returns **-1** rather than clamping to the first step. `FIELD_LIFECYCLE` and
   `REVIEW_PHASE` are separate, because the review phase is a *loop* and cannot be drawn as a sixth
   step on a one-way rail. `src/lib/status.test.ts` guards all of it. The old failure — a case under
   investigation rendered as "Pending — awaiting triage" with the stepper reset to step 1 — is gone.

### 0.3 Known contract gaps the UI does **not** paper over

*(Long-standing gaps, all still accurate. The newer, more urgent drift is §0.4.)*

| Gap | How the UI handles it |
|---|---|
| **`GetOfficersByUnit` is implemented but never routed** — assigning a case needs an `officers`-table id and nothing on this API lists officers | `AssignOfficerDialog` treats a 404 as a first-class "contract gap" state: it says so plainly and offers a manual officer-ID field. Never an empty roster, never a generic error. Mocks mirror the *unregistered handler's* shape so the flow is walkable today. One-line backend fix: register the route |
| `POST /cases/:id/assign` takes an **`officers` id, not a user id** (`officer.UnitID` must equal the case's unit) | `lib/assignment.ts` mirrors the guard (`officerBelongsToUnit`, `assignmentBlocker`) so the dialog blocks a wrong-unit pick before the round trip; officers are a separate entity from `User` in `types/api.ts` |
| `GET /cases/:id/assignments` may not exist on a given build | `AssignmentPanel` splits 404 (`missing` — endpoint not exposed) from network failure (`failed` — explicitly *does not* mean unassigned). Conflating them would let a dropped connection read as "nobody is assigned" |
| `GET /auth/profile` carries no `unitId` | `inferAdminUnitId` derives the admin's unit from the scoped case list, and only when **exactly one** unit is present — a super admin's cross-unit list assumes no roster |
| `GET /cases` may not preload `evidence` | The queue's "evidence unverified" counter is gated on any case actually carrying the array; otherwise the card is hidden rather than reporting a confident zero |
| `GET /cases/:id/progress` has no server-side authorization | Case shows only what it is given; gap documented, backend untouched |
| `GET /cases/analytics` counts `status="resolved"`, which the workflow never sets → `resolutionRate` always 0 | Analytics surfaces are not built on it; KPIs will use `closed` |
| No binary upload endpoint — evidence takes a hosted `fileUrl` | `EvidenceUpload` asks for a link and says so plainly; the report wizard's evidence step does the same, and caps it at three links rather than implying a file picker is coming |
| Notification reads live under `/mobile/notifications*` only | `useNotifications` uses the mobile endpoints |
| **`GetCaseAccountability` is implemented but never routed** — `handlers.GetCaseAccountability` exists in the former review implementation and calls `services.GetCaseAccountability`, but no route registers it. Its model is `models.CaseAccountabilityEvent` | Nothing calls it. Treat it as *available to design against*, not as a live endpoint — see §0.4 item 5 |
| **`POST /cases` accepts a `unitId` that attaches nothing**, and `models.Case.UnitID` is `not null`, so a case whose unit does not parse is stored against the zero UUID. `GetAllCases` scopes officers and unit admins by `unit_id`, so **that case is returned to nobody but its reporter and a super admin** — no unit's queue shows it, and no admin can triage it | The wizard requires a unit (and a location pin) before it will submit, and says why: the endpoint allows omitting both, but a report no unit can see is not a feature. If the backend ever grows a triage pool for unattached cases, the requirement can be relaxed — not before |
| **`GET /cases/:id` returns a reporter more than they should read** — the progress feed, the evidence list, and timeline `description` strings that name officers and administrators ("Assigned to Officer Tunde Balogun.") | The citizen view curates by **omission**: it does not fetch progress or evidence, and `lib/caseLog.ts` renders an actor *role* instead of a name. This is presentation, not enforcement — the same token gets the rest with `curl`. **A real boundary means the backend stops sending it**; until then, do not describe the citizen view as private |

**Backend readiness:** the Python routes do not match the mock contract. Live integration is pending; turning mocks off does not connect these screens.

### 0.4 Lifecycle drift — the backend changed the case workflow under this frontend

**This was the M4.5 milestone, and M4.5 is complete.** It was not new feature work; it was the
frontend catching up to a contract that already shipped. What remains below is the record of what
moved and the one item still blocked on the backend.

This section describes the former frontend target contract; it has not been verified against the
Python routers in this checkout.

**What moved**

- `POST /cases/:id/close` **is no longer registered**, though `handlers.CloseCase` still exists in
  the former workflow implementation. The shipped **Close case** button
  (`OfficerCasePage` → `useCaseActions(id).close` → `POST /cases/:id/close`) therefore 404s against
  the live API. Closure is now an *approval*, not an action.
- Three states were added. Nothing else in the workflow reaches `closed` directly any more:

```
pending → assigned → dispatched → on_scene → investigating
              ↑                                   │
              │                          submit-review (assigned officer)
              │                                   ↓
              │                        pending_admin_review
              │                          │              │
              │     request-changes ─────┘              └───── approve (case admin)
              │            ↓                                       ↓
              └── admin_changes_requested ──┐                  closed
                   (resubmit via submit-review)
```

**The eight states**

| Status | Who sets it | Notes |
|---|---|---|
| `pending` | `POST /cases` | unchanged |
| `assigned` | `POST /cases/:id/assign` (case admin) | unchanged |
| `dispatched` | `POST /cases/:id/dispatch` | unchanged |
| `on_scene` | `POST /cases/:id/arrive` | unchanged |
| `investigating` | *see caveat below* | new — the entry condition for review |
| `pending_admin_review` | `POST /cases/:id/submit-review` | new |
| `admin_changes_requested` | `POST /cases/:id/review/request-changes` | new — can be resubmitted |
| `closed` | `POST /cases/:id/review/approve` | new — sets `closedAt`, `closedBy`, `approvedBy` |

> **Verified 2026-09-14** against `backend/models/` — the literals above are contract, not inference:
> the former case-workflow model defines all eight `CaseStatus*` constants exactly as written. The JSON tags
> are **camelCase**, not snake_case: `json:"citizenVisible"`, `json:"weekStart"`, `json:"weekEnd"`.
> And there are **three** review decisions, not two — `approve`, `request_changes`, and
> `deescalate` (the former case-review model). See "Still open with the backend" below for what
> `deescalate` implies.

**The review loop, exactly**

| Endpoint | Rule |
|---|---|
| `POST /cases/:id/submit-review` | Assigned officer only. Requires status `investigating` **or** `admin_changes_requested` (else **409**, and the response body carries the current `status`). Requires a non-empty `finalReport` on the case (else **400**). → `pending_admin_review` |
| `GET /cases/:id/review` | Readable by case admin, assigned officer, or the reporter. Returns `{ caseId, status, reviews[] }`, oldest decision first |
| `POST /cases/:id/review/request-changes` | Case admin only. Body `{ comment }` **required**. Requires `pending_admin_review`. Records a `CaseReview` decision and → `admin_changes_requested` |
| `POST /cases/:id/review/approve` | Case admin only. Body `{ comment }` **required**. Requires `pending_admin_review`. **403 if the approver is the case's own assigned officer** ("An assigned officer cannot approve their own case closure"). Transactionally → `closed` + `closedAt`/`closedBy`/`approvedBy` |

*"Case admin"* = a super admin, **or** any user with an active `UnitMembership` on the case's unit
whose role is `admin`. Unit membership alone is **not** enough — that is the same principle as
§0.3's unrouted-roster gap, applied to authorization.

**Weekly updates are now real**

| Endpoint | Rule |
|---|---|
| `POST /cases/:id/weekly-update` | Assigned officer only. `summary` + `investigation` required; `actionsTaken`, `findings`, `evidenceSummary`, `outstandingActions`, `nextSteps` optional. Rejected on a `closed` case. **One per officer per reporting week** — the week is Monday 00:00 **UTC** to Sunday, and a second submission returns **409 with the existing update in the body**, not an error string. Responds **201** `{ message, update }` |
| `GET /cases/:id/weekly-updates` | Readable by case admin, assigned officer, or reporter. Returns `{ caseId, updates[] }` ordered by `weekStart`. A reporter who is neither admin nor assigned officer **only** receives entries with `citizenVisible = true` |

That last clause is a **privacy boundary the UI must not blur**: the same case can legitimately yield
different weekly updates to two different roles. Render what the endpoint returns; never merge a
cached officer view into a citizen view.

**What the frontend has to do**

1. ~~**Widen the vocabulary**~~ ✅ **Done.** `CaseStatus` is derived from one `CASE_STATUS` const object
   (`types/api.ts`); `FIELD_LIFECYCLE` / `REVIEW_PHASE` / `CASE_STATUS_ORDER` / `CASE_STATUS_META`
   live in `lib/status.ts` with three new `--color-status-*` tokens in `index.css`. The silent
   fallback is gone: `statusMeta()` reports `known: false` and the UI says **"Unrecognised state"**
   with the raw value shown, and `statusIndex()` returns **-1** instead of clamping to step 1.
2. ~~**Replace the closure flow**~~ ✅ **Done.** `useCaseActions(id).close` is deleted, the mock's
   fabricated `POST /cases/:id/close` route is deleted, and the officer's Close modal is replaced by
   **Submit for review** (gated on a non-empty final report, disabled *with the reason shown* when
   there isn't one). The 409's returned `status` is read out of `ApiError.body` and surfaced as
   "the case is now *X*".
3. ~~**Build the admin decision surface**~~ ✅ **Done.** `AdminCaseReviewPage` now owns the closure
   decision. The panel sits **above the tabs**, not inside one: an administrator opening a case that
   is waiting on them should not have to go looking for the thing they came to do. Both outcomes open
   a required-comment modal whose primary action is disabled *with the reason written out* until a
   comment exists. A second **Review** tab carries the full decision history
   (`GET /cases/:id/review`), the next-step line, and the same actions.
   The self-approval rule is an *explained* disabled action: "You are the assigned officer on this
   case, so a second administrator must approve its closure — nobody signs off the investigation they
   ran. You can still request changes." That last sentence matters; the rule blocks approval, **not**
   the request-changes path, so an admin who is also the assigned officer is not stuck with a case
   they cannot move at all. `selfAssigned` is derived from `case.assignedTo` rather than the
   assignment list, because that list is a separate endpoint allowed to 404 — the guard must not
   vanish when it does.
4. ~~**Rewire `WeeklyUpdates`**~~ ✅ **Done.** The tab now reads the real entity through
   `useWeeklyUpdates` / `useFileWeeklyUpdate` (`GET|POST /cases/:id/weekly-update(s)`) and the
   fabricated write is gone: no more summaries filed as a `weekly_summary` *progress* entry, and
   `WEEKLY_SUMMARY_ACTION` is deleted from `activity.ts` (legacy rows would still render as "Weekly
   summary" via the Title Case fallback).
   **The old design's actual bug:** a filed narrative was stored as a progress record, so the week's
   summary panel and the week's timeline listing rendered *the same text twice*, while
   `models.CaseWeeklyUpdate` — seven fields, a server-computed week, a privacy flag — went unread.
   The tab is now two clearly separated sections: **Officer narratives** (filed records) and
   **Activity by week** (assembled from progress + status events, labelled as a reading aid so nobody
   mistakes it for something a person filed).
   The duplicate-week **409 is a designed state**, not an error: `conflictingWeeklyUpdate()` extracts
   the existing update from `{ error, update }` and the screen shows what is on file. The ordinary
   case is caught *before* the click — the composer is replaced by "This week's update is already
   filed" as soon as the officer has one for the current UTC week — and the 409 handler covers the
   race where the week rolls over with the form open. A **closed** case answers 409 too, so
   `isClosedCaseRefusal()` distinguishes it; conflating the two would tell an officer they had filed
   a narrative they never wrote.
   `lib/week.ts` stays display-only, exactly as this item required.
5. ~~**Extend the mocks**~~ ✅ **Done.** `src/mocks/handlers.ts` now serves the full review loop
   (submit-review, review, request-changes, approve, weekly-update, weekly-updates) with real
   authorization checks, and `seed.ts` seeds a case in each new state — including one in
   `pending_admin_review` **with a prior request-changes round already in its history**, so the admin
   surface has something to decide on and a decision record to render.

**Still open with the backend** (do not design against these yet):

- **What performs `on_scene → investigating`?** No registered route sets it explicitly, and it is the
  review workflow's entry condition, so *something* must. The only candidate is `PUT /cases/:id`
  (`handlers.UpdateCaseStatus`), which has since been read and does **zero status validation** —
  `caseObj.Status = input.Status`. That is also an authorization hole: the same call can set
  `"closed"` directly and bypass the whole approval workflow. The frontend must not depend on that
  path as a blessed transition; `investigating` is reachable in the demo only because the seed sets it.
- **`CaseReviewDecisionDeescalate` (`"deescalate"`)** exists in the former case-review model next to
  `approve` and `request_changes`, but no route that records it has been found. Do not build a
  de-escalation affordance; the review history humanises it ("De-escalated") rather than pretending
  only two decisions exist.
- Whether the orphaned `CloseCase` handler, the unrouted `GetCaseAccountability`
  (model: `models.CaseAccountabilityEvent`), and `GetOfficersByUnit` are pending registration or
  deliberately retired.

---

### 0.5 Visual direction — Dawn Canopy

**Status: first pass implemented in this branch; visual review on mobile and desktop pending.**
Replace the old navy-blue impression with a sheltered forest at dawn: deep evergreen surfaces,
warm sunlight for the primary action, a soft sky glow and a few still points of light on the public
front door. The scene should suggest visibility, calm and a place to return to. The system is a
public-safety tool, so the atmospheric treatment belongs in hero and welcome areas; reporting,
maps, forms and case decisions retain quiet, solid, highly legible surfaces.

| Role | Token | Colour | Use |
|---|---|---|---|
| Background | `--color-base` | `#10221d` | App canvas; replaces navy |
| Panels | `--color-surface` / `--color-surface-hi` | `#193129` / `#254137` | Cards and raised controls |
| Primary action | `--color-signal` / `--color-signal-ink` | `#f4cb78` / `#19251c` | Warm dawn light, with dark button text |
| Text | `--color-ink` / `--color-ink-muted` | `#f8f5e9` / `#c2d4c7` | Main and supporting copy |
| Boundaries | `--color-border` / `--color-border-hi` | `#355348` / `#527466` | Structure without bright outlines |
| Emergency | `--color-emergency` | `#ef4444` | SOS and urgent states only; never decorative |

The first pass updates shared CSS tokens and gives the public landing hero a lightweight CSS dawn
sky and forest horizon. It adds no network image or animation. Keep lifecycle colours distinct and
keep all status labels visible so meaning never depends on colour. Verify text, focus, buttons,
status chips and maps for contrast on a phone in daylight; respect reduced motion if atmosphere is
extended. Do not apply stars or scenery behind emergency controls or input fields.

**Sequence:** finish and review this visual direction before starting another feature from §3.
Keep the theme separate from the SOS, governance, communication and PWA milestones.

---

## 1. Product intent & engagement goals

Nativity Guard must be **chosen** by communities, not mandated. That only happens if the
experience is trustworthy, fast, and human. Engagement is a design output, not a growth hack.

**Engagement goals**

| Goal | How the UI earns it |
|---|---|
| Trust | Transparent case status, timestamps, assigned unit/officer always visible |
| Return visits | Notifications that matter; visible progress on every case |
| Fast reporting | ≤ 3 taps from home to a submitted report; ≤ 1 tap to SOS |
| Field speed | Officer actions reachable in one hand, high contrast, offline-tolerant |
| Local relevance | Multilingual copy, local units, local alerts |
| Safety | One-tap SOS with confirm/cancel; clear privacy and consent |

**Anti-goals:** dark patterns, notification spam, manufactured urgency, vanity gamification,
surveillance aesthetics.

---

## 2. Roles & information architecture

Four role-gated consoles share one design system.

| Role | Primary nav | Home screen |
|---|---|---|
| **Citizen** | Home · Report · Track · Alerts · Profile | SOS-first dashboard |
| **Officer** | Queue · Active · Map · Comms · Profile | Assigned-case queue |
| **Unit Admin** | Overview · Cases · Officers · Analytics · Config | Ops overview + dispatch board |
| **Super Admin** | Governance · Units · Users · Audit · Analytics | Platform control + health |

**Route map (target)**

```
/                     → role router → role home
/auth/login · /auth/register · /auth/otp · /auth/forgot
/onboarding/*         → profile, unit application, gov ID, medical info, prefs

# Citizen
/report · /report/new · /track · /track/:caseId · /alerts · /alerts/:id
/community · /community/posts/:id · /community/events
/map · /sos · /ai-assistant · /settings · /profile

# Officer
/officer/queue · /officer/cases/:id · /officer/cases/:id/evidence
/officer/map · /officer/comms · /officer/comms/:roomId

# Unit admin
/admin/overview · /admin/cases · /admin/cases/:id · /admin/officers
/admin/verification · /admin/analytics · /admin/finance · /admin/settings

# Super admin
/super/users · /super/users/:id · /super/units · /super/units/:id
/super/audit · /super/analytics · /super/settings · /super/health
```

---

## 3. Feature catalogue

Each feature lists: **screens**, **required elements**, **states**, and **APIs**.
Checklist marks build progress. `[ ]` to build · `[~]` partial · `[x]` done.

### F1 — Authentication & onboarding

**Screens:** login, citizen signup, forgot/reset; OTP verify and onboarding wizard pending.

- [x] Email/phone + password login; role-aware redirect
- [x] Register a citizen at `/auth/signup` (no role selector; backend assigns the role)
- [ ] Officer application and onboarding
- [x] Password recovery screens (6-digit code; live delivery and reset still to verify)
- [ ] OTP send / verify / resend (countdown, rate-limit messaging)
- [ ] Unit application flow (search nearby units, select, apply)
- [ ] Government ID submission (camera/file, status pending/verified/rejected)
- [ ] Medical info intake (explicitly optional, privacy notice)
- [ ] Profile completion + onboarding checklist
- [x] Session: JWT storage, session restore on boot, logout, 401 → login
- [x] Refresh token flow via `POST /auth/refresh` with one shared in-flight rotation attempt;
      rejected refresh signs out

**States:** idle · loading · invalid credentials · OTP expired · rate-limited · pending verification
**APIs:** `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`,
`POST /auth/refresh`, `POST /auth/change-password`, `GET /auth/profile`,
`POST /otp/send|verify|resend`, `GET /units/nearby`, `POST /units/apply`,
`POST /units/government-id`, `GET|PUT /settings/onboarding`

---

### F2 — Citizen home & SOS

**Screens:** citizen home, SOS console, SOS active/status, SOS history.

- [x] Emergency SOS entry opens the console; explicit confirm/cancel before `POST /sos/send`
- [x] Geolocation with permission-denied fallback, map pin and manual coordinates
- [x] Receipt with server-returned tracking ID and status; `/sos/my` refreshes every 15 seconds
- [x] Optional emergency-contact details and medical information, cleared from the form after success
- [x] Priority select (default high) and optional preferred unit; preference never claims dispatch
- [x] SOS history and per-alert status (`pending`, `dispatched`, `resolved`, `escalated`)
- [x] Persistent, labelled SOS entry in the citizen shell on desktop and mobile

**Live integration gate:** these seven frontend items work against the in-browser mock. The
checkout contains no Python SOS submission route to verify the mock contract, so the live payload and
response shape, consent/storage handling for medical details, and real dispatch lifecycle must be
checked before enabling SOS in a pilot. With `VITE_USE_MOCKS=false`, send is disabled and the page
explains that the service is unavailable. The mock stores new SOS requests as `pending` and never
pretends a unit was dispatched. A receipt confirms submission only. The optional details stay out
of local storage. The `GET /sos/:id` mock exists for contract exploration; the console reads the
history endpoint. `PUT /sos/:id/status` remains a staff-side API, not a citizen control.

**Demo check:** run `npm run dev`, sign in as the citizen mock account, open **SOS**, enter a manual
location or place the pin, press **Prepare SOS**, then cancel once to verify nothing is sent. Confirm
on the second attempt and check that a `pending` receipt appears in history. The in-memory mock
history resets on page reload; it is not an offline or durable emergency queue.

**States:** locating · active · escalated · resolved · permission-denied · offline-queued
**APIs:** `POST /sos/send`, `GET /sos/my`, `GET /sos/:id`, `PUT /sos/:id/status`

---

### F3 — Incident reporting

**Screens:** report wizard (multi-step), report review, success/receipt, my reports.
*(All four exist: the wizard and its review step and receipt are `/report`; "my reports" is the citizen
home, documented under F4.)*

The wizard shipped as **four steps, not the five sketched below**, and every departure is a contract
fact rather than a design preference:

- **The category/type step is gone.** `POST /cases` has no category field and `models.Case` has no
  category column, so a step asking for one would collect an answer that is discarded on the way out
  and unreadable on the way back. A title and a description are what triage actually reads, which is
  why those two are the only required answers.
- **"Where" also picks the responding unit**, because the pin is exactly what decides which units are
  candidates (`GET /units/nearby`, sorted by distance, split by whether the unit's coverage reaches the
  point).
- **Media capture became evidence *links*, collected before the case exists and attached after it.**
  There is no binary-upload endpoint, so a file must already be hosted somewhere the reporter can
  share. The links are staged in step 3, but `POST /evidence/upload` needs a `caseId` — so they are
  sent in step 4's wake, each reported separately on the receipt and each retryable on its own. A
  failed link never becomes a failed report.
- **The public/private choice is gone.** `IsPublic` is hardcoded `true` on create, so the reporter has
  no such choice to make and the review step says the record is public instead of offering a toggle the
  server would ignore.

- [x] Step 1 — **what happened**: title + description, sized and hinted for a report rather than a form
- [x] Step 2 — **where**: `MapView` in `pick` mode with "use my location", an optional landmark field
      for when the pin is slightly off, and the responding unit. **Both halves are required, and the
      refusal states the real reason** — the endpoint accepts neither as mandatory, but a case with no
      `unit_id` is returned by no unit's `GET /cases` (see §0.3), and one pinned at 0,0 sends a unit to
      the Gulf of Guinea
- [x] Step 3 — **evidence**: up to three hosted links, each with a type and a remove control, each
      validated as `http(s)` before Submit will enable. The copy says the file must be hosted already;
      it does not imply a file picker is coming
- [x] Step 4 — **review**: every answer shown with an **Edit** that returns to its step, a
      "what happens when you send this" block (recorded as `pending`, a tracking ID is issued, reports
      are public records), and the flow's single Send button
- [x] Location picker (map + "use my location" + manual landmark)
- [x] Priority hint (advisory only — triage is server-side): the amber "someone is in danger or injured
      now" checkbox maps to `priority: "high"` → **P2**. It is **not** an SOS path — `isSOS` bands P1
      and triggers an immediate dispatch attempt, and that belongs to an arming flow with its own
      confirm step. The copy says the checkbox marks the report for faster triage and does not send
      anyone
- [x] Draft auto-save — `localStorage` via `lib/reportDraft.ts` (pure and tested), written as the form
      is typed and cleared the moment a case is created, so a failed evidence attach can never leave a
      reporter able to file the same report twice. A restored draft reopens at step 1 and says so; a
      corrupt or foreign value opens a fresh form rather than half-restoring one
- [–] **Offline queue with sync — not built, deliberately.** A draft is not a report: it has no tracking
      ID and no unit has seen it, and nothing in the UI may let a saved draft read as "sent". A real
      queue needs ordering, retry and dedupe semantics, and a local draft is not a partial version of
      one
- [x] Receipt with **Tracking ID** + "track this report" CTA, rendered from the create response rather
      than a refetched list, so a slow refresh cannot make a filed report look unfiled
- [x] Validation inline, never a raw error toast: field-level errors, a disabled Continue/Submit whose
      reason names the missing thing, and a failed create that says outright that nothing was created
      and every answer is still on screen

**States:** drafting · validating · submitting · partial-evidence-failed · submitted · create-failed
**APIs:** `POST /cases`, `POST /evidence/upload`, `GET /units/nearby`, `GET /cases`, `GET /cases/:id`

---

### F4 — Case tracking & feedback

**Screens:** my cases list, case detail, timeline, evidence viewer, feedback form.
*(The list and the case detail exist; the evidence viewer and feedback form do not.)*

- [x] Case list with status chips, priority, last-updated — each card is the link to the detail view
- [x] **Case detail (`/cases/:id`, citizen-only) — a curated record, built as a *subset* of the staff
      case page by omission.** It renders status + stepper, the final report, the closure decisions
      with their comments, the weekly narratives shared with the reporter, and a **case log of status
      changes only**. It calls three endpoints (`GET /cases/:id`, `/review`, `/weekly-updates`) and
      never fetches the progress feed or the evidence list, so the "Activity by week" reading aid is
      *absent* rather than hidden — there is no code path that could render a progress note. Names are
      withheld: the log shows an **actor role** (`You` / `Assigned officer` / `Unit staff`) derived
      from ids the case already carries, and never renders timeline `description` text, which is
      written for an internal audience and names people ("Assigned to Officer Tunde Balogun."). The
      responding **unit** is named once in the header, not per row — the timeline says which *user*
      acted, never which unit, so a per-row unit stamp would be a guess
- [x] **Lifecycle stepper** — rebuilt as two instruments: a one-way *field rail* over
      `FIELD_LIFECYCLE` (pending → assigned → dispatched → on scene → investigating) and a separate
      *review phase* block, because approve/request-changes is a loop and cannot be drawn as a sixth
      step. An unrecognised status renders its raw value instead of pretending to be step 1
- [x] Progress timeline (action, description, officer, time) — **officer and admin views only**
- [~] Assigned unit/officer visibility (respecting privacy) — the staff views show the officer when
      named on the timeline; the citizen view shows the **unit** and no officer identity at all
- [x] Evidence gallery (thumbnails, type, verification badge) — **officer and admin views only**
- [x] Resolution summary (final report) — rendered whenever present, headed by the case's actual
      review position (draft / submitted for review / changes requested), not only when closed
- [x] Weekly updates — the real entity: `GET|POST /cases/:id/weekly-update(s)` over
      `models.CaseWeeklyUpdate`, with all seven fields, the server-computed week, and the
      duplicate-week 409 as a designed "already filed" state. Display-only helpers
      (`lib/week.ts`) remain for grouping activity into weeks and labelling a supplied `weekStart`.
      **The reporter reads the same endpoint, server-filtered to `citizenVisible = true`** — that
      boundary now has a caller, and the empty state says "nothing has been shared with you yet"
      rather than "nothing was filed", because on a filtered feed those are different facts
- [x] Review history — the shared `<ReviewHistory>` renders every decision with comment, actor and
      timestamp, and is wired into the **officer** page (where "what were you asked to change?"
      matters most), the **admin** Review tab, and now the **citizen case detail**, so "why is my case
      still open" has an answer addressed to the person who asked it. A closed case with no recorded
      decision is described as such rather than as "not decided yet"
- [x] Feedback: a reporter can rate and comment on their closed case once in the demo; the mock
      rejects non-reporters, open cases and duplicate submissions. Live Python integration pending
- [ ] Push/in-app updates on every status change

**States:** empty · loading · not-found · forbidden · closed-readonly · awaiting-review · changes-requested
**APIs:** `GET /cases`, `GET /cases/:id`, `GET /cases/:id/timeline`, `GET /cases/:id/progress`,
`GET /cases/:id/review`, `GET /cases/:id/weekly-updates`, `GET /evidence/case/:caseId`,
`POST /cases/:id/feedback`, `POST /ratings`, `GET /ratings/units/:unitId`

> **The citizen view is presentation, not a privacy boundary.** `GET /cases/:id` already returns the
> reporter the progress feed, the evidence list and timeline descriptions that name officers and
> administrators, and the mock's `canSeeCase` permits it. Not fetching two endpoints and not rendering
> a third field is a decision about what a reporter should have to read; it is not a control on what
> they can obtain, because the same token gets the rest with `curl`. If this must be enforced, the
> backend has to stop sending it.

**States:** empty · loading · not-found · forbidden · closed-readonly · awaiting-review · changes-requested
**APIs:** `GET /cases`, `GET /cases/:id`, `GET /cases/:id/timeline`, `GET /cases/:id/progress`,
`GET /cases/:id/review`, `GET /cases/:id/weekly-updates`, `GET /evidence/case/:caseId`,
`POST /cases/:id/feedback`, `POST /ratings`, `GET /ratings/units/:unitId`

---

### F5 — Alerts, news & notifications

**Screens:** alert feed, alert detail, news feed, notification center, subscriptions.

- [x] Demo community alert feed with severity labels and empty/loading/error states
- [x] Demo alert detail with location, idempotent confirm action and explicit demo-link sharing
- [x] Demo news feed with `news`/`alert` item types and clear source labeling
- [x] Notification center — bell + popover and `/notifications` page with all entries, individual
      mark-read, bulk mark-read and loading/empty/error states
- [x] Demo subscription management for areas, categories and preferred channels
- [x] Browser permission opt-in and demo device register/unregister; no push delivery until a
      Python service implements it

**APIs:** `GET /alerts`, `GET /alerts/:id`, `POST /alerts/:id/confirm`, `POST /alerts/subscribe`,
`GET /alerts/subscriptions`, `GET /alerts/news`, `GET /news`, `GET /news/:id`,
`POST /notifications/register`, `DELETE /notifications/unregister`

---

### F6 — Community & prevention

**Screens:** community hub, forum post, announcements, events, event detail/RSVP.

- [x] Demo forum: post list, inline detail, replies and create post
- [x] Demo announcements feed
- [x] Demo events with RSVP toggle and attendee count
- [ ] Prevention/safety tips surface (AI-assisted, cached)
- [x] Demo post-report action, visibly recorded only in memory; live moderation queue pending

**APIs:** `POST|GET /community/posts`, `GET /community/posts/:id`, `POST /community/replies`,
`POST|GET /community/announcements`, `POST|GET /community/events`, `POST /community/events/:id/rsvp`

---

### F7 — Officer console

**Screens:** queue, active case, case actions, progress log, evidence capture/verify, map, comms.

- [x] Assigned-case queue (priority-sorted, SLA badges)
- [x] Case detail with one-tap **Dispatch** and **Arrive** (state-gated)
- [x] Progress logging (action type + description)
- [x] Evidence attach (by link) + verify evidence (with badge)
- [ ] **Investigate** — the `on_scene → investigating` step. Nothing reaches the review workflow
      without it, and the frontend has no action for it at all (§0.4, open question). Deliberately
      still unbuilt: no registered route performs this transition, so any button would be inventing
      one. It is reachable in the demo only because the seed places cases in `investigating`
- [x] **Submit case for review** — captures the final report then `POST /cases/:id/submit-review`
      from `investigating` or `admin_changes_requested`. Surfaces the 409's returned `status` ("the
      case is now *X*") via `ApiError.body`, and the primary action is disabled **with the reason
      written out** when there is no final report — an administrator cannot approve a closure they
      cannot read
- [x] **Read the admin's decision** — a case returning as `admin_changes_requested` shows the
      reviewer's instruction as a notice above the case, not as a log line. `<ReviewNotice>` names the
      admin, the time, and the full comment; the closure-review section repeats it in history
- [x] Weekly officer narrative — filed through `POST /cases/:id/weekly-update` by the **assigned
      officer only** (the endpoint is gated to it, so `canFile` is false for an administrator and the
      composer is absent rather than refused). Required `summary` + `investigation`; the other five
      fields sit behind an "Add detail" disclosure
- [x] ~~Close case with final report~~ — **removed.** `POST /cases/:id/close` is no longer routed, so
      the button could only ever 404. Deleted from `useCaseActions`, `OfficerCasePage` and the mock;
      superseded by *submit for review* (§0.4 item 2). Do not re-add it
- [~] Team view (primary/investigator/support) — assignment list + roles rendered in admin case
      review; the officer-facing team view is open
- [x] Responder map with lawful unit/location context
- [ ] Comms: rooms, messages, calling entry points

**States:** action-not-allowed (wrong state) surfaced clearly, never silently disabled
**APIs:** `POST /cases/:id/dispatch`, `POST /cases/:id/arrive`, `POST /cases/:id/progress`,
`GET /cases/:id/progress`, `POST /cases/:id/submit-review`, `POST /cases/:id/weekly-update`,
`POST /evidence/upload`, `PATCH /evidence/:id/verify`, `POST /cases/:id/assign`,
`GET /cases/:id/assignments`

**Open question for the backend:** `canAddProgress` in `lib/status.ts` was widened to include
`investigating` and `admin_changes_requested`, on the assumption that the backend's guard was widened
too when those states were added. **That assumption is unconfirmed.** A deployed guard matching only
`dispatched | on_scene` would mean an investigating officer cannot log progress at all — which would
empty the very evidence the reviewing admin is asked to judge. The mock mirrors the widened version,
so the demo will *not* reveal a mismatch. Confirm against the handler before relying on the tab.

---

**Demo boundary for the next ten items:** The checked admin and profile workflows below run only with `VITE_USE_MOCKS=true`; they are session-scoped sample screens. Their actions do not persist to Python, move money, provision real officers, change live dispatch, or provide live monitoring. Profile photo uses an HTTPS URL, not an uploaded image. Live integrations remain open.

### F8 — Unit admin console

**Screens:** overview, dispatch board, case management, officers, verification queue, analytics,
finance, unit settings.

- [x] Ops overview: open/assigned/dispatched cards, response-time KPIs **(session demo only; Python API integration pending)**
- [x] Dispatch board: unassigned queue → assign to officer
      *(`/admin/cases` — attention counters are filters, not decoration: each one toggles the list)*
- [~] Case review: facts, progress, weekly, evidence, assignment, **the closure decision and its
      history** — done; the per-case audit trail (`CaseAccountabilityEvent`, unrouted) is open
- [x] **Review decisions — the admin's half of the accountability loop.** A case arriving as
      `pending_admin_review` gets a decision panel above the tabs: approve closure, or request
      changes, each with a **required comment** the contract will not accept otherwise, plus the
      decision history (`GET /cases/:id/review`) in a dedicated **Review** tab so a second
      administrator can see what the first one decided and why. This is where "someone is accountable
      for closing this" becomes visible instead of merely true
- [x] **Self-approval guard, explained.** An administrator who is also the case's assigned officer
      sees **Approve closure** disabled, with the rule stated underneath — "you are the assigned
      officer on this case, so a second administrator must approve its closure" — plus the reminder
      that **Request changes** still works. A 403 from the server is still handled, but as a
      fallback for a disagreement between server and UI, not as the primary path. Never a silent
      grey-out (§3 law 5)
- [x] `pending_admin_review` in the triage counters — **"Awaiting your decision"** is now the first
      attention card on `/admin/cases`, with the hint "Submitted for closure — only an admin can move
      these", and it filters the list like the other counters (`Focus = 'to_decide'`)
- [x] Officer roster: add/edit, status, workload *(blocked by the unrouted `GetOfficersByUnit`)* **(session demo only; Python API integration pending)**
- [~] Verification queue: evidence verify shipped inside case review; memberships and gov IDs open
- [x] Analytics: volume, response/dispatch/arrival, resolution, workload **(session demo only; Python API integration pending)**
- [x] Finance: accounts, donations, transactions + approvals, budgets, reports **(session demo only; Python API integration pending)**
- [x] Unit config, operational radius, contact info **(session demo only; Python API integration pending)**

**APIs:** `GET /cases`, `POST /cases/:id/assign`, `GET /cases/:id/review`,
`POST /cases/:id/review/approve`, `POST /cases/:id/review/request-changes`, `GET /cases/analytics`,
`GET|POST|PUT /units`, `GET|POST /bank/*`, `GET|POST /finance/*`, `GET /audit/*`,
`POST|GET /unit-verification`* *(verification endpoints per the former mock target contract)*

---

### F9 — Super admin console

**Screens:** governance dashboard, users, user detail, units, audit search, analytics, health, settings.

- [ ] Users: list, search, role change, suspend/activate, impersonate/stop
- [x] Units: registry CRUD **(session demo only; Python API integration pending)**
- [x] Audit: searchable activity + audit logs (actor, action, entity, time) **(session demo only; Python API integration pending)**
- [x] Platform analytics + system health **(session demo only; Python API integration pending)**
- [x] System settings, templates, data exports **(session demo only; Python API integration pending)**
- [ ] Impersonation banner (unmissable "you are impersonating") + one-click exit

**APIs:** `GET /admin/users`, `GET /admin/users/:id`, `PUT /admin/users/:id/role`,
`POST /admin/users/:id/suspend|activate|impersonate`, `POST /admin/stop-impersonate`,
`GET /admin/stats`, `GET /audit/activities`, `GET /audit/logs`, `GET /audit/health`,
`GET|PUT /settings/:key`, `GET /settings/templates/:name`, `POST|GET /settings/exports`

---

### F10 — Maps & GIS

**Screens:** incident map, unit coverage, hotspot view, case location picker.

- [x] Incident markers grouped at close screen positions, with count and click-to-zoom; individual
      status-coloured markers appear when separated
- [x] Unit coverage radii
- [x] Optional aggregated activity overlay (at least five cases per area, low zoom only); explicitly
      labelled as observed report concentration, never predicted danger or individual profiling
- [x] Case location picker — `MapView mode="pick"` is wired into the report wizard
- [~] Filters: period, category, status, unit — status + free-text search shipped

**States:** map-unavailable fallback to list; tiles offline-degraded
**APIs:** `GET /cases` (geo fields), `GET /units/nearby`, `GET /units/by-location`,
`POST /ai/map-insights`, `POST /ai/predict-hotspots`

---

### F11 — Communication

**Screens:** room list, chat room, call UI, walkie-talkie mode.

- [ ] Room list per unit; create room
- [ ] Real-time messaging (WebSocket) with optimistic send + delivery state
- [ ] Voice call + WebRTC signaling UI
- [ ] Presence/room-status indicators
- [ ] Offline message queue + sync on reconnect

**APIs:** `POST|GET /communication/rooms`, `POST /communication/messages`,
`GET /communication/rooms/:roomId/messages`, `POST /communication/calls`,
`PUT /communication/calls/:id/end`, `POST /communication/sync`, `/ws`

---

### F12 — AI assistant

**Screens:** assistant chat, tips panel, security warning banner, case summary.

- [ ] Chat assistant with role-aware answers
- [ ] Context-aware safety tips (location, time, role)
- [ ] Security warning surface (banner/card)
- [ ] Case summary (officer-facing, clearly labelled AI-generated)
- [ ] AI outputs always labelled + human-in-the-loop; never authoritative

**APIs:** `POST /ai/chatbot`, `POST /ai/smart-tips`, `POST /ai/security-warning`,
`POST /ai/analyze-location`, `POST /ai/analyze-news`

---

### F13 — Settings, profile & privacy

**Screens:** profile, notification prefs, privacy, language, data export, delete account.

- [x] Profile edit (name, contact, photo) **(session demo only; Python API integration pending)**
- [ ] Notification channel prefs (push/SMS/email/in-app)
- [ ] Privacy controls + consent explanations in plain language
- [ ] Language selector (English + Nigerian languages, Pidgin)
- [ ] Request data export; request deletion
- [ ] Medical info (separate, protected, explicit consent)

**APIs:** `GET|PUT /auth/profile`, `GET|PUT /settings/onboarding`, `GET|PUT /settings/:key`,
`POST|GET /settings/exports`

---

### F14 — Offline & field mode (PWA)

- [ ] Installable PWA + service worker
- [ ] Offline read cache for tracked cases + alerts
- [ ] Offline write queue (reports, progress, messages) with sync indicator
- [ ] Low-bandwidth mode; image compression before upload
- [ ] Mobile sync endpoint integration

**APIs:** `GET /mobile/config`, `GET /mobile/dashboard`, `GET /mobile/sync`,
`GET|PUT /mobile/notifications*`, `POST /mobile/crash-report`

---

## 4. Cross-cutting requirements

| Area | Requirement |
|---|---|
| **Design system** | Tokens, type scale, spacing, elevation, motion, iconography, component library |
| **States** | Every data surface: loading (skeleton), empty (illustrated + action), error (retry), unauthorized, offline |
| **Accessibility** | WCAG AA contrast, keyboard nav, focus-visible, screen-reader labels, large touch targets |
| **i18n** | Externalized strings, RTL-safe layout, multilingual copy |
| **Performance** | Route-level code splitting, image optimization, list virtualization, budget (see `frontagent`) |
| **Security (client)** | No secrets in client, token hygiene, safe rendering (no `dangerouslySetInnerHTML` on user content) |
| **Observability** | Error boundary + crash report, key funnel events (report start→submit, SOS→resolved) |
| **Motion** | Purposeful only; respects `prefers-reduced-motion` |

---

## 5. API mapping summary

| Feature | Key endpoints |
|---|---|
| Auth | `/auth/*`, `/otp/*` |
| Cases | `/cases`, `/cases/:id`, `/cases/:id/{timeline,progress,assign,dispatch,arrive,feedback}` |
| Case review | `/cases/:id/submit-review`, `/cases/:id/review`, `/cases/:id/review/{approve,request-changes}` |
| Weekly updates | `/cases/:id/weekly-update`, `/cases/:id/weekly-updates` |
| Evidence | `/evidence/upload`, `/evidence/case/:caseId`, `/evidence/:id/verify` |
| SOS | `/sos/send`, `/sos/my`, `/sos/:id/status` |
| Alerts/News | `/alerts/*`, `/news/*` |
| Community | `/community/*` |
| Units | `/units/*`, `/units/nearby`, `/units/apply` |
| Admin | `/admin/*`, `/audit/*`, `/settings/*` |
| Comms | `/communication/*`, `/ws` |
| AI | `/ai/*` |
| Mobile | `/mobile/*` |
| Finance | `/bank/*`, `/finance/*` |

**Contract rule:** the OpenAPI/typed contract is the source of truth. Freeze it before building;
generate or hand-write types from it. No hand-typed endpoint strings in components.

---

## 6. Build roadmap (milestones)

- [x] **M0 — Foundations:** scaffold, router, API client, auth/session, design tokens, component
      primitives *(lint/build/CI wiring still to run — see §9)*
- [~] **M1 — Auth & onboarding:** login, role routing, guards, citizen signup, password recovery
      screens and token refresh shipped; OTP, officer applications, onboarding, session management
      and live signup/recovery verification open
- [x] **M2 — Citizen core: reporting and tracking — COMPLETE.** `/report` is a four-step wizard
      against `POST /cases` — what happened, where (pin + unit), optional evidence links, review —
      ending in a receipt that carries the tracking ID and attaches the links, each with its own
      outcome. `/` lists the reporter's cases and `/cases/:id` renders one as a curated record: the
      final report, the closure decisions and their comments, the shared weekly narratives, and a
      status-change log with names and internal notes withheld. Draft auto-save is in; **SOS UI is
      mock-verified with a live integration gate; feedback is now available in the demo only**
- [~] **M3 — Awareness:** case/unit map, notification centre, demo alerts/news and subscription
      controls shipped in the browser; Python integration and real push delivery remain open
- [~] **M4 — Officer console:** queue and case workspace (details/progress/weekly/evidence) shipped,
      with weekly narratives served by the real endpoints (M4.5); dispatch → arrive shipped;
      **closure moved to the review workflow**; team view and comms open
- [x] **M4.5 — Lifecycle & accountability re-alignment — COMPLETE.** The backend's review workflow is
      fully adopted: the eight states and their tokens, submit-for-review, the officer's read of a
      changes-requested decision, the admin's approve/request-changes surface with its required
      comment and explained self-approval guard, the real weekly-update endpoints, and mocks covering
      every new state. **The only piece left is `on_scene → investigating`, which is not a frontend
      task** — no registered route performs it, so any button would be inventing a transition. It is
      reachable in the demo via the seed
- [~] **M5 — Unit admin:** case review + dispatch board shipped (`/admin/cases`,
      `/admin/cases/:id`) — triage counters, assignment, evidence verification, **and the closure
      decision (delivered in M4.5)**; overview, roster, analytics, finance and settings open.
      **Super admin:** F9 not started
- [~] **M6 — Depth:** aggregated activity map overlay shipped; F11 (comms) and F12 (AI) open
- [ ] **M7 — Field hardening:** F14 offline/PWA, performance budget, a11y audit, E2E suite
- [ ] **M8 — Pilot polish:** empty/error states everywhere, i18n, analytics funnel, real-device testing

Each milestone ships only when the **Definition of Done** below is met.

**Why M4.5 is a milestone and not a bug fix.** A lifecycle that renders `investigating` as "Pending"
does not look broken — it looks *finished*, and wrong. The cost of the drift is not a crash; it is a
citizen reading a confident, incorrect status, and an administrator with no way to close a case. That
is a correctness problem in a public-safety product, so it is scheduled like one.

---

## 7. Definition of Done (frontend)

```
CONTRACT → ROUTE → COMPONENT → STATE (loading/empty/error/offline)
        → A11Y → MOBILE → I18N → TEST → PERF BUDGET → DOCS
```

A screen is done only when all applicable layers are addressed and it has been reviewed against
[`frontagent.md`](./frontagent.md). If a layer does not apply, record why.

---

## 8. How to use this with the design agent

1. Pick a feature (e.g. **F3 — Incident reporting**).
2. Load [`frontagent.md`](./frontagent.md) as the design brief.
3. Agent produces: information architecture → wireframes → tokens/components → implementation.
4. Review against the engagement goals (§1) and the Definition of Done (§7).
5. Ship, instrument the funnel, and feed learnings back into this spec.

> `frontReadme.md` = **what and why**. `frontagent.md` = **how it looks, feels, and hooks users.**

---

## 9. Verifying a change (frontend)

```bash
cd frontend
npm install
npm run lint         # eslint, TS + hooks rules
npm run typecheck    # tsc --noEmit
npm run test         # vitest (ISO weeks, API client, assignment rules, status vocabulary,
                     #         reporter case-log disclosure, report draft) — 83 tests
npm run build        # tsc --noEmit && vite build
npm run dev          # mocks on by default
```

Manual walk — officer (with mocks): sign in as **Officer** → `/officer/queue` → open the P1 robbery
case → **Details / Progress / Weekly / Evidence** tabs → add a progress update and confirm it lands
under **This week** in the *Activity by week* section → open **Weekly** and confirm the filed
narrative from last week renders with its full field set, and that the two sections are visibly
different things → on a case you are assigned to that has no narrative this week, **File this week's
update** → confirm the primary action is disabled until both *Summary* and *Investigation* are
filled → file it, then reopen the tab and confirm the composer has been **replaced** by "This week's
update is already filed" quoting what you wrote (one per officer per week) → verify an evidence item
→ **Dispatch** or **Mark on scene** on a case in the right state → `/map`, toggle status filters and
unit coverage → resize to phone width and confirm the stepper and tabs stay one-hand usable.

Manual walk — officer closure path (the M4.5 half that is built): open the seeded case in
`investigating` with **no final report** → confirm **Submit for review** is present but disabled and
that the screen *says why* ("Add the final report first…") → add a final report → submit → confirm
the case lands on `pending_admin_review` and that the submit action is replaced by a statement that an
administrator holds it → open the seeded `admin_changes_requested` case and confirm the reviewer's
instruction is the first thing on the screen, above the facts.

> **This walk stops at `pending_admin_review` on the officer side** — the decision belongs to an
> administrator, and the walk for that is below. Do not "fix" the stop by re-adding a Close button;
> `POST /cases/:id/close` is not routed.

Manual walk — administrator (with mocks): sign in as **Unit admin** → you land on `/admin/cases` →
confirm the attention counters and that clicking one filters the list → confirm **"Awaiting your
decision"** is the first counter and that it isolates the submitted-for-closure cases → open the
pending assault case → **Assign officer** → pick an officer and confirm the case flips
**pending → assigned** → sign back in as **Officer** and confirm that case is now dispatchable. That
round trip is the end-to-end proof that `pending → assigned → dispatched` works, and it is the only
path to the first transition.

Manual walk — administrator, the closure decision (the M4.5 loop, now built): open the case sitting
in `pending_admin_review` → confirm the decision panel is the **first thing on the page**, above the
tabs, and that the Review tab carries the history (including the earlier request-changes round) →
click **Request changes** → confirm the primary action is **disabled** and says a comment is required
→ type a comment and send it → confirm the case flips to `admin_changes_requested` and that the panel
is replaced by the instruction notice → as **Officer**, confirm that same comment is the first thing
on the case → **Submit for review** again → as **Unit admin**, **Approve closure** with a comment →
confirm the case reaches `closed`, shows `closedAt` / `closedBy` / `approvedBy`, and that the Review
tab now holds both decisions in order.

> Then the guard: open the seeded case **"Repeated vandalism of the street lighting on Ogunlana
> Drive"** (`CS-2026-0043`) — it is in `pending_admin_review` **and** assigned to the unit admin, so
> signing in as **Unit admin** makes the collision visible immediately. Confirm **Approve closure** is
> disabled and explains the self-approval rule, and that **Request changes** still works — the rule
> blocks approval, not the whole decision. That seed exists *only* for this: without it the guard is
> unreachable until a real deployment happens to produce the collision, and an untestable privacy
> rule is one that quietly rots.

Manual walk — citizen, filing a report (F3, with mocks): sign in as **Citizen** → `/` shows a **New
report** button and the nav's **Report** item is live (its "Soon" chip is gone) → both that button and
the empty-state button lead to `/report` → on **step 1**, leave the form empty and confirm **Continue
is disabled** and says a title is needed; add a title and confirm the reason changes to name the
description → **reload mid-form**: the answers come back with "We kept the answers you had started.
Nothing has been sent." → on **step 2**, with no pin the unit list says to place the pin first; press
**Use my location** or tap the map, then confirm a coordinate readout appears and the units list by
distance with the nearest marked **Closest** → confirm **Continue** stays disabled until *both* a pin
and a unit are chosen, and that the refusal names whichever is missing → on **step 3**, add a link and
type `myphoto`: it must show "Must be a link starting with http:// or https://" and block Continue →
fix it to `https://example.com/p.jpg` → on **step 4** the summary shows the title, the description,
the unit by name and the link, plus a **Marked urgent** badge if you ticked the danger box. That badge
must be **amber, not red** — red is SOS only.

Send it, and read the receipt: **"Report filed"**, a `CS-YYYYMMDD-NNNN` tracking ID with a copy
button, a **Pending** chip, **P3** (or **P2** if you ticked urgent), and your links attaching one at a
time. **Track this report** opens the reporter's case page, whose log's first row reads **Case
reported** with the actor **You** — a report you just filed must never open on an empty log.

> Then the failure paths, which matter more here than the happy one:
>
> **Nothing created** — stop the dev server and press Send: the message must read *"Your report was
> not sent. Nothing was created, and every answer is still here"*, with every field still filled. A
> create failure that reads as a success, or that costs the reporter their typing, is the worst bug
> this flow can have.
>
> **Report filed, link failed** — file a report with the server up, then stop it before the uploads
> finish: the receipt must still read **"Report filed"**, the failed link must say *"Not attached"*
> with its own **Retry**, and nothing may describe the *report* as having failed.
>
> **Draft cleared** — after a successful send, reload `/report`: it must open blank. The draft is
> cleared the instant the case is created, and a draft that survives a successful submit is how the
> same report gets filed twice.

Manual walk — citizen (with mocks): sign in as **Citizen** → `/` lists your reports and **every card
is a link** (hover highlights the border) → open **"Armed robbery in progress at Adeniran Ogunsanya
Plaza"** (`CS-2026-0041`, `on_scene`) → confirm the header shows the **responding unit by name**
("Surulere Central Response Unit"), that there is **no officer name and no evidence gallery
anywhere on the page**, and **no "Activity by week" section** — only *Officer narratives* → the
**Case log** reads oldest first: *Case reported* (You) → *Officer assigned* (Unit staff) → *Officer
dispatched* (Assigned officer) → *Officer on scene* (Assigned officer), each with the status chip it
moved to. **If any row shows a description sentence or a name — "Assigned to Officer Tunde Balogun."
is the one the seed carries — the redaction has been lost.**

> Then the two ends of the record: open **"Vandalised street lights along Herbert Macaulay Way"**
> (`CS-2026-0031`, `closed`) — final report, **Closure approved** with the administrator's comment, a
> one-row log. Open **"Contraband goods offloaded at night at a warehouse on Apapa Road"**
> (`CS-2026-0036`) — **two** shared narratives and a **Changes requested** decision with its comment,
> which is the case that answers "why is my case still open".

> `CS-2026-0028` (the Ikeja loitering case) is closed with **no** recorded decision, on purpose: it is
> the legacy-closed state, and the review section must say the record is absent rather than that a
> decision is still coming.

If you are working on files that predate this rebuild, note that stale Vite entry points
(`src/App.jsx`, `src/main.jsx`, `src/App.css`, `vite.config.js`, `src/assets/`) must **not** exist:
Vite resolves `.jsx` before `.tsx`, so they would shadow the current app. Delete them if they
reappear.


---

## Appendix A — Historical target-contract notes (not Python backend evidence)

This appendix records every change made to the platform after the M4.5 milestone
was written. The catalogue above reflects the state at M4.5; this section tracks
what shipped after it.

### A1 — Rebrand

Platform renamed **Nativity Guard → Nativity Guard**.
Logo, favicon, and wordmark shipped in `frontend/public/logo.svg` and `favicon.svg`.
Backend display strings, email templates, SMS copy, and AI prompts all say Nativity Guard.

### A2 — Authentication & session hardening

Backend changes that the frontend will need to wire:

| Feature | Endpoint | Status |
|---|---|---|
| Rate limiting on auth | (middleware — no endpoint change) | ✅ live |
| JWT revocation on logout | `POST /auth/logout` (unchanged, now actually revokes) | ✅ live |
| Password change revokes all sessions | `POST /auth/change-password` | ✅ live |
| Refresh token issuance | `POST /auth/login` returns `refreshToken` alongside `token` | ✅ live |
| Refresh endpoint with rotation | `POST /auth/refresh` | ✅ live |
| Rate limiting on OTP | (middleware) | ✅ live |
| Rate limiting on vote | (middleware) | ✅ live |
| Rate limiting on invite | (middleware) | ✅ live |

**What the frontend must do:** store the refresh token, call `/auth/refresh` when the
access token expires, and treat a revoked-token 401 as a hard logout. See F1 above.

### A3 — Case authorization hardening

The former target contract described the rule **unit access is NOT case access**. The mounted Python
routes must implement and verify their own access control before this can be relied on:

- Officers only see cases they are assigned to (as `primary` or `paired`)
- Support officers see a limited view
- Admins only see cases submitted to them via `CaseAdminAssignment`
- Head Admin sees all cases in their unit
- Reporter sees public-safe fields only (`GetCaseByID` returns a curated DTO for them)
- Direct officer `CloseCase` is blocked; closure requires admin review

**What the frontend must do:** if a screen calls `GET /cases/:id` for a reporter, expect
the public-safe DTO (no evidence list, no progress list, no officer identity). The
existing citizen case page already renders a subset — no code change required, but the
data returned is now also safe by contract, not just by UI curation.

### A4 — Governance UI (historical target; Python availability unverified)

The former target described a governance layer. The Python app inspected here does not mount
these endpoints, and no corresponding frontend screens exist yet.

| Feature | Endpoints | UI status |
|---|---|---|
| Admin elections | `POST /units/:unitId/elections`, `POST /elections/:id/vote`, `POST /elections/:id/close`, `GET /elections/:id/results` | ❌ not built |
| Head-admin elections | `POST /units/:unitId/head-admin-elections` | ❌ not built |
| Revocation cycles | `POST /units/:unitId/revocations`, `POST /revocations/:id/vote`, `POST /revocations/:id/close`, `GET /revocations/:id` | ❌ not built |
| UnitAuth policy | `GET /units/:unitId/auth`, `PUT /units/:unitId/auth` | ❌ not built |

**New UI work needed** (added to feature catalogue as **F7 — Governance UI**):

- Election list per unit
- Voting interface (one vote per verified member)
- Head-admin election interface (admins only)
- Revocation cycle interface (open cycle, vote, view tally)
- UnitAuth policy editor (head admin / admin only)
- Term and cooling-off status per admin

### A5 — Suspect self-view (historical target; Python availability unverified)

A citizen linked as a suspect via `SuspectCase` can fetch their own case list:

| Endpoint | Returns |
|---|---|
| `GET /suspects/me/cases` | `{ active: [...], history: [...] }` — case ID, tracking ID, title, category, status, role, timestamps, last progress. **No officer identity, no evidence, no admin notes, no other suspects.** |

Resolved cases move from `active` to `history`. The citizen never sees the case as an
investigative record — only as a status.

**Frontend work:** add a "Suspect cases" tab to the citizen profile that renders the
two lists. Presumed-innocence framing — "you are linked to this case" — not "you are
accused".

### A6 — Invite scoping

Two invite types now exist:

| Scope | Who creates | Purpose |
|---|---|---|
| `platform` | Any registered user | Invite a new person to register on Nativity Guard (no unit) |
| `unit` | Admin / Head Admin, post-verification | Invite a citizen to join a specific unit |

| Endpoint | Purpose |
|---|---|
| `POST /invites` | Create an invite (scope in body) — returns the raw code once |
| `POST /invites/validate` | Public — checks code, returns safe metadata (scope, unit name if applicable) |

**Frontend work:** invitation generator UI + a public registration landing page that
uses `/invites/validate` to show the join-or-stay-citizen choice.

### A7 — What remains open from M4.5

- `on_scene → investigating` — still no registered route (backend gap, not frontend)
- `PUT /cases/:id` — still performs zero status validation (authorization hole,
  tracked for Wave 4b of the backend roadmap)
- `CaseReviewDecisionDeescalate` — still no route records it
- `GetCaseAccountability` — still unrouted
- `GetOfficersByUnit` — still unrouted

### A8 — Outstanding frontend gaps

- `npm install` must be run once — `node_modules/` is required for `npm run build`
- No component tests, only pure-function tests
- Map marker grouping and aggregated activity overlay are built; advanced map filters remain open
- SOS frontend is mock-verified; live service contract and dispatch behavior remain unverified
- Feedback submission form still unbuilt


---

## Appendix B — Historical integration brief (not applicable to the current Python routes)

This appendix preserves an older integration brief. It was not validated against the Python
service in this checkout. **The Python backend contract at the top of this file takes precedence**
over all endpoint, deployment and authorization claims below.

### B1 — Former live environment (historical)

The old deployment URLs and CORS instructions were removed because they do not establish that
the current Python service implements the frontend API. Verify any deployment separately.

### B2 — Former target contract (unverified against Python)

| Item | Former target behavior, unverified for Python | Historical source |
|---|---|---|
| **Role strings** | **Underscore** — `citizen` · `officer` · `unit_admin` · `super_admin`. Compared in ~40 places. Super-admin is *also* a separate boolean, `models.User.IsSuperAdmin`; several guards accept either. A hyphenated DB value is an outlier to fix in the data, not a convention to adopt. `head_admin` is **not** a user role — it is `UnitMembership.IsHeadAdmin`. | former identity contract; not a Python source reference |
| **`POST /auth/register`** | Returns **201 `{ message, user }` and NO token.** It also **ignores the `role` input** — `Role: "citizen"` is hardcoded — so a role selector on signup would collect an answer the server discards. `dateOfBirth` (YYYY-MM-DD) is `binding:"required"`. | former auth contract; no mounted Python equivalent |
| **`POST /auth/login`** | Accepts **either** `identifier` **or** legacy `email`. The brief implies `identifier` is required; the existing `{ email, password }` call is fine and was never a bug. | former auth contract; no mounted Python equivalent |
| **Route params** | `:id` everywhere — never `:unitId`, never `:userId`. | former route registry; not in this checkout |
| **`GET /units/:id/officers`** | **Unrouted**, but the handler is fully implemented, already reads `c.Param("id")`, already authorises via `canViewOfficersInUnit`, and already returns `{ officers: [...] }` — the exact shape `useOfficers.ts` expects. Registering it is one line. Only `/:id/officers/ranking` is registered today. | former officers contract; not in this checkout |
| **`/cases` pagination** | `?limit=` is capped at **100**, default **50**. Never assume "get all"; page through. Other list endpoints follow in a later wave. | brief §5 |
| **Error envelope** | `{ "error": "..." }` — that is the whole message. Do not add or invent a wrapper. | handlers |

`ranking` is **not** a substitute for the roster: it is an ordered leaderboard, whereas
`AssignOfficerDialog` needs every assignable officer in the unit.

### B3 — Historical bug triage for the mock target

The integration brief listed eight bugs. Checked against the source, **three are not real, one has
the right symptom and the wrong cause, and three real defects are missing from it entirely.**

| # | Brief's claim | Verdict |
|---|---|---|
| 1 | `/login` renders blank black | **Partly.** The route is `/auth/login` (`App.tsx:48`), and `/login` redirects there rather than going blank. An alias is still worth adding; the stated symptom is wrong. |
| 2 | `/` blank when logged out | **Not a bug.** `RequireRole.tsx:23-25` navigates to `/auth/login` with `state.from`. There *is* a redirect target. |
| 3 | Flash then blank on every page | **Real symptom, wrong cause** — see B3.1. |
| 4 | No signup page, no landing page | **Real.** `src/pages/auth/` contains only `LoginPage.tsx`; there is no `/signup` route. This is the genuine blocking gap. |
| 5 | No catch-all; unknown URLs blank | **Not a bug.** `App.tsx:143` → `<Route path="*">` → `NotFoundPage`, which exists. |
| 6 | Role string mismatch | **Real — and it is the root cause of #3.** |
| 7 | `/units/:id/officers` does not exist | **Real** (see B2). |
| 8 | Mock param drift `:unitId` → `:id` | **Cosmetic, not a bug.** The mock matches its own pattern and reads its own `params.unitId`, so it works. Convention alignment only. |
| **+** | *(absent from the brief)* | **`POST /auth/register` issues no token** — a signup page built to the brief's stated response shape fails silently on auto-login. |
| **+** | *(absent from the brief)* | **No error boundary.** `main.tsx` renders `<App />` bare, so any render throw becomes an unexplained blank page. |
| **+** | *(absent from the brief)* | **`/public/*` serialises whole models to anonymous callers** — see B3.2. Why the landing page ships without its map preview. |

#### B3.1 — Why the app actually goes blank

`types/api.ts` types `Role` as the underscore union, and **nothing normalises what the API
returns**. With a hyphenated role from the database:

1. `RequireRole.tsx:26` — the role is not in `ALL_ROLES` → `Navigate` to `homePathForRole('super-admin')`
2. `RequireRole.tsx:41-43` — no switch match → `default: return '/'`
3. `/` → `HomeRoute` (`App.tsx:37`) — the role is truthy and is not `'citizen'` → `Navigate` to `homePathForRole(...)` → `'/'`
4. **Infinite self-redirect** → React throws "Maximum update depth exceeded" → no error boundary → black screen.

That is the reported symptom exactly. **The brief's priority order is therefore inverted: this is
the first thing to fix, not the last item to reconcile.**

**Two independent defects have to be fixed, and fixing only one leaves the app breakable:**

- **The data** — Agene sets the database role to `super_admin` (underscore).
- **The code** — the frontend must survive a role it does not recognise. A correct database today
  does not stop the next unexpected value (a new role, a typo, a legacy row) from blanking the app
  again. This is `frontagent` §2 rule 3, and its fallback is a designed screen.

#### B3.2 — Why the landing page has no map preview

`GET /public/cases` and `GET /public/units` are registered with **no auth middleware** — only
`RateLimitGeneral()` (former route registry; not in this checkout) — and neither handler narrows the query with a `Select`. So
every `json`-tagged field on the model serialises to an anonymous caller:

- from `models.Case` — `reportedBy`, the citizen's `title` and `description`, `latitude` /
  `longitude` **and** `gisLatitude` / `gisLongitude`, `assignedTo`, `closedBy`, `approvedBy`;
- from `models.SecurityUnit` — `contactPerson`, `contactPhone`, `contactEmail`, `hostUserId`,
  `verifiedBy`, `verificationNotes`, `adminCount`, `memberCount`.

The handler's own comment reads `// GetPublicUnits - Public endpoint for landing page (no auth
required)` — these were written *for* the page T3 builds. The leak is real and bounded, and it is
**not a frontend fix**: narrowing the response is a backend change (a curated public DTO), so T3
ships without the map block and `LandingPage` makes no API call at all. Reported to Agene; see the
open question in B2.

### B4 — Task backlog

The Python API agreement in the opening section now precedes all tasks below. Their endpoint
assumptions must be rechecked before implementation.

Ordered. One task per branch-push, each verified locally before it ships. Every task ships only
when its Definition of Done (§7) is met.

**P0 — blocking a usable site**

- [x] **T1 — Role reconciliation + unrecognised-role state** — *done, verified 2026-09-24 (`lib/role.ts` + 7 tests, 90 passing).*
      New `lib/role.ts` (pure, tested): `normaliseRole()` maps database spellings onto the canonical
      union and returns `null` for anything it does not know. `AuthContext` normalises at the
      boundary and exposes the raw value alongside. `RequireRole` renders a designed
      **Unrecognised role** screen instead of redirecting, and `homePathForRole` returns null rather
      than defaulting to `'/'`.
      *DoD:* no role value can produce a redirect loop; an unknown role is visible, explained and
      escapable; the normaliser is unit-tested.
      *Blocked on:* nothing — defensive regardless of the database fix.

- [x] **T2 — Signup page** (`/signup`) wired to `POST /auth/register` — *done, verified 2026-09-24
      (`lib/signup.ts` + 13 tests, 103 passing). Live-backend half of the DoD still to confirm.*
      Fields: email, phone, firstName, lastName, password (min 6), **dateOfBirth (required)**.
      **No role selector** — the endpoint ignores it. **No auto-login** — the response carries no
      token, so success redirects to `/auth/login` with the email prefilled and a notice.
      *DoD:* a real account is created against the live backend and then signed in successfully.

- [x] **T3 — Public landing page at `/` for logged-out visitors** — *done, verified 2026-09-24
      (103 tests, build clean). `LandingPage.tsx`; `/` moved outside `ProtectedShell` and decided by
      `RootRoute` from auth status. **The map-preview block is deliberately unbuilt** — see B3.2.*
      Hero, value proposition, feature sections, **Sign in** / **Get started**.
      Requires moving `/` outside `ProtectedShell`. *DoD:* logged-out `/` renders the landing page;
      signed-in `/` behaves exactly as today. Design brief: `frontagent` §11.

- [x] **T4 — `/login` alias** → redirect to `/auth/login`, preserving `state.from`.
      *Done, verified 2026-09-24. `LoginAlias` carries `location.state` — a bare `<Navigate>` drops it.*

- [x] **T5 — Error boundary.** *Done, verified 2026-09-24. `ErrorBoundary.tsx` mounted full-page in
      `main.tsx` and per-page in `AppShell` so a route throw keeps the chrome.*
      Any render throw becomes an explained, recoverable screen.
      *DoD:* a deliberate throw in a page component shows the boundary, not a blank page.

**P1 — contract alignment**

- [ ] **T6 — Officer roster.** Ask Agene to register `GET /units/:id/officers` (handler exists, shape
      already matches). `AssignOfficerDialog` then loses its contract-gap branch (§0.3). Rename the
      mock's `:unitId` → `:id` in the same change (cosmetic; B3 #8).

- [x] **T7 — Refresh token flow** (F1 / A2) — *done, verified 2026-09-24 (`apiClient.test.ts` 9 → 15
      tests, 109 passing). Rotating refresh, one shared attempt, hard logout on a rejected refresh.*

- [ ] **T8 — Password recovery.** `POST /auth/forgot-password`, `POST /auth/reset-password`.
      Built but **not yet verified**. `/auth/forgot-password` → `/auth/reset-password`, plus
      `lib/passwordReset.ts` (pure validators + `maskIdentifier`). Two corrections to the brief,
      both read off the former reset service:
      1. **The token is a 6-digit code, not a link.** `generateResetCode` mints six digits and both
         notifiers put the code in the message body. The handler's 200 still says "we've sent a
         password reset link" — the screens do not repeat that, because someone told to expect a
         link waits for one that never arrives. **Backend copy is wrong** (report to Agene).
      2. **Reset requires 8 characters; registration accepts 6.** `ResetPassword` binds `min=8` and
         `ResetWithToken` refuses anything shorter again. Stated as-is in the reset hint rather than
         reusing signup's 6 — a 6-character password can never be *reset*, only set at signup.
         **Backend asymmetry** (report to Agene).
      Success revokes every session and refresh token, so the page calls `logout()` before handing
      off to `/auth/login` with a `reset` notice — a local token left behind would fail its next
      request with no explanation.

- [ ] **T9 — Session management.** `GET /auth/sessions`, `DELETE /auth/sessions/:jti`,
      `DELETE /auth/sessions`.

**P2 — surfaces the brief exposes that have no UI yet**

- [ ] **T10** — Governance UI (A4)
- [ ] **T11** — Suspect self-view (A5)
- [ ] **T12** — Invites (A6)
- [~] **T13** — F2 SOS frontend built against mocks; live integration and responder-side workflow
      must be verified before pilot use
- [ ] **T14** — Feedback submission (F4)
- [ ] **T15** — Audit / finance / bank-account / public endpoints from the brief's §4 reference

**Explicitly not tasks:** the brief's "BUG 2", "BUG 3" and "BUG 5" (not real — see B3), and BUG 8
beyond the cosmetic rename. Do not "fix" a redirect guard that is already correct.

### B5 — Doc drift, tracked and deferred

Found while confirming these docs. Not corrected here, to keep this change to a single task.

1. §A1 reads "renamed **Nativity Guard → Nativity Guard**" — the source name was lost in the edit.
   It was almost certainly **Community Shield → Nativity Guard**.
2. §A4 files governance UI as "**F7** — Governance UI", but F7 is already the Officer console.
3. F4's `**States:**` / `**APIs:**` block is duplicated verbatim (~L553–556 and ~L565–568).
4. §A3 and §0.3/F4 disagree on the reporter's privacy boundary. **Verified truth:** the curated
   `case` DTO, the evidence list and the progress feed *are* now withheld server-side
   (the former case handler `GetCaseByID`), but **`timeline` is still returned in full** — so
   `lib/caseLog.ts`'s role-not-name redaction remains load-bearing, not cosmetic. §A3's "safe by
   contract" is half-right; §0.3's warning is half-stale.
5. `frontagent` §2 rule 2 described a `lib/status.ts` fallback bug that is already fixed.
