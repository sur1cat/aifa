CREATE TABLE IF NOT EXISTS savings_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    monthly_target DECIMAL(12, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_user_savings_goal UNIQUE (user_id)
);

CREATE INDEX IF NOT EXISTS idx_savings_goals_user_id ON savings_goals(user_id);

CREATE OR REPLACE FUNCTION update_savings_goals_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_savings_goals_updated_at
    BEFORE UPDATE ON savings_goals
    FOR EACH ROW
    EXECUTE FUNCTION update_savings_goals_updated_at();
