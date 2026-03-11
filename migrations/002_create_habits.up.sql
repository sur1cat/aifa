CREATE TABLE habits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    icon VARCHAR(50) NOT NULL DEFAULT '🎯',
    color VARCHAR(20) NOT NULL DEFAULT 'green',
    period VARCHAR(20) NOT NULL DEFAULT 'daily', -- daily, weekly, monthly
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_habits_user ON habits(user_id);

CREATE TABLE habit_completions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    completed_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_habit_completion UNIQUE (habit_id, completed_date)
);

CREATE INDEX idx_habit_completions_habit ON habit_completions(habit_id);
CREATE INDEX idx_habit_completions_date ON habit_completions(completed_date);
