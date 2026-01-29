# Database Schema Design

**Issue:** #8 - Create initial database schema and migrations
**Date:** 2026-01-28
**Status:** Approved

## Overview

Initial database schema for DevTools Sync, combining requirements from Issue #8 with the v4 project plan's detailed specifications.

## Table Summary

| Table | Purpose |
|-------|---------|
| users | Dashboard administrators |
| groups | Teams/organizational units |
| extensions | Extension catalog |
| agents | Registered machines |
| api_keys | API key management with rotation |
| group_extensions | Junction: groups to extensions |
| reports | Sync session / compliance data |
| blocklist | Extension blocking rules |
| audit_log | Security event logging |

## Schema Design

### Foundation Tables (No Dependencies)

#### users
Dashboard administrators with role-based access.

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(255),
  role VARCHAR(20) DEFAULT 'viewer',      -- admin, manager, viewer
  is_active BOOLEAN DEFAULT true,
  last_login_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
```

#### groups
Teams/organizational units for extension management.

```sql
CREATE TABLE groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT,
  settings JSONB DEFAULT '{}',            -- group-specific config
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_groups_name ON groups(name);
```

#### extensions
Extension catalog with version pinning and integrity verification.

```sql
CREATE TABLE extensions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  extension_id VARCHAR(255) NOT NULL UNIQUE,  -- e.g., "ms-python.python"
  name VARCHAR(255) NOT NULL,
  publisher VARCHAR(255) NOT NULL,
  pinned_version VARCHAR(50),                 -- NULL = latest
  sha256_checksum VARCHAR(64),                -- integrity verification
  marketplace_url TEXT,
  is_blocked BOOLEAN DEFAULT false,           -- quick block flag
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_extensions_extension_id ON extensions(extension_id);
CREATE INDEX idx_extensions_publisher ON extensions(publisher);
```

### Dependent Tables

#### agents
Registered machines with heartbeat tracking.

```sql
CREATE TABLE agents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  machine_id VARCHAR(64) NOT NULL,            -- SHA256(hostname + hardware ID)
  user_hash VARCHAR(64) NOT NULL,             -- HMAC of username
  display_name VARCHAR(255),
  platform VARCHAR(20) NOT NULL,              -- windows, darwin, linux
  agent_version VARCHAR(50),
  last_heartbeat TIMESTAMPTZ,
  last_sync TIMESTAMPTZ,
  status VARCHAR(20) DEFAULT 'active',        -- active, stale, revoked
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ,
  UNIQUE(group_id, machine_id)
);
CREATE INDEX idx_agents_group_id ON agents(group_id);
CREATE INDEX idx_agents_last_heartbeat ON agents(last_heartbeat);
CREATE INDEX idx_agents_status ON agents(status);
```

#### api_keys
API key management with rotation support.

```sql
CREATE TABLE api_keys (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  key_hash VARCHAR(64) NOT NULL UNIQUE,       -- SHA256 of actual key
  key_prefix VARCHAR(8) NOT NULL,             -- first 8 chars for identification
  name VARCHAR(255),                          -- human-friendly label
  scopes VARCHAR(255) DEFAULT 'agent',        -- agent, admin
  expires_at TIMESTAMPTZ,
  last_used_at TIMESTAMPTZ,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_api_keys_group_id ON api_keys(group_id);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
```

#### group_extensions
Junction table linking groups to extensions.

```sql
CREATE TABLE group_extensions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  extension_id UUID NOT NULL REFERENCES extensions(id) ON DELETE CASCADE,
  is_required BOOLEAN DEFAULT true,           -- required vs recommended
  added_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(group_id, extension_id)
);
CREATE INDEX idx_group_extensions_group_id ON group_extensions(group_id);
CREATE INDEX idx_group_extensions_extension_id ON group_extensions(extension_id);
```

### Operational Tables

#### reports
Sync session / compliance data from agents.

```sql
CREATE TABLE reports (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
  group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  extension_id VARCHAR(255) NOT NULL,         -- marketplace ID (denormalized for speed)
  user_hash VARCHAR(64) NOT NULL,
  installed BOOLEAN NOT NULL,
  installed_version VARCHAR(50),
  extension_sha256 VARCHAR(64),
  client_ip INET,
  reported_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_hash, group_id, extension_id)   -- upsert on re-sync
);
CREATE INDEX idx_reports_agent_id ON reports(agent_id);
CREATE INDEX idx_reports_group_id ON reports(group_id);
CREATE INDEX idx_reports_reported_at ON reports(reported_at);
CREATE INDEX idx_reports_user_extension ON reports(user_hash, extension_id);
```

#### blocklist
Extension blocking rules by pattern.

```sql
CREATE TABLE blocklist (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  pattern_type VARCHAR(20) NOT NULL,          -- exact, publisher, regex
  pattern VARCHAR(255) NOT NULL,              -- e.g., "evil.malware" or "sketchy-publisher.*"
  reason TEXT,
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_blocklist_pattern_type ON blocklist(pattern_type);
CREATE INDEX idx_blocklist_is_active ON blocklist(is_active);
```

#### audit_log
Security event logging for compliance.

```sql
CREATE TABLE audit_log (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  event_type VARCHAR(50) NOT NULL,            -- auth.login, agent.register, extension.add, etc.
  actor_type VARCHAR(20) NOT NULL,            -- user, agent, system
  actor_id UUID,                              -- user or agent ID
  group_id UUID REFERENCES groups(id) ON DELETE SET NULL,
  target_type VARCHAR(50),                    -- agent, extension, api_key, etc.
  target_id UUID,
  details JSONB DEFAULT '{}',                 -- event-specific data
  client_ip INET,
  user_agent TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_audit_log_event_type ON audit_log(event_type);
CREATE INDEX idx_audit_log_actor ON audit_log(actor_type, actor_id);
CREATE INDEX idx_audit_log_group_id ON audit_log(group_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
```

## Migration File Structure

```
server/migrations/
├── 000001_create_users_table.up.sql
├── 000001_create_users_table.down.sql
├── 000002_create_groups_table.up.sql
├── 000002_create_groups_table.down.sql
├── 000003_create_extensions_table.up.sql
├── 000003_create_extensions_table.down.sql
├── 000004_create_agents_table.up.sql
├── 000004_create_agents_table.down.sql
├── 000005_create_api_keys_table.up.sql
├── 000005_create_api_keys_table.down.sql
├── 000006_create_group_extensions_table.up.sql
├── 000006_create_group_extensions_table.down.sql
├── 000007_create_reports_table.up.sql
├── 000007_create_reports_table.down.sql
├── 000008_create_blocklist_table.up.sql
├── 000008_create_blocklist_table.down.sql
├── 000009_create_audit_log_table.up.sql
├── 000009_create_audit_log_table.down.sql
├── 000010_seed_data.up.sql
├── 000010_seed_data.down.sql
```

## Seed Data

| Table | Seed Data |
|-------|-----------|
| users | 1 admin (`admin@example.com`), 1 viewer |
| groups | 2 groups: "Engineering", "Design" |
| extensions | 5 extensions (Python, Go, Prettier, ESLint, GitLens) |
| group_extensions | 3-4 extensions per group |
| agents | 2 sample agents (one per group) |
| api_keys | 1 key per group |
| reports | Sample compliance data |
| blocklist | 1 example blocked extension |
| audit_log | Empty (populated by application) |

## Acceptance Criteria

- [x] Design schema with relationships
- [x] Create initial migration files
- [x] Add seed data for development
- [x] Document schema in docs/
- [x] Add database diagram
