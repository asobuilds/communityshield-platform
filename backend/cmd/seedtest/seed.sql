\timing on

BEGIN;

TRUNCATE TABLE audit_logs, financial_ledgers, ratings, cases, security_units, users RESTART IDENTITY CASCADE;

INSERT INTO users (email, first_name, last_name, password, role, status, created_at)
SELECT
  'user_' || i || '@example.com',
  'First' || i,
  'Last' || i,
  'hashed_' || i,
  (ARRAY['citizen','officer','unit_admin'])[1 + (i % 3)],
  'active',
  now() - (interval '365 days' * random())
FROM generate_series(1, 5000) AS g(i);

INSERT INTO security_units (name, type, status, created_at)
SELECT
  'Unit ' || i,
  (ARRAY['police','vigilante','neighborhood_watch'])[1 + (i % 3)],
  'active',
  now() - (interval '730 days' * random())
FROM generate_series(1, 500) AS g(i);

WITH unit_ids AS MATERIALIZED (SELECT array_agg(id) AS ids FROM security_units),
     user_ids AS MATERIALIZED (SELECT array_agg(id) AS ids FROM users)
INSERT INTO cases (unit_id, reported_by, title, description, tracking_id, status, is_public, created_at, latitude, longitude)
SELECT
  (SELECT ids[1 + (i % array_length(ids, 1))] FROM unit_ids),
  (SELECT ids[1 + (i % array_length(ids, 1))] FROM user_ids),
  'Case ' || i,
  'Description for case ' || i,
  'TRK-' || LPAD(i::text, 10, '0'),
  (ARRAY['pending','open','closed','under_review'])[1 + (i % 4)],
  (i % 2 = 0),
  now() - (interval '730 days' * random()),
  (random() * 20 - 10)::numeric(10,6),
  (random() * 20 - 10)::numeric(10,6)
FROM generate_series(1, 20000) AS g(i);

WITH user_ids AS MATERIALIZED (SELECT array_agg(id) AS ids FROM users),
     unit_ids AS MATERIALIZED (SELECT array_agg(id) AS ids FROM security_units)
INSERT INTO ratings (user_id, target_id, target_type, rating, status, created_at)
SELECT
  (SELECT ids[1 + (i % array_length(ids, 1))] FROM user_ids),
  CASE WHEN i % 2 = 0
       THEN (SELECT ids[1 + (i % array_length(ids, 1))] FROM user_ids)
       ELSE (SELECT ids[1 + (i % array_length(ids, 1))] FROM unit_ids)
  END,
  CASE WHEN i % 2 = 0 THEN 'officer' ELSE 'unit' END,
  1 + (i % 5),
  'active',
  now() - (interval '365 days' * random())
FROM generate_series(1, 50000) AS g(i);

WITH unit_ids AS MATERIALIZED (SELECT array_agg(id) AS ids FROM security_units),
     user_ids AS MATERIALIZED (SELECT array_agg(id) AS ids FROM users)
INSERT INTO financial_ledgers (reference, unit_id, year, sequence_number, direction, entry_type, amount, currency, status, created_by, created_at)
SELECT
  'NG-' || (2024 + (i % 3)) || '-' || LPAD(i::text, 8, '0'),
  (SELECT ids[1 + (i % array_length(ids, 1))] FROM unit_ids),
  2024 + (i % 3),
  i,
  CASE WHEN i % 2 = 0 THEN 'in' ELSE 'out' END,
  (ARRAY['donation','expense','allocation'])[1 + (i % 3)],
  (random() * 10000)::numeric(14,2),
  'NGN',
  'posted',
  (SELECT ids[1 + (i % array_length(ids, 1))] FROM user_ids),
  now() - (interval '730 days' * random())
FROM generate_series(1, 50000) AS g(i);

WITH user_ids AS MATERIALIZED (SELECT array_agg(id) AS ids FROM users)
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, created_at)
SELECT
  (SELECT ids[1 + (i % array_length(ids, 1))] FROM user_ids),
  (ARRAY['create','update','delete','login'])[1 + (i % 4)],
  (ARRAY['case','user','unit','ledger'])[1 + (i % 4)],
  'entity-' || i,
  now() - (interval '365 days' * random())
FROM generate_series(1, 100000) AS g(i);

COMMIT;

SELECT 'users' AS t, COUNT(*) FROM users
UNION ALL SELECT 'security_units', COUNT(*) FROM security_units
UNION ALL SELECT 'cases', COUNT(*) FROM cases
UNION ALL SELECT 'ratings', COUNT(*) FROM ratings
UNION ALL SELECT 'financial_ledgers', COUNT(*) FROM financial_ledgers
UNION ALL SELECT 'audit_logs', COUNT(*) FROM audit_logs;