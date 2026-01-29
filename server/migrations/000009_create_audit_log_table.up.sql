-- 000009_create_audit_log_table.up.sql
CREATE TABLE audit_log (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_type VARCHAR(50) NOT NULL,
  actor_type VARCHAR(20) NOT NULL,
  actor_id UUID,
  group_id UUID REFERENCES groups(id) ON DELETE SET NULL,
  target_type VARCHAR(50),
  target_id UUID,
  details JSONB DEFAULT '{}',
  client_ip INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_log_event_type ON audit_log(event_type);
CREATE INDEX idx_audit_log_actor ON audit_log(actor_type, actor_id);
CREATE INDEX idx_audit_log_group_id ON audit_log(group_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
