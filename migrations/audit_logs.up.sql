CREATE TABLE IF NOT EXISTS audit_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT,
    username    VARCHAR(100),
    method      VARCHAR(10) NOT NULL,
    path        TEXT NOT NULL,
    resource    VARCHAR(50),
    record_id   BIGINT,
    status_code INT NOT NULL,
    ip          VARCHAR(64),
    request     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created  ON audit_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs (resource);

-- Permission baru untuk melihat audit log, langsung diberikan ke admin
INSERT INTO permissions (name, description) VALUES ('audit:read', 'View audit logs')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name = 'audit:read'
ON CONFLICT DO NOTHING;