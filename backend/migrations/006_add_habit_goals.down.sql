-- Remove habit_progress table
DROP TABLE IF EXISTS habit_progress;

-- Remove columns from habits table
ALTER TABLE habits DROP COLUMN IF EXISTS target_value;
ALTER TABLE habits DROP COLUMN IF EXISTS unit;
