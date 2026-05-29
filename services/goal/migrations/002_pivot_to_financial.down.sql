DROP INDEX IF EXISTS idx_goals_type;

ALTER TABLE goals
    ADD COLUMN IF NOT EXISTS target_value INTEGER,
    ADD COLUMN IF NOT EXISTS unit         TEXT;

UPDATE goals
SET target_value = target_amount::INTEGER
WHERE target_value IS NULL AND target_amount IS NOT NULL;

ALTER TABLE goals
    DROP COLUMN IF EXISTS goal_type,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS current_amount,
    DROP COLUMN IF EXISTS target_amount;
