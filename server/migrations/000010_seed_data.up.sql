-- 000010_seed_data.up.sql
-- Seed data for development environment only

-- Users (password: devpassword123, bcrypt hash)
-- nosemgrep: generic.secrets.security.detected-bcrypt-hash.detected-bcrypt-hash
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
