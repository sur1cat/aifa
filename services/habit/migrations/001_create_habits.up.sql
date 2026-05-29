-- Consolidated from monolith migrations 002, 006, 010, 012.
-- No cross-schema FK to users — ownership is enforced in SQL (WHERE user_id=...)
-- and deletions cascade via the user.deleted NATS event.
-- goal_id is a plain UUID (goals live in goal-service); the goal.deleted event
-- nulls it out.

CREATE TABLE IF NOT EXISTS habits (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL,
    goal_id       UUID,
    title         TEXT NOT NULL,
    icon          TEXT NOT NULL DEFAULT '🎯',
    color         TEXT NOT NULL DEFAULT 'green',
    period        TEXT NOT NULL DEFAULT 'daily',
    target_value  INTEGER,
    unit          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_habits_user ON habits (user_id);
CREATE INDEX IF NOT EXISTS idx_habits_goal ON habits (goal_id);
CREATE INDEX IF NOT EXISTS idx_habits_archived ON habits (archived_at);

CREATE TABLE IF NOT EXISTS habit_completions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    habit_id        UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    completed_date  DATE NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT habit_completions_unique UNIQUE (habit_id, completed_date)
);

CREATE INDEX IF NOT EXISTS idx_habit_completions_habit ON habit_completions (habit_id);
CREATE INDEX IF NOT EXISTS idx_habit_completions_date ON habit_completions (completed_date);

CREATE TABLE IF NOT EXISTS habit_progress (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    habit_id        UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    progress_date   DATE NOT NULL,
    progress_value  INTEGER NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT habit_progress_unique UNIQUE (habit_id, progress_date)
);

CREATE INDEX IF NOT EXISTS idx_habit_progress_habit ON habit_progress (habit_id);
CREATE INDEX IF NOT EXISTS idx_habit_progress_date ON habit_progress (progress_date);
