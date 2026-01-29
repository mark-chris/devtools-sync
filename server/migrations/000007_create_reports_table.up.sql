-- 000007_create_reports_table.up.sql
CREATE TABLE reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
  group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  extension_id VARCHAR(255) NOT NULL,
  user_hash VARCHAR(64) NOT NULL,
  installed BOOLEAN NOT NULL,
  installed_version VARCHAR(50),
  extension_sha256 VARCHAR(64),
  client_ip INET,
  reported_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_hash, group_id, extension_id)
);

CREATE INDEX idx_reports_agent_id ON reports(agent_id);
CREATE INDEX idx_reports_group_id ON reports(group_id);
CREATE INDEX idx_reports_reported_at ON reports(reported_at);
CREATE INDEX idx_reports_user_extension ON reports(user_hash, extension_id);
