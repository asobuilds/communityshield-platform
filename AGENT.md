# CommunityShield — AGENT.md

## Master Architecture, Product Blueprint, Engineering State & Execution Plan

> Durable blueprint for the CommunityShield project. Update Current State, Next Actions, and Decision Log at major milestones.

## 1. Product Identity

**CommunityShield** is a community-centered public-safety and security operations platform connecting grassroots residents, community structures, security units/officers, and administrators through a common digital workflow.

### Problem

Grassroots security information is often fragmented across calls, messaging apps, social media, paper records, informal networks, and disconnected agency workflows. This can cause delayed reporting/response, weak case visibility, poor evidence continuity, weak accountability, poor coordination, limited location intelligence, and loss of institutional memory.

### Core outcome

```text
Community signal
 -> secure intake
 -> triage/prioritization
 -> unit routing
 -> officer assignment
 -> dispatch
 -> arrival
 -> investigation/progress
 -> evidence
 -> resolution/closure
 -> audit/analytics
 -> prevention feedback
```

CommunityShield should improve coordination without replacing lawful security institutions or encouraging vigilantism.

## 2. Product Principles

1. Grassroots-first: communities are an intelligence and prevention layer.
2. Institutional workflow: reports become structured, trackable cases.
3. Accountability by design: actions are attributable to authenticated users.
4. Evidence integrity: evidence remains linked to cases and actors.
5. Least privilege: citizens, officers, unit admins and super admins have different powers.
6. Location-aware operations.
7. Explicit assignment/dispatch/arrival lifecycle.
8. Human-in-the-loop AI.
9. Privacy and safety by default.
10. Auditability.
11. Local adaptability.
12. Mobile-first field operation.

## 3. Positioning and Novelty

Do **not** claim that CommunityShield invented digital crime reporting, GIS, evidence management, or case management. Existing products and research already demonstrate those capabilities.

The defensible differentiation is the integrated, grassroots-to-operational architecture:

- community-originated reports;
- security-unit ownership;
- officer assignment;
- dispatch/arrival lifecycle;
- progress timeline;
- case-linked evidence;
- location intelligence;
- controlled community visibility;
- role/unit/resource authorization;
- operational analytics;
- future AI-assisted triage and pattern analysis;
- adaptation to Nigerian grassroots structures.

The novelty is therefore primarily **system integration, workflow design, local context, governance, and deployment approach**.

## 4. Research Context

Research and current systems show that this is an active field:

- The Nigeria Police Force has maintained community-policing structures and describes community policing as an established strategy.
- Nigerian policy statements continue to emphasize local participation and grassroots ownership.
- Nigerian research identifies police-public trust, responsiveness, technology and digital engagement as important factors.
- The Nigeria Police National Cybercrime Centre already provides a guided e-reporting portal with case tracking.
- African commercial systems already combine incident/case management, evidence, GIS, community reporting and field-officer workflows.
- Academic prototypes also combine mobile reporting, GPS, multimedia evidence, notifications and crime mapping.

Implication: CommunityShield should differentiate through workflow quality, local usability, interoperability, trust, accountability and governance rather than simply feature count.

## 5. System Actors

### Citizen / Community User
Registration, reporting, location/description submission, permitted evidence submission, permitted case tracking, notifications, feedback, SOS.

### Officer
Assigned-case access, operational actions, dispatch, arrival, progress, evidence actions and team participation according to permissions.

### Unit Administrator
Unit officers, allocation, operational oversight, local analytics, verification and unit configuration.

### Super Administrator
Platform-wide governance, units, permissions, verification, exceptional administration and system-wide analytics.

## 6. Backend Architecture

```text
backend/
├── cmd/migrate/
├── config/
├── handlers/
├── middleware/
├── models/
├── routes/
├── services/
└── websocket/
```

### models/
GORM domain/database models.

Important current models:
- User
- Officer
- Case
- Evidence
- Progress
- CaseOfficer
- SecurityUnit and related unit structures

### handlers/
HTTP request/response layer.

Current important case handlers:
- AssignCase
- GetCaseAssignments
- UpdateCaseStatus
- DispatchCase
- ArriveAtCase
- AddCaseProgress
- GetCaseProgress

Evidence handlers:
- UploadEvidence
- GetEvidenceByCase
- VerifyEvidence
- DeleteEvidence

### middleware/
Actual authentication file:

```text
middleware/auth_middleware.go
```

It:
- reads Bearer JWT;
- validates token;
- resolves database user;
- stores `user`, `user_id`, `role` in Gin context;
- provides RoleMiddleware.

Do not refer to `middleware/auth.go`.

### routes/
Central API registration.

Current case workflow routes:

```text
GET    /cases
POST   /cases
GET    /cases/analytics
GET    /cases/:id
PUT    /cases/:id
POST   /cases/:id/timeline
GET    /cases/:id/timeline
POST   /cases/:id/feedback
POST   /cases/:id/assign
GET    /cases/:id/assignments
POST   /cases/:id/dispatch
POST   /cases/:id/arrive
POST   /cases/:id/progress
GET    /cases/:id/progress
```

Evidence routes should be:

```text
POST   /evidence/upload
GET    /evidence/case/:caseId
DELETE /evidence/:id
PATCH  /evidence/:id/verify
```

## 7. Core Data Model

### User
UUID, Email, Phone, FirstName, LastName, Password, Role, UnitID, Status, IsSuperAdmin, Impersonating, MedicalInfo, LastLogin, timestamps, soft delete.

### Officer
UUID, UnitID, Name, Rank, BadgeNumber, Role, Phone, Email, JoinedDate, Status, timestamps, soft delete, Unit relationship.

### Case
Important fields:
- ID
- UnitID
- ReportedBy
- AssignedTo
- Title
- Description
- IncidentDate
- Location
- Latitude/Longitude
- Status
- Priority
- TransferDetails
- IsPublic
- TrackingID
- PriorityLevel
- GISLatitude/GISLongitude
- AssignedAt
- DispatchedAt
- ArrivedAt
- ClosedAt
- ClosedBy
- ApprovedBy
- FinalReport
- timestamps
- Evidence relationship
- Progress relationship

### Evidence
Current fields:
- ID
- CaseID
- UploadedBy
- Type
- FileURL
- Description
- Latitude
- Longitude
- IsVerified
- UploadedAt
- timestamps
- soft delete

### Progress
Current fields:
- ID
- CaseID
- OfficerID
- Action
- Description
- timestamps
- soft delete

### CaseOfficer
Supports multiple officers on a case:
- ID
- CaseID
- OfficerID
- Role
- timestamps
- soft delete

Roles: primary, investigator, support.

## 8. Canonical Case Lifecycle

```text
REPORTED
  -> PENDING / TRIAGE
  -> ASSIGNED
  -> DISPATCHED
  -> ARRIVED
  -> INVESTIGATING / IN_PROGRESS
  -> EVIDENCE + PROGRESS
  -> RESOLVED
  -> CLOSED
  -> APPROVED / ARCHIVED
```

Important timestamps:
- AssignedAt
- DispatchedAt
- ArrivedAt
- ClosedAt

These enable response-time metrics.

## 9. Assignment Architecture

`Case.AssignedTo` = primary officer.

`CaseOfficer` = complete operational team.

Future assignment rules:
1. Case exists.
2. Officer exists.
3. Officer is active.
4. Officer belongs to relevant unit where required.
5. Requester has authority.
6. Assignment timestamp is recorded.
7. Assignment is auditable.

## 10. Evidence Architecture

```text
UPLOAD
 -> STORE FILE REFERENCE
 -> ATTACH TO CASE
 -> OPTIONAL GEOLOCATION
 -> VERIFY
 -> USE IN INVESTIGATION
 -> RETAIN / ARCHIVE
```

Current handler is functional but requires hardening:
- authenticate uploader;
- verify uploader has case access;
- validate evidence type;
- validate storage reference;
- eventually use controlled object storage instead of arbitrary file URLs;
- enforce size/type restrictions;
- restrict verification;
- restrict deletion;
- record verifier and timestamp;
- eventually add cryptographic hash metadata;
- never expose sensitive evidence through public endpoints.

## 11. Progress Timeline

Progress records represent operational activity such as:
- investigation_started
- interview_conducted
- evidence_collected
- suspect_identified
- follow_up_required
- unit_notified
- case_resolved

Each event should contain case, officer, action, description and timestamp.

The progress duplicate-handler issue was resolved and the backend subsequently passed `go test ./...` and `go vet ./...`.

## 12. Authorization

Current authentication is JWT-based.

Target authorization model:

```text
Authentication
 -> Identity
 -> Role
 -> Unit membership
 -> Resource access
 -> Action permission
```

Role alone must not automatically grant access to every case.

## 13. GIS

Cases already support textual and coordinate fields.

Future GIS:
- incident maps;
- unit coverage;
- lawful responder location;
- clustering/hotspots;
- response-time geography;
- community risk patterns.

GIS must support prevention/resource allocation, not profiling individuals.

## 14. Realtime

Existing WebSocket infrastructure should eventually publish permission-aware events:

```text
case.created
case.assigned
case.dispatched
case.arrived
case.progress_added
evidence.uploaded
evidence.verified
case.status_changed
sos.created
alert.created
notification.created
```

## 15. Notifications

Channels:
- in-app
- push
- SMS
- email
- future approved messaging integrations

Use background processing/queues as scale requires.

## 16. SOS

SOS is an emergency workflow:

```text
SOS
 -> validation/rate control
 -> priority escalation
 -> appropriate unit
 -> realtime notification
 -> dispatch
 -> arrival
 -> resolution
```

Must include abuse controls, escalation, location handling and audit.

## 17. Analytics

Operational:
- reports by period
- open/closed cases
- response time
- dispatch time
- arrival time
- resolution time
- officer workload
- unit workload

Community:
- reporting volume
- recurring areas
- issue categories
- satisfaction
- unresolved concerns

Prevention:
- recurring locations
- trends
- hotspot changes
- intervention outcomes

Do not interpret correlations as proof of criminality.

## 18. AI

AI is downstream of trustworthy data.

Potential uses:
- incident classification;
- priority recommendation;
- duplicate detection;
- summarization;
- trend detection;
- GIS pattern analysis;
- anomaly detection;
- metadata extraction;
- operator assistance.

AI must not autonomously declare guilt, authorize force, make irreversible enforcement decisions, or expose protected information.

## 19. Frontend

React/TypeScript frontend includes routing, AppLayout, providers, map functionality, dashboards, case/incident views and citizen/operational UI.

Temporary map recovery files should remain untracked unless deliberately promoted:

```text
IncidentMap.stage1.tsx
IncidentMap.stage3.tsx
IncidentMap.tsx.SAFE-BACKUP.tsx
```

Do not blindly commit backup artifacts.

## 20. Git Rules

Repository root:

```text
C:\Users\USER\Desktop\communityshield-platform
```

Backend:

```text
C:\Users\USER\Desktop\communityshield-platform\backend
```

After reboot:

```powershell
Set-Location "C:\Users\USER\Desktop\communityshield-platform\backend"
git status
go test ./...
go vet ./...
```

Never paste Go expressions such as `handlers.AssignCase` or `cases.POST(...)` into PowerShell. They belong in `.go` source files.

Open a file:

```powershell
notepad .\handlers\evidence_handler.go
```

## 21. Current Blocker

Latest reported error:

```text
handlers\evidence_handler.go:1:1: illegal character U+0040 '@'
```

The evidence handler has an accidental `@` at the beginning (or otherwise contains non-Go text).

Immediate repair:

```powershell
Set-Location "C:\Users\USER\Desktop\communityshield-platform\backend"
notepad .\handlers\evidence_handler.go
```

Delete the accidental marker/non-Go content and restore valid Go source.

Then:

```powershell
gofmt -w .\handlers\evidence_handler.go
go test ./...
go vet ./...
```

Do not start the next backend feature until both checks pass.

## 22. Evidence Route Consistency

The handler uses:

```go
caseID := c.Param("caseId")
```

Therefore the route must be:

```go
evidence.GET("/case/:caseId", middleware.AuthMiddleware(), handlers.GetEvidenceByCase)
```

Not:

```go
evidence.GET("/:id", handlers.GetEvidenceByCase)
```

The latter supplies `id` and currently bypasses authentication.

## 23. Engineering Sequence After Evidence

1. Evidence hardening.
2. Explicit case-state transition rules.
3. Unit/resource access enforcement.
4. Audit system.
5. Notifications/WebSocket integration.
6. Frontend operational workflow.
7. SOS escalation.
8. GIS/analytics.
9. AI assistance.
10. Security hardening.
11. Staging/production deployment.

## 24. Definition of Done

A feature is complete only when applicable layers are addressed:

```text
MODEL
 -> MIGRATION
 -> SERVICE
 -> HANDLER
 -> AUTHORIZATION
 -> ROUTE
 -> TEST
 -> FRONTEND
 -> REALTIME/NOTIFICATION
 -> AUDIT
 -> ANALYTICS
 -> DOCUMENTATION
```

If a layer is not applicable, record why.

## 25. Testing Gate

Minimum backend gate:

```powershell
go test ./...
go vet ./...
```

API tests should cover:
- unauthenticated;
- wrong role;
- wrong unit;
- authorized success;
- invalid UUID;
- missing fields;
- missing resource;
- duplicate action;
- unauthorized mutation.

Later add integration, frontend, E2E, load and security testing.

## 26. Security Requirements

Critical:
- password hashing;
- JWT validation;
- least privilege;
- secure evidence storage;
- input validation;
- rate limiting;
- audit logs;
- secret management;
- HTTPS;
- CORS policy;
- backup/restore;
- privacy controls;
- retention policy.

Medical information and other sensitive data require stronger controls than ordinary profile data.

## 27. Deployment

```text
LOCAL
 -> DEVELOPMENT
 -> STAGING
 -> PILOT COMMUNITY / UNIT
 -> MULTI-UNIT
 -> STATE / REGIONAL SCALE
```

Do not scale before authorization, evidence safety, auditing, backups and monitoring are reliable.

## 28. Grassroots Integration

Target model:

```text
Community
 -> Ward / Local Structure
 -> Security Unit
 -> Officer / Response Team
 -> Command / Administration
```

Community participation should enable reporting, safety information, feedback, prevention and local problem identification. It must not encourage vigilantism or unauthorized enforcement.

## 29. Long-Term Roadmap

### Phase A — Foundation
Authentication, users, units, cases, officers, permissions.

### Phase B — Operational Workflow
Assignment, dispatch, arrival, progress, evidence, closure.

### Phase C — Community Layer
Reporting, public safety map, notifications, feedback, announcements.

### Phase D — Intelligence
GIS, analytics, trends, dashboards, AI assistance.

### Phase E — Institutional Integration
Agency interoperability, SMS/IVR, approved communication channels, identity and records integrations.

### Phase F — Scale
Multi-state/multi-agency, disaster recovery, high availability, enterprise governance.

## 30. Research-Based Positioning

CommunityShield aligns with established community-oriented policing principles: partnership, problem solving, local context and prevention.

Nigeria's police materials describe community policing as an established strategy, while current policy discussion continues to emphasize grassroots participation and local ownership. Research also shows that digital engagement can improve reporting access while responsiveness and trust remain important.

Recommended positioning:

> **A digital operating layer for community-centered public safety, connecting grassroots observations and citizen participation to accountable, permission-controlled operational case management.**

## 31. Decision Log

- Go/Gin/GORM backend.
- PostgreSQL-style UUID entities.
- JWT authentication with database-backed authorization.
- `Case.AssignedTo` for primary officer.
- `CaseOfficer` for team assignments.
- `Progress` for operational timeline.
- `Evidence` for case-linked evidence.
- Explicit dispatch/arrival/closure timestamps.
- AI remains downstream of reliable operational data.
- Grassroots integration is a core architecture principle.

## 32. Current Checkpoint

Completed/working according to the latest project state:
- case assignment handlers exist;
- case assignment routes exist;
- dispatch/arrival handlers exist;
- progress add/get handlers exist;
- evidence handlers exist but currently contain a source corruption blocker;
- evidence route parameter mismatch identified;
- authentication middleware is `auth_middleware.go`;
- Git synchronization/push was successful at the latest checkpoint.

Current task:

```text
Repair evidence handler
 -> gofmt
 -> go test
 -> go vet
 -> evidence authorization hardening
 -> case lifecycle hardening
```

## 33. Agent Operating Rules

1. Check `git status` first.
2. Confirm repository root after every reboot.
3. Fix compile errors before adding features.
4. Never paste Go code directly into PowerShell.
5. Replace corrupted files completely when necessary.
6. Run gofmt after source changes.
7. Run `go test ./...` and `go vet ./...` after each logical backend milestone.
8. Commit coherent milestones only.
9. Push only after local verification.
10. Never commit database dumps, ZIP backups, credentials or temporary recovery files.
11. Update this document at architectural milestones.
12. Optimize for a trustworthy end-to-end workflow, not feature count.

## 34. North Star

```text
COMMUNITY
 -> REPORT / SOS / INTELLIGENCE
 -> SECURE INTAKE
 -> TRIAGE + PRIORITY
 -> UNIT ROUTING
 -> OFFICER ASSIGNMENT
 -> DISPATCH
 -> ARRIVAL
 -> INVESTIGATION
    -> PROGRESS
    -> EVIDENCE
    -> COMMUNICATION
    -> COLLABORATION
 -> RESOLUTION
 -> APPROVAL / CLOSURE
 -> AUDIT + ANALYTICS
 -> COMMUNITY FEEDBACK
 -> PREVENTION
 -> BETTER LOCAL SECURITY
```

**North-star principle:** Build digital infrastructure that lets local people, legitimate security institutions and accountable technology work together around real incidents quickly, transparently, safely and with measurable outcomes.

## 35. Research References Used for This Blueprint

- Nigeria Police Force — community policing and strategy development.
- Nigeria Federal Ministry of Information — grassroots/community participation in security.
- Nigeria Police National Cybercrime Centre — e-reporting workflow.
- OJJDP/Office of Justice Programs — community-oriented/problem-oriented policing.
- Research on police-public engagement and case management in Nigeria.
- Research on digital policing in Nigeria.
- Research on technology-enabled community policing in Nigeria.
- Existing African public-safety technology products and community-reporting systems.

This research supports the problem framing and landscape analysis; it does not prove that CommunityShield itself is novel in every individual feature.
