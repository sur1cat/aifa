-- Add target and unit columns to habits table
ALTER TABLE habits ADD COLUMN IF NOT EXISTS target_value INTEGER;
ALTER TABLE habits ADD COLUMN IF NOT EXISTS unit VARCHAR(50);

-- Create habit_progress table for tracking progress values
CREATE TABLE IF NOT EXISTS habit_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    habit_id UUID NOT NULL REFERENCES habits(id) ON DELETE CASCADE,
    progress_date DATE NOT NULL,
    progress_value INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(habit_id, progress_date)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_habit_progress_habit_id ON habit_progress(habit_id);
CREATE INDEX IF NOT EXISTS idx_habit_progress_date ON habit_progress(progress_date);
