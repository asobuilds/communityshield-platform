# CommunityShield — Frontend Feature & Build Specification

> **This is the build contract for the CommunityShield frontend.**
> It lists every feature, screen, state, and API mapping needed to ship the product.
> Pair it with [`frontagent.md`](./frontagent.md) — the design agent that governs *how* the UI
> should look and feel. `frontReadme` says **what to build**; `frontagent` says **how to make it
> compelling**.

---

## 0. Current reality (read first)

The original 126-line `App.jsx` SOS mock has been **replaced** (milestone **M0** is done). The
frontend is now a typed React 19 + Vite + Tailwind v4 application with routing, an auth/session
layer, a server-state cache, a design system, a reusable map, and the first real workflows.

| Item | Today | Target |
|---|---|---|
| Screens | 8 implemented | 40+ across 4 role consoles |
| Routing | ✅ React Router v7, role-gated | Complete |
| API layer | ✅ typed client, mock adapter, 401 handling | Complete |
| State | ✅ React Query (server) + auth context | Complete |
| Design system | ✅ primitives + tokens + 4 states | Grow with features |
| Tests | ✅ unit (ISO weeks, API client, assignment rules) | Component + E2E |
| Mocks | ✅ in-browser mock API (`VITE_USE_MOCKS`) | Contract tests |

### 0.1 What exists right now

**Run it:** `cd frontend && npm install && npm run dev` — mocks are on by default
(`.env.development`), so every screen below works with **no backend**. Sign in with any demo
account: `officer@shield.ng`, `admin@shield.ng`, `citizen@shield.ng`, `super@shield.ng`
(password `password`).

| Route | Screen | State |
|---|---|---|
| `/auth/login` | Sign-in (role quick-fill under mocks) | ✅ |
| `/` | Citizen: your reports + lifecycle stepper | ✅ (detail view next) |
| `/officer/queue` | Case queue — priority sort, filters, search, SLA badge | ✅ |
| `/officer/cases/:id` | Case workspace — Details · Progress · Weekly · Evidence | ✅ |
| `/admin/cases` | Admin triage board — attention counters, search, assign/reassign | ✅ |
| `/admin/cases/:id` | Admin case review — facts · progress · weekly · evidence · assignment | ✅ |
| `/map` | Operations map — cases + unit coverage + filters | ✅ |
| `/admin/*`, `/super/*` | Remaining admin surfaces — named placeholders | ⏳ M5 / M7 |
| `*` | 404 | ✅ |

The admin console stops at case review on purpose: **assignment is the one lifecycle transition an
officer cannot perform for themselves.** `pending → assigned` is an administrator's decision, and
without it the officer workspace can only exercise the second half of the case lifecycle. That is
why `/admin/cases` leads the unit-admin nav and is where `homePathForRole('unit_admin')` lands.

**Key source files**

```
src/
  App.tsx                  route table (public → protected shell → role-gated)
  main.tsx                 installs the mock API, then renders
  types/api.ts             hand-written contract from the Go handlers/models
  lib/apiClient.ts         typed fetch: bearer token, ApiError, 401 → logout
  lib/queryClient.ts       React Query defaults (no retry on 4xx)
  lib/status.ts            case-status + priority metadata (single source of truth)
  lib/week.ts              ISO-week grouping — powers the Weekly interface
  lib/assignment.ts        assignment rules mirroring the backend guard (pure, tested)
  auth/                    AuthContext (session restore) + RequireRole guard
  components/ui/           Button, Card, Chips, Field, Tabs, Modal, Toast, States
  components/layout/       AppShell (sidebar ⇄ bottom nav) + NotificationBell
  components/map/MapView   the one map: view mode + location-pick mode
  components/case/         CaseHeader, CaseFacts, CaseStatusStepper, ProgressTimeline,
                           WeeklyUpdates, EvidenceGallery, EvidenceUpload
  components/admin/        AssignOfficerDialog
  hooks/                   useCases, useProgress, useEvidence, useUnits,
                           useOfficers, useNotifications
  mocks/                   fetch-level mock API + Lagos seed data
```

`CaseFacts` is shared by the officer workspace and the admin review on purpose: **an administrator
reviewing a decision must not be looking at a different rendering of it.** Both consoles read the
same case facts from one component (`components/case/CaseFacts.tsx`), which also exports the
`SystemTimeline` used by both.

### 0.2 Two decisions worth knowing

1. **The mock layer is a fetch adapter, not MSW.** MSW needs an npm install *and* a generated
   `public/mockServiceWorker.js`; that file cannot be produced without network access, which would
   have left "works offline against mocks" unverifiable. `src/mocks/adapter.ts` + `install.ts`
   match routes in MSW's shape (`method`, `path` with `:params`, `respond({ request, params })`),
   so porting later is mechanical. **Add `msw` back only when the worker file can be generated.**
2. **There is no weekly-update backend entity.** `WeeklyUpdates` derives weeks client-side from
   progress/timeline timestamps (`lib/week.ts`). Officers can file a narrative summary, stored as a
   progress entry with action `weekly_summary` — reusing an endpoint that exists rather than
   inventing one. If the backend later grows a weekly entity, this is what to replace.

### 0.3 Known contract gaps the UI does **not** paper over

| Gap | How the UI handles it |
|---|---|
| **`GetOfficersByUnit` is implemented but never routed** — assigning a case needs an `officers`-table id and nothing on this API lists officers | `AssignOfficerDialog` treats a 404 as a first-class "contract gap" state: it says so plainly and offers a manual officer-ID field. Never an empty roster, never a generic error. Mocks mirror the *unregistered handler's* shape so the flow is walkable today. One-line backend fix: register the route |
| `POST /cases/:id/assign` takes an **`officers` id, not a user id** (`officer.UnitID` must equal the case's unit) | `lib/assignment.ts` mirrors the guard (`officerBelongsToUnit`, `assignmentBlocker`) so the dialog blocks a wrong-unit pick before the round trip; officers are a separate entity from `User` in `types/api.ts` |
| `GET /cases/:id/assignments` may not exist on a given build | `AssignmentPanel` splits 404 (`missing` — endpoint not exposed) from network failure (`failed` — explicitly *does not* mean unassigned). Conflating them would let a dropped connection read as "nobody is assigned" |
| `GET /auth/profile` carries no `unitId` | `inferAdminUnitId` derives the admin's unit from the scoped case list, and only when **exactly one** unit is present — a super admin's cross-unit list assumes no roster |
| `GET /cases` may not preload `evidence` | The queue's "evidence unverified" counter is gated on any case actually carrying the array; otherwise the card is hidden rather than reporting a confident zero |
| `GET /cases/:id/progress` has no server-side authorization | Case shows only what it is given; gap documented, backend untouched |
| `GET /cases/analytics` counts `status="resolved"`, which the workflow never sets → `resolutionRate` always 0 | Analytics surfaces are not built on it; KPIs will use `closed` |
| No binary upload endpoint — evidence takes a hosted `fileUrl` | `EvidenceUpload` asks for a link and says so plainly |
| Notification reads live under `/mobile/notifications*` only | `useNotifications` uses the mobile endpoints |

**Prerequisite still open:** the Go backend does not build from this repo (`go.sum` is git-ignored),
so `VITE_USE_MOCKS=false` has nothing to talk to yet.

---

## 1. Product intent & engagement goals

CommunityShield must be **chosen** by communities, not mandated. That only happens if the
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

**Screens:** login, register (role select), OTP verify, forgot/reset, onboarding wizard.

- [x] Email/phone + password login; role-aware redirect
- [ ] Register as citizen / officer-application
- [ ] OTP send / verify / resend (countdown, rate-limit messaging)
- [ ] Unit application flow (search nearby units, select, apply)
- [ ] Government ID submission (camera/file, status pending/verified/rejected)
- [ ] Medical info intake (explicitly optional, privacy notice)
- [ ] Profile completion + onboarding checklist
- [x] Session: JWT storage, session restore on boot, logout, 401 → login
      *(no silent refresh — the backend issues no refresh token)*

**States:** idle · loading · invalid credentials · OTP expired · rate-limited · pending verification
**APIs:** `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`, `GET /auth/profile`,
`POST /otp/send|verify|resend`, `GET /units/nearby`, `POST /units/apply`,
`POST /units/government-id`, `GET|PUT /settings/onboarding`

---

### F2 — Citizen home & SOS

**Screens:** citizen home, SOS console, SOS active/status, SOS history.

- [ ] One-tap SOS with confirm + cancel (never accidental)
- [ ] Geolocation with graceful permission-denied fallback + manual pin
- [ ] Live tracking token + status once active
- [ ] Emergency-contacts capture; medical-info attach
- [ ] Priority select (default high); optional unit target
- [ ] SOS history + per-alert status (`pending → dispatched → resolved → escalated`)
- [ ] Prominent, always-reachable SOS affordance (persistent action)

**States:** locating · active · escalated · resolved · permission-denied · offline-queued
**APIs:** `POST /sos/send`, `GET /sos/my`, `GET /sos/:id`, `PUT /sos/:id/status`

---

### F3 — Incident reporting

**Screens:** report wizard (multi-step), report review, success/receipt, my reports.

- [ ] Step 1: category/type; Step 2: title + description; Step 3: media; Step 4: location; Step 5: review
- [ ] Media capture (photo/video/voice note) with upload + preview + remove
- [ ] Location picker (map + manual address + "use my location")
- [ ] Priority hint (advisory only — triage is server-side)
- [ ] Public/private visibility choice with plain-language explanation
- [ ] Draft auto-save; offline queue with sync on reconnect
- [ ] Receipt with **Tracking ID** + "track this case" CTA
- [ ] Validation inline, never a raw error toast

**States:** drafting · validating · uploading · submitting · queued-offline · submitted · failed-retry
**APIs:** `POST /cases`, `POST /evidence/upload`, `GET /cases`, `GET /cases/:id`

---

### F4 — Case tracking & feedback

**Screens:** my cases list, case detail, timeline, evidence viewer, feedback form.

- [x] Case list with status chips, priority, last-updated
- [x] **Lifecycle stepper** (pending → assigned → dispatched → on-scene → closed) with timestamps
- [x] Progress timeline (action, description, officer, time)
- [~] Assigned unit/officer visibility (respecting privacy) — officer shown when named on the timeline
- [x] Evidence gallery (thumbnails, type, verification badge)
- [x] Resolution summary (final report) when closed
- [~] Feedback: rating + comment rendered; **submission form not built**
- [ ] Push/in-app updates on every status change

**States:** empty · loading · not-found · forbidden · closed-readonly
**APIs:** `GET /cases`, `GET /cases/:id`, `GET /cases/:id/timeline`, `GET /cases/:id/progress`,
`GET /evidence/case/:caseId`, `POST /cases/:id/feedback`, `POST /ratings`,
`GET /ratings/units/:unitId`

---

### F5 — Alerts, news & notifications

**Screens:** alert feed, alert detail, news feed, notification center, subscriptions.

- [ ] Community alert feed with severity styling
- [ ] Alert detail with location, confirm action, share
- [ ] News feed + news-alert items
- [~] Notification center (unread badge, mark read / mark all) — bell + popover in the shell;
      full-page centre not built
- [ ] Subscription management (areas, categories, channels)
- [ ] Push opt-in with clear value framing; device register/unregister

**APIs:** `GET /alerts`, `GET /alerts/:id`, `POST /alerts/:id/confirm`, `POST /alerts/subscribe`,
`GET /alerts/subscriptions`, `GET /alerts/news`, `GET /news`, `GET /news/:id`,
`POST /notifications/register`, `DELETE /notifications/unregister`

---

### F6 — Community & prevention

**Screens:** community hub, forum post, announcements, events, event detail/RSVP.

- [ ] Forum: post list, post detail, replies, create post
- [ ] Announcements feed
- [ ] Events with RSVP + attendee count
- [ ] Prevention/safety tips surface (AI-assisted, cached)
- [ ] Moderation affordances (report content) — visible, simple

**APIs:** `POST|GET /community/posts`, `GET /community/posts/:id`, `POST /community/replies`,
`POST|GET /community/announcements`, `POST|GET /community/events`, `POST /community/events/:id/rsvp`

---

### F7 — Officer console

**Screens:** queue, active case, case actions, progress log, evidence capture/verify, map, comms.

- [x] Assigned-case queue (priority-sorted, SLA badges)
- [x] Case detail with one-tap **Dispatch** and **Arrive** (state-gated)
- [x] Progress logging (action type + description)
- [x] Evidence attach (by link) + verify evidence (with badge)
- [x] Close case with final report (guarded, requires on-scene)
- [~] Team view (primary/investigator/support) — assignment list + roles rendered in admin case
      review; the officer-facing team view is open
- [x] Responder map with lawful unit/location context
- [ ] Comms: rooms, messages, calling entry points

**States:** action-not-allowed (wrong state) surfaced clearly, never silently disabled
**APIs:** `POST /cases/:id/dispatch`, `POST /cases/:id/arrive`, `POST /cases/:id/progress`,
`GET /cases/:id/progress`, `POST /cases/:id/close`, `POST /evidence/upload`,
`PATCH /evidence/:id/verify`, `POST /cases/:id/assign`, `GET /cases/:id/assignments`

---

### F8 — Unit admin console

**Screens:** overview, dispatch board, case management, officers, verification queue, analytics,
finance, unit settings.

- [ ] Ops overview: open/assigned/dispatched cards, response-time KPIs
- [x] Dispatch board: unassigned queue → assign to officer
      *(`/admin/cases` — attention counters are filters, not decoration: each one toggles the list)*
- [~] Case review: facts, progress, weekly, evidence, assignment — done; per-case audit trail open
- [ ] Officer roster: add/edit, status, workload *(blocked by the unrouted `GetOfficersByUnit`)*
- [~] Verification queue: evidence verify shipped inside case review; memberships and gov IDs open
- [ ] Analytics: volume, response/dispatch/arrival, resolution, workload
- [ ] Finance: accounts, donations, transactions + approvals, budgets, reports
- [ ] Unit config, operational radius, contact info

**APIs:** `GET /cases`, `POST /cases/:id/assign`, `GET /cases/analytics`, `GET|POST|PUT /units`,
`GET|POST /bank/*`, `GET|POST /finance/*`, `GET /audit/*`, `POST|GET /unit-verification`*
*(verification endpoints per `unit_verification_handler.go`)*

---

### F9 — Super admin console

**Screens:** governance dashboard, users, user detail, units, audit search, analytics, health, settings.

- [ ] Users: list, search, role change, suspend/activate, impersonate/stop
- [ ] Units: registry CRUD
- [ ] Audit: searchable activity + audit logs (actor, action, entity, time)
- [ ] Platform analytics + system health
- [ ] System settings, templates, data exports
- [ ] Impersonation banner (unmissable "you are impersonating") + one-click exit

**APIs:** `GET /admin/users`, `GET /admin/users/:id`, `PUT /admin/users/:id/role`,
`POST /admin/users/:id/suspend|activate|impersonate`, `POST /admin/stop-impersonate`,
`GET /admin/stats`, `GET /audit/activities`, `GET /audit/logs`, `GET /audit/health`,
`GET|PUT /settings/:key`, `GET /settings/templates/:name`, `POST|GET /settings/exports`

---

### F10 — Maps & GIS

**Screens:** incident map, unit coverage, hotspot view, case location picker.

- [x] Incident markers (clustered) with severity/status color *(status-coloured; no clustering yet)*
- [x] Unit coverage radii
- [ ] Hotspot/heatmap layer (prevention framing, never individual profiling)
- [~] Case location picker (drag pin) — `MapView mode="pick"` is built; report wizard not wired
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

- [ ] Profile edit (name, contact, photo)
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
| Cases | `/cases`, `/cases/:id`, `/cases/:id/{timeline,progress,assign,dispatch,arrive,close,feedback}` |
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
- [~] **M1 — Auth & onboarding:** login + role routing + guards done; register/OTP/onboarding open
- [~] **M2 — Citizen core:** F4 case tracking shipped; SOS and reporting wizards open
- [~] **M3 — Awareness:** F10 case/unit map shipped; alerts, news, notifications centre open
- [~] **M4 — Officer console:** queue, case workspace (details/progress/weekly/evidence), dispatch →
      arrive → close all shipped; team view and comms open
- [~] **M5 — Unit admin:** case review + dispatch board shipped (`/admin/cases`,
      `/admin/cases/:id`) — triage counters, assignment, evidence verification; overview, roster,
      analytics, finance and settings open. **Super admin:** F9 not started
- [ ] **M6 — Depth:** F11 (comms), F12 (AI), F10 hotspot layer
- [ ] **M7 — Field hardening:** F14 offline/PWA, performance budget, a11y audit, E2E suite
- [ ] **M8 — Pilot polish:** empty/error states everywhere, i18n, analytics funnel, real-device testing

Each milestone ships only when the **Definition of Done** below is met.

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
npm run test         # vitest (ISO weeks + API client)
npm run build        # tsc --noEmit && vite build
npm run dev          # mocks on by default
```

Manual walk — officer (with mocks): sign in as **Officer** → `/officer/queue` → open the P1 robbery
case → **Details / Progress / Weekly / Evidence** tabs → add a progress update and confirm it lands
under **This week** → open **Weekly** and file a weekly summary → verify an evidence item →
**Dispatch** or **Mark on scene** on a case in the right state → `/map`, toggle status filters and
unit coverage → resize to phone width and confirm the stepper and tabs stay one-hand usable.

Manual walk — administrator (with mocks): sign in as **Unit admin** → you land on `/admin/cases` →
confirm the attention counters and that clicking one filters the list → open the pending assault case
→ **Assign officer** → pick an officer and confirm the case flips **pending → assigned** → sign back
in as **Officer** and confirm that case is now dispatchable. That round trip is the end-to-end proof
that `pending → assigned → dispatched` works, and it is the only path to the first transition.

If you are working on files that predate this rebuild, note that stale Vite entry points
(`src/App.jsx`, `src/main.jsx`, `src/App.css`, `vite.config.js`, `src/assets/`) must **not** exist:
Vite resolves `.jsx` before `.tsx`, so they would shadow the current app. Delete them if they
reappear.
