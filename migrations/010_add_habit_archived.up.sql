ALTER TABLE habits ADD COLUMN archived_at TIMESTAMPTZ;

CREATE INDEX idx_habits_archived ON habits(archived_at);
