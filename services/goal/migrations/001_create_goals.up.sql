-- goals live in goal-service schema. No FK to users (cleanup via user.deleted).
CREATE TABLE IF NOT EXISTS goals (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL,
    title         TEXT NOT NULL,
    icon          TEXT NOT NULL DEFAULT '🎯',
    target_value  INTEGER,
    unit          TEXT,
    deadline      TIMESTAMPTZ,
    archived_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_goals_user ON goals (user_id);
CREATE INDEX IF NOT EXISTS idx_goals_archived ON goals (archived_at);
