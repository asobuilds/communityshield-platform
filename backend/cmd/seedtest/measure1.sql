\echo '=== unit_scores schema ==='
\d unit_scores

\echo '=== officer_scores schema ==='
\d officer_scores

\echo '=== Q1: public cases list (public_handler.go) ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM cases
WHERE is_public = true AND status <> 'closed'
ORDER BY created_at DESC
LIMIT 20;

\echo '=== Q2: cases by unit_id (case_handler list) ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM cases
WHERE unit_id = (SELECT id FROM security_units LIMIT 1)
ORDER BY created_at DESC;

\echo '=== Q3: cases by reported_by (case_handler mine) ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM cases
WHERE reported_by = (SELECT id FROM users LIMIT 1);

\echo '=== Q4: active security units (unit_handler list) ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM security_units
WHERE status = 'active';

\echo '=== Q5: recent audit logs (audit_handler list) ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM audit_logs
ORDER BY created_at DESC
LIMIT 200;

\echo '=== Q6: ratings by target (rating_service) ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM ratings
WHERE target_id = (SELECT id FROM security_units LIMIT 1)
  AND target_type = 'unit'
  AND status = 'active';

\echo '=== Q7: ledger by unit (finance_handler list) ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM financial_ledgers
WHERE unit_id = (SELECT id FROM security_units LIMIT 1)
ORDER BY created_at DESC;

\echo '=== Q8: per-row PK lookup (the N+1 inner query) ==='
EXPLAIN (ANALYZE, BUFFERS)
SELECT * FROM security_units
WHERE id = (SELECT id FROM security_units LIMIT 1);