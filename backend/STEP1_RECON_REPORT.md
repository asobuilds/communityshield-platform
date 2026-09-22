# Wave 8f — Step 1 Recon Report

Generated: 2026-09-22

---

## A. ENDPOINT INVENTORY — Top 10 Read Endpoints by Expected Traffic

| # | Method | Path | Handler | File:Line | Description |
|---|--------|------|---------|-----------|-------------|
| 1 | GET | /api/v1/public/cases | GetPublicCases | public_handler.go:13 | Landing page cases feed (no auth, public) |
| 2 | GET | /api/v1/public/units | GetPublicUnits | public_handler.go:25 | Landing page units list (no auth, public) |
| 3 | GET | /api/v1/public/leaderboard | GetPublicLeaderboard | rating_handler.go:136 | Public leaderboard (top 50 units) |
| 4 | GET | /api/v1/cases | GetAllCases | case_handler.go:184 | Authenticated user's cases list |
| 5 | GET | /api/v1/cases/:id | GetCaseByID | case_handler.go:214 | Single case detail with timeline/feedback |
| 6 | GET | /api/v1/units | GetAllUnits | unit_handler.go:119 | All active units (no auth on public variant) |
| 7 | GET | /api/v1/units/:id | GetUnitByID | unit_handler.go:132 | Single unit detail |
| 8 | GET | /api/v1/public/units/:id/ledger | GetPublicBankAccounts | public_bank_handler.go:16 | Public unit bank accounts |
| 9 | GET | /api/v1/finance/units/:id/transactions | GetTransactions | finance_handler.go:288 | Unit transaction history |
| 10 | GET | /api/v1/mobile/dashboard | MobileDashboard | mobile_handler.go:50 | Mobile app dashboard (cases, SOS, recent) |

**Honorable mentions (high traffic but auth-gated):**
- GET /api/v1/public/units/suggest → SuggestUnits (rating_handler.go:270)
- GET /api/v1/audit/activities → GetActivityLogs (audit_handler.go:57)
- GET /api/v1/audit/logs → GetAuditLogs (audit_handler.go:89)
- GET /api/v1/notifications → GetUserNotifications (notification_handler.go:51)
- GET /api/v1/units/:id/officers/ranking → GetOfficersInUnitRanking (rating_handler.go:172)

---

## B. KNOWN CANDIDATES FROM 8e (Carry-Forward)

| Candidate | Status | Evidence |
|-----------|--------|----------|
| `ranking_service.go:98` — `AggregateForUnit` called twice | **CONFIRMED** | Lines 96–97: `_, directCount, directBayes := s.Ratings.AggregateForUnit(unitID)` then `dSum, _, _ := s.Ratings.AggregateForUnit(unitID)` — identical call, second ignores count/bayes |
| `user_media_handler.go:44,50,92,98` — `_ = storageSvc.Delete(...)` failures ignored | **HYGIENE ONLY** | Correctness issue (orphaned files on DB failure), not a read-path performance problem. Drop from 8f scope. |

---

## C. N+1 PATTERNS

| # | Endpoint | File:Line | Loop Pattern | Expected Query Count |
|---|----------|-----------|--------------|----------------------|
| 1 | GET /api/v1/public/leaderboard | rating_handler.go:154–164 | `for _, s := range scores { config.DB.First(&unit, "id = ?", s.UnitID) }` | 1 (scores) + N (units) = **51** (limit 50) |
| 2 | GET /api/v1/units/:id/officers/ranking | rating_handler.go:218–228 | `for _, s := range scores { config.DB.First(&officer, "id = ?", s.OfficerID) }` | 1 (scores) + N (officers) = **1 + officer_count** |
| 3 | GET /api/v1/public/units/suggest | rating_handler.go:285–299 | `for _, s := range scores { config.DB.First(&unit, "id = ?", s.UnitID) }` | 1 (scores) + N (units) = **11** (limit 10) |
| 4 | GET /api/v1/cases | case_handler.go:193 | `Preload("Evidence").Preload("Progress")` — GORM may issue separate queries per association if not using JOIN | 1 (cases) + N×2 (evidence + progress per case) |
| 5 | GET /api/v1/cases/:id | case_handler.go:232 | `Preload("Evidence").Preload("Progress")` — same as above | 1 + N×2 |
| 6 | GET /api/v1/finance/units/:id/transactions | finance_handler.go:311 | `Preload("Initiator").Preload("Approver")` — same risk | 1 + N×2 |
| 7 | GET /api/v1/audit/activities | audit_handler.go:66 | `Preload("User")` with JOIN — **likely OK** (uses Joins) | 1 |
| 8 | GET /api/v1/audit/logs | audit_handler.go:103 | `Preload("User")` with JOIN — **likely OK** | 1 |
| 9 | rating_service.go:142–152 | AggregateForOfficer | Two separate queries: `SUM(rating)` then `COUNT(*)` on same WHERE | **2** per officer (could be 1) |
| 10 | rating_service.go:157–167 | AggregateForUnit | Two separate queries: `SUM(rating)` then `COUNT(*)` on same WHERE | **2** per unit (could be 1) |

---

## D. MISSING INDEXES (Hot WHERE/ORDER BY on Tables >10k Expected Rows)

| Table | Column(s) | Query Using It | Index Exists? | Model Tag / Evidence |
|-------|-----------|----------------|---------------|----------------------|
| `cases` | `is_public, status, created_at` | `GetPublicCases` (public_handler.go:15): `WHERE is_public = ? AND status != ? ORDER BY created_at desc LIMIT 20` | **NO** | Case model: no composite index tag |
| `cases` | `unit_id` | `GetAllCases` (case_handler.go:198), `GetCaseAnalytics` (553), `RecomputeAllUnits` (106), mobile dashboard (63) | **NO** | UnitID field: `gorm:"type:uuid;not null"` — no `index` tag |
| `cases` | `reported_by` | `GetAllCases` citizen branch (196) | **NO** | ReportedBy field: no index tag |
| `cases` | `status` | `GetAllCases` (implicit), `GetCaseAnalytics` (557–558), `MobileDashboard` (64) | **NO** | Status field: no index tag |
| `security_units` | `status` | `GetPublicUnits` (27), `GetNearbyUnits` (46), `GetAllUnits` (121), `RecomputeAllUnits` (68) | **NO** | Status field: `gorm:"default:active"` — no index tag |
| `ratings` | `(target_id, target_type, status)` | `AggregateForOfficer` (143), `AggregateForUnit` (158) — composite WHERE | **PARTIAL** | Separate indexes on each column, but **no composite index** for the 3-column predicate |
| `audit_logs` | `created_at` | `GetAuditLogs` (103): `ORDER BY created_at desc LIMIT 200` | **NO** | No index tag on CreatedAt |
| `activity_logs` | `created_at` | `GetActivityLogs` (66): `ORDER BY created_at desc LIMIT 100` | **NO** | No index tag on CreatedAt |
| `notifications` | `user_id, status` | `GetUserNotifications` (60): `WHERE user_id = ? ORDER BY created_at desc`; `MobileNotifications` (104): `WHERE user_id = ? AND status = ?` | **NO** | No composite index on (user_id, status, created_at) |
| `case_weekly_updates` | `case_id, week_start` | `GetCaseWeeklyUpdates` (529): `WHERE case_id = ? ORDER BY week_start ASC` | **YES** | Unique index `idx_case_week_officer` on (case_id, officer_id, week_start) — covers case_id prefix |

---

## E. REDUNDANT WORK

| # | File:Line | Pattern | Evidence |
|---|-----------|---------|----------|
| 1 | ranking_service.go:96–97 | Same function called twice with same arg | `_, directCount, directBayes := s.Ratings.AggregateForUnit(unitID)` then `dSum, _, _ := s.Ratings.AggregateForUnit(unitID)` — second call discards 2/3 of result |
| 2 | rating_service.go:143–149 | Two queries for SUM and COUNT on same WHERE | `AggregateForOfficer`: separate `Select("COALESCE(SUM(rating), 0)").Scan(&sum)` and `Count(&count)` |
| 3 | rating_service.go:158–164 | Two queries for SUM and COUNT on same WHERE | `AggregateForUnit`: same pattern as above |
| 4 | leaderboard_handler.go:95–96 | Two COUNT queries on same table/WHERE | `Count(&caseCount)` then `Where(...).Count(&resolvedCount)` — could be single conditional aggregation |
| 5 | mobile_handler.go:63–65 | Three independent COUNT queries | `totalCases`, `pendingCases`, `sosCount` — no shared WHERE, but each is a full table scan |
| 6 | audit_handler.go:196,199 | Two COUNT queries | `activeCount` on users, `totalRequests` on audit_logs — separate tables, acceptable |
| 7 | finance_handler.go:382–394 | Multiple SUM queries with overlapping WHERE | `totalIncome`, `totalExpenses`, `pendingCount`, `categoryBreakdown`, `monthlyTrends` — 5 separate scans of `transactions` table |
| 8 | case_handler.go:556–558 | Three COUNT queries on same base query | `totalCases`, `resolvedCases`, `pendingCases` — could be single conditional aggregation |

---

## F. PAGINATION DEFAULTS

| Endpoint | Handler | Default Page Size | Max Page Size | Server-Enforced? | Notes |
|----------|---------|-------------------|---------------|------------------|-------|
| GET /api/v1/public/cases | GetPublicCases | 20 (hardcoded) | 20 (hardcoded) | Yes | No pagination params accepted |
| GET /api/v1/public/units | GetPublicUnits | **Unbounded** (no LIMIT) | None | **NO** | DoS vector — returns all active units |
| GET /api/v1/public/leaderboard | GetPublicLeaderboard | 50 | 50 | Yes | Hardcoded LIMIT 50 |
| GET /api/v1/cases | GetAllCases | **Unbounded** | None | **NO** | `Find(&cases)` with no limit — DoS vector |
| GET /api/v1/units | GetAllUnits | **Unbounded** | None | **NO** | `Find(&units)` with no limit — DoS vector |
| GET /api/v1/units/:id/officers/ranking | GetOfficersInUnitRanking | **Unbounded** | None | **NO** | `Find(&scores)` with no limit |
| GET /api/v1/audit/activities | GetActivityLogs | 100 | 100 | Yes | Hardcoded LIMIT 100 |
| GET /api/v1/audit/logs | GetAuditLogs | 200 | 200 | Yes | Hardcoded LIMIT 200 |
| GET /api/v1/audit/notifications | GetNotificationLogs | 100 | 100 | Yes | Hardcoded LIMIT 100 |
| GET /api/v1/notifications | GetUserNotifications | **Unbounded** | None | **NO** | `Find(&notifications)` with no limit |
| GET /api/v1/mobile/notifications | MobileNotifications | 20 | 20 | Yes | Hardcoded LIMIT 20 |
| GET /api/v1/mobile/sync | MobileSync | 50 | 50 | Yes | Hardcoded LIMIT 50 |
| GET /api/v1/finance/units/:id/transactions | GetTransactions | **Unbounded** | None | **NO** | `Find(&transactions)` with no limit |
| GET /api/v1/bank/:id/accounts | GetBankAccounts | **Unbounded** | None | **NO** | `Find(&bankAccounts)` with no limit |
| GET /api/v1/bank/:id/donations | GetDonations | **Unbounded** | None | **NO** | `Find(&donations)` with no limit |

**Findings:** 7 list endpoints have **no LIMIT at all** (unbounded), creating DoS risk and unpredictable latency. 4 more have hardcoded limits but no client-controlled pagination (no `page`/`page_size` params).

---

## SUMMARY — Prioritized Fix Candidates for Step 3 Measurement

1. **ranking_service.go:96–97** — Double `AggregateForUnit` call (redundant work, easy fix)
2. **rating_handler.go:154–164** — N+1 on public leaderboard (51 queries for 50 rows)
3. **rating_handler.go:218–228** — N+1 on officer ranking (1 + officer_count queries)
4. **rating_handler.go:285–299** — N+1 on unit suggestions (11 queries for 10 rows)
5. **Missing composite index on `ratings(target_id, target_type, status)`** — Affects all rating aggregates
6. **Missing index on `cases(is_public, status, created_at)`** — Public cases landing page
7. **Missing index on `cases(unit_id)`** — Hot filter in multiple endpoints
8. **Missing index on `security_units(status)`** — Hot filter in multiple endpoints
9. **Missing index on `audit_logs(created_at)`** — Admin audit logs ordering
10. **Pagination bounds** — 7 unbounded list endpoints

**Ready for Step 2 (seeding) and Step 3 (baseline measurement).**