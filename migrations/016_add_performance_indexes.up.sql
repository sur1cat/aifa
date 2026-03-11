CREATE INDEX IF NOT EXISTS idx_habits_user_id ON habits(user_id);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);

CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, date);

CREATE INDEX IF NOT EXISTS idx_recurring_transactions_user_id ON recurring_transactions(user_id);

CREATE INDEX IF NOT EXISTS idx_recurring_transactions_active ON recurring_transactions(user_id, is_active) WHERE is_active = true;

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);

CREATE INDEX IF NOT EXISTS idx_tasks_user_due_date ON tasks(user_id, due_date);

CREATE INDEX IF NOT EXISTS idx_habit_completions_habit_date ON habit_completions(habit_id, completed_date);

CREATE INDEX IF NOT EXISTS idx_goals_user_id ON goals(user_id);

CREATE INDEX IF NOT EXISTS idx_savings_goals_user_id ON savings_goals(user_id);

CREATE INDEX IF NOT EXISTS idx_device_tokens_user_id ON device_tokens(user_id);

CREATE INDEX IF NOT EXISTS idx_otp_codes_phone ON otp_codes(phone, verified, expires_at);
