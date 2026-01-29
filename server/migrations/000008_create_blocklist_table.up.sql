-- 000008_create_blocklist_table.up.sql
CREATE TABLE blocklist (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  pattern_type VARCHAR(20) NOT NULL,
  pattern VARCHAR(255) NOT NULL,
  reason TEXT,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_blocklist_pattern_type ON blocklist(pattern_type);
CREATE INDEX idx_blocklist_is_active ON blocklist(is_active);
