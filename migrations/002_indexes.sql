CREATE INDEX IF NOT EXISTS idx_entries_vault_status ON entries(vault_id,status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_entries_creator ON entries(creator_id,created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_entries_tags ON entries USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_comments_entry_time ON comments(entry_id,created_at);
CREATE INDEX IF NOT EXISTS idx_audit_vault_time ON audit_events(vault_id,created_at DESC);
CREATE INDEX IF NOT EXISTS idx_refresh_family ON refresh_tokens(family_id) WHERE revoked_at IS NULL;
