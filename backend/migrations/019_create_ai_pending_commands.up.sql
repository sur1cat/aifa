CREATE TABLE IF NOT EXISTS ai_pending_commands (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_pending_commands_updated_at ON ai_pending_commands(updated_at);
