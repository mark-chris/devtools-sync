-- 000003_create_extensions_table.up.sql
CREATE TABLE extensions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  extension_id VARCHAR(255) NOT NULL UNIQUE,
  name VARCHAR(255) NOT NULL,
  publisher VARCHAR(255) NOT NULL,
  pinned_version VARCHAR(50),
  sha256_checksum VARCHAR(64),
  marketplace_url TEXT,
  is_blocked BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_extensions_extension_id ON extensions(extension_id);
CREATE INDEX idx_extensions_publisher ON extensions(publisher);
