-- 000004_create_agents_table.up.sql
CREATE TABLE agents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  machine_id VARCHAR(64) NOT NULL,
  user_hash VARCHAR(64) NOT NULL,
  display_name VARCHAR(255),
  platform VARCHAR(20) NOT NULL,
  agent_version VARCHAR(50),
  last_heartbeat TIMESTAMPTZ,
  last_sync TIMESTAMPTZ,
  status VARCHAR(20) DEFAULT 'active',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  UNIQUE(group_id, machine_id)
);

CREATE INDEX idx_agents_group_id ON agents(group_id);
CREATE INDEX idx_agents_last_heartbeat ON agents(last_heartbeat);
CREATE INDEX idx_agents_status ON agents(status);
