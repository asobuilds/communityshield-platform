CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_cases_unit_created ON cases (unit_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_cases_reported_by ON cases (reported_by);
CREATE INDEX IF NOT EXISTS idx_cases_public_list ON cases (is_public, status, created_at DESC);