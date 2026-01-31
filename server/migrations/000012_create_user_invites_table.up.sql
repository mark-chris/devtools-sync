-- 000012_create_user_invites_table.up.sql
CREATE TABLE user_invites (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL,
  token_hash VARCHAR(64) NOT NULL UNIQUE,
  role VARCHAR(20) NOT NULL,
  invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  accepted_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(email, accepted_at)
);

CREATE INDEX idx_user_invites_token_hash ON user_invites(token_hash);
CREATE INDEX idx_user_invites_email ON user_invites(email);
CREATE INDEX idx_user_invites_expires_at ON user_invites(expires_at);
