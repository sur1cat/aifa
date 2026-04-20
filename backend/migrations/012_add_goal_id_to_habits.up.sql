-- Add goal_id to habits (nullable for optional relationship)
-- ON DELETE SET NULL: when goal is deleted, habits remain but lose their goal reference
ALTER TABLE habits ADD COLUMN goal_id UUID REFERENCES goals(id) ON DELETE SET NULL;

CREATE INDEX idx_habits_goal ON habits(goal_id);
