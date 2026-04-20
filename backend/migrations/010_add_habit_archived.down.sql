-- Remove archived_at field from habits
DROP INDEX IF EXISTS idx_habits_archived;
ALTER TABLE habits DROP COLUMN IF EXISTS archived_at;
