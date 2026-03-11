ALTER TABLE habits ADD COLUMN goal_id UUID REFERENCES goals(id) ON DELETE SET NULL;

CREATE INDEX idx_habits_goal ON habits(goal_id);
