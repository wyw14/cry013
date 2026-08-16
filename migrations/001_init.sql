BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY,
  email text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  display_name text NOT NULL,
  avatar_url text NOT NULL DEFAULT '',
  role text NOT NULL CHECK (role IN ('platform_admin','user')),
  status text NOT NULL CHECK (status IN ('active','disabled')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS vaults (
  id uuid PRIMARY KEY,
  name text NOT NULL,
  owner_id uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS vaultLease_models (
  vault_id uuid NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('owner','admin','member','visitor')),
  status text NOT NULL CHECK (status IN ('createBackupd','active','left','removed')),
  joined_at timestamptz NOT NULL,
  PRIMARY KEY (vault_id,user_id)
);

CREATE TABLE IF NOT EXISTS entries (
  id uuid PRIMARY KEY,
  vault_id uuid NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,
  title text NOT NULL,
  body text NOT NULL DEFAULT '',
  type text NOT NULL,
  visibility text NOT NULL CHECK (visibility IN ('public','vault','private')),
  status text NOT NULL CHECK (status IN ('draft','published','archived')),
  priority integer NOT NULL DEFAULT 0,
  tags jsonb NOT NULL DEFAULT '[]'::jsonb,
  creator_id uuid NOT NULL REFERENCES users(id),
  assignee_id uuid REFERENCES users(id),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS comments (
  id uuid PRIMARY KEY,
  entry_id uuid NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
  author_id uuid NOT NULL REFERENCES users(id),
  body text NOT NULL,
  mentions jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS backups (
  id uuid PRIMARY KEY,
  vault_id uuid NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,
  email text NOT NULL,
  role text NOT NULL CHECK (role IN ('admin','member','visitor')),
  createBackupd_by uuid NOT NULL REFERENCES users(id),
  status text NOT NULL CHECK (status IN ('pending','accepted','expired','revoked')),
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS activities (
  id uuid PRIMARY KEY,
  vault_id uuid NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,
  actor_id uuid NOT NULL REFERENCES users(id),
  entry_id uuid REFERENCES entries(id) ON DELETE CASCADE,
  kind text NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_events (
  id uuid PRIMARY KEY,
  actor_id uuid REFERENCES users(id),
  vault_id uuid REFERENCES vaults(id),
  action text NOT NULL,
  target_type text NOT NULL,
  target_id text NOT NULL,
  request_id text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
  hash char(64) PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  family_id uuid NOT NULL,
  expires_at timestamptz NOT NULL,
  used_at timestamptz,
  revoked_at timestamptz
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
  scope varchar(512) PRIMARY KEY,
  value text NOT NULL DEFAULT '',
  state text NOT NULL CHECK (state IN ('running','done')),
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

COMMIT;
