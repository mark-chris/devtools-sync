-- 000005_create_api_keys_table.up.sql
CREATE TABLE api_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  key_hash VARCHAR(64) NOT NULL UNIQUE,
  key_prefix VARCHAR(8) NOT NULL,
  name VARCHAR(255),
  scopes VARCHAR(255) DEFAULT 'agent',
  expires_at TIMESTAMPTZ,
  last_used_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_api_keys_group_id ON api_keys(group_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
