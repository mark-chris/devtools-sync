-- 000006_create_group_extensions_table.up.sql
CREATE TABLE group_extensions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  extension_id UUID NOT NULL REFERENCES extensions(id) ON DELETE CASCADE,
  is_required BOOLEAN DEFAULT true,
  added_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(group_id, extension_id)
);

CREATE INDEX idx_group_extensions_group_id ON group_extensions(group_id);
CREATE INDEX idx_group_extensions_extension_id ON group_extensions(extension_id);
