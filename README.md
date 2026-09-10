# CommunityShield

> A digital operating layer for community-centered public safety — connecting grassroots
> observations and citizen participation to accountable, permission-controlled operational
> case management.

CommunityShield is a community-oriented public-safety and security operations platform. It links
residents, community structures, security units/officers and administrators through a single,
auditable workflow: **signal → intake → triage → routing → assignment → dispatch → arrival →
investigation → evidence → resolution → audit → prevention.**

It does **not** replace lawful security institutions, and it is explicitly designed not to
encourage vigilantism.

---

## Table of contents

1. [What the system is](#1-what-the-system-is)
2. [Current status snapshot](#2-current-status-snapshot)
3. [Target architecture](#3-target-architecture)
4. [Python intelligence & automation tier](#4-python-intelligence--automation-tier)
5. [Case lifecycle](#5-case-lifecycle-as-actually-implemented)
6. [Data model](#6-data-model-at-a-glance)
7. [Build-stage roadmap](#7-build-stage-roadmap)
8. [Implementation backlog (consolidated findings)](#8-implementation-backlog-consolidated-findings)
9. [Professional UI/UX structure](#9-professional-uiux-structure)
10. [Security, privacy & compliance](#10-security-privacy--compliance)
11. [Testing strategy](#11-testing-strategy)
12. [Deployment & environments](#12-deployment--environments)
13. [Definition of Done](#13-definition-of-done-per-feature)
14. [North star](#14-north-star)
15. [Decision log](#15-decision-log)

---

## 1. What the system is

| Layer | Purpose |
|---|---|
| **Community / Citizen** | Report incidents, raise SOS, submit permitted evidence, track cases, give feedback |
| **Security Unit** | Own cases for a geographic/operational area, allocate officers, run local oversight |
| **Officer** | Work assigned cases: dispatch, arrival, progress, evidence, closure |
| **Unit Admin** | Manage officers, allocation, verification, unit configuration, local analytics |
| **Super Admin** | Platform governance, units, roles, verification, system-wide analytics, impersonation |

### Domains present in the codebase

Cases · Evidence · Progress timeline · Case officers/teams · Units & memberships · Officers ·
Users & roles · SOS alerts · Notifications (in-app/push/SMS/email) · Ratings · Transfers ·
Community (forum, announcements, events) · Alerts & subscriptions · News · Suspects & sightings ·
Finance (accounts, donations, transactions, budgets, reports) · Peacebuilding & trust scores ·
Video analytics & social monitoring · Communication (rooms, messages, calls) · Audit & activity logs ·
Settings, templates, exports, onboarding · AI assistance · Static/public endpoints · WebSocket hub.

### Stack

| Tier | Technology | Status |
|---|---|---|
| API / system of record | Go 1.21 (module `security-solution`) · Gin · GORM · PostgreSQL (UUID PKs) · JWT | Present |
| Web client | React 19 · Vite · Tailwind · lucide-react | Minimal (single-screen mock) |
| Intelligence / automation | **Python · FastAPI · Celery · scikit-learn · GeoPandas · MLflow** | **Planned (§4)** |
| Infra | PostgreSQL · Redis · Docker Compose · Render.com | Partial |
| AI (today) | OpenRouter chat completions (prompt-only) | Present, unguarded |

---

## 2. Current status snapshot

**Strong:** a genuinely broad backend domain model and API surface; a coherent, increasingly
authorized case workflow (assign → dispatch → arrive → progress → close); evidence handling with
access control; a working (prompt-only) AI integration; real deployment scaffolding.

**Weak / blocking:**

- The **frontend is a mock SOS screen**, not the application `AGENT.md` describes.
- **Zero test files** despite the testing gate in `AGENT.md`.
- **`go.sum` is git-ignored** → the backend cannot build from a clean clone, CI, or Docker.
- **Several read endpoints lack case-level authorization.**
- The **audit trail is coarse** (method + path, no structured before/after).
- **SOS escalation is non-durable** (in-process `time.Sleep` goroutine).
- **Analytics query a `resolved` state that never occurs**, so those metrics are always zero.
- **The WebSocket hub is a chat/WebRTC transport**, not a permission-aware domain event bus.
- **`AGENT.md` describes capabilities** (rich TS frontend, domain event bus, testing gate) **not yet in the tree.**

Legend for checkboxes throughout: `[x]` implemented · `[~]` partial / not hardened · `[ ]` to do.

---

## 3. Target architecture

Go remains the **API gateway, system of record and single authorization boundary**. Python is
added as a **separate intelligence & automation tier** that augments Go — it never replaces it and
never mutates core case state directly.

```
                          ┌──────────────────────────────────────────────┐
   React client  ───────▶ │  Go API (Gin)  ·  system of record           │
   (web / mobile)         │  auth · case workflow · authorization · audit │
                          └───────┬───────────────┬──────────────┬───────┘
                                  │               │              │
                     sync HTTP (timeout-guarded)  │              │ publish events
                                  ▼               │              ▼
                    ┌───────────────────────┐     │      ┌───────────────────────┐
                    │  Python intelligence  │     │      │  Broker (Redis /      │
                    │  API (FastAPI)        │     │      │  RabbitMQ)            │
                    │  triage · NLP · geo · │     │      └──────────┬────────────┘
                    │  risk · summaries     │     │                 │ consume
                    └───────────┬───────────┘     │                 ▼
                                │                 │      ┌───────────────────────┐
                                │                 │      │  Python workers       │
                                │                 │      │  (Celery/Dramatiq)    │
                                │                 │      │  notifications ·      │
                                │                 │      │  escalation · reports │
                                │                 │      │  ETL · model jobs     │
                                │                 │      └──────────┬────────────┘
                                ▼                 ▼                 ▼
                    ┌────────────────────────────────────────────────────────┐
                    │  PostgreSQL (source of truth)  ·  read replica /       │
                    │  analytics schema  ·  Redis cache  ·  object storage    │
                    └────────────────────────────────────────────────────────┘
```

### Repository layout (target)

```
communityshield-platform/
├── backend/                 Go · Gin · GORM · PostgreSQL · JWT
│   ├── cmd/migrate/         Explicit schema migration entrypoint
│   ├── config/              DB connection
│   ├── handlers/            ~40 HTTP handler files
│   ├── middleware/          auth_middleware.go · permission_middleware.go · audit_middleware.go
│   ├── models/              ~47 GORM domain models
│   ├── routes/              Central route registration
│   ├── services/            auth · AI client · SMS · audit
│   └── websocket/           Room-based hub (chat / WebRTC / presence)
├── intelligence/            Python · FastAPI · Celery  (NEW — §4)
│   ├── api/                 FastAPI app: triage, classify, summarize, risk, hotspots
│   ├── workers/             Celery/Dramatiq tasks: notifications, escalation, reports, ETL
│   ├── ml/                  models, training, evaluation, registry (MLflow)
│   ├── nlp/                 classification · NER · sentiment · multilingual · ASR
│   ├── geo/                 clustering · heatmaps · coverage · routing
│   ├── pipelines/           ETL · analytics aggregations · scheduled reporting
│   └── shared/              DB (read replica) · schemas · config
├── frontend/                React 19 · Vite · Tailwind
├── docs/                    MOBILE_API.md · contracts/
├── AGENT.md                 Master architecture & product blueprint
├── README.md                This document
├── docker-compose.yml
├── render.yaml
└── deploy.sh
```

### Architectural principles

1. **One authorization boundary.** All permission checks and case-state mutations live in Go.
2. **Python recommends, Go decides.** Python returns recommendations/events; Go applies them after
   re-checking authorization. No direct state writes from Python.
3. **Fail-open for UX, fail-closed for security.** A dropped Python call must never block reporting
   or SOS; it degrades to a non-AI path. A failed *authorization* check always denies.
4. **Events, not RPC, for side effects.** Notifications, escalation, reporting and ETL are async.
5. **Derived data is disposable.** Python-written analytics can be rebuilt from Postgres at any time.

---

## 4. Python intelligence & automation tier

**Recommendation: Python is a separate tier — never embedded in the Go monolith.** Go is excellent
at the transactional core (auth, state machine, authorization) and poor at the things Python
dominates (ML, NLP, geospatial, dataframes, durable async orchestration). Splitting them keeps Go's
security boundary single and lets Python scale, deploy and upgrade independently.

### 4.1 Service map — where each piece lives and why

| Python service | Augments (Go section) | Responsibility | Why Python |
|---|---|---|---|
| `intelligence-api` (FastAPI) | `handlers/ai_handler.go`, `services/ai_service.go` | Triage scoring, classification, summarisation, risk, duplicate detection | Mature ML/NLP ecosystem; replaces prompt-only OpenRouter calls |
| `workers` (Celery/Dramatiq) | SOS escalation, `notification_handler.go`, SMS/email/push | Durable, retryable, scheduled async work | Go goroutines + `time.Sleep` are lost on restart (§7 Stage 7) |
| `analytics` (pipelines) | `analytics_handler.go` | ETL, aggregation, KPI/report generation | Correct the broken `resolved` queries; pandas/polars/BI |
| `geo` (GeoPandas/PostGIS) | Case GIS fields, SOS proximity | Hotspot clustering, coverage, heatmaps, routing | Geospatial tooling is Python-native |
| `ml-lifecycle` (MLflow) | AI model management | Training, evaluation, versioning, drift monitoring | Reproducible model registry |

### 4.2 Communication contracts

- **Synchronous (Go → Python):** request-time inference only — e.g. a triage/priority suggestion
  when a case is created. Strict client timeout (e.g. 1500 ms), circuit breaker, and a
  deterministic fallback (rule-based priority) so case creation never fails because Python is down.
- **Asynchronous (Python → Go):** Python publishes domain events to the broker; the Go consumer
  performs the state change after re-verifying authorization and emitting its own audit record.
- **Async (Go → Python):** Go enqueues jobs (generate report, run clustering, send notification);
  workers execute and write results back through the broker, not by touching case tables.

### 4.3 Shared infrastructure

- **PostgreSQL** — source of truth. Python uses a **read replica** plus its own `analytics` schema;
  it has no write grant on core case/evidence/audit tables.
- **Redis** — broker for Celery + cache for hot reads.
- **Object storage** — evidence and generated reports (see Stage 4 hardening).
- **Schema ownership** — Postgres stays **Go-owned** (GORM migrations). Python reads, and writes
  only its own schema. Cross-schema changes go through a review.

### 4.4 Python release gates

- [ ] Contract tests: Go client ⇄ FastAPI schema compatibility in CI
- [ ] Every Python endpoint authenticated (service-to-service token / mTLS)
- [ ] Pinned, scanned dependencies; no secrets in images
- [ ] Model inputs/outputs logged for auditability, with PII redaction
- [ ] Python unavailable ⇒ documented Go fallback path per call site
- [ ] No write access from Python to core case/evidence/audit tables (enforced at DB grant level)

---

## 5. Case lifecycle (as actually implemented)

The code implements a **five-state linear pipeline** with hardened transitions:

```
pending ──assign──▶ assigned ──dispatch──▶ dispatched ──arrive──▶ on_scene ──close──▶ closed
   │                    │                       │                      │
   │ AssignedTo/At      │ DispatchedAt          │ ArrivedAt            │ ClosedAt/By/FinalReport
```

- `AssignCase` → sets primary officer + `AssignedAt`, moves `pending → assigned`
- `DispatchCase` → `pending|assigned → dispatched`, sets `DispatchedAt`
- `ArriveAtCase` → `dispatched → on_scene`, sets `ArrivedAt`
- `CloseCase` → `on_scene → closed`, records `ClosedAt`, `ClosedBy`, `FinalReport`
- `AddCaseProgress` → permitted only while `dispatched | on_scene`
- Evidence upload/delete → blocked on `closed` cases

Each of `Assigned/Dispatched/Arrived/Closed` timestamps enables **response-time metrics**.

> ⚠️ **Naming note:** `AGENT.md` describes a richer model (`REPORTED → PENDING → TRIAGE →
> ASSIGNED → DISPATCHED → ARRIVED → INVESTIGATING → RESOLVED → CLOSED → APPROVED/ARCHIVED`).
> The code implements only the linear pipeline above — no triage, `investigating`, `resolved`, or
> approval/archive state exists. Reconcile the two (either extend the state machine or amend the
> blueprint).

---

## 6. Data model at a glance

| Model | Key fields |
|---|---|
| **User** | UUID, email, phone, name, role, unitID, status, super-admin flag, impersonation, medical info, timestamps, soft delete |
| **Officer** | UUID, unitID, name, rank, badge, role, contact, joined date, status, unit relation |
| **Case** | unitID, reportedBy, assignedTo, title, description, incident date, location, lat/lng, status, priority, trackingID, priorityLevel, GIS lat/lng, assigned/dispatched/arrived/closed timestamps, closedBy, approvedBy, final report, evidence + progress relations |
| **Evidence** | caseID, uploadedBy, type, fileUrl, description, lat/lng, isVerified, uploadedAt |
| **Progress** | caseID, officerID, action, description |
| **CaseOfficer** | caseID, officerID, role (`primary`/`investigator`/`support`) |

Plus the extended domain: unit/membership/verification, SOS, notifications, ratings, transfers,
community, alerts, news, suspects, finance, peacebuilding, video analytics, communication, audit,
settings, AI analysis.

> **Data-model risk:** `Evidence.FileURL` is client-supplied and `Evidence` has no verifier/timestamp
> or hash — see Stage 4 and backlog item **P0-4**.

---

## 7. Build-stage roadmap

### Stage 0 — Repository & engineering foundations

- [x] Git repository with commit history
- [x] Go module + Gin/GORM backend scaffold
- [x] PostgreSQL model layer with UUID entities
- [x] Explicit migration entrypoint (`cmd/migrate`)
- [x] Docker Compose definition (postgres + backend + frontend)
- [x] Render.com deployment manifest
- [x] `AGENT.md` master blueprint
- [x] Root `README.md` (this file)
- [ ] **Backend builds reproducibly** — `go.sum` is git-ignored (`.gitignore:44`); `go build ./...`
      fails for any fresh clone / CI / Docker build → **P0-1**
- [ ] Consistent environment story — `render.yaml` sets `ALLOWED_ORIGINS` but `main.go` hard-codes
      `AllowOrigins: ["*"]`; `Dockerfile` expects an absent `.env.production` → **P0-2**
- [ ] Remove stray artifacts (`backend/curl`, `backend/Logs`, `hash.go.bak`, tracked
      `backend/uploads/**`, `.gitignore` line `h origin mai`) → **P1-6**

### Stage 1 — Foundation: identity, users, units, roles

- [x] User model + registration/login/logout/profile
- [x] Password hashing (bcrypt via `services/auth_service.go`)
- [x] JWT issuance + validation
- [x] DB-backed auth middleware injecting `user`/`user_id`/`role`
- [x] `RoleMiddleware` + `RequireRole` primitives
- [x] Security unit model, membership, government-ID verification
- [x] Officer model + endpoints
- [x] OTP send/verify/resend
- [x] Super-admin routes with `SuperAdminMiddleware`
- [~] Unit/resource authorization — `CanAccessCase`/`CanAccessUnit` exist but are **not wired onto
      most routes**; most handlers gate on role only → **P0-3**
- [ ] Refresh tokens / token revocation
- [ ] Rate limiting on auth + OTP endpoints
- [ ] Field-level protection for sensitive data (medical info) → **P1-2**

### Stage 2 — Case intake & lifecycle

- [x] Case model with location, GIS coords, priority, tracking ID, lifecycle timestamps
- [x] Create/list/get case endpoints
- [x] Case analytics endpoint
- [~] Case state transitions enforced in workflow handlers, but the state machine is linear and
      diverges from the blueprint
- [x] Timeline add/get
- [x] Case feedback endpoint
- [ ] Approval/archival transition (`ApprovedBy` exists, no handler sets it)
- [ ] Triage / priority-scoring workflow → **Python `intelligence-api`** (§4)
- [ ] Duplicate detection → **Python** (§4)

### Stage 3 — Assignment, dispatch, arrival, progress

- [x] `AssignCase` — primary officer via `Case.AssignedTo`, team via `CaseOfficer`
- [x] `GetCaseAssignments`
- [x] `DispatchCase` / `ArriveAtCase` with unit + assignment authorization
- [x] `AddCaseProgress` with unit + assignment authorization
- [~] `GetCaseProgress` — **no case-level authorization check**; any authenticated user can read
      any case's progress → **P0-3**
- [~] `GetCaseAssignments` — **no case-level authorization check** → **P0-3**
- [ ] Assignment rules fully enforced (officer active, requester authority, auditability)
- [ ] Full team-role semantics (`primary`/`investigator`/`support`) surfaced in workflow

### Stage 4 — Evidence

- [x] Evidence model (case link, uploader, type, file ref, geo, verify flag, timestamps)
- [x] Upload / list-by-case / verify / delete handlers
- [x] Case-access authorization on upload + list; officer authorization on verify + delete
- [x] Closed-case mutation guard
- [x] Route parameter/authentication mismatch repaired (`/case/:caseId`)
- [ ] Controlled object storage (currently arbitrary client-supplied `fileUrl`) → **P0-4**
- [ ] Size/type restriction + content validation → **P0-4**
- [ ] Verifier identity + timestamp recorded → **P0-4**
- [ ] Cryptographic hash / chain-of-custody metadata → **P0-4**
- [ ] Guarantee evidence is never exposed via public endpoints → **P1-1**

### Stage 5 — Authorization & data-integrity hardening

- [x] Case workflow endpoints hardened (dispatch/arrive/close)
- [x] Progress-add hardened
- [x] Evidence endpoints hardened
- [x] Super-admin route group gated
- [~] Read endpoints — case list/get, timeline, feedback, mobile dashboard/sync, analytics; several
      leak cross-unit or global counts → **P0-3**
- [~] `GetPublicCases` exposes non-closed `isPublic` cases with no auth → **P1-1**
- [ ] Uniform authorization middleware applied across every case/unit route → **P0-3**
- [ ] Unit-scoped query filtering in every list/DTO → **P0-3**

### Stage 6 — Audit, notifications, realtime

- [x] Audit middleware (user, method, path, IP, user-agent, timestamp)
- [x] Audit/activity/health/notification-log endpoints
- [x] Notification model + in-app notifications
- [x] Push subscription + FCM token handlers
- [x] Email (SMTP) + SMS handlers
- [x] WebSocket hub (rooms, broadcast, presence, WebRTC signaling, chat)
- [~] Audit quality — `EntityID` = raw query string; `OldValue` reused for errors/slow-requests;
      not structured or tamper-evident → **P0-5**
- [ ] **Permission-aware domain events** (`case.assigned`, `case.dispatched`, `evidence.verified`,
      `sos.created`, …) — current hub is a chat/WebRTC transport → **P1-3**
- [ ] Durable background queue for notifications → **Python `workers`** (§4)
- [ ] SMS/email actually dispatched (some SOS paths log only) → **P1-4**

### Stage 7 — SOS emergency workflow

- [x] SOS model + send/list/get/status endpoints with role-filtered reads
- [x] Priority defaulting, optional unit routing, emergency-contact + medical-info intake
- [x] Nearest-unit notification via Haversine radius
- [x] 5-minute escalation → notifies super admins
- [~] Escalation runs as an in-process goroutine + `time.Sleep` — **lost on restart**, not durable
      → **P0-6 (Python `workers`)**
- [~] Emergency-contact notification logs only → **P1-4**
- [ ] Abuse / rate controls
- [ ] Realtime dispatch + arrival integration
- [ ] Full SOS lifecycle audit trail

### Stage 8 — Community layer

- [x] Community forum (posts/replies), announcements, events + RSVP
- [x] Community alerts + subscriptions
- [x] News + news-alert feeds
- [x] Ratings for units
- [x] Public cases/units endpoints for landing pages
- [ ] Community reporting UX wired to the frontend
- [ ] Moderated, controlled community visibility of cases
- [ ] Feedback loop from resolution back to community
- [ ] Multilingual + Pidgin/Hausa/Yoruba/Igbo support → **Python `nlp`** (§4)

### Stage 9 — Extended operations (present in code)

- [x] Finance: accounts, donations, transactions + approvals, budgets, reports
- [x] Peacebuilding: committees, conflict resolution, trust scores, metrics
- [x] Suspects: records, sightings, associations, linked cases
- [x] Video analytics: cameras, alerts, review; social-media monitoring
- [x] Transfers: request/approve/reject + approvals
- [~] These modules share the same authorization/validation gaps as the core → **P0-3**
- [ ] Product-scope decision: gate foundation vs. expansion features

### Stage 10 — Frontend & professional UI/UX

> **Reality check:** the frontend is currently a **single 126-line mobile-style SOS console**
> (`frontend/src/App.jsx`) with mock geolocation and no API wiring, routing, auth or state
> management. `AGENT.md` describes a rich TypeScript app that is **not in this tree.** The plan
> below is the target.

- [x] Tactical dark theme tokens (`tailwind.config.js`, `index.css`) and SOS console shell
- [x] Component/icon foundation (lucide-react)
- [~] Tailwind wiring — Tailwind **v4** with **v3-style** `@tailwind` directives, config at repo
      root instead of `frontend/`, no `postcss.config.js` → styling not reliably built → **P1-5**
- [ ] API client + auth/session handling (`VITE_API_URL` defined but unused) → **P0-7**
- [ ] Client-side routing (react-router) + layout shells
- [ ] Design system: tokens, type scale, spacing, components, empty/loading/error states
- [ ] Role-based information architecture (citizen / officer / unit admin / super admin)
- [ ] Core screens: onboarding → report → track → SOS → unit map → case detail → progress → evidence
- [ ] Operational dashboards: dispatch board, case queue, analytics
- [ ] Accessibility (WCAG AA), light/dark parity, i18n, offline/PWA field mode
- [ ] Engagement mechanics (§9)

### Stage 11 — GIS, analytics & AI

- [x] GIS coordinate fields on cases/units; Haversine proximity for SOS
- [x] Case analytics + unit/weekly aggregation queries
- [~] Analytics status vocabulary mismatch — queries count `status = "resolved"`, a state the
      workflow never sets → those metrics are always zero → **P0-8**
- [x] AI service integrated with OpenRouter — chatbot, image analysis, location risk, news
      sentiment, warnings, hotspot prediction, case summary
- [~] AI operates on prompt-only guardrails; no output validation or human-in-the-loop enforcement
      → **P1-7**
- [ ] Incident/coverage/hotspot map UI → **Python `geo`** (§4)
- [ ] Response-time geography + prevention analytics → **Python `analytics`** (§4)
- [ ] AI triage wired into intake (downstream of trustworthy data) → **Python `intelligence-api`**
- [ ] Documented prohibition: AI never declares guilt or authorizes enforcement

### Stage 12 — Testing gate & CI

- [ ] **Backend test suite** — currently **zero `_test.go` files** despite the `AGENT.md` gate.
      Required: unauthenticated, wrong role, wrong unit, authorized success, invalid UUID, missing
      fields, missing resource, duplicate action, unauthorized mutation → **P0-9**
- [ ] `go test ./...` + `go vet ./...` enforced in CI → **P0-9**
- [ ] Frontend unit + component tests
- [ ] Integration / E2E tests
- [ ] Python contract tests (Go ⇄ FastAPI) → **P0-9**
- [ ] Load and security testing
- [ ] Restore `go.sum` to version control so CI can build → **P0-1**

### Stage 13 — Staging & deployment

- [x] Local Docker Compose topology
- [x] Render manifest (backend, static frontend, managed Postgres) with health check
- [~] Build/deploy pipeline — `deploy.sh` commits **all** changes with a timestamped message;
      `Dockerfile` depends on missing `go.sum`/`.env.production` → **P0-2**
- [ ] Environment separation (local → dev → staging → pilot → multi-unit)
- [ ] Secret management (no hard-coded fallback secrets in compose) → **P1-8**
- [ ] HTTPS, CORS allow-list (currently `*`), backup/restore, monitoring & alerting
- [ ] Python tier deployment (separate service + worker dynos) → **§4**
- [ ] Pilot community/unit rollout

### Stage 14 — Institutional integration & scale

- [ ] Agency interoperability
- [ ] SMS/USSD/IVR channels (handlers exist; production integration pending)
- [ ] Approved messaging integrations
- [ ] Identity and records integrations
- [ ] Multi-state / multi-agency, disaster recovery, high availability, enterprise governance
- [ ] Data warehouse / BI layer fed by Python pipelines → **§4**

---

## 8. Implementation backlog (consolidated findings)

Every finding from the audit, prioritised. `P0` = blocks a trustworthy product; `P1` = needed
before pilot; `P2` = scale/quality.

| ID | Priority | Finding | Fix | Layer |
|---|---|---|---|---|
| **P0-1** | P0 | `go.sum` git-ignored → backend unbuildable from clean checkout | Remove from `.gitignore`, commit `go.sum`, verify clean-clone build | Repo/CI |
| **P0-2** | P0 | Build/deploy broken: missing `go.sum`, absent `.env.production`, `ALLOWED_ORIGINS` ignored, CORS `*` | Fix Dockerfile/render.yaml, wire CORS allow-list, secret management | Deploy |
| **P0-3** | P0 | Authorization gaps: `GetCaseProgress`, `GetCaseAssignments`, case list/get, timeline, feedback, mobile/analytics scoping | Apply `CanAccessCase`/unit-scoped middleware uniformly; add tests | AuthZ |
| **P0-4** | P0 | Evidence integrity: client-supplied `fileUrl`, no type/size limits, no verifier/timestamp/hash | Object storage, validation, verifier fields, hash/chain-of-custody | Evidence |
| **P0-5** | P0 | Audit trail coarse/incorrect (`EntityID`=query, `OldValue` overloaded) | Structured before/after, entity id, actor, outcome; tamper-evident store | Audit |
| **P0-6** | P0 | SOS escalation non-durable (`time.Sleep` goroutine) | Durable scheduler via Python `workers` | Realtime |
| **P0-7** | P0 | Frontend has no API client, auth, routing or state | Build real client per §9 | Frontend |
| **P0-8** | P0 | Analytics reference non-existent `resolved` state | Fix vocab + SQL; move aggregation to Python `analytics` | Analytics |
| **P0-9** | P0 | Zero tests; no CI gate | Test matrix (§11) + CI enforcing build/vet/test | Testing |
| **P1-1** | P1 | `GetPublicCases` exposes non-closed cases unauthenticated | Restrict fields/state, explicit allow-list DTO | Privacy |
| **P1-2** | P1 | Sensitive fields (medical info) lack field-level protection | Encrypt at rest, restrict access, audit reads | Privacy |
| **P1-3** | P1 | No permission-aware domain event bus | Event schema + broker; Go publishes, clients subscribe by permission | Realtime |
| **P1-4** | P1 | SOS/notification paths log instead of send | Wire SMS/email/push providers; delivery status + retries | Notifications |
| **P1-5** | P1 | Tailwind v4/v3 mismatch; config location; no postcss config | Align Tailwind setup in `frontend/` | Frontend |
| **P1-6** | P1 | Stray artifacts + tracked binary; `.gitignore` `h origin mai` | Clean repo | Repo |
| **P1-7** | P1 | AI guardrails are prompt-only | Output validation, human-in-the-loop, prohibition policy | AI |
| **P1-8** | P1 | Hard-coded fallback secrets in compose | Secret manager / env injection | Deploy |
| **P2-1** | P2 | Python tier: API, workers, analytics, geo, ML lifecycle | Stand up per §4 | Python |
| **P2-2** | P2 | Load, security, E2E, accessibility testing | Test pyramid extension | Testing |
| **P2-3** | P2 | Multi-unit → multi-state scale, DR, HA | Infra hardening | Scale |

---

## 9. Professional UI/UX structure

The goal is a system people **trust and return to** — not a feature showcase. Design is the trust
layer: citizens must feel safe reporting, officers must move fast under pressure, and admins must
see the whole picture without being overwhelmed.

### 9.1 Design principles

1. **Calm under stress** — high-contrast, low-noise emergency surfaces; one primary action per screen.
2. **Trust through transparency** — always show case status, assigned officer/unit, and last update.
3. **Progress you can see** — a visible 5-stage lifecycle tracker on every case.
4. **One-tap safety** — SOS reachable from anywhere, with confirm + cancel (no accidental triggers).
5. **Role-appropriate density** — citizens see clarity; officers see throughput; admins see control.
6. **Mobile-first field operation**, responsive to desktop command centres.
7. **Honest states** — real loading/empty/error/skeleton states; never fake "API Live" indicators.

### 9.2 Information architecture (by role)

| Role | Primary nav | Key screens |
|---|---|---|
| Citizen | Home · Report · Track · Alerts · Profile | SOS console, report wizard, case tracker, safety map, notifications |
| Officer | Queue · Active · Map · Comms · Profile | Assigned cases, dispatch/arrival actions, progress log, evidence capture, chat |
| Unit Admin | Overview · Cases · Officers · Analytics · Config | Dispatch board, workload, response-time metrics, verification queue |
| Super Admin | Governance · Units · Users · Audit · Analytics | Role/permission control, unit registry, audit search, platform metrics |

### 9.3 Core component inventory

- **Layout:** app shell, role switcher, notification center, toasts
- **Input:** multi-step report wizard, media capture, geolocation picker, search/filter bar
- **Status:** lifecycle stepper, priority chip, SLA/response-time badge, verification badge
- **Case:** case card, case detail header, timeline/activity feed, evidence gallery, assignment panel
- **Map:** incident map, unit coverage, hotspot clustering, heatmap layer
- **Data:** KPI tiles, trend charts, workload bars, response-time distributions
- **Feedback:** rating, resolution summary, follow-up prompt

### 9.4 Engagement mechanics (ethical, not manipulative)

- **Progress transparency** — visible stage advancement + timestamps builds trust and return visits.
- **Notifications that matter** — status change, officer assigned, resolution, local alert.
- **Community participation** — announcements, events, forum, prevention tips, with moderation.
- **Recognition where lawful** — unit response-time/quality dashboards (no individual profiling).
- **Graceful degradation** — offline queue + sync for low-connectivity field use.
- **Safety by default** — clear consent, privacy controls, visible abuse-reporting.

### 9.5 Accessibility & inclusion

- WCAG AA contrast (critical for a dark tactical theme used in the field).
- Large touch targets; works one-handed, in sunlight, on low-end Android.
- Multilingual + Pidgin/Hausa/Yoruba/Igbo copy (Python `nlp` for classification/summaries).
- Works on low bandwidth; SMS/USSD fallbacks in Stage 14.

---

## 10. Security, privacy & compliance

**Critical requirements**

- Password hashing; validated JWTs; least privilege throughout
- Secure evidence storage; input validation; rate limiting on auth/OTP/SOS
- Structured audit logs; secret management; HTTPS; explicit CORS policy
- Backup/restore; privacy controls; data-retention policy

**Principles**

- Medical and other sensitive data require stronger controls than ordinary profile data.
- Evidence must never leak through public endpoints.
- AI must never declare guilt, authorize force, make irreversible enforcement decisions, or expose
  protected information.
- GIS must support prevention and resource allocation — never individual profiling.

**Checklist**

- [ ] Rate limiting (auth, OTP, SOS)
- [ ] Secret manager (no hard-coded secrets)
- [ ] CORS allow-list
- [ ] Evidence object storage + validation + hashing
- [ ] Field-level encryption for medical info
- [ ] Data-retention + export + deletion (GDPR/NDPA-style) flows
- [ ] Structured, tamper-evident audit trail
- [ ] Security review before pilot

---

## 11. Testing strategy

**Minimum backend gate**

```bash
go build ./...
go vet ./...
go test ./...
```

**API test matrix** (per case/evidence/progress endpoint):

- [ ] Unauthenticated → 401
- [ ] Wrong role → 403
- [ ] Wrong unit → 403
- [ ] Authorized success → 2xx
- [ ] Invalid UUID → 400
- [ ] Missing required fields → 400
- [ ] Missing resource → 404
- [ ] Duplicate action → 409
- [ ] Unauthorized mutation → 403

**Later:** integration, frontend component, E2E, load, security and **Go ⇄ Python contract** tests.

---

## 12. Deployment & environments

```
LOCAL → DEVELOPMENT → STAGING → PILOT COMMUNITY/UNIT → MULTI-UNIT → STATE/REGIONAL SCALE
```

- [x] Local Docker Compose (postgres + backend + frontend)
- [x] Render manifest (managed Postgres, health check at `/health`)
- [ ] Fix build prerequisites (`go.sum`, `.env.production`) before any deploy
- [ ] Deploy Python tier as separate API + worker services
- [ ] Do not scale before authorization, evidence safety, auditing, backups and monitoring are reliable

---

## 13. Definition of Done (per feature)

```
MODEL → MIGRATION → SERVICE → HANDLER → AUTHORIZATION → ROUTE
      → TEST → FRONTEND → REALTIME/NOTIFICATION → AUDIT
      → ANALYTICS → DOCUMENTATION
```

If a layer is not applicable, record why.

---

## 14. North star

```
COMMUNITY → REPORT / SOS / INTELLIGENCE → SECURE INTAKE → TRIAGE + PRIORITY
  → UNIT ROUTING → OFFICER ASSIGNMENT → DISPATCH → ARRIVAL → INVESTIGATION
     (PROGRESS · EVIDENCE · COMMUNICATION · COLLABORATION)
  → RESOLUTION → APPROVAL / CLOSURE → AUDIT + ANALYTICS
  → COMMUNITY FEEDBACK → PREVENTION → BETTER LOCAL SECURITY
```

**North-star principle:** build digital infrastructure that lets local people, legitimate security
institutions and accountable technology work together around real incidents — quickly,
transparently, safely and with measurable outcomes.

---

## 15. Decision log

- Go/Gin/GORM backend as the system of record and single authorization boundary.
- PostgreSQL UUID entities; GORM-owned schema.
- JWT authentication with database-backed authorization.
- `Case.AssignedTo` = primary officer; `CaseOfficer` = full team.
- `Progress` = operational timeline; `Evidence` = case-linked evidence.
- Explicit dispatch/arrival/closure timestamps for response-time metrics.
- **Python added as a separate intelligence & automation tier — never embedded in Go, never a
  writer of core case state** (§4).
- AI remains downstream of reliable operational data and is human-in-the-loop.
- Grassroots integration is a core architecture principle.

See [`AGENT.md`](./AGENT.md) for the full architecture and product blueprint.
