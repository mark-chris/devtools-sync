# Database Migrations Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create PostgreSQL migration files for the DevTools Sync database schema with reversible migrations and seed data.

**Architecture:** Using golang-migrate for migrations. Each table gets a numbered up/down pair. Seed data in final migration. All migrations are idempotent and reversible.

**Tech Stack:** PostgreSQL 16, golang-migrate, SQL

---

## Prerequisites

Before starting, ensure Docker is running and the database is up:

```bash
docker compose -f docker-compose.dev.yml up -d postgres
```

Verify connection:
```bash
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "SELECT 1"
```

---

### Task 1: Create Users Table Migration

**Files:**
- Create: `server/migrations/000001_create_users_table.up.sql`
- Create: `server/migrations/000001_create_users_table.down.sql`

**Step 1: Write the up migration**

```sql
-- 000001_create_users_table.up.sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(255),
  role VARCHAR(20) DEFAULT 'viewer',
  is_active BOOLEAN DEFAULT true,
  last_login_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
```

**Step 2: Write the down migration**

```sql
-- 000001_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

**Step 3: Test migration up**

```bash
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -f /migrations/000001_create_users_table.up.sql
```

Expected: `CREATE TABLE`, `CREATE INDEX` (x2)

**Step 4: Verify table exists**

```bash
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "\d users"
```

Expected: Table schema displayed

**Step 5: Test migration down**

```bash
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -f /migrations/000001_create_users_table.down.sql
```

Expected: `DROP TABLE`

**Step 6: Commit**

```bash
git add server/migrations/000001_create_users_table.*.sql
git commit -m "feat(db): add users table migration"
```

---

### Task 2: Create Groups Table Migration

**Files:**
- Create: `server/migrations/000002_create_groups_table.up.sql`
- Create: `server/migrations/000002_create_groups_table.down.sql`

**Step 1: Write the up migration**

```sql
-- 000002_create_groups_table.up.sql
CREATE TABLE groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT,
  settings JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_groups_name ON groups(name);
```

**Step 2: Write the down migration**

```sql
-- 000002_create_groups_table.down.sql
DROP TABLE IF EXISTS groups;
```

**Step 3: Test migration up/down cycle**

```bash
# Up
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -f /migrations/000002_create_groups_table.up.sql
# Verify
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "\d groups"
# Down
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -f /migrations/000002_create_groups_table.down.sql
```

**Step 4: Commit**

```bash
git add server/migrations/000002_create_groups_table.*.sql
git commit -m "feat(db): add groups table migration"
```

---

### Task 3: Create Extensions Table Migration

**Files:**
- Create: `server/migrations/000003_create_extensions_table.up.sql`
- Create: `server/migrations/000003_create_extensions_table.down.sql`

**Step 1: Write the up migration**

```sql
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
```

**Step 2: Write the down migration**

```sql
-- 000003_create_extensions_table.down.sql
DROP TABLE IF EXISTS extensions;
```

**Step 3: Test migration up/down cycle**

```bash
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -f /migrations/000003_create_extensions_table.up.sql
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "\d extensions"
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -f /migrations/000003_create_extensions_table.down.sql
```

**Step 4: Commit**

```bash
git add server/migrations/000003_create_extensions_table.*.sql
git commit -m "feat(db): add extensions table migration"
```

---

### Task 4: Create Agents Table Migration

**Files:**
- Create: `server/migrations/000004_create_agents_table.up.sql`
- Create: `server/migrations/000004_create_agents_table.down.sql`

**Step 1: Write the up migration**

```sql
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
```

**Step 2: Write the down migration**

```sql
-- 000004_create_agents_table.down.sql
DROP TABLE IF EXISTS agents;
```

**Step 3: Test migration (requires groups table)**

```bash
# First run migrations 1-2 to set up dependencies
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -f /migrations/000002_create_groups_table.up.sql
# Then test agents
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -f /migrations/000004_create_agents_table.up.sql
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "\d agents"
```

**Step 4: Commit**

```bash
git add server/migrations/000004_create_agents_table.*.sql
git commit -m "feat(db): add agents table migration"
```

---

### Task 5: Create API Keys Table Migration

**Files:**
- Create: `server/migrations/000005_create_api_keys_table.up.sql`
- Create: `server/migrations/000005_create_api_keys_table.down.sql`

**Step 1: Write the up migration**

```sql
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
```

**Step 2: Write the down migration**

```sql
-- 000005_create_api_keys_table.down.sql
DROP TABLE IF EXISTS api_keys;
```

**Step 3: Commit**

```bash
git add server/migrations/000005_create_api_keys_table.*.sql
git commit -m "feat(db): add api_keys table migration"
```

---

### Task 6: Create Group Extensions Table Migration

**Files:**
- Create: `server/migrations/000006_create_group_extensions_table.up.sql`
- Create: `server/migrations/000006_create_group_extensions_table.down.sql`

**Step 1: Write the up migration**

```sql
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
```

**Step 2: Write the down migration**

```sql
-- 000006_create_group_extensions_table.down.sql
DROP TABLE IF EXISTS group_extensions;
```

**Step 3: Commit**

```bash
git add server/migrations/000006_create_group_extensions_table.*.sql
git commit -m "feat(db): add group_extensions table migration"
```

---

### Task 7: Create Reports Table Migration

**Files:**
- Create: `server/migrations/000007_create_reports_table.up.sql`
- Create: `server/migrations/000007_create_reports_table.down.sql`

**Step 1: Write the up migration**

```sql
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
```

**Step 2: Write the down migration**

```sql
-- 000007_create_reports_table.down.sql
DROP TABLE IF EXISTS reports;
```

**Step 3: Commit**

```bash
git add server/migrations/000007_create_reports_table.*.sql
git commit -m "feat(db): add reports table migration"
```

---

### Task 8: Create Blocklist Table Migration

**Files:**
- Create: `server/migrations/000008_create_blocklist_table.up.sql`
- Create: `server/migrations/000008_create_blocklist_table.down.sql`

**Step 1: Write the up migration**

```sql
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
```

**Step 2: Write the down migration**

```sql
-- 000008_create_blocklist_table.down.sql
DROP TABLE IF EXISTS blocklist;
```

**Step 3: Commit**

```bash
git add server/migrations/000008_create_blocklist_table.*.sql
git commit -m "feat(db): add blocklist table migration"
```

---

### Task 9: Create Audit Log Table Migration

**Files:**
- Create: `server/migrations/000009_create_audit_log_table.up.sql`
- Create: `server/migrations/000009_create_audit_log_table.down.sql`

**Step 1: Write the up migration**

```sql
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
```

**Step 2: Write the down migration**

```sql
-- 000009_create_audit_log_table.down.sql
DROP TABLE IF EXISTS audit_log;
```

**Step 3: Commit**

```bash
git add server/migrations/000009_create_audit_log_table.*.sql
git commit -m "feat(db): add audit_log table migration"
```

---

### Task 10: Create Seed Data Migration

**Files:**
- Create: `server/migrations/000010_seed_data.up.sql`
- Create: `server/migrations/000010_seed_data.down.sql`

**Step 1: Write the up migration**

```sql
-- 000010_seed_data.up.sql
-- Seed data for development environment only

-- Users (password: devpassword123, bcrypt hash)
INSERT INTO users (id, email, password_hash, display_name, role) VALUES
  ('11111111-1111-1111-1111-111111111111', 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.QWHKjNC.5RILXPnZGW', 'Admin User', 'admin'),
  ('22222222-2222-2222-2222-222222222222', 'viewer@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.QWHKjNC.5RILXPnZGW', 'Viewer User', 'viewer');

-- Groups
INSERT INTO groups (id, name, description) VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Engineering', 'Software engineering team'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Design', 'Design and UX team');

-- Extensions
INSERT INTO extensions (id, extension_id, name, publisher) VALUES
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'ms-python.python', 'Python', 'Microsoft'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'golang.go', 'Go', 'Go Team at Google'),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'esbenp.prettier-vscode', 'Prettier', 'Prettier'),
  ('ffffffff-ffff-ffff-ffff-ffffffffffff', 'dbaeumer.vscode-eslint', 'ESLint', 'Microsoft'),
  ('00000000-0000-0000-0000-000000000001', 'eamodio.gitlens', 'GitLens', 'GitKraken');

-- Group Extensions
INSERT INTO group_extensions (group_id, extension_id, is_required, added_by) VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'cccccccc-cccc-cccc-cccc-cccccccccccc', true, '11111111-1111-1111-1111-111111111111'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'dddddddd-dddd-dddd-dddd-dddddddddddd', true, '11111111-1111-1111-1111-111111111111'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', true, '11111111-1111-1111-1111-111111111111'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '00000000-0000-0000-0000-000000000001', false, '11111111-1111-1111-1111-111111111111'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', true, '11111111-1111-1111-1111-111111111111'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'ffffffff-ffff-ffff-ffff-ffffffffffff', true, '11111111-1111-1111-1111-111111111111'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '00000000-0000-0000-0000-000000000001', false, '11111111-1111-1111-1111-111111111111');

-- Agents
INSERT INTO agents (id, group_id, machine_id, user_hash, display_name, platform, status) VALUES
  ('33333333-3333-3333-3333-333333333333', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'abc123def456abc123def456abc123def456abc123def456abc123def456abcd', 'user1hash', 'Dev Laptop 1', 'linux', 'active'),
  ('44444444-4444-4444-4444-444444444444', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'xyz789xyz789xyz789xyz789xyz789xyz789xyz789xyz789xyz789xyz789xyza', 'user2hash', 'Design Mac 1', 'darwin', 'active');

-- API Keys (key_hash is SHA256 of "devkey-engineering-001" and "devkey-design-001")
INSERT INTO api_keys (id, group_id, created_by, key_hash, key_prefix, name, scopes) VALUES
  ('55555555-5555-5555-5555-555555555555', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 'a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2', 'devkey-e', 'Engineering Dev Key', 'agent'),
  ('66666666-6666-6666-6666-666666666666', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', 'f6e5d4c3b2a1f6e5d4c3b2a1f6e5d4c3b2a1f6e5d4c3b2a1f6e5d4c3b2a1f6e5', 'devkey-d', 'Design Dev Key', 'agent');

-- Reports (sample compliance data)
INSERT INTO reports (agent_id, group_id, extension_id, user_hash, installed, installed_version) VALUES
  ('33333333-3333-3333-3333-333333333333', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'ms-python.python', 'user1hash', true, '2024.1.0'),
  ('33333333-3333-3333-3333-333333333333', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'golang.go', 'user1hash', true, '0.40.0'),
  ('33333333-3333-3333-3333-333333333333', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'esbenp.prettier-vscode', 'user1hash', false, NULL),
  ('44444444-4444-4444-4444-444444444444', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'esbenp.prettier-vscode', 'user2hash', true, '10.0.0'),
  ('44444444-4444-4444-4444-444444444444', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'dbaeumer.vscode-eslint', 'user2hash', true, '2.4.0');

-- Blocklist (example blocked extension)
INSERT INTO blocklist (pattern_type, pattern, reason, created_by, is_active) VALUES
  ('exact', 'malicious.bad-extension', 'Known malware - reported by security team', '11111111-1111-1111-1111-111111111111', true);
```

**Step 2: Write the down migration**

```sql
-- 000010_seed_data.down.sql
-- Remove seed data in reverse dependency order
DELETE FROM reports WHERE agent_id IN ('33333333-3333-3333-3333-333333333333', '44444444-4444-4444-4444-444444444444');
DELETE FROM blocklist WHERE pattern = 'malicious.bad-extension';
DELETE FROM api_keys WHERE id IN ('55555555-5555-5555-5555-555555555555', '66666666-6666-6666-6666-666666666666');
DELETE FROM agents WHERE id IN ('33333333-3333-3333-3333-333333333333', '44444444-4444-4444-4444-444444444444');
DELETE FROM group_extensions WHERE group_id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');
DELETE FROM extensions WHERE id IN ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'ffffffff-ffff-ffff-ffff-ffffffffffff', '00000000-0000-0000-0000-000000000001');
DELETE FROM groups WHERE id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');
DELETE FROM users WHERE id IN ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222');
```

**Step 3: Commit**

```bash
git add server/migrations/000010_seed_data.*.sql
git commit -m "feat(db): add seed data migration"
```

---

### Task 11: Run Full Migration Suite

**Step 1: Reset database**

```bash
docker compose -f docker-compose.dev.yml down -v
docker compose -f docker-compose.dev.yml up -d postgres
sleep 3  # Wait for postgres to be ready
```

**Step 2: Run all migrations with golang-migrate**

```bash
cd /home/mark/Projects/devtools-sync/server
migrate -path ./migrations -database "postgres://devtools:devtools-local-dev@localhost:5432/devtools_sync?sslmode=disable" up
```

Expected: All 10 migrations applied successfully

**Step 3: Verify tables exist**

```bash
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "\dt"
```

Expected: 9 tables listed (users, groups, extensions, agents, api_keys, group_extensions, reports, blocklist, audit_log)

**Step 4: Verify seed data**

```bash
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "SELECT email, role FROM users"
```

Expected: admin@example.com (admin), viewer@example.com (viewer)

**Step 5: Test full rollback**

```bash
cd /home/mark/Projects/devtools-sync/server
migrate -path ./migrations -database "postgres://devtools:devtools-local-dev@localhost:5432/devtools_sync?sslmode=disable" down -all
```

Expected: All migrations rolled back

**Step 6: Re-apply for final state**

```bash
migrate -path ./migrations -database "postgres://devtools:devtools-local-dev@localhost:5432/devtools_sync?sslmode=disable" up
```

---

### Task 12: Create Database Diagram

**Files:**
- Create: `docs/database-diagram.md`

**Step 1: Create the diagram using Mermaid**

```markdown
# Database Schema Diagram

```mermaid
erDiagram
    users ||--o{ api_keys : "created_by"
    users ||--o{ group_extensions : "added_by"
    users ||--o{ blocklist : "created_by"

    groups ||--o{ agents : "contains"
    groups ||--o{ api_keys : "has"
    groups ||--o{ group_extensions : "has"
    groups ||--o{ reports : "receives"
    groups ||--o{ audit_log : "tracked_in"

    extensions ||--o{ group_extensions : "assigned_to"

    agents ||--o{ reports : "submits"

    users {
        uuid id PK
        varchar email UK
        varchar password_hash
        varchar display_name
        varchar role
        boolean is_active
        timestamptz last_login_at
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    groups {
        uuid id PK
        varchar name UK
        text description
        jsonb settings
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    extensions {
        uuid id PK
        varchar extension_id UK
        varchar name
        varchar publisher
        varchar pinned_version
        varchar sha256_checksum
        text marketplace_url
        boolean is_blocked
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    agents {
        uuid id PK
        uuid group_id FK
        varchar machine_id
        varchar user_hash
        varchar display_name
        varchar platform
        varchar agent_version
        timestamptz last_heartbeat
        timestamptz last_sync
        varchar status
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }

    api_keys {
        uuid id PK
        uuid group_id FK
        uuid created_by FK
        varchar key_hash UK
        varchar key_prefix
        varchar name
        varchar scopes
        timestamptz expires_at
        timestamptz last_used_at
        timestamptz revoked_at
        timestamptz created_at
    }

    group_extensions {
        uuid id PK
        uuid group_id FK
        uuid extension_id FK
        boolean is_required
        uuid added_by FK
        timestamptz created_at
    }

    reports {
        uuid id PK
        uuid agent_id FK
        uuid group_id FK
        varchar extension_id
        varchar user_hash
        boolean installed
        varchar installed_version
        varchar extension_sha256
        inet client_ip
        timestamptz reported_at
    }

    blocklist {
        uuid id PK
        varchar pattern_type
        varchar pattern
        text reason
        uuid created_by FK
        boolean is_active
        timestamptz created_at
        timestamptz updated_at
    }

    audit_log {
        uuid id PK
        varchar event_type
        varchar actor_type
        uuid actor_id
        uuid group_id FK
        varchar target_type
        uuid target_id
        jsonb details
        inet client_ip
        text user_agent
        timestamptz created_at
    }
```
```

**Step 2: Commit**

```bash
git add docs/database-diagram.md
git commit -m "docs: add database schema diagram"
```

---

### Task 13: Update Design Document and Close Issue

**Step 1: Update acceptance criteria in design doc**

Mark all criteria as complete in `docs/plans/2026-01-28-database-schema-design.md`.

**Step 2: Final commit**

```bash
git add docs/plans/2026-01-28-database-schema-design.md
git commit -m "docs: mark Issue #8 acceptance criteria complete"
```

**Step 3: Push to remote**

```bash
git push origin main
```

---

## Summary

| Task | Description |
|------|-------------|
| 1 | Create users table migration |
| 2 | Create groups table migration |
| 3 | Create extensions table migration |
| 4 | Create agents table migration |
| 5 | Create api_keys table migration |
| 6 | Create group_extensions table migration |
| 7 | Create reports table migration |
| 8 | Create blocklist table migration |
| 9 | Create audit_log table migration |
| 10 | Create seed data migration |
| 11 | Run full migration suite |
| 12 | Create database diagram |
| 13 | Update docs and close issue |
